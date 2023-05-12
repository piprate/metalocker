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

package wallet

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/piprate/metalocker/index"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/model/dataset"
	"github.com/piprate/metalocker/model/expiry"
	"github.com/piprate/metalocker/model/slip10"
	"github.com/piprate/metalocker/storage"
	"github.com/piprate/metalocker/utils"
	"github.com/piprate/metalocker/utils/measure"
	"github.com/piprate/metalocker/utils/zero"
	"github.com/rs/zerolog/log"
)

const (
	// saltSize is the number of bytes of the salt used when hashing private passphrases.
	saltSize = 32
)

type (
	LocalDataWallet struct {
		acct      *account.Account
		lockLevel model.AccessLevel
		lockMtx   sync.RWMutex

		hostedCryptoKey  *model.AESKey
		managedCryptoKey *model.AESKey

		// privPassphraseSalt and hashedPrivPassphrase allow for the secure
		// detection of a correct passphrase on wallet unlock when the
		// wallet is already unlocked.  The hash is zeroed each lock.
		privPassphraseSalt   [saltSize]byte
		hashedPrivPassphrase [sha512.Size]byte

		secretStoreValues map[model.AccessLevel]*account.SecretStorePayload

		managedHMACKey       []byte
		managedEncryptionKey *model.AESKey
		managedRootLockerID  string

		hostedHMACKey       []byte
		hostedEncryptionKey *model.AESKey
		hostedRootLockerID  string

		node slip10.Node

		dataMtx        sync.RWMutex
		identities     map[string]Identity
		identityHashes map[string]bool
		lockers        map[string]Locker
		lockerHashes   map[string]bool

		nodeClient   NodeClient
		indexClient  index.Client
		dataStoreFn  DataSetStoreConstructor
		datasetStore DataStore

		confirmAccountUpdates bool
	}
)

var _ DataWallet = (*LocalDataWallet)(nil)

func NewLocalDataWallet(acct *account.Account, nodeClient NodeClient, dataStoreFn DataSetStoreConstructor, indexClient index.Client) (*LocalDataWallet, error) {

	if dataStoreFn == nil {
		dataStoreFn = newLocalDataStore
	}

	wallet := &LocalDataWallet{
		acct:        acct,
		lockLevel:   model.AccessLevelNone,
		nodeClient:  nodeClient,
		indexClient: indexClient,
		dataStoreFn: dataStoreFn,

		identities:     make(map[string]Identity),
		identityHashes: make(map[string]bool),
		lockers:        make(map[string]Locker),
		lockerHashes:   make(map[string]bool),

		confirmAccountUpdates: false,
	}

	datasetStore, err := dataStoreFn(wallet, nodeClient)
	if err != nil {
		return nil, err
	}
	wallet.datasetStore = datasetStore

	switch acct.AccessLevel {
	case model.AccessLevelHosted:
		// Generate private passphrase salt.
		var privPassphraseSalt [saltSize]byte
		_, err := rand.Read(privPassphraseSalt[:])
		if err != nil {
			return nil, errors.New("failed to read random source for passphrase salt")
		}
		wallet.privPassphraseSalt = privPassphraseSalt
		wallet.lockLevel = model.AccessLevelNone
	case model.AccessLevelManaged:
		wallet.lockLevel = model.AccessLevelNone
	case model.AccessLevelRestricted:
		wallet.lockLevel = model.AccessLevelRestricted
	default:
		return nil, fmt.Errorf("unsupported access level: %d", acct.AccessLevel)
	}

	return wallet, nil
}

func (dw *LocalDataWallet) Close() error {
	_ = dw.Lock()

	if err := dw.nodeClient.Close(); err != nil {
		log.Warn().AnErr("error", err).Msg("Error when closing node client")
	}

	// help garbage collector
	dw.nodeClient = nil
	dw.datasetStore = nil

	return nil
}

func (dw *LocalDataWallet) ID() string {
	return dw.acct.ID
}

func (dw *LocalDataWallet) Account() *account.Account {
	return dw.acct.Copy()
}

// lock performs a best try effort to remove and zero all secret keys associated
// with the wallet.
//
// This function MUST be called with the wallet lock held for writes.
func (dw *LocalDataWallet) lock() {

	for _, idy := range dw.identities {
		// clear all the identity's private keys
		idy.Raw().DID.Zero()
	}

	for _, l := range dw.lockers {
		// clear all the locker's secrets
		l.Raw().Zero()
	}

	// then erase all identity information
	dw.identities = make(map[string]Identity)
	dw.identityHashes = make(map[string]bool)

	// then erase all locker information
	dw.lockers = make(map[string]Locker)
	dw.lockerHashes = make(map[string]bool)

	if dw.acct.AccessLevel == model.AccessLevelHosted {
		// Remove clear text private master and crypto keys from memory.
		if dw.hostedCryptoKey != nil {
			dw.hostedCryptoKey.Zero()
		}

		// Zero the hashed passphrase.
		zero.Bytea64(&dw.hashedPrivPassphrase)
	}

	for _, p := range dw.secretStoreValues {
		p.Zero()
	}
	dw.secretStoreValues = nil

	if dw.managedHMACKey != nil {
		zero.Bytes(dw.managedHMACKey)
		dw.managedHMACKey = nil
		dw.managedEncryptionKey.Zero()
		dw.managedEncryptionKey = nil
		dw.managedRootLockerID = ""
	}
	if dw.hostedHMACKey != nil {
		zero.Bytes(dw.hostedHMACKey)
		dw.hostedHMACKey = nil
		dw.hostedEncryptionKey.Zero()
		dw.hostedEncryptionKey = nil
		dw.hostedRootLockerID = ""
	}
	if dw.node != nil {
		dw.node.Zero()
	}

	dw.lockLevel = model.AccessLevelNone
}

// LockLevel returns the current level of wallet access.
func (dw *LocalDataWallet) LockLevel() model.AccessLevel {
	dw.lockMtx.RLock()
	defer dw.lockMtx.RUnlock()

	return dw.lockLevel
}

// Lock performs a best try effort to remove and zero all secret keys associated
// with the wallet.
//
// This function will return an error if invoked on a watching-only wallet.
func (dw *LocalDataWallet) Lock() error {
	dw.lockMtx.Lock()
	defer dw.lockMtx.Unlock()

	// Error on attempt to lock an already locked wallet.
	if dw.lockLevel == model.AccessLevelNone {
		return errors.New("data wallet already locked")
	}

	dw.lock()

	return nil
}

func (dw *LocalDataWallet) Unlock(passphrase string) error {
	defer measure.ExecTime("wallet.Unlock")()

	if dw.lockLevel != model.AccessLevelNone {
		// Avoid actually unlocking if the wallet is already unlocked
		// and the passphrases match.
		if dw.lockLevel == model.AccessLevelHosted {
			saltedPassphrase := append(dw.privPassphraseSalt[:],
				passphrase...)
			hashedPassphrase := sha512.Sum512(saltedPassphrase)
			zero.Bytes(saltedPassphrase)
			if hashedPassphrase != dw.hashedPrivPassphrase {
				dw.lock()
				return errors.New("invalid passphrase for master key")
			}
			return nil
		}

		return errors.New("data wallet already unlocked")
	}

	passphraseHash := account.HashUserPassword(passphrase)

	if dw.acct.AccessLevel == model.AccessLevelManaged {

		key, err := dw.acct.ManagedSecretStore.ExtractPayloadKey(passphraseHash)
		if err != nil {
			return err
		}
		return dw.UnlockAsManaged(key)
	}

	if dw.acct.AccessLevel != model.AccessLevelHosted && dw.acct.AccessLevel != model.AccessLevelLocal {
		return errors.New("can't unlock data wallet with passphrase: wrong wallet type")
	}

	dw.lockMtx.Lock()
	defer dw.lockMtx.Unlock()

	managedKey, err := dw.acct.ManagedSecretStore.ExtractPayloadKey(passphraseHash)
	if err != nil {
		return err
	}

	dw.managedCryptoKey = managedKey

	hostedKey, err := dw.acct.HostedSecretStore.ExtractPayloadKey(passphrase)
	if err != nil {
		return err
	}

	if err = dw.decryptHostedSecrets(hostedKey); err != nil {
		return err
	}

	saltedPassphrase := append(dw.privPassphraseSalt[:], passphrase...)
	dw.hashedPrivPassphrase = sha512.Sum512(saltedPassphrase)
	zero.Bytes(saltedPassphrase)

	dw.lockLevel = dw.acct.AccessLevel

	return nil
}

func (dw *LocalDataWallet) decryptHostedSecrets(hostedKey *model.AESKey) error {
	dw.hostedCryptoKey = hostedKey

	payload, err := dw.acct.HostedSecretStore.GetPayload(hostedKey)
	if err != nil {
		dw.lock()
		return err
	}

	dw.managedHMACKey, _ = base64.StdEncoding.DecodeString(payload.ManagedHMACKey)
	val, _ := base64.StdEncoding.DecodeString(payload.ManagedEncryptionKey)
	dw.managedEncryptionKey = model.NewAESKey(val)
	dw.hostedHMACKey, _ = base64.StdEncoding.DecodeString(payload.HostedHMACKey)
	val, _ = base64.StdEncoding.DecodeString(payload.HostedEncryptionKey)
	dw.hostedEncryptionKey = model.NewAESKey(val)
	dw.node, _ = slip10.NewNodeFromString(payload.AccountRootKey)
	dw.managedRootLockerID = payload.ManagedRootLocker
	dw.hostedRootLockerID = payload.HostedRootLocker

	dw.dataMtx.Lock()
	for _, idy := range payload.Identities {
		cleanIdy := idy.Copy()
		cleanIdy.Lockers = nil
		dw.identities[cleanIdy.ID()] = newIdentityWrapper(dw, cleanIdy)
	}
	dw.dataMtx.Unlock()

	// add included lockers

	for _, idy := range payload.Identities {
		for _, locker := range idy.Lockers {
			if err = dw.addLocker(newLockerWrapper(dw, locker)); err != nil {
				return err
			}
		}
	}

	dw.secretStoreValues = map[model.AccessLevel]*account.SecretStorePayload{
		model.AccessLevelHosted: payload,
	}

	return nil
}

func (dw *LocalDataWallet) decryptManagedSecrets(managedKey *model.AESKey) error {
	dw.managedCryptoKey = managedKey

	payload, err := dw.acct.ManagedSecretStore.GetPayload(managedKey)
	if err != nil {
		dw.lock()
		return err
	}

	dw.managedHMACKey, _ = base64.StdEncoding.DecodeString(payload.ManagedHMACKey)
	val, _ := base64.StdEncoding.DecodeString(payload.ManagedEncryptionKey)
	dw.managedEncryptionKey = model.NewAESKey(val)
	dw.node, _ = slip10.NewNodeFromString(payload.AccountRootKey)
	dw.managedRootLockerID = payload.ManagedRootLocker

	dw.secretStoreValues = map[model.AccessLevel]*account.SecretStorePayload{
		model.AccessLevelManaged: payload,
	}

	return nil
}

func (dw *LocalDataWallet) UnlockAsManaged(managedKey *model.AESKey) error {
	defer measure.ExecTime("wallet.UnlockAsManaged")()

	if dw.lockLevel != model.AccessLevelNone {
		return errors.New("data wallet already unlocked")
	}

	if managedKey == nil {
		return errors.New("managed key not provided")
	}

	if dw.acct.AccessLevel != model.AccessLevelManaged && dw.acct.AccessLevel != model.AccessLevelHosted {
		return errors.New("can't unlock data wallet with client secret: wrong wallet type")
	}

	dw.lockMtx.Lock()
	defer dw.lockMtx.Unlock()

	if err := dw.decryptManagedSecrets(managedKey); err != nil {
		return err
	}

	dw.lockLevel = model.AccessLevelManaged

	return nil
}

func (dw *LocalDataWallet) UnlockWithAccessKey(apiKey, apiSecret string) error {
	defer measure.ExecTime("wallet.UnlockWithAccessKey")()

	if dw.lockLevel != model.AccessLevelNone {
		return errors.New("data wallet already unlocked")
	}

	accessKey, err := dw.GetAccessKey(apiKey)
	if err != nil {
		return err
	}
	err = accessKey.Hydrate(apiSecret)
	if err != nil {
		return err
	}

	if dw.acct.AccessLevel != model.AccessLevelManaged && dw.acct.AccessLevel != model.AccessLevelHosted {
		return errors.New("can't unlock data wallet with an access key: wrong wallet type")
	}

	dw.lockMtx.Lock()
	defer dw.lockMtx.Unlock()

	if accessKey.EncryptedManagedKey != "" {
		var err error
		dw.managedCryptoKey, err = model.DecodeAESKey(accessKey.EncryptedManagedKey, accessKey.ManagementKeyPrv)
		if err != nil {
			return err
		}
		if accessKey.EncryptedHostedKey == "" {
			if err := dw.decryptManagedSecrets(dw.managedCryptoKey); err != nil {
				return err
			}

			dw.lockLevel = model.AccessLevelManaged
		}
	}

	if accessKey.EncryptedHostedKey != "" {
		var err error
		hostedKey, err := model.DecodeAESKey(accessKey.EncryptedHostedKey, accessKey.ManagementKeyPrv)
		if err != nil {
			return err
		}

		if err = dw.decryptHostedSecrets(hostedKey); err != nil {
			return err
		}

		dw.lockLevel = model.AccessLevelHosted
	}

	if dw.lockLevel < model.AccessLevelManaged {
		return errors.New("access key doesn't have sufficient permissions to unlock the wallet")
	}

	return nil
}

func (dw *LocalDataWallet) UnlockAsChild(parentNode slip10.Node) error {
	defer measure.ExecTime("wallet.UnlockAsChild")()

	if dw.lockLevel != model.AccessLevelNone {
		return errors.New("data wallet already unlocked")
	}

	dw.lockMtx.Lock()
	defer dw.lockMtx.Unlock()

	childNode, err := parentNode.Derive(dw.acct.DerivationIndex)
	if err != nil {
		return err
	}

	hostedCryptoKey := account.GenerateHostedKeyFromNode(childNode)

	dw.managedCryptoKey = account.GenerateManagedFromHostedKey(hostedCryptoKey)

	switch dw.acct.AccessLevel {
	case model.AccessLevelManaged:
		if err = dw.decryptManagedSecrets(dw.managedCryptoKey); err != nil {
			return err
		}
	case model.AccessLevelHosted:
		dw.hostedCryptoKey = hostedCryptoKey
		if err = dw.decryptHostedSecrets(hostedCryptoKey); err != nil {
			return err
		}
	default:
		return errors.New("unsupported account access level when unlocking a child node")
	}

	dw.lockLevel = dw.acct.AccessLevel

	return nil
}

func (dw *LocalDataWallet) addLocker(locker Locker) error {
	if !locker.Raw().IsHydrated() {
		// hydrate locker
		for _, party := range locker.Raw().Participants {
			var signKey ed25519.PrivateKey
			if party.Self {
				idy, err := dw.GetIdentity(party.ID)
				if err != nil {
					return err
				}
				signKey = idy.DID().SignKeyValue()
			}
			if err := party.Hydrate(signKey); err != nil {
				return err
			}
		}
	}

	dw.dataMtx.Lock()
	dw.lockers[locker.ID()] = locker
	dw.dataMtx.Unlock()

	return nil
}

func (dw *LocalDataWallet) GetRootIdentity() (Identity, error) {
	return dw.GetIdentity(dw.acct.ID)
}

func (dw *LocalDataWallet) GetRootLocker(level model.AccessLevel) (Locker, error) {
	if level == model.AccessLevelManaged {
		return dw.GetLocker(dw.managedRootLockerID)
	} else if level >= model.AccessLevelHosted {
		return dw.GetLocker(dw.hostedRootLockerID)
	} else {
		return nil, fmt.Errorf("access level not supported for root lockers: %d", level)
	}
}

func (dw *LocalDataWallet) getRootLockerID(level model.AccessLevel) string {
	if level == model.AccessLevelManaged {
		return dw.managedRootLockerID
	} else if level >= model.AccessLevelHosted {
		return dw.hostedRootLockerID
	} else {
		return ""
	}
}

func (dw *LocalDataWallet) sendAccountUpdate(au *AccountUpdate, wait bool) (int64, error) {
	rootLockerID := dw.getRootLockerID(au.AccessLevel)
	lb, err := dw.DataStore().NewDataSetBuilder(rootLockerID,
		dataset.WithVault(dw.acct.DefaultVault))
	if err != nil {
		return 0, err
	}

	mrid, err := lb.AddMetaResource(au)
	if err != nil {
		return 0, err
	}

	now := time.Now().UTC()
	err = lb.AddProvenance(mrid, &model.ProvEntity{
		ID:              mrid,
		Type:            model.ProvTypeEntity,
		GeneratedAtTime: &now,
		WasGeneratedBy: &model.ProvActivity{
			Type:              model.ProvTypeActivity,
			WasAssociatedWith: lb.CreatorID(),
			Algorithm:         "DataWallet",
		},
	}, false)
	if err != nil {
		return 0, err
	}

	f := lb.Submit(expiry.Years(MetaLeaseDurationYears))

	if wait && f.Error() == nil {
		if err = f.Wait(time.Minute); err != nil {
			return 0, err
		}

		return f.DataSet().BlockNumber(), nil
	} else {
		return 0, f.Error()
	}
}

func (dw *LocalDataWallet) AddLocker(locker *model.Locker) (Locker, error) {
	if locker.AccessLevel == model.AccessLevelNone {
		return nil, errors.New("locker access level is not provided")
	}

	if locker.AccessLevel == model.AccessLevelLocal {
		if dw.lockLevel < model.AccessLevelHosted {
			return nil, ErrInsufficientLockLevel
		}
	} else if dw.lockLevel < locker.AccessLevel {
		return nil, ErrInsufficientLockLevel
	}

	locker = locker.Copy()

	acceptedAt, err := dw.sendAccountUpdate(&AccountUpdate{
		Type:          AccountUpdateType,
		AccountID:     dw.acct.ID,
		AccessLevel:   locker.AccessLevel,
		LockersOpened: []string{locker.ID},
	}, true)
	if err != nil {
		log.Err(err).Msg("Error when sending account update message")
		return nil, err
	}

	locker.SetAcceptedAtBlock(acceptedAt)

	if locker.FirstBlock == 0 {
		log.Warn().Msg("Missing first block for locker. Setting it to the acceptedAt block value")
		locker.FirstBlock = acceptedAt
	}

	wrapper := newLockerWrapper(dw, locker)
	err = dw.addLocker(wrapper)
	if err != nil {
		return nil, err
	}

	switch locker.AccessLevel {
	case model.AccessLevelManaged:
		envelope, err := dw.encryptLocker(locker)
		if err != nil {
			return nil, err
		}
		if err = dw.nodeClient.StoreLocker(envelope); err != nil {
			return nil, err
		}
	case model.AccessLevelHosted:
		envelope, err := dw.encryptLocker(locker)
		if err != nil {
			return nil, err
		}
		if err = dw.nodeClient.StoreLocker(envelope); err != nil {
			return nil, err
		}
	case model.AccessLevelLocal:
		if err := dw.flushToHostedSecretStore(); err != nil {
			return nil, errors.New("failed to build encrypted payload")
		}
		if err := dw.nodeClient.UpdateAccount(dw.acct); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("locker access level not supported: %d", locker.AccessLevel)
	}

	return wrapper, nil
}

func (dw *LocalDataWallet) AddIdentity(idy *account.Identity) error {
	if idy.AccessLevel == model.AccessLevelLocal {
		if dw.lockLevel < model.AccessLevelHosted {
			return ErrInsufficientLockLevel
		}
	} else if dw.lockLevel < idy.AccessLevel {
		return ErrInsufficientLockLevel
	}

	dw.dataMtx.RLock()
	_, alreadyExists := dw.identities[idy.ID()]
	dw.dataMtx.RUnlock()

	if alreadyExists {
		return fmt.Errorf("identity already exists: %s", idy.ID())
	}

	if idy.DID.SignKey == "" {
		return errors.New("DID sign key is not provided")
	}

	if idy.AccessLevel == model.AccessLevelNone {
		return errors.New("identity access level is not provided")
	}

	// check if identity exists
	publishDID := false
	existingDDoc, err := dw.nodeClient.DIDProvider().GetDIDDocument(idy.ID())
	if err != nil {
		if errors.Is(err, storage.ErrDIDNotFound) {
			publishDID = true
		} else {
			return err
		}
	} else {
		existingDid, err := existingDDoc.ExtractIndyStyleDID()
		if err != nil {
			log.Error().Str("did", idy.DID.ID).Msg("Existing and new identities can't be compared")
			return errors.New("identity mismatch")
		}
		if existingDid.VerKey != idy.DID.VerKey {
			log.Error().Str("orig", existingDid.VerKey).Str("new",
				idy.DID.VerKey).Msg("VerKey for new identity doesn't match")
			return errors.New("identity mismatch")
		}
	}

	cleanIdy := idy.Copy()
	cleanIdy.Lockers = nil

	dw.dataMtx.Lock()
	dw.identities[cleanIdy.ID()] = newIdentityWrapper(dw, cleanIdy)
	dw.dataMtx.Unlock()

	// save identity

	switch idy.AccessLevel {
	case model.AccessLevelManaged:
		envelope, err := dw.encryptIdentity(cleanIdy)
		if err != nil {
			return err
		}
		if err = dw.nodeClient.StoreIdentity(envelope); err != nil {
			return err
		}
	case model.AccessLevelHosted:
		envelope, err := dw.encryptIdentity(cleanIdy)
		if err != nil {
			return err
		}
		if err = dw.nodeClient.StoreIdentity(envelope); err != nil {
			return err
		}
	case model.AccessLevelLocal:
		if err := dw.flushToHostedSecretStore(); err != nil {
			return errors.New("failed to build encrypted payload")
		}
		if err = dw.nodeClient.UpdateAccount(dw.acct); err != nil {
			return err
		}
	default:
		return fmt.Errorf("identity access level not supported: %d", idy.AccessLevel)
	}

	_, err = dw.sendAccountUpdate(&AccountUpdate{
		Type:            AccountUpdateType,
		AccountID:       dw.acct.ID,
		AccessLevel:     cleanIdy.AccessLevel,
		IdentitiesAdded: []string{cleanIdy.ID()},
	}, dw.confirmAccountUpdates)
	if err != nil {
		log.Err(err).Msg("Error when sending account update message")
		return err
	}

	// add included lockers

	for _, locker := range idy.Lockers {
		if _, err = dw.AddLocker(locker); err != nil {
			return err
		}
	}

	if publishDID {
		dDoc, err := model.SimpleDIDDocument(idy.DID, idy.Created)
		if err != nil {
			return err
		}

		err = dw.nodeClient.DIDProvider().CreateDIDDocument(dDoc)
		if err != nil {
			return err
		}

		log.Debug().Str("did", idy.DID.ID).Msg("DID document published")
	}
	log.Debug().Str("did", idy.ID()).Msg("Added new identity")

	return nil
}

func (dw *LocalDataWallet) GetIdentities() (map[string]Identity, error) {
	if dw.lockLevel == model.AccessLevelNone {
		return nil, ErrWalletLocked
	}

	if dw.acct.AccessLevel == model.AccessLevelRestricted {
		return dw.identities, nil
	}

	idyEnvList, err := dw.nodeClient.ListIdentities()
	if err != nil {
		return nil, err
	}

	dw.dataMtx.Lock()
	defer dw.dataMtx.Unlock()

	for _, envelope := range idyEnvList {
		if envelope.AccessLevel <= dw.lockLevel {
			_, err := dw.loadIdentityEnvelope(envelope)
			if err != nil {
				return nil, err
			}
		}
	}

	return dw.identities, nil
}

func (dw *LocalDataWallet) GetIdentity(iid string) (Identity, error) {
	if dw.lockLevel == model.AccessLevelNone {
		return nil, ErrWalletLocked
	}

	dw.dataMtx.RLock()
	idy, found := dw.identities[iid]
	dw.dataMtx.RUnlock()

	if !found {
		managedHash := account.HashID(iid, dw.managedHMACKey)
		envelope, err := dw.nodeClient.GetIdentity(managedHash)
		if errors.Is(err, storage.ErrIdentityNotFound) {
			hostedHash := account.HashID(iid, dw.hostedHMACKey)
			envelope, err = dw.nodeClient.GetIdentity(hostedHash)
		}
		if err != nil {
			return nil, err
		}

		dw.dataMtx.Lock()
		defer dw.dataMtx.Unlock()

		if idy, err = dw.loadIdentityEnvelope(envelope); err != nil {
			return nil, err
		} else if idy == nil {
			dw.dataMtx.RLock()
			idy = dw.identities[iid]
			dw.dataMtx.RUnlock()
		}
	}
	return idy, nil
}

func (dw *LocalDataWallet) GetDID(iid string) (*model.DID, error) {
	dw.dataMtx.RLock()
	idy, found := dw.identities[iid]
	dw.dataMtx.RUnlock()

	if found {
		return idy.Raw().DID.NeuteredCopy(), nil
	} else {
		didDoc, err := dw.nodeClient.DIDProvider().GetDIDDocument(iid)
		if err != nil {
			return nil, err
		}

		return didDoc.ExtractIndyStyleDID()
	}
}

func (dw *LocalDataWallet) GetLockers() ([]*model.Locker, error) {
	if dw.lockLevel == model.AccessLevelNone {
		return nil, ErrWalletLocked
	}

	lockerEnvList, err := dw.nodeClient.ListLockers()
	if err != nil {
		return nil, err
	}

	for _, envelope := range lockerEnvList {
		if envelope.AccessLevel <= dw.lockLevel {
			_, err = dw.loadLockerEnvelope(envelope)
			if err != nil {
				return nil, err
			}
		}
	}

	dw.dataMtx.RLock()
	list := make([]*model.Locker, 0, len(dw.lockers))
	for _, l := range dw.lockers {
		list = append(list, l.Raw())
	}
	dw.dataMtx.RUnlock()

	return list, nil
}

func (dw *LocalDataWallet) GetLocker(lockerID string) (Locker, error) {
	if dw.lockLevel == model.AccessLevelNone {
		return nil, ErrWalletLocked
	}

	var locker Locker
	var found bool

	dw.dataMtx.RLock()
	locker, found = dw.lockers[lockerID]
	dw.dataMtx.RUnlock()

	if !found {
		managedHash := account.HashID(lockerID, dw.managedHMACKey)
		envelope, err := dw.nodeClient.GetLocker(managedHash)
		if errors.Is(err, storage.ErrLockerNotFound) {
			hostedHash := account.HashID(lockerID, dw.hostedHMACKey)
			envelope, err = dw.nodeClient.GetLocker(hostedHash)
		}
		if err != nil {
			return nil, err
		}

		if locker, err = dw.loadLockerEnvelope(envelope); err != nil {
			return nil, err
		} else if locker == nil {
			// the locker may have been loaded since this function call started
			dw.dataMtx.RLock()
			locker = dw.lockers[lockerID]
			dw.dataMtx.RUnlock()
		}
	}

	return locker, nil
}

func (dw *LocalDataWallet) GetProperty(key string) (string, error) {
	if dw.lockLevel == model.AccessLevelNone {
		return "", ErrWalletLocked
	}

	var envelope *account.DataEnvelope
	var err error

	if dw.lockLevel == model.AccessLevelHosted {
		hostedHash := account.HashID(key, dw.hostedHMACKey)
		envelope, err = dw.nodeClient.GetProperty(hostedHash)
	}
	if errors.Is(err, storage.ErrPropertyNotFound) || dw.lockLevel == model.AccessLevelManaged {
		managedHash := account.HashID(key, dw.managedHMACKey)
		envelope, err = dw.nodeClient.GetProperty(managedHash)
	}
	if err != nil {
		return "", err
	}
	if envelope == nil {
		return "", storage.ErrPropertyNotFound
	}

	return dw.decryptProperty(envelope, nil)
}

func (dw *LocalDataWallet) SetProperty(key string, value string, lvl model.AccessLevel) error {
	if dw.lockLevel < lvl {
		return ErrInsufficientLockLevel
	}

	env, err := dw.encryptProperty(key, value, lvl)
	if err != nil {
		return err
	}

	return dw.nodeClient.StoreProperty(env)
}

func (dw *LocalDataWallet) GetProperties() (map[string]string, error) {
	if dw.lockLevel == model.AccessLevelNone {
		return nil, ErrWalletLocked
	}

	envList, err := dw.nodeClient.ListProperties()
	if err != nil {
		return nil, err
	}

	res := make(map[string]string, len(envList))
	for _, envelope := range envList {
		if envelope.AccessLevel <= dw.lockLevel {
			var id string
			val, err := dw.decryptProperty(envelope, &id)
			if err != nil {
				return nil, err
			}

			if _, exists := res[id]; exists && envelope.AccessLevel < dw.lockLevel {
				// skip property values from less secure levels
				continue
			}
			res[id] = val
		}
	}

	return res, nil
}

func (dw *LocalDataWallet) DeleteProperty(key string, lvl model.AccessLevel) error {
	if dw.lockLevel == model.AccessLevelNone {
		return ErrWalletLocked
	}

	if dw.lockLevel < lvl {
		return ErrInsufficientLockLevel
	}

	if lvl == model.AccessLevelHosted {
		hostedHash := account.HashID(key, dw.hostedHMACKey)
		return dw.nodeClient.DeleteProperty(hostedHash)
	}
	if lvl == model.AccessLevelManaged {
		managedHash := account.HashID(key, dw.managedHMACKey)
		return dw.nodeClient.DeleteProperty(managedHash)
	}

	return fmt.Errorf("unsupported property access level: %d", lvl)
}

func (dw *LocalDataWallet) flushToHostedSecretStore() error {
	if dw.lockLevel < model.AccessLevelHosted {
		return ErrInsufficientLockLevel
	}

	iiList := make([]*account.Identity, 0)
	dw.dataMtx.RLock()
	for _, idy := range dw.identities {
		if idy.AccessLevel() == model.AccessLevelLocal {
			iiList = append(iiList, idy.Raw())
		}
	}
	dw.dataMtx.RUnlock()

	payload, found := dw.secretStoreValues[model.AccessLevelHosted]
	if !found {
		return errors.New("hosted secret store not found")
	}
	payload.Identities = iiList

	if err := dw.acct.HostedSecretStore.UpdatePayload(payload, dw.hostedCryptoKey); err != nil {
		return err
	}

	return nil
}

func (dw *LocalDataWallet) ChangeEmail(email string) error {
	dw.acct.Email = strings.ToLower(email)

	err := dw.nodeClient.PatchAccount(dw.acct.Email, "", "", "", "", "")
	if err != nil {
		return err
	}

	return nil
}

func (dw *LocalDataWallet) ChangePassphrase(oldPassphrase, newPassphrase string, isHash bool) (DataWallet, error) {
	if dw.acct.AccessLevel > dw.lockLevel {
		return nil, ErrInsufficientLockLevel
	}

	acct, err := account.ChangePassphrase(dw.acct, oldPassphrase, newPassphrase, isHash)
	if err != nil {
		return nil, err
	}

	err = dw.nodeClient.UpdateAccount(acct)
	if err != nil {
		return nil, err
	}

	newNodeClient, err := dw.nodeClient.NewInstance(acct.Email, newPassphrase, isHash)
	if err != nil {
		return nil, err
	}

	newDataWallet, err := NewLocalDataWallet(acct, newNodeClient, dw.dataStoreFn, dw.indexClient)
	if err != nil {
		return nil, err
	}

	if isHash {
		managedKey, err := acct.ExtractManagedKey(newPassphrase)
		if err != nil {
			return nil, err
		}

		if err = newDataWallet.UnlockAsManaged(managedKey); err != nil {
			return nil, errors.New("failed unlock the new wallet")
		}
	} else {
		if err = newDataWallet.Unlock(newPassphrase); err != nil {
			return nil, errors.New("failed unlock the new wallet")
		}
	}

	if err = dw.nodeClient.Close(); err != nil {
		log.Warn().AnErr("error", err).Msg("Error when closing node client")
	}

	return newDataWallet, nil
}

func (dw *LocalDataWallet) Recover(cryptoKey *model.AESKey, newPassphrase string) (DataWallet, error) {
	if dw.lockLevel != model.AccessLevelNone {
		return nil, errors.New("can't recover passphrase for unlocked wallet")
	}

	acct, err := account.Recover(dw.acct, cryptoKey, newPassphrase)
	if err != nil {
		return nil, err
	}

	newWallet, err := NewLocalDataWallet(acct, dw.nodeClient, dw.dataStoreFn, dw.indexClient)
	if err != nil {
		return nil, err
	}

	if err = newWallet.Unlock(newPassphrase); err != nil {
		return nil, err
	}

	// return a new wallet (unlocked)

	return newWallet, nil
}

func (dw *LocalDataWallet) CreateSubAccount(accessLevel model.AccessLevel, name string, opts ...account.Option) (DataWallet, error) {
	if accessLevel > dw.lockLevel {
		return nil, ErrInsufficientLockLevel
	}

	if accessLevel > dw.acct.AccessLevel {
		return nil, errors.New("requested sub-account access level is too high")
	}

	idxBytes := make([]byte, 4)
	_, err := rand.Read(idxBytes)
	if err != nil {
		return nil, err
	}
	derivationIndex := utils.BytesToUint32(idxBytes)
	if derivationIndex < slip10.FirstHardenedIndex {
		derivationIndex += slip10.FirstHardenedIndex
	}

	subNode, err := dw.node.Derive(derivationIndex)
	if err != nil {
		return nil, err
	}

	if name == "" {
		name = fmt.Sprintf("Sub-Account #%d", derivationIndex)
	}

	tb, err := dw.nodeClient.Ledger().GetTopBlock()
	if err != nil {
		return nil, err
	}

	acctTemplate := &account.Account{
		Name:            name,
		AccessLevel:     accessLevel,
		DerivationIndex: derivationIndex,
		DefaultVault:    dw.acct.DefaultVault,
	}

	opts = append(opts, account.WithMaster(dw.acct, subNode), account.WithFirstBlock(tb.Number))

	resp, err := account.GenerateAccount(acctTemplate, opts...)
	if err != nil {
		return nil, err
	}

	subNodeClient, err := dw.nodeClient.SubAccountInstance(resp.Account.ID)
	if err != nil {
		return nil, err
	}

	err = SaveNewAccount(resp, subNodeClient, "", nil)
	if err != nil {
		return nil, err
	}

	_, err = dw.sendAccountUpdate(&AccountUpdate{
		Type:             AccountUpdateType,
		AccountID:        dw.acct.ID,
		AccessLevel:      resp.Account.AccessLevel,
		SubAccountsAdded: []string{resp.Account.ID},
	}, dw.confirmAccountUpdates)
	if err != nil {
		log.Err(err).Msg("Error when sending account update message")
		return nil, err
	}

	return dw.GetSubAccountWallet(resp.Account.ID)
}

func (dw *LocalDataWallet) GetSubAccount(id string) (*account.Account, error) {
	return dw.nodeClient.GetAccount(id)
}

func (dw *LocalDataWallet) GetSubAccountWallet(id string) (DataWallet, error) {
	acct, err := dw.nodeClient.GetAccount(id)
	if err != nil {
		return nil, err
	}

	newNodeClient, err := dw.nodeClient.SubAccountInstance(id)
	if err != nil {
		return nil, err
	}

	subWallet, err := NewLocalDataWallet(acct, newNodeClient, dw.dataStoreFn, dw.indexClient)
	if err != nil {
		return nil, err
	}

	err = subWallet.UnlockAsChild(dw.node)
	if err != nil {
		return nil, err
	}

	return subWallet, nil
}

func (dw *LocalDataWallet) DeleteSubAccount(id string) error {
	return dw.nodeClient.DeleteAccount(id)
}

func (dw *LocalDataWallet) SubAccounts() ([]*account.Account, error) {
	return dw.nodeClient.ListSubAccounts(dw.acct.ID)
}

func (dw *LocalDataWallet) CreateAccessKey(accessLevel model.AccessLevel, duration time.Duration) (*model.AccessKey, error) {

	// we allow to create a key with managed secrets only for hosted accounts
	if dw.lockLevel < accessLevel {
		return nil, ErrInsufficientLockLevel
	}

	key, err := model.GenerateAccessKey(dw.acct.ID, accessLevel)
	if err != nil {
		return nil, err
	}

	if accessLevel >= model.AccessLevelManaged {
		key.AddManagedKey(dw.managedCryptoKey)
		if accessLevel >= model.AccessLevelHosted {
			key.AddHostedKey(dw.hostedCryptoKey)
		}
	}

	return dw.nodeClient.CreateAccessKey(key)
}

func (dw *LocalDataWallet) GetAccessKey(keyID string) (*model.AccessKey, error) {
	return dw.nodeClient.GetAccessKey(keyID)
}

func (dw *LocalDataWallet) RevokeAccessKey(keyID string) error {
	return dw.nodeClient.DeleteAccessKey(keyID)
}

func (dw *LocalDataWallet) AccessKeys() ([]*model.AccessKey, error) {
	return dw.nodeClient.ListAccessKeys()
}

func (dw *LocalDataWallet) encryptIdentity(idy *account.Identity) (*account.DataEnvelope, error) {
	switch idy.AccessLevel {
	case model.AccessLevelManaged:
		return account.EncryptIdentity(idy, dw.managedHMACKey, dw.managedEncryptionKey)
	case model.AccessLevelHosted:
		return account.EncryptIdentity(idy, dw.hostedHMACKey, dw.hostedEncryptionKey)
	default:
		return nil, fmt.Errorf("encryption not unsupported for access level %d", idy.AccessLevel)
	}
}

func (dw *LocalDataWallet) decryptIdentity(envelope *account.DataEnvelope) (*account.Identity, error) {
	if envelope.AccessLevel > dw.lockLevel {
		return nil, ErrInsufficientLockLevel
	}
	switch envelope.AccessLevel {
	case model.AccessLevelManaged:
		return account.DecryptIdentity(envelope, dw.managedEncryptionKey)
	case model.AccessLevelHosted:
		return account.DecryptIdentity(envelope, dw.hostedEncryptionKey)
	default:
		return nil, fmt.Errorf("decryption not unsupported for access level %d", envelope.AccessLevel)
	}
}

func (dw *LocalDataWallet) loadIdentityEnvelope(envelope *account.DataEnvelope) (Identity, error) {
	if _, found := dw.identityHashes[envelope.Hash]; !found {

		idy, err := dw.decryptIdentity(envelope)
		if err != nil {
			return nil, err
		}

		iw := newIdentityWrapper(dw, idy)

		dw.identities[idy.ID()] = iw

		dw.identityHashes[envelope.Hash] = true

		return iw, nil
	}

	return nil, nil
}

func (dw *LocalDataWallet) encryptLocker(locker *model.Locker) (*account.DataEnvelope, error) {
	switch locker.AccessLevel {
	case model.AccessLevelManaged:
		return account.EncryptLocker(locker, dw.managedHMACKey, dw.managedEncryptionKey)
	case model.AccessLevelHosted:
		return account.EncryptLocker(locker, dw.hostedHMACKey, dw.hostedEncryptionKey)
	default:
		return nil, fmt.Errorf("encryption not unsupported for access level %d", locker.AccessLevel)
	}
}

func (dw *LocalDataWallet) decryptLocker(envelope *account.DataEnvelope) (*model.Locker, error) {
	if envelope.AccessLevel > dw.lockLevel {
		return nil, ErrInsufficientLockLevel
	}
	switch envelope.AccessLevel {
	case model.AccessLevelManaged:
		return account.DecryptLocker(envelope, dw.managedEncryptionKey)
	case model.AccessLevelHosted:
		return account.DecryptLocker(envelope, dw.hostedEncryptionKey)
	default:
		return nil, fmt.Errorf("decryption not unsupported for access level %d", envelope.AccessLevel)
	}
}

func (dw *LocalDataWallet) loadLockerEnvelope(envelope *account.DataEnvelope) (Locker, error) {
	dw.dataMtx.Lock()
	_, found := dw.lockerHashes[envelope.Hash]
	dw.dataMtx.Unlock()

	if !found {
		lockerData, err := dw.decryptLocker(envelope)
		if err != nil {
			return nil, err
		}

		locker := newLockerWrapper(dw, lockerData)
		if err = dw.addLocker(locker); err != nil {
			return nil, err
		}

		dw.dataMtx.Lock()
		dw.lockerHashes[envelope.Hash] = true
		dw.dataMtx.Unlock()

		return locker, nil
	}

	return nil, nil
}

func (dw *LocalDataWallet) encryptProperty(key string, val string, lvl model.AccessLevel) (*account.DataEnvelope, error) {
	switch lvl {
	case model.AccessLevelManaged:
		return account.EncryptValue(key, val, lvl, dw.managedHMACKey, dw.managedEncryptionKey)
	case model.AccessLevelHosted:
		return account.EncryptValue(key, val, lvl, dw.hostedHMACKey, dw.hostedEncryptionKey)
	default:
		return nil, fmt.Errorf("encryption not unsupported for access level %d", lvl)
	}
}

func (dw *LocalDataWallet) decryptProperty(envelope *account.DataEnvelope, id *string) (string, error) {
	if envelope.AccessLevel > dw.lockLevel {
		return "", ErrInsufficientLockLevel
	}
	switch envelope.AccessLevel {
	case model.AccessLevelManaged:
		return account.DecryptValue(envelope, dw.managedEncryptionKey, id)
	case model.AccessLevelHosted:
		return account.DecryptValue(envelope, dw.hostedEncryptionKey, id)
	default:
		return "", fmt.Errorf("decryption not unsupported for access level %d", envelope.AccessLevel)
	}
}

func (dw *LocalDataWallet) Services() Services {
	return dw.nodeClient
}

type ForceSyncMessage struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

type AccountUpdateMessage struct {
	Type   string `json:"type"`
	UserID string `json:"id"`
}

func (dw *LocalDataWallet) RestrictedWallet(identities []string) (DataWallet, error) {
	//restrictedNodeClient := NewRestrictedNodeClient(identities, dw.nodeClient)

	newWallet, err := NewLocalDataWallet(dw.acct.RestrictedCopy(), dw.nodeClient, dw.dataStoreFn,
		dw.indexClient)
	if err != nil {
		return nil, err
	}

	for _, iid := range identities {
		idy, found := dw.identities[iid]
		if found {
			newWallet.identities[iid] = idy
			for _, l := range dw.lockers {
				if l.Raw().Us() != nil && l.Raw().Us().ID == iid {
					newWallet.lockers[l.ID()] = l
				}
			}
		}
	}

	return newWallet, nil
}

func (dw *LocalDataWallet) Backend() AccountBackend {
	return dw.nodeClient
}

func (dw *LocalDataWallet) CreateRootIndex(indexStoreName string) (index.RootIndex, error) {
	ix, err := dw.CreateIndex(indexStoreName, index.TypeRoot)
	if err != nil {
		return nil, err
	} else {
		return ix.(index.RootIndex), nil
	}
}

func (dw *LocalDataWallet) CreateIndex(indexStoreName, indexType string, opts ...index.Option) (index.Index, error) {
	if dw.lockLevel == model.AccessLevelNone {
		return nil, ErrWalletLocked
	}

	if dw.lockLevel < model.AccessLevelManaged {
		return nil, ErrInsufficientLockLevel
	}

	indexStore, err := dw.indexClient.IndexStore(indexStoreName)
	if err != nil {
		return nil, err
	}

	if dw.lockLevel != dw.acct.AccessLevel {
		log.Warn().Str("userID", dw.acct.ID).
			Msg("Creating root index at access level different to the account's level")
	}

	if indexStore.Properties().EncryptionMode == index.ModeClientEncryption {
		key, err := dw.EncryptionKey(dw.acct.ID, dw.lockLevel)
		if err != nil {
			return nil, err
		}

		opts = append(opts, index.WithEncryption(key[:]))
	}

	ix, err := indexStore.CreateIndex(dw.acct.ID, indexType, dw.lockLevel, opts...)
	if err != nil {
		return nil, err
	}

	// add root lockers

	iw, _ := ix.Writer()

	l, err := dw.GetLocker(dw.managedRootLockerID)
	if err != nil {
		return nil, err
	}
	if err = iw.AddLockerState(dw.ID(), l.ID(), l.Raw().FirstBlock); err != nil {
		return nil, err
	}

	if dw.lockLevel >= model.AccessLevelHosted {
		l, err := dw.GetLocker(dw.hostedRootLockerID)
		if err != nil {
			return nil, err
		}
		if err = iw.AddLockerState(dw.ID(), l.ID(), l.Raw().FirstBlock); err != nil {
			return nil, err
		}
	}

	if ai, ok := ix.(AccountIndex); ok {
		// initialise account index
		if err = InitAccountIndex(ai, dw); err != nil {
			return nil, err
		}
	}

	return ix, nil
}

func (dw *LocalDataWallet) EncryptionKey(tag string, accessLevel model.AccessLevel) (*model.AESKey, error) {
	if dw.lockLevel == model.AccessLevelNone {
		return nil, ErrWalletLocked
	}

	if dw.lockLevel < accessLevel {
		return nil, ErrInsufficientLockLevel
	}

	var cryptoKey *model.AESKey
	switch accessLevel {
	case model.AccessLevelManaged:
		cryptoKey = dw.managedCryptoKey
	case model.AccessLevelHosted:
		cryptoKey = dw.hostedCryptoKey
	default:
		return nil, errors.New("unsupported access level when generating an encryption key")
	}

	return model.NewAESKey(model.Hash(tag, cryptoKey[:])), nil
}

func (dw *LocalDataWallet) RootIndex() (index.RootIndex, error) {
	if dw.lockLevel == model.AccessLevelNone {
		return nil, ErrWalletLocked
	}

	if dw.lockLevel < model.AccessLevelManaged {
		return nil, ErrInsufficientLockLevel
	}

	ix, err := dw.indexClient.RootIndex(dw.acct.ID, dw.acct.AccessLevel)
	if err != nil {
		return nil, err
	}
	if ix.IsLocked() {
		key, err := dw.EncryptionKey(dw.acct.ID, dw.lockLevel)
		if err != nil {
			return nil, err
		}

		if err = ix.Unlock(key[:]); err != nil {
			return nil, err
		}
	}

	return ix, nil
}

func (dw *LocalDataWallet) Index(id string) (index.Index, error) {
	if dw.lockLevel == model.AccessLevelNone {
		return nil, ErrWalletLocked
	}

	if dw.lockLevel < model.AccessLevelManaged {
		return nil, ErrInsufficientLockLevel
	}

	ix, err := dw.indexClient.Index(dw.acct.ID, id)
	if err != nil {
		return nil, err
	}
	if ix.IsLocked() {
		key, err := dw.EncryptionKey(dw.acct.ID, dw.lockLevel)
		if err != nil {
			return nil, err
		}

		if err = ix.Unlock(key[:]); err != nil {
			return nil, err
		}
	}

	return ix, nil
}

func (dw *LocalDataWallet) IndexUpdater(indexes ...index.Index) (*IndexUpdater, error) {
	updater := NewIndexUpdater(dw.nodeClient.Ledger())
	err := updater.AddIndexes(dw, indexes...)
	if err != nil {
		_ = updater.Close()
		return nil, err
	}
	return updater, nil
}

func (dw *LocalDataWallet) DataStore() DataStore {
	return dw.datasetStore
}
