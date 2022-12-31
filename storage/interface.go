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

package storage

import (
	"errors"
	"io"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
)

var (
	ErrAccountExists        = errors.New("account exists")
	ErrAccountNotFound      = errors.New("account not found")
	ErrDIDExists            = errors.New("DID document exists")
	ErrDIDNotFound          = errors.New("DID document not found")
	ErrIdentityNotFound     = errors.New("identity not found")
	ErrLockerNotFound       = errors.New("locker not found")
	ErrAccessKeyNotFound    = errors.New("access key not found")
	ErrPropertyNotFound     = errors.New("property not found")
	ErrRecoveryCodeNotFound = errors.New("recovery code not found")
)

type (
	// IdentityBackend stores MetaLocker accounts and related entities
	// such as encrypted identities, lockers, access keys, properties,
	// DID documents and recovery codes. This is the main node-specific
	// storage layer.
	IdentityBackend interface {
		io.Closer
		DIDBackend
		AccountBackend
		RecoveryBackend

		IsNew() bool
	}

	DIDBackend interface {
		CreateDIDDocument(ddoc *model.DIDDocument) error
		GetDIDDocument(iid string) (*model.DIDDocument, error)
		ListDIDDocuments() ([]*model.DIDDocument, error)
	}

	AccountBackend interface {
		CreateAccount(acct *account.Account) error
		UpdateAccount(acct *account.Account) error
		GetAccount(id string) (*account.Account, error)
		DeleteAccount(id string) error
		ListAccounts(parentAccountID, stateFilter string) ([]*account.Account, error)

		HasAccountAccess(accountID, targetAccountID string) (bool, error)

		ListAccessKeys(accountID string) ([]*model.AccessKey, error)
		StoreAccessKey(accessKey *model.AccessKey) error
		GetAccessKey(keyID string) (*model.AccessKey, error)
		DeleteAccessKey(keyID string) error

		StoreIdentity(accountID string, idy *account.DataEnvelope) error
		GetIdentity(accountID string, hash string) (*account.DataEnvelope, error)
		ListIdentities(accountID string, lvl model.AccessLevel) ([]*account.DataEnvelope, error)

		StoreLocker(accountID string, l *account.DataEnvelope) error
		GetLocker(accountID string, hash string) (*account.DataEnvelope, error)
		ListLockers(accountID string, lvl model.AccessLevel) ([]*account.DataEnvelope, error)
		ListLockerHashes(accountID string, lvl model.AccessLevel) ([]string, error)

		StoreProperty(accountID string, prop *account.DataEnvelope) error
		GetProperty(accountID string, hash string) (*account.DataEnvelope, error)
		ListProperties(accountID string, lvl model.AccessLevel) ([]*account.DataEnvelope, error)
		DeleteProperty(accountID string, hash string) error
	}

	RecoveryBackend interface {
		CreateRecoveryCode(c *account.RecoveryCode) error
		GetRecoveryCode(code string) (*account.RecoveryCode, error)
		DeleteRecoveryCode(code string) error
	}
)
