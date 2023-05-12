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
	"errors"
	"io"
	"time"

	"github.com/piprate/metalocker/index"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/model/dataset"
	"github.com/piprate/metalocker/model/slip10"
	"github.com/piprate/metalocker/services/notification"
)

var (
	ErrInsufficientLockLevel = errors.New("insufficient wallet lock level")
	ErrWalletLocked          = errors.New("data wallet is locked")
)

type (
	DataWalletBackendBuilderFn func(acct *account.Account) (NodeClient, error)

	// Identity is an interface to a specific identity, one of many, stored in the account's data wallet.
	Identity interface {
		// ID returns the identity's ID
		ID() string
		// DID returns the identity's full DID definition, including its keys.
		DID() *model.DID
		// CreatedAt returns the time when the identity was created.
		CreatedAt() *time.Time
		// Name returns the name of the identity (only accessible to the account owner
		// for navigation/documentation purposes).
		Name() string
		// SetName is NOT SUPPORTED YET.
		SetName(name string) error
		// AccessLevel returns the identity's access level. Data wallet needs to
		// be unlocked to a specific access level to gain access to identities
		// at this level or higher.
		AccessLevel() model.AccessLevel
		// Raw returns the raw identity definition (as stored in the backend).
		Raw() *account.Identity
		// NewLocker creates a new locker for the identity. Use Participant option
		// to add other participants to the locker.
		NewLocker(name string, options ...LockerOption) (Locker, error)
	}

	// Locker is an interface to the account's lockers (secure, persistent, bidirectional communication
	// channels between two or more participants).
	Locker interface {
		// ID returns the locker ID.
		ID() string
		// CreatedAt returns the locker's creation time. For documentation purposes only.
		CreatedAt() *time.Time
		// Name returns the locker's name. These names are useful for locker documentation purposes.
		// They aren't used in any data processing.
		Name() string
		// SetName is NOT SUPPORTED YET.
		SetName(name string) error
		// AccessLevel returns the locker's access level. Data wallet needs to be unlocked
		// to a specific access level to gain access to lockers at this level or higher.
		AccessLevel() model.AccessLevel
		// Raw returns the raw locker definition (as stored in the backend).
		Raw() *model.Locker

		// IsUniLocker returns true, if the locker has just one participant (is a 'uni-locker').
		IsUniLocker() bool
		// IsThirdParty returns true, if the account doesn't have control over any of the locker
		// participants, but has access to the locker's secrets (a delegated access).
		IsThirdParty() bool
		// Us returns the account controlled locker participant (if any).
		Us() *model.LockerParticipant
		// Them returns a list of all locker participants that aren't controlled by the account.
		Them() []*model.LockerParticipant

		// NewDataSetBuilder returns an instance of dataset.Builder that enables interactive construction
		// of a dataset. This builder assumes the dataset will be stored in this locker.
		NewDataSetBuilder(opts ...dataset.BuilderOption) (dataset.Builder, error)
		// Store is a convenience method that submits a dataset with no attachments to this locker.
		Store(meta any, expiryTime time.Time, opts ...dataset.BuilderOption) dataset.RecordFuture
		// Share shares the dataset from the record with the given id (we assume the account has access
		// to this record) through the locker.
		Share(id, vaultName string, expiryTime time.Time) dataset.RecordFuture
		// HeadID returns the ID of the dataset head for the given asset ID and head name (and linked
		// to the locker).
		HeadID(assetID string, headName string) string
		// SetAssetHead sets the record with the given ID as a head for the dataset with the given asset ID.
		SetAssetHead(assetID, headName, recordID string) dataset.RecordFuture

		// Seal closes the locker. NOT CURRENTLY SUPPORTED.
		Seal() error
	}

	// Services is an interface to MetaLocker services that are necessary for data wallet operations.
	// It is assumed all the operations with these services will be authenticated against the data wallet's
	// account.
	Services interface {
		DIDProvider() model.DIDProvider

		OffChainStorage() model.OffChainStorage
		Ledger() model.Ledger
		BlobManager() model.BlobManager
		NotificationService() (notification.Service, error)
	}

	// DataStore is a direct interface to dataset management operations for the enclosing data wallet.
	DataStore interface {
		// NewDataSetBuilder returns an instance of dataset.Builder that enables interactive construction
		// of a dataset.
		NewDataSetBuilder(lockerID string, opts ...dataset.BuilderOption) (dataset.Builder, error)
		// Load returns an interface to interact with the dataset behind the given record ID.
		Load(id string, opts ...dataset.LoadOption) (model.DataSet, error)
		// Revoke revokes for the lease for the dataset behind the given record ID.
		Revoke(id string) dataset.RecordFuture

		// AssetHead returns the dataset that is a head with the given ID.
		AssetHead(headID string, opts ...dataset.LoadOption) (model.DataSet, error)
		// SetAssetHead sets the record with the given ID as a head for the dataset with the given asset ID,
		// name and for the given locker.
		SetAssetHead(assetID string, locker *model.Locker, headName string, recordID string) dataset.RecordFuture

		// Share shares the dataset from the record with the given id (we assume the account has access
		// to this record) through the locker.
		Share(ds model.DataSet, locker Locker, vaultName string, expiryTime time.Time) dataset.RecordFuture

		// PurgeDataAssets purges all data assets (resources) for the given revoked lease.
		PurgeDataAssets(recordID string) error
	}

	DataSetStoreConstructor func(dataWallet DataWallet, services Services) (DataStore, error)

	// DataWallet is the main interface to the user's account and its data stored in MetaLocker.
	// It incorporates all the complexity of interacting with encrypted resources, the main
	// MetaLocker ledger, indexes, etc.
	DataWallet interface {
		io.Closer

		// ID returns the account ID.
		ID() string
		// Account returns the full account definition.
		Account() *account.Account
		// ChangePassphrase updates the passphrase for the account. If isHash is true,
		// the provided passphrase is a double SHA256 of the passphrase, not the cleartext
		// passphrase.
		ChangePassphrase(oldPassphrase, newPassphrase string, isHash bool) (DataWallet, error)
		// ChangeEmail changes the email of the account.
		ChangeEmail(email string) error
		// Recover enables account recovery, in the passphrase has been lost.
		Recover(cryptoKey *model.AESKey, newPassphrase string) (DataWallet, error)

		// EncryptionKey derives a deterministic AES key for the given tag. We assume that this derivation
		// can be repeated by the user at any time, producing the same key. Only a party in possession of
		// the user's secrets can produce a key.
		// This is useful for encrypting data stored outside the main MetaLocker platform. For instance,
		// external indexes can rely on this function.
		EncryptionKey(tag string, accessLevel model.AccessLevel) (*model.AESKey, error)

		// LockLevel returns the wallet's current lock level
		LockLevel() model.AccessLevel
		// Lock locks the data wallet and clears all sensitive information held in memory.
		Lock() error
		// Unlock unlocks the data wallet using a passphrase. Data wallet needs to be unlocked
		// to perform the majority of operations with the underlying account and its data.
		Unlock(passphrase string) error
		// UnlockAsManaged unlocks the data wallet at 'managed' level using the provided key.
		UnlockAsManaged(managedKey *model.AESKey) error
		// UnlockWithAccessKey unlocks the data wallet using an access key. Access level depends on the underlying
		// key's access level.
		UnlockWithAccessKey(apiKey, apiSecret string) error
		// UnlockAsChild unlock the data wallet for sub-account using its parent secret.
		UnlockAsChild(parentNode slip10.Node) error

		CreateSubAccount(accessLevel model.AccessLevel, name string, opts ...account.Option) (DataWallet, error)
		GetSubAccount(id string) (*account.Account, error)
		DeleteSubAccount(id string) error
		SubAccounts() ([]*account.Account, error)
		GetSubAccountWallet(id string) (DataWallet, error)

		CreateAccessKey(accessLevel model.AccessLevel, duration time.Duration) (*model.AccessKey, error)
		GetAccessKey(keyID string) (*model.AccessKey, error)
		RevokeAccessKey(keyID string) error
		AccessKeys() ([]*model.AccessKey, error)

		RestrictedWallet(identities []string) (DataWallet, error)

		NewIdentity(accessLevel model.AccessLevel, name string, options ...IdentityOption) (Identity, error)
		AddIdentity(idy *account.Identity) error
		GetIdentities() (map[string]Identity, error)
		GetIdentity(iid string) (Identity, error)
		GetDID(iid string) (*model.DID, error)
		GetRootIdentity() (Identity, error)

		AddLocker(l *model.Locker) (Locker, error)
		GetLockers() ([]*model.Locker, error)
		GetLocker(lockerID string) (Locker, error)
		GetRootLocker(level model.AccessLevel) (Locker, error)

		GetProperty(key string) (string, error)
		SetProperty(key string, value string, lvl model.AccessLevel) error
		GetProperties() (map[string]string, error)
		DeleteProperty(key string, lvl model.AccessLevel) error

		CreateRootIndex(indexStoreName string) (index.RootIndex, error)
		RootIndex() (index.RootIndex, error)

		CreateIndex(indexStoreName, indexType string, opts ...index.Option) (index.Index, error)
		Index(id string) (index.Index, error)

		IndexUpdater(indexes ...index.Index) (*IndexUpdater, error)

		DataStore() DataStore

		Services() Services

		// Backend function is used to access raw identity and locker storage operations
		// in downstream infrastructure such as Digital Twins.
		Backend() AccountBackend
	}

	RecoveryDetails struct {
		RecoveryPhrase          string
		SecondLevelRecoveryCode string
	}

	// Factory provides an interface for creating Data Wallets for the given API key ID and secret.
	// This interface can hide details how the wallet is constructed and whether it's local or remote.
	Factory interface {
		// GetWalletWithAccessKey returns an unlocked data wallet instance for the given access key and secret.
		GetWalletWithAccessKey(apiKey, apiSecret string) (DataWallet, error)
	}
)
