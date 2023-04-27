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

package account

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/utils/zero"
	"github.com/rs/zerolog/log"
	"github.com/tyler-smith/go-bip39"
)

const (
	// Secret phrase for mnemonic based crypto key generation
	secretPhrase = "Piprate u5SPXFiNqfxYfZU2V23k4s9HsiN44UFhzPk5t9QLvNt" //nolint:gosec
)

type RecoveryCode struct {
	Code      string     `json:"code"`
	UserID    string     `json:"userID"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

func (rc RecoveryCode) Bytes() []byte {
	b, _ := jsonw.Marshal(rc)
	return b
}

func NewRecoveryCode(userID string, secondsTTL int64) (*RecoveryCode, error) {
	idBuffer := make([]byte, 32)
	_, err := rand.Read(idBuffer)
	if err != nil {
		return nil, err
	}

	expiryTime := time.Now().Add(time.Duration(secondsTTL) * time.Second).UTC()

	rc := &RecoveryCode{
		Code:      base58.Encode(idBuffer),
		UserID:    userID,
		ExpiresAt: &expiryTime,
	}
	return rc, nil
}

type RecoveryRequest struct {
	UserID                string `json:"userID"`
	RecoveryCode          string `json:"recoveryCode"`
	VerificationSignature string `json:"signature"`
	EncryptedPassword     string `json:"encryptedPassword"`
	ManagedCryptoKey      string `json:"managedCryptoKey,omitempty"`
}

func (req *RecoveryRequest) Valid(recoveryPublicKey []byte) bool {
	codeHash := sha256.Sum256([]byte(req.RecoveryCode))
	sig := base58.Decode(req.VerificationSignature)
	return ed25519.Verify(recoveryPublicKey, codeHash[:], sig)
}

// BuildRecoveryRequest creates a recovery request structure that can be sent to /v1/recover-account endpoint
// to regain access to a MetaLocker account.
// if cryptoKey is passed, the request will contain the account's managed crypto key in a cleartext form.
// This enables server side recovery for managed accounts for clients that don't have access to advanced
// cryptography.
func BuildRecoveryRequest(userID, recoveryCode string, privKey ed25519.PrivateKey, newPassphrase string, cryptoKey *model.AESKey) *RecoveryRequest {

	codeHash := sha256.Sum256([]byte(recoveryCode))

	sig := ed25519.Sign(privKey, codeHash[:])

	req := &RecoveryRequest{
		UserID:                userID,
		RecoveryCode:          recoveryCode,
		VerificationSignature: base58.Encode(sig),
		EncryptedPassword:     HashUserPassword(newPassphrase),
	}

	if cryptoKey != nil {
		req.ManagedCryptoKey = base64.StdEncoding.EncodeToString(GenerateManagedFromHostedKey(cryptoKey)[:])
	}

	return req
}

func GenerateKeysFromRecoveryPhrase(recoveryPhrase string) (*model.AESKey, ed25519.PublicKey, ed25519.PrivateKey, error) {
	mnemonicSeed := bip39.NewSeed(recoveryPhrase, secretPhrase)[:model.KeySize]

	cryptoKeyPriv := model.NewAESKey(mnemonicSeed[:model.KeySize])

	pubk, privk, err := ed25519.GenerateKey(bytes.NewReader(mnemonicSeed))
	if err != nil {
		return nil, nil, nil, err
	}

	zero.Bytes(mnemonicSeed)

	return cryptoKeyPriv, pubk, privk, nil
}

func Recover(acct *Account, cryptoKey *model.AESKey, newPassphrase string) (*Account, error) {
	var err error

	acct = acct.Copy()
	acct.EncryptedPassword = HashUserPassword(newPassphrase)

	switch acct.AccessLevel {
	case model.AccessLevelHosted:
		hostedMasterPassphrase := []byte(newPassphrase)
		newHostedMasterKey, err := newSecretKey(&hostedMasterPassphrase, hostedAccountConfig)
		if err != nil {
			log.Err(err).Msg("Failed to create master key")
			return nil, err
		}

		// Encrypt the crypto keys with the associated master keys.
		cryptoKeyPrivEncrypted, err := newHostedMasterKey.Encrypt(cryptoKey.Bytes())
		if err != nil {
			return nil, err
		}

		acct.HostedSecretStore.MasterKeyParams = base64.StdEncoding.EncodeToString(newHostedMasterKey.Marshal())
		acct.HostedSecretStore.EncryptedPayloadKey = base64.StdEncoding.EncodeToString(cryptoKeyPrivEncrypted)
	case model.AccessLevelManaged:
		// managed account, do nothing
	default:
		return nil, fmt.Errorf("account with access level %d can't be recovered", acct.AccessLevel)
	}

	// update managed secret store

	managedMasterPassphrase := []byte(HashUserPassword(newPassphrase))
	newManagedMasterKey, err := newSecretKey(&managedMasterPassphrase, managedAccountConfig)
	if err != nil {
		log.Err(err).Msg("Failed to create master key")
		return nil, err
	}

	managedCryptoKey := GenerateManagedFromHostedKey(cryptoKey)

	// Encrypt the crypto keys with the associated master keys.
	managedCryptoKeyEncrypted, err := newManagedMasterKey.Encrypt(managedCryptoKey.Bytes())
	if err != nil {
		return nil, err
	}

	acct.ManagedSecretStore.MasterKeyParams = base64.StdEncoding.EncodeToString(newManagedMasterKey.Marshal())
	acct.ManagedSecretStore.EncryptedPayloadKey = base64.StdEncoding.EncodeToString(managedCryptoKeyEncrypted)

	acct.State = StateActive

	return acct, nil
}

// RecoverManaged recovers a managed account for clients that don't have access to advanced cryptography.
func RecoverManaged(acct *Account, managedCryptoKey *model.AESKey, hashedNewPassphrase string) (*Account, error) {
	var err error

	if acct.AccessLevel != model.AccessLevelManaged {
		return nil, fmt.Errorf("attempted to recoved account with access level %d as managed", acct.AccessLevel)
	}

	acct = acct.Copy()
	acct.EncryptedPassword = hashedNewPassphrase

	// update managed secret store

	managedMasterPassphrase := []byte(hashedNewPassphrase)
	newManagedMasterKey, err := newSecretKey(&managedMasterPassphrase, managedAccountConfig)
	if err != nil {
		log.Err(err).Msg("Failed to create master key")
		return nil, err
	}

	// Encrypt the crypto keys with the associated master keys.
	managedCryptoKeyEncrypted, err := newManagedMasterKey.Encrypt(managedCryptoKey.Bytes())
	if err != nil {
		return nil, err
	}

	acct.ManagedSecretStore.MasterKeyParams = base64.StdEncoding.EncodeToString(newManagedMasterKey.Marshal())
	acct.ManagedSecretStore.EncryptedPayloadKey = base64.StdEncoding.EncodeToString(managedCryptoKeyEncrypted)

	acct.State = StateActive

	return acct, nil
}
