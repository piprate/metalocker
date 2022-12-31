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
	"compress/gzip"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/slip10"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/utils/snacl"
	"github.com/piprate/metalocker/utils/zero"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tyler-smith/go-bip39"
)

// Options is used to hold the optional parameters passed to Create or Load.
type Options struct {
	ScryptN int
	ScryptR int
	ScryptP int
}

// defaultNewSecretKey returns a new secret key.  See newSecretKey.
func newSecretKey(passphrase *[]byte, config *Options) (*snacl.SecretKey, error) {
	return snacl.NewSecretKey(passphrase, config.ScryptN, config.ScryptR, config.ScryptP)
}

const (
	CurrentAccountVersion uint32 = 4

	Type = "Account"

	StateActive    = "active"
	StateSuspended = "suspended"
	StateDeleted   = "deleted"
	StateRecovery  = "recovery"
)

var (
	Version = CurrentAccountVersion

	// hostedAccountConfig is an instance of the Options struct is used as a default
	// for 'hosted' accounts.
	hostedAccountConfig = &Options{
		ScryptN: 2048, // original value is 262144 (2^18) - may be slow in the browser
		ScryptR: 8,
		ScryptP: 1,
	}

	// managedAccountConfig is an instance of the Options struct is used as a default
	// for 'managed' accounts.
	managedAccountConfig = &Options{
		ScryptN: 2048, // original value is 262144 (2^18) - may be slow in the browser
		ScryptR: 8,
		ScryptP: 1,
	}

	ErrInvalidPassphrase = errors.New("invalid passphrase")
)

// Account represents a MetaLocker account. Its JSON representation can be used to store
// accounts in the MetaLocker backend. Generally, it doesn't contain any secrets that
// may give access to the account's data, but some fields, such as EncryptedPassword,
// should be protected to avoid dictionary attacks. It's recommended to store account
// definition in an encrypted form.
type Account struct {
	ID                      string            `json:"id,omitempty"`
	Type                    string            `json:"type"`
	Version                 uint32            `json:"version,omitempty"`
	Email                   string            `json:"email"`
	EncryptedPassword       string            `json:"encryptedPassword"`
	MasterAccount           string            `json:"master,omitempty"`
	ParentAccount           string            `json:"parent,omitempty"`
	State                   string            `json:"state,omitempty"`
	RegisteredAt            *time.Time        `json:"registeredAt"`
	Name                    string            `json:"name"`
	GivenName               string            `json:"givenName,omitempty"`
	FamilyName              string            `json:"familyName,omitempty"`
	AccessLevel             model.AccessLevel `json:"level"`
	RecoveryPublicKey       string            `json:"recoveryPublicKey,omitempty"`
	EncryptedRecoverySecret string            `json:"encryptedRecoverySecret,omitempty"`
	DefaultVault            string            `json:"defaultVault,omitempty"`

	ManagedSecretStore *SecretStore `json:"managedSecretStore,omitempty"`
	HostedSecretStore  *SecretStore `json:"hostedSecretStore,omitempty"`

	DerivationIndex uint32 `json:"derivationIndex,omitempty"`
}

type SecretStore struct {
	AccessLevel         model.AccessLevel `json:"level"`
	MasterKeyParams     string            `json:"masterKeyParams,omitempty"`
	EncryptedPayloadKey string            `json:"encryptedPayloadKey,omitempty"`
	EncryptedPayload    string            `json:"encryptedPayload,omitempty"`
}

type SecretStorePayload struct {
	Identities           []*Identity `json:"ii,omitempty"`
	ManagedHMACKey       string      `json:"mhk,omitempty"`
	ManagedEncryptionKey string      `json:"mek,omitempty"`
	HostedHMACKey        string      `json:"hhk,omitempty"`
	HostedEncryptionKey  string      `json:"hek,omitempty"`
	AccountRootKey       string      `json:"ark,omitempty"`
	ManagedRootLocker    string      `json:"marl,omitempty"`
	HostedRootLocker     string      `json:"harl,omitempty"`
}

func (ss *SecretStore) UpdatePayload(payload *SecretStorePayload, key *model.AESKey) error {
	payloadBytes, err := jsonw.Marshal(payload)
	if err != nil {
		return err
	}

	// gzip payload

	var buf bytes.Buffer
	gw, _ := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	_, _ = gw.Write(payloadBytes)
	_ = gw.Close()

	// encrypt payload

	payloadEnc, err := model.EncryptAESCGM(buf.Bytes(), key)
	if err != nil {
		return err
	}

	// encode payload in base64

	ss.EncryptedPayload = base64.StdEncoding.EncodeToString(payloadEnc)

	return nil
}

func (ss *SecretStore) ExtractPayloadKey(passphrase string) (*model.AESKey, error) {
	var masterKeyPriv snacl.SecretKey
	masterKeyParamsBytes, err := base64.StdEncoding.DecodeString(ss.MasterKeyParams)
	if err != nil {
		return nil, errors.New("failed to decode master private key")
	}
	err = masterKeyPriv.Unmarshal(masterKeyParamsBytes)
	if err != nil {
		return nil, errors.New("failed to unmarshal master private key")
	}

	passphraseBytes := []byte(passphrase)
	if err = masterKeyPriv.DeriveKey(&passphraseBytes); err != nil {
		if errors.Is(err, snacl.ErrInvalidPassword) {
			return nil, ErrInvalidPassphrase
		} else {
			return nil, errors.New("failed to derive master private key")
		}
	}

	// Use the master private key to decrypt the crypto private key.
	cryptoKeyPrivEncrypted, err := base64.StdEncoding.DecodeString(ss.EncryptedPayloadKey)
	if err != nil {
		return nil, errors.New("failed to decode crypto key")
	}
	decryptedKey, err := masterKeyPriv.Decrypt(cryptoKeyPrivEncrypted)
	if err != nil {
		return nil, errors.New("failed to decrypt crypto private key")
	}
	cryptoKeyPriv := model.NewAESKey(decryptedKey)
	zero.Bytes(decryptedKey)

	return cryptoKeyPriv, nil
}

func (ss *SecretStore) GetPayload(key *model.AESKey) (*SecretStorePayload, error) {
	payload := &SecretStorePayload{}

	if ss.EncryptedPayload != "" {
		var payloadBytes []byte
		encBytes, _ := base64.StdEncoding.DecodeString(ss.EncryptedPayload)

		payloadBytes, err := model.DecryptAESCGM(encBytes, key)
		if err != nil {
			return nil, errors.New("failed to decrypt account payload")
		}

		gr, err := gzip.NewReader(bytes.NewBuffer(payloadBytes))
		if err != nil {
			return nil, err
		}
		payloadBytes, err = io.ReadAll(gr)
		if err != nil {
			return nil, err
		}
		_ = gr.Close()

		err = jsonw.Unmarshal(payloadBytes, payload)
		if err != nil {
			return nil, errors.New("failed to unmarshal account payload")
		}
	}

	return payload, nil
}

func (ss *SecretStore) Copy() *SecretStore {
	cpy := *ss
	return &cpy
}

func (ss *SecretStore) Validate() error {
	if ss.AccessLevel == 0 {
		return errors.New("empty access level in Secret Store")
	}
	if ss.EncryptedPayload == "" {
		return errors.New("empty payload in Secret Store")
	}
	return nil
}

func (ssp *SecretStorePayload) Zero() {
	for _, idy := range ssp.Identities {
		// Clear all the identity private keys.
		idy.DID.Zero()
	}
}

func (a *Account) Copy() *Account {
	cp := *a
	if cp.HostedSecretStore != nil {
		cp.HostedSecretStore = cp.HostedSecretStore.Copy()
	}
	if cp.ManagedSecretStore != nil {
		cp.ManagedSecretStore = cp.ManagedSecretStore.Copy()
	}
	return &cp
}

func (a *Account) Bytes() []byte {
	b, _ := jsonw.Marshal(a)
	return b
}

func (a *Account) RestrictedCopy() *Account {
	return &Account{
		ID:                a.ID,
		Type:              a.Type,
		Version:           a.Version,
		Email:             "",
		EncryptedPassword: "",
		MasterAccount:     a.MasterAccount,
		ParentAccount:     a.ParentAccount,
		State:             a.State,
		RegisteredAt:      a.RegisteredAt,
		Name:              a.Name,
		GivenName:         a.GivenName,
		FamilyName:        a.FamilyName,
		AccessLevel:       model.AccessLevelRestricted,
	}
}

func (a *Account) Validate() error {
	if a.ID == "" {
		return errors.New("empty account ID")
	}
	if a.DefaultVault == "" {
		return errors.New("no default vault")
	}
	switch a.AccessLevel {
	case model.AccessLevelHosted:
		if a.HostedSecretStore == nil || a.ManagedSecretStore == nil {
			return errors.New("both hostedSecretStore and managedSecretStore should be specified in the account")
		}
		if err := a.HostedSecretStore.Validate(); err != nil {
			return err
		}
		if err := a.ManagedSecretStore.Validate(); err != nil {
			return err
		}
	case model.AccessLevelManaged:
		if a.ManagedSecretStore == nil {
			return errors.New("empty secret store in a managed account")
		}
		if err := a.ManagedSecretStore.Validate(); err != nil {
			return err
		}
	case model.AccessLevelRestricted:
		// not implemented yet
	default:
		if a.AccessLevel == 0 {
			return errors.New("account access level not specified")
		} else {
			return fmt.Errorf("unknown account access level: %d", a.AccessLevel)
		}
	}
	return nil
}

func (a *Account) ExtractManagedKey(hashedPassphrase string) (*model.AESKey, error) {
	var masterKeyPriv snacl.SecretKey
	masterKeyParamsBytes, err := base64.StdEncoding.DecodeString(a.ManagedSecretStore.MasterKeyParams)
	if err != nil {
		return nil, errors.New("failed to decode master private key")
	}
	err = masterKeyPriv.Unmarshal(masterKeyParamsBytes)
	if err != nil {
		return nil, errors.New("failed to unmarshal master private key")
	}

	passphraseBytes := []byte(hashedPassphrase)
	if err = masterKeyPriv.DeriveKey(&passphraseBytes); err != nil {
		if errors.Is(err, snacl.ErrInvalidPassword) {
			return nil, ErrInvalidPassphrase
		} else {
			return nil, errors.New("failed to derive master private key")
		}
	}

	// Use the master private key to decrypt the crypto private key.
	cryptoKeyPrivEncrypted, err := base64.StdEncoding.DecodeString(a.ManagedSecretStore.EncryptedPayloadKey)
	if err != nil {
		return nil, errors.New("failed to decode crypto key")
	}
	decryptedKey, err := masterKeyPriv.Decrypt(cryptoKeyPrivEncrypted)
	if err != nil {
		return nil, errors.New("failed to decrypt crypto private key")
	}
	cryptoKeyPriv := model.NewAESKey(decryptedKey)
	zero.Bytes(decryptedKey)

	return cryptoKeyPriv, nil
}

const (
	IdentityTypeRoot        = "Root"
	IdentityTypeVerinym     = "Verinym"
	IdentityTypePersona     = "Persona"
	IdentityTypeDigitalTwin = "DigitalTwin"
	IdentityTypePairwise    = "PairwiseIdentity"
	IdentityTypeAnonymous   = "AnonymousIdentity"
)

func IsCorrectIdentityType(val string) bool {
	return val == IdentityTypeVerinym || val == IdentityTypePersona || val == IdentityTypeDigitalTwin || val == IdentityTypePairwise || val == IdentityTypeAnonymous
}

type Identity struct {
	// DID is the identity's full DID definition, including its keys.
	DID *model.DID `json:"did"`
	// Created is the time when the identity was created.
	Created *time.Time `json:"created"`
	// Name is the name of the identity (only accessible to the account owner
	// for navigation/documentation purposes).
	Name string `json:"name,omitempty"`
	// Type is the identity's type.
	Type string `json:"type"`
	// AccessLevel is the identity's access level. Data wallet needs to
	// be unlocked to a specific access level to gain access to identities
	// at this level or higher.
	AccessLevel model.AccessLevel `json:"level"`
	// Lockers field is only used for imports to consolidate
	// the data in one structure (Identity). This field is always
	// empty, when Data Wallet returns the identity.
	Lockers []*model.Locker `json:"lockers,omitempty"`
}

func (idy *Identity) ID() string {
	return idy.DID.ID
}

func (idy *Identity) Copy() *Identity {
	cp := *idy
	return &cp
}

func (idy *Identity) Bytes() []byte {
	b, _ := jsonw.Marshal(idy)
	return b
}

type accountOptions struct {
	passphrase             string
	hashedPassphrase       string
	registrationCode       string
	parentAccount          *Account
	masterNode             slip10.Node
	didMethod              string
	rootIdentity           *model.DID
	firstBlock             int64
	entropyFunc            EntropyFunction
	secondLevelRecoveryKey []byte
	log                    *zerolog.Logger
}

// Option is for defining parameters when creating new accounts
type Option func(opts *accountOptions) error

func WithPassphraseAuth(passphrase string) Option {
	return func(opts *accountOptions) error {
		opts.passphrase = passphrase
		return nil
	}
}

func WithHashedPassphraseAuth(hashedPassphrase string) Option {
	return func(opts *accountOptions) error {
		opts.hashedPassphrase = hashedPassphrase
		return nil
	}
}

func WithRegistrationCode(regCode string) Option {
	return func(opts *accountOptions) error {
		opts.registrationCode = regCode
		return nil
	}
}

func WithCustomEntropy(entropyFunc EntropyFunction) Option {
	return func(opts *accountOptions) error {
		opts.entropyFunc = entropyFunc
		return nil
	}
}

func WithSLRK(secondLevelRecoveryKey []byte) Option {
	return func(opts *accountOptions) error {
		opts.secondLevelRecoveryKey = secondLevelRecoveryKey
		return nil
	}
}

func WithLogger(logInstance *zerolog.Logger) Option {
	return func(opts *accountOptions) error {
		opts.log = logInstance
		return nil
	}
}

func WithMaster(parentAcct *Account, masterNode slip10.Node) Option {
	return func(opts *accountOptions) error {
		opts.parentAccount = parentAcct
		opts.masterNode = masterNode
		return nil
	}
}

func WithDIDMethod(method string) Option {
	return func(opts *accountOptions) error {
		opts.didMethod = method
		return nil
	}
}

func WithRootIdentity(rootIdentity *model.DID) Option {
	return func(opts *accountOptions) error {
		opts.rootIdentity = rootIdentity
		return nil
	}
}

func WithFirstBlock(firstBlock int64) Option {
	return func(opts *accountOptions) error {
		opts.firstBlock = firstBlock
		return nil
	}
}

type GenerationResponse struct {
	Account                 *Account
	RegistrationCode        string
	RecoveryPhrase          string
	SecondLevelRecoveryCode string
	RootIdentities          []*Identity
	EncryptedIdentities     []*DataEnvelope
	EncryptedLockers        []*DataEnvelope
}

//nolint:gocyclo
func GenerateAccount(acctTemplate *Account, opts ...Option) (*GenerationResponse, error) {
	var options accountOptions
	for _, fn := range opts {
		if err := fn(&options); err != nil {
			return nil, err
		}
	}

	var logger zerolog.Logger
	if options.log != nil {
		logger = *options.log
	} else {
		logger = log.Logger
	}

	if options.entropyFunc == nil {
		options.entropyFunc = DefaultEntropyFunction()
	}

	if options.passphrase != "" {
		options.hashedPassphrase = HashUserPassword(options.passphrase)
	}

	if acctTemplate.AccessLevel != model.AccessLevelManaged && acctTemplate.AccessLevel != model.AccessLevelHosted {
		return nil, fmt.Errorf("can't create an account with access level %d", acctTemplate.AccessLevel)
	}

	acct := acctTemplate.Copy()

	// fill missing and mandatory fields

	if acct.RegisteredAt == nil {
		now := time.Now().UTC()
		acct.RegisteredAt = &now
	}

	var err error
	rootDID := options.rootIdentity
	if rootDID == nil {
		rootDID, err = model.GenerateDID(model.WithMethod(options.didMethod))
		if err != nil {
			return nil, err
		}
	}

	rootIdy := &Identity{
		Type:        IdentityTypeRoot,
		Name:        "Root Identity",
		AccessLevel: model.AccessLevelManaged,
		DID:         rootDID,
		Created:     acct.RegisteredAt,
	}
	acct.ID = rootDID.ID

	if acct.Type == "" {
		acct.Type = Type
	}

	acct.Version = Version
	acct.State = StateActive

	var recoveryPhrase string
	var secondLevelRecoveryCode string
	var hostedCryptoKey *model.AESKey
	var masterNode slip10.Node
	if options.masterNode == nil {
		// Generate a mnemonic recovery phrase for memorization
		entropy := options.entropyFunc()
		recoveryPhrase, err = bip39.NewMnemonic(entropy)
		if err != nil {
			return nil, err
		}

		cryptoKey, recoveryPublicKey, privKey, err := GenerateKeysFromRecoveryPhrase(recoveryPhrase)
		if err != nil {
			return nil, err
		}

		hostedCryptoKey = cryptoKey

		acct.RecoveryPublicKey = base64.StdEncoding.EncodeToString(recoveryPublicKey)
		acct.EncryptedRecoverySecret = base64.StdEncoding.EncodeToString(
			model.AnonEncrypt(hostedCryptoKey[:], recoveryPublicKey),
		)

		if options.secondLevelRecoveryKey != nil {
			secondLevelRecoveryCode = base58.Encode(
				model.AnonEncrypt(privKey, options.secondLevelRecoveryKey))
		}

		seed, err := slip10.GenerateSeed(slip10.RecommendedSeedLen)
		if err != nil {
			return nil, err
		}

		// Derive the master extended key from the seed.
		masterNode, err = slip10.NewMasterNode(seed)
		if err != nil {
			return nil, err
		}
	} else {
		masterNode = options.masterNode

		hostedCryptoKey = GenerateHostedKeyFromNode(masterNode)
	}

	if options.parentAccount != nil {
		master := options.parentAccount.MasterAccount
		if master == "" {
			master = options.parentAccount.ID
		}
		acct.MasterAccount = master
		acct.ParentAccount = options.parentAccount.ID
	}

	managedCryptoKey := GenerateManagedFromHostedKey(hostedCryptoKey)

	managedHMACKey := GenerateIDHMACKey()
	managedEncryptionKey := model.NewEncryptionKey()

	result := &GenerationResponse{
		Account:                 acct,
		RegistrationCode:        options.registrationCode,
		RecoveryPhrase:          recoveryPhrase,
		SecondLevelRecoveryCode: secondLevelRecoveryCode,
	}

	idyEnvelope, err := EncryptIdentity(rootIdy, managedHMACKey, managedEncryptionKey)
	if err != nil {
		return nil, err
	}
	result.RootIdentities = append(result.RootIdentities, rootIdy)
	result.EncryptedIdentities = append(result.EncryptedIdentities, idyEnvelope)

	// generate root managed locker

	rootManagedLocker, err := model.GenerateLocker(model.AccessLevelManaged, "Root Managed Locker",
		nil, options.firstBlock, model.Us(rootIdy.DID, nil))
	if err != nil {
		return nil, err
	}

	lockerEnv, err := EncryptLocker(rootManagedLocker, managedHMACKey, managedEncryptionKey)
	if err != nil {
		return nil, err
	}
	result.EncryptedLockers = append(result.EncryptedLockers, lockerEnv)

	// Generate new master keys.  These master keys are used to protect the
	// crypto keys that will be generated next.

	switch acct.AccessLevel {
	case model.AccessLevelHosted:
		hostedSecretStore := &SecretStore{
			AccessLevel: model.AccessLevelHosted,
		}

		if options.passphrase != "" {
			hostedMasterPassphrase := []byte(options.passphrase)

			masterKeyPriv, err := newSecretKey(&hostedMasterPassphrase, hostedAccountConfig)
			zero.Bytes(hostedMasterPassphrase)
			if err != nil {
				log.Err(err).Msg("Failed to create master key")
				return nil, err
			}

			// Encrypt the crypto keys with the associated master keys.
			cryptoKeyPrivEncrypted, err := masterKeyPriv.Encrypt(hostedCryptoKey.Bytes())
			if err != nil {
				return nil, err
				//return nil, errors.New("failed to encrypt crypto private key")
			}

			hostedSecretStore.MasterKeyParams = base64.StdEncoding.EncodeToString(masterKeyPriv.Marshal())
			hostedSecretStore.EncryptedPayloadKey = base64.StdEncoding.EncodeToString(cryptoKeyPrivEncrypted)

			masterKeyPriv.Zero()

		} else if options.hashedPassphrase != "" {
			return nil, errors.New("only hashed passphrase provided for hosted account")
		}

		hostedHMACKey := GenerateIDHMACKey()
		hostedEncryptionKey := model.NewEncryptionKey()

		// generate root hosted locker

		rootHostedLocker, err := model.GenerateLocker(model.AccessLevelHosted, "Root Hosted Locker",
			nil, options.firstBlock, model.Us(rootIdy.DID, nil))
		if err != nil {
			return nil, err
		}

		lockerEnv, err := EncryptLocker(rootHostedLocker, hostedHMACKey, hostedEncryptionKey)
		if err != nil {
			return nil, err
		}
		result.EncryptedLockers = append(result.EncryptedLockers, lockerEnv)

		err = hostedSecretStore.UpdatePayload(&SecretStorePayload{
			ManagedHMACKey:       base64.StdEncoding.EncodeToString(managedHMACKey),
			ManagedEncryptionKey: base64.StdEncoding.EncodeToString(managedEncryptionKey[:]),
			HostedHMACKey:        base64.StdEncoding.EncodeToString(hostedHMACKey),
			HostedEncryptionKey:  base64.StdEncoding.EncodeToString(hostedEncryptionKey[:]),
			AccountRootKey:       masterNode.Serialize(),
			ManagedRootLocker:    rootManagedLocker.ID,
			HostedRootLocker:     rootHostedLocker.ID,
		}, hostedCryptoKey)

		hostedEncryptionKey.Zero()
		zero.Bytes(hostedHMACKey)

		if err != nil {
			return nil, err
		}

		acct.HostedSecretStore = hostedSecretStore

	case model.AccessLevelManaged:
		if options.hashedPassphrase != "" {
			acct.EncryptedPassword = options.hashedPassphrase
		}
	default:
		return nil, fmt.Errorf("can't create an account with access level %d", acctTemplate.AccessLevel)
	}

	managedSecretStore := &SecretStore{
		AccessLevel: model.AccessLevelManaged,
	}

	if options.hashedPassphrase != "" {
		acct.EncryptedPassword = options.hashedPassphrase

		managedMasterPassphrase := []byte(options.hashedPassphrase)

		managedMasterKeyPriv, err := newSecretKey(&managedMasterPassphrase, managedAccountConfig)
		if err != nil {
			logger.Err(err).Msg("Failed to create master key")
			return nil, err
		}

		// Encrypt the crypto keys with the associated master keys.
		managedCryptoKeyPrivEncrypted, err := managedMasterKeyPriv.Encrypt(managedCryptoKey.Bytes())
		if err != nil {
			return nil, err
		}

		managedSecretStore.MasterKeyParams = base64.StdEncoding.EncodeToString(managedMasterKeyPriv.Marshal())
		managedSecretStore.EncryptedPayloadKey = base64.StdEncoding.EncodeToString(managedCryptoKeyPrivEncrypted)

		managedMasterKeyPriv.Zero()
	}

	err = managedSecretStore.UpdatePayload(&SecretStorePayload{
		ManagedHMACKey:       base64.StdEncoding.EncodeToString(managedHMACKey),
		ManagedEncryptionKey: base64.StdEncoding.EncodeToString(managedEncryptionKey[:]),
		AccountRootKey:       masterNode.Serialize(),
		ManagedRootLocker:    rootManagedLocker.ID,
	}, managedCryptoKey)
	if err != nil {
		return nil, err
	}

	managedEncryptionKey.Zero()
	zero.Bytes(managedHMACKey)

	acct.ManagedSecretStore = managedSecretStore

	return result, nil
}

type EntropyFunction func() []byte

func DefaultEntropyFunction() EntropyFunction {
	return func() []byte {
		entropy, _ := bip39.NewEntropy(256)
		return entropy
	}
}

func GenerateIDHMACKey() []byte {
	secret := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, secret)
	if err != nil {
		panic("failed to generate random secret")
	}
	return secret
}

func GenerateHostedKeyFromNode(node slip10.Node) *model.AESKey {
	return model.NewAESKey(model.Hash("crypto key", node.PrivateKey()))
}

func GenerateManagedFromHostedKey(hostedKey *model.AESKey) *model.AESKey {
	return model.NewAESKey(model.Hash("managed key", hostedKey.Bytes()))
}
