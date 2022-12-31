// Copyright 2022 Piprate Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/jamesruan/sodium"
	"github.com/piprate/metalocker/utils"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/utils/zero"
	"github.com/rs/zerolog/log"
)

const (
	RequestingCommitmentTag = "requesting commitment"
)

type (
	// LockerParticipant is a definition of locker participant. It contains sensitive secrets, such as SharedSecret,
	// and should be stored securely.
	LockerParticipant struct {
		// ID is the participant's identity ID (DID)
		ID string `json:"id"`
		// SharedSecret is a Base64-encoded secret used to encrypt operations in the given locker (leases, etc)
		SharedSecret string `json:"sharedSecret,omitempty"`
		Self         bool   `json:"self,omitempty"`
		// RootPublicKey is a Base64-encoded root public key that can be used to identify if the specific record
		// was issued by this participant.
		RootPublicKey string `json:"rootPublicKey,omitempty"`
		// RootPrivateKeyEnc is a Base64-encoded, encrypted root HD key used to generate record's routing keys.
		RootPrivateKeyEnc string `json:"encryptedRootPrivateKey,omitempty"`
		// AcceptedAtBlock is the number of the block when the locker was accepted by the party
		// and registered in its root locker.
		AcceptedAtBlock int64 `json:"acceptedAtBlock,omitempty"`

		rootKeyPriv       *hdkeychain.ExtendedKey
		rootKeyPub        *hdkeychain.ExtendedKey
		sharedSecretBytes []byte
	}

	// Locker is a secure, persistent, bidirectional communication channel between two or more participants.
	// A special type of locker with just one participant is called a uni-locker.
	Locker struct {
		// ID is the unique locker ID.
		ID string `json:"id"`
		// Name is the locker's name. These names are useful for locker documentation purposes.
		// They aren't used in any data processing.
		Name string `json:"name"`
		// AccessLevel is the locker's access level. Data wallet needs to be unlocked to a specific access level
		// to gain access to lockers at this level or higher.
		AccessLevel AccessLevel `json:"level"`
		// Participants is a list of locker participants.
		Participants []*LockerParticipant `json:"participants"`
		// Created is the locker's creation time. For documentation purposes only.
		Created *time.Time `json:"created"`
		// Expires is the time when the locker will expire. NOT SUPPORTED.
		Expires *time.Time `json:"expires,omitempty"`
		// Sealed is the time when the locker was sealed (closed). NOT SUPPORTED.
		Sealed *time.Time `json:"sealed,omitempty"`
		// FirstBlock is the block number that was the height of the chain when the locker was created.
		// It is guaranteed that all records for this locker will be in blocks AFTER this block.
		FirstBlock int64 `json:"firstBlock"`
		// LastBlock is the block number that was the height of the chain when the locker was sealed.
		// It is guaranteed that all records for this locker will be in blocks BEFORE this block.
		// NOT SUPPORTED.
		LastBlock int64 `json:"lastBlock,omitempty"`
		// ThirdPartyAcceptedAtBlock is the number of the block when the locker was accepted by the owner
		// when the owner acts as a third party (is not a participant on the locker)
		ThirdPartyAcceptedAtBlock int64 `json:"acceptedAtBlock,omitempty"`
	}
)

func (lp *LockerParticipant) Zero() {
	if lp.rootKeyPriv != nil {
		// zero the key
		lp.rootKeyPriv.Zero()
		// delete the key
		lp.rootKeyPriv = nil
	}
	if lp.rootKeyPub != nil {
		// zero the key
		lp.rootKeyPub.Zero()
		// delete the key
		lp.rootKeyPub = nil
	}
	if lp.sharedSecretBytes != nil {
		zero.Bytes(lp.sharedSecretBytes)
	}
}

func (lp *LockerParticipant) IsHydrated() bool {
	if lp.Self {
		return lp.rootKeyPriv != nil && lp.rootKeyPub != nil
	} else {
		return lp.rootKeyPub != nil
	}
}

// Hydrate decrypts (if needed) and instantiates ExtendedKey fields from Base64 encoded values
func (lp *LockerParticipant) Hydrate(pk ed25519.PrivateKey) error {
	if lp.Self && pk != nil {
		b, err := base64.StdEncoding.DecodeString(lp.RootPrivateKeyEnc)
		if err != nil {
			return err
		}
		keyBytes, err := AnonDecrypt(b, pk)
		if err != nil {
			return err
		}
		privKey, err := hdkeychain.NewKeyFromString(string(keyBytes))
		if err != nil {
			return err
		}
		lp.rootKeyPriv = privKey
		lp.rootKeyPub, err = privKey.Neuter()
		if err != nil {
			return err
		}
	} else {
		// only process the public key
		pubKey, err := hdkeychain.NewKeyFromString(lp.RootPublicKey)
		if err != nil {
			return err
		}
		lp.rootKeyPub = pubKey
	}

	sharedSecretBytes, err := base64.StdEncoding.DecodeString(lp.SharedSecret)
	if err != nil {
		return err
	}

	lp.sharedSecretBytes = sharedSecretBytes

	return nil
}

func (lp *LockerParticipant) GetRecordPublicKey(idx uint32) (*btcec.PublicKey, error) {
	k, err := lp.rootKeyPub.Derive(idx)
	if err != nil {
		return nil, err
	}
	return k.ECPubKey()
}

func (lp *LockerParticipant) GetRecordPrivateKey(idx uint32) (*hdkeychain.ExtendedKey, error) {
	return lp.rootKeyPriv.Derive(idx)
}

func (lp *LockerParticipant) GetRootPrivateKey() string {
	return lp.rootKeyPriv.String()
}

func (lp *LockerParticipant) IsRecordOwner(routingKey string, idx uint32) (*btcec.PublicKey, *AESKey, error) {
	k, err := lp.rootKeyPub.Derive(idx)
	if err != nil {
		return nil, nil, err
	}

	pk, err := k.ECPubKey()
	if err != nil {
		return nil, nil, err
	}

	rk, _ := BuildRoutingKey(pk)

	if rk == routingKey {
		return pk, DeriveSymmetricalKey(lp.sharedSecretBytes, pk), nil
	} else {
		return nil, nil, nil
	}
}

func (lp *LockerParticipant) GetOperationSymKey(idx uint32) *AESKey {
	recordPubKey, _ := lp.GetRecordPublicKey(idx)
	return DeriveSymmetricalKey(lp.sharedSecretBytes, recordPubKey)
}

func (l *Locker) Bytes() []byte {
	b, _ := jsonw.Marshal(l)
	return b
}

func (l *Locker) Us() *LockerParticipant {
	var us *LockerParticipant
	for _, p := range l.Participants {
		if p.Self {
			if us != nil {
				if us.ID == p.ID {
					return us
				} else {
					log.Warn().
						Msg("Attempting to find the 'self' participant in a locker between two 'self' identities")
					return nil
				}
			} else {
				us = p
			}
		}
	}
	return us
}

func (l *Locker) Them() *LockerParticipant {
	var them *LockerParticipant
	for _, p := range l.Participants {
		if !p.Self {
			if them != nil {
				if them.ID == p.ID {
					return them
				} else {
					log.Warn().Msg("Attempting to find the counterparty in a locker between two 'self' identities")
					return nil
				}
			} else {
				them = p
			}
		}
	}
	return them
}

func (l *Locker) IsUnilocker() bool {
	return len(l.Participants) == 1
}

func (l *Locker) GetParticipant(participantID string) *LockerParticipant {
	for _, p := range l.Participants {
		if p.ID == participantID {
			return p
		}
	}

	return nil
}

func (l *Locker) AcceptedAtBlock() int64 {
	us := l.Us()
	if us != nil && us.AcceptedAtBlock != 0 {
		return us.AcceptedAtBlock
	} else {
		return l.ThirdPartyAcceptedAtBlock
	}
}

func (l *Locker) SetAcceptedAtBlock(block int64) {
	us := l.Us()
	if us != nil {
		us.AcceptedAtBlock = block
	} else {
		l.ThirdPartyAcceptedAtBlock = block
	}
}

func (l *Locker) Perspective(iid string) *Locker {
	newLocker := &Locker{
		ID:          l.ID,
		Name:        l.Name,
		AccessLevel: l.AccessLevel,
		Created:     l.Created,
		Expires:     l.Expires,
		FirstBlock:  l.FirstBlock,
	}

	parties := make([]*LockerParticipant, len(l.Participants))
	for i, party := range l.Participants {
		if party.ID == iid {
			parties[i] = &LockerParticipant{
				ID:                party.ID,
				SharedSecret:      party.SharedSecret,
				Self:              true,
				AcceptedAtBlock:   party.AcceptedAtBlock,
				RootPublicKey:     party.RootPublicKey,
				RootPrivateKeyEnc: party.RootPrivateKeyEnc,
			}
		} else {
			parties[i] = &LockerParticipant{
				ID:              party.ID,
				SharedSecret:    party.SharedSecret,
				Self:            false,
				AcceptedAtBlock: party.AcceptedAtBlock,
				RootPublicKey:   party.RootPublicKey,
			}
		}
	}

	newLocker.Participants = parties

	return newLocker
}

func (l *Locker) Copy() *Locker {
	newLocker := &Locker{
		ID:                        l.ID,
		Name:                      l.Name,
		AccessLevel:               l.AccessLevel,
		Created:                   l.Created,
		Expires:                   l.Expires,
		FirstBlock:                l.FirstBlock,
		ThirdPartyAcceptedAtBlock: l.ThirdPartyAcceptedAtBlock,
	}

	parties := make([]*LockerParticipant, 0)
	for _, party := range l.Participants {
		parties = append(parties, &LockerParticipant{
			ID:                party.ID,
			SharedSecret:      party.SharedSecret,
			Self:              party.Self,
			AcceptedAtBlock:   party.AcceptedAtBlock,
			RootPublicKey:     party.RootPublicKey,
			RootPrivateKeyEnc: party.RootPrivateKeyEnc,
		})
	}

	newLocker.Participants = parties

	return newLocker
}

func (l *Locker) IsHydrated() bool {
	if len(l.Participants) == 0 {
		return false
	}

	for _, party := range l.Participants {
		if !party.IsHydrated() {
			return false
		}
	}

	return true
}

func (l *Locker) Hydrate(pk ed25519.PrivateKey) error {
	for _, party := range l.Participants {
		if err := party.Hydrate(pk); err != nil {
			return err
		}
	}

	return nil
}

func (l *Locker) Zero() {
	for _, p := range l.Participants {
		p.Zero()
	}
	l.Participants = nil
}

func BuildRoutingKey(key *btcec.PublicKey) (string, error) {
	return base58.Encode(key.SerializeCompressed()), nil
}

func BuildSharedSecret(key *hdkeychain.ExtendedKey) string {
	return base64.StdEncoding.EncodeToString(Hash("ledger shared secret", []byte(key.String())))
}

func DeriveSymmetricalKey(secret []byte, pubKey *btcec.PublicKey) *AESKey {
	sk := Hash("Symmetrical key", append(secret, pubKey.SerializeCompressed()...))

	symKey := &AESKey{}
	copy(symKey[:], sk)

	return symKey
}

func BuildAuthorisingCommitmentInput(privKey *hdkeychain.ExtendedKey, opAddress string) []byte {
	return Hash(
		"authorising commitment",
		append(
			[]byte(privKey.String()),
			[]byte(opAddress)...,
		),
	)
}

func DeriveStorageAccessKey(leaseID string) (ed25519.PublicKey, ed25519.PrivateKey) {
	privateKey := ed25519.NewKeyFromSeed(Hash(
		"storage access key", []byte(leaseID),
	),
	)
	publicKey := make([]byte, ed25519.PublicKeySize)
	copy(publicKey, privateKey[32:])

	return publicKey, privateKey
}

func BuildRequestingCommitmentInput(leaseID string, expiresAt *time.Time) []byte {
	saPubKey, _ := DeriveStorageAccessKey(leaseID)

	rcInput := saPubKey
	if expiresAt != nil {
		rcInput = append(rcInput, utils.Int64ToBytes(expiresAt.Unix())...)
	}
	return Hash(RequestingCommitmentTag, rcInput)
}

func GenerateNewHDKey(seed []byte) (*hdkeychain.ExtendedKey, *hdkeychain.ExtendedKey, error) {
	if len(seed) == 0 {
		var err error
		seed, err = hdkeychain.GenerateSeed(hdkeychain.RecommendedSeedLen)
		if err != nil {
			return nil, nil, errors.New("failed to generate new seed")
		}
	}

	// Generate the BIP0044 HD key structure to ensure the provided seed
	// can generate the required structure with no issues.

	// Derive the master extended key from the seed.
	rootPriv, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, nil, errors.New("failed to derive master extended key")
	}

	// Encrypt the root HD keys with the associated crypto keys.
	rootPub, err := rootPriv.Neuter()
	if err != nil {
		return nil, nil, errors.New("failed to convert root private key")
	}
	return rootPriv, rootPub, nil
}

// AnonEncrypt encrypts a message by anonymous-encryption scheme.
// Sealed boxes are designed to anonymously send messages to a Recipient given its public key.
// Only the Recipient can decrypt these messages, using its private key.
// While the Recipient can verify the integrity of the message, it cannot verify the identity of the Sender.
func AnonEncrypt(msg, publicKey []byte) []byte {
	boxPublicKey := sodium.SignPublicKey{Bytes: publicKey}.ToBox()
	return sodium.Bytes(msg).SealedBox(boxPublicKey)
}

func AnonDecrypt(cypherText, privateKey []byte) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Warn().Str("reason", fmt.Sprintf("%s", r)).Msg("Recovered in model.AnonDecrypt")
		}
	}()

	publicKey := privateKey[32:]

	spk := sodium.SignPublicKey{Bytes: publicKey}
	sk := sodium.SignSecretKey{Bytes: privateKey}
	bsk := sk.ToBox()
	bspk := spk.ToBox()

	decryptedMsg, err := sodium.Bytes(cypherText).SealedBoxOpen(sodium.BoxKP{
		PublicKey: bspk,
		SecretKey: bsk,
	})
	if err != nil {
		return nil, err
	} else {
		return []byte(decryptedMsg), nil
	}
}

type PartyOption func() (*LockerParticipant, error)

func withParty(did *DID, seed []byte, us bool) PartyOption {
	return func() (*LockerParticipant, error) {
		if len(seed) == 0 {
			var err error
			seed, err = hdkeychain.GenerateSeed(hdkeychain.RecommendedSeedLen)
			if err != nil {
				return nil, errors.New("failed to generate new seed")
			}
		} else if len(seed) < hdkeychain.RecommendedSeedLen { // pad the seed, if needed
			seed = append(bytes.Repeat([]byte("0"), hdkeychain.RecommendedSeedLen-len(seed)), seed...)
		}

		privHD, pubHD, err := GenerateNewHDKey(seed)
		if err != nil {
			return nil, err
		}

		verKeyVal := did.VerKeyValue()
		if verKeyVal == nil {
			return nil, fmt.Errorf("participant %s doesn't have a verkey", did.ID)
		}
		return &LockerParticipant{
			ID:                did.ID,
			SharedSecret:      BuildSharedSecret(privHD),
			Self:              us,
			RootPublicKey:     pubHD.String(),
			RootPrivateKeyEnc: base64.StdEncoding.EncodeToString(AnonEncrypt([]byte(privHD.String()), verKeyVal)),
			rootKeyPriv:       privHD,
			rootKeyPub:        pubHD,
		}, nil
	}
}

func Us(did *DID, seed []byte) PartyOption {
	return withParty(did, seed, true)
}

func Them(did *DID, seed []byte) PartyOption {
	return withParty(did, seed, false)
}

func GenerateLocker(accessLevel AccessLevel, name string, expires *time.Time, firstBlock int64,
	parties ...PartyOption) (*Locker, error) {

	participants := make([]*LockerParticipant, len(parties))
	var err error
	for i := 0; i < len(parties); i++ {
		participants[i], err = parties[i]()
		if err != nil {
			return nil, err
		}
	}

	id := make([]byte, 32)
	_, err = rand.Read(id)
	if err != nil {
		return nil, errors.New("failed to generate new locker id")
	}

	now := time.Now().UTC()

	return &Locker{
		ID:           base58.Encode(id),
		Name:         name,
		AccessLevel:  accessLevel,
		Participants: participants,
		Created:      &now,
		Expires:      expires,
		FirstBlock:   firstBlock,
	}, nil
}
