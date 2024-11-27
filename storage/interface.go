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
	"context"
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
	}

	DIDBackend interface {
		CreateDIDDocument(ctx context.Context, ddoc *model.DIDDocument) error
		GetDIDDocument(ctx context.Context, iid string) (*model.DIDDocument, error)
		ListDIDDocuments(ctx context.Context) ([]*model.DIDDocument, error)
	}

	AccountBackend interface {
		CreateAccount(ctx context.Context, acct *account.Account) error
		UpdateAccount(ctx context.Context, acct *account.Account) error
		GetAccount(ctx context.Context, id string) (*account.Account, error)
		DeleteAccount(ctx context.Context, id string) error
		ListAccounts(ctx context.Context, parentAccountID, stateFilter string) ([]*account.Account, error)

		HasAccountAccess(ctx context.Context, accountID, targetAccountID string) (bool, error)

		ListAccessKeys(ctx context.Context, accountID string) ([]*model.AccessKey, error)
		StoreAccessKey(ctx context.Context, accessKey *model.AccessKey) error
		GetAccessKey(ctx context.Context, keyID string) (*model.AccessKey, error)
		DeleteAccessKey(ctx context.Context, keyID string) error

		StoreIdentity(ctx context.Context, accountID string, idy *account.DataEnvelope) error
		GetIdentity(ctx context.Context, accountID string, hash string) (*account.DataEnvelope, error)
		ListIdentities(ctx context.Context, accountID string, lvl model.AccessLevel) ([]*account.DataEnvelope, error)

		StoreLocker(ctx context.Context, accountID string, l *account.DataEnvelope) error
		GetLocker(ctx context.Context, accountID string, hash string) (*account.DataEnvelope, error)
		ListLockers(ctx context.Context, accountID string, lvl model.AccessLevel) ([]*account.DataEnvelope, error)

		StoreProperty(ctx context.Context, accountID string, prop *account.DataEnvelope) error
		GetProperty(ctx context.Context, accountID string, hash string) (*account.DataEnvelope, error)
		ListProperties(ctx context.Context, accountID string, lvl model.AccessLevel) ([]*account.DataEnvelope, error)
		DeleteProperty(ctx context.Context, accountID string, hash string) error
	}

	RecoveryBackend interface {
		CreateRecoveryCode(ctx context.Context, c *account.RecoveryCode) error
		GetRecoveryCode(ctx context.Context, code string) (*account.RecoveryCode, error)
		DeleteRecoveryCode(ctx context.Context, code string) error
	}
)
