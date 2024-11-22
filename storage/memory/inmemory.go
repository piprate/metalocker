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

package memory

import (
	"context"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/sdk/cmdbase"
	"github.com/piprate/metalocker/storage"
	"github.com/piprate/metalocker/utils"
)

func init() {
	storage.Register("memory", CreateIdentityBackend)
}

type InMemoryBackend struct {
	accounts            map[string]*account.Account
	accountsByEmail     map[string]*account.Account
	accessKeys          map[string]*model.AccessKey
	accessKeysByAccount map[string]map[string]*model.AccessKey
	identities          map[string]map[string]*account.DataEnvelope
	lockers             map[string]map[string]*account.DataEnvelope
	properties          map[string]map[string]*account.DataEnvelope
	dids                map[string]*model.DIDDocument
	recoveryCodes       map[string]*account.RecoveryCode
}

var _ storage.IdentityBackend = (*InMemoryBackend)(nil)

func (be *InMemoryBackend) Close() error {
	return nil
}

func (be *InMemoryBackend) UpdateAccount(ctx context.Context, acct *account.Account) error {
	if a, found := be.accountsByEmail[acct.Email]; found && a.ID != acct.ID {
		return storage.ErrAccountExists
	}
	acct = copyAccount(acct)
	be.accounts[acct.ID] = acct
	be.accountsByEmail[acct.Email] = acct
	return nil
}

func (be *InMemoryBackend) GetAccount(ctx context.Context, id string) (*account.Account, error) {
	acct, found := be.accounts[id]
	if !found {
		acct, found = be.accountsByEmail[id]
		if !found {
			return nil, storage.ErrAccountNotFound
		}
	}
	return copyAccount(acct), nil
}

func (be *InMemoryBackend) HasAccountAccess(ctx context.Context, accountID, targetAccountID string) (bool, error) {
	for {
		acct, found := be.accounts[targetAccountID]
		if !found {
			return false, storage.ErrAccountNotFound
		}
		if acct.ID == accountID {
			return true, nil
		} else if acct.ParentAccount == "" {
			return false, nil
		} else {
			targetAccountID = acct.ParentAccount
		}
	}
}

func (be *InMemoryBackend) DeleteAccount(ctx context.Context, id string) error {
	delete(be.accounts, id)
	for email, acct := range be.accountsByEmail {
		if acct.ID == id {
			delete(be.accountsByEmail, email)
		}
	}
	delete(be.accessKeysByAccount, id)
	delete(be.identities, id)
	delete(be.lockers, id)
	delete(be.properties, id)
	delete(be.properties, id)
	for keyID, key := range be.accessKeys {
		if key.AccountID == id {
			delete(be.accessKeys, keyID)
		}
	}
	for code, codeVal := range be.recoveryCodes {
		if codeVal.UserID == id {
			delete(be.recoveryCodes, code)
		}
	}

	return nil
}

func (be *InMemoryBackend) ListAccessKeys(ctx context.Context, accountID string) ([]*model.AccessKey, error) {
	acctMap, found := be.accessKeysByAccount[accountID]
	if !found {
		return []*model.AccessKey{}, nil
	}
	res := make([]*model.AccessKey, 0, len(acctMap))
	for _, key := range acctMap {
		res = append(res, key)
	}
	return res, nil
}

func (be *InMemoryBackend) StoreAccessKey(ctx context.Context, accessKey *model.AccessKey) error {
	acctMap, found := be.accessKeysByAccount[accessKey.AccountID]
	if !found {
		acctMap = make(map[string]*model.AccessKey)
		be.accessKeysByAccount[accessKey.AccountID] = acctMap
	}
	acctMap[accessKey.ID] = accessKey
	be.accessKeys[accessKey.ID] = accessKey

	return nil
}

func (be *InMemoryBackend) GetAccessKey(ctx context.Context, keyID string) (*model.AccessKey, error) {
	key, found := be.accessKeys[keyID]
	if found {
		return key, nil
	}
	return nil, storage.ErrAccessKeyNotFound
}

func (be *InMemoryBackend) DeleteAccessKey(ctx context.Context, keyID string) error {
	key, found := be.accessKeys[keyID]
	if found {
		delete(be.accessKeysByAccount, key.AccountID)
		delete(be.accessKeys, keyID)
		return nil
	} else {
		return storage.ErrAccessKeyNotFound
	}
}

func (be *InMemoryBackend) StoreIdentity(ctx context.Context, accountID string, idy *account.DataEnvelope) error {
	acctMap, found := be.identities[accountID]
	if !found {
		acctMap = make(map[string]*account.DataEnvelope)
		be.identities[accountID] = acctMap
	}
	acctMap[idy.Hash] = idy
	return nil
}

func (be *InMemoryBackend) GetIdentity(ctx context.Context, accountID string, hash string) (*account.DataEnvelope, error) {
	acctMap, found := be.identities[accountID]
	if found {
		idy, found := acctMap[hash]
		if found {
			return idy, nil
		}
	}
	return nil, storage.ErrIdentityNotFound
}

func (be *InMemoryBackend) ListIdentities(ctx context.Context, accountID string, lvl model.AccessLevel) ([]*account.DataEnvelope, error) {
	acctMap, found := be.identities[accountID]
	if !found {
		return []*account.DataEnvelope{}, nil
	}
	res := make([]*account.DataEnvelope, 0, len(acctMap))
	for _, idy := range acctMap {
		if lvl == 0 || idy.AccessLevel == lvl {
			res = append(res, idy)
		}
	}
	return res, nil
}

func (be *InMemoryBackend) StoreLocker(ctx context.Context, accountID string, l *account.DataEnvelope) error {
	acctMap, found := be.lockers[accountID]
	if !found {
		acctMap = make(map[string]*account.DataEnvelope)
		be.lockers[accountID] = acctMap
	}
	acctMap[l.Hash] = l
	return nil
}

func (be *InMemoryBackend) GetLocker(ctx context.Context, accountID string, hash string) (*account.DataEnvelope, error) {
	acctMap, found := be.lockers[accountID]
	if found {
		locker, found := acctMap[hash]
		if found {
			return locker, nil
		}
	}
	return nil, storage.ErrLockerNotFound
}

func (be *InMemoryBackend) ListLockers(ctx context.Context, accountID string, lvl model.AccessLevel) ([]*account.DataEnvelope, error) {
	acctMap, found := be.lockers[accountID]
	if !found {
		return []*account.DataEnvelope{}, nil
	}
	res := make([]*account.DataEnvelope, 0, len(acctMap))
	for _, locker := range acctMap {
		if lvl == 0 || locker.AccessLevel == lvl {
			res = append(res, locker)
		}
	}
	return res, nil
}

func (be *InMemoryBackend) ListLockerHashes(ctx context.Context, accountID string, lvl model.AccessLevel) ([]string, error) {
	acctMap, found := be.lockers[accountID]
	if !found {
		return []string{}, nil
	}
	res := make([]string, 0, len(acctMap))
	for _, locker := range acctMap {
		if lvl == 0 || locker.AccessLevel == lvl {
			res = append(res, locker.Hash)
		}
	}
	return res, nil
}

func (be *InMemoryBackend) StoreProperty(ctx context.Context, accountID string, prop *account.DataEnvelope) error {
	acctMap, found := be.properties[accountID]
	if !found {
		acctMap = make(map[string]*account.DataEnvelope)
		be.properties[accountID] = acctMap
	}
	acctMap[prop.Hash] = prop
	return nil
}

func (be *InMemoryBackend) GetProperty(ctx context.Context, accountID string, hash string) (*account.DataEnvelope, error) {
	acctMap, found := be.properties[accountID]
	if found {
		prop, found := acctMap[hash]
		if found {
			return prop, nil
		}
	}
	return nil, storage.ErrPropertyNotFound
}

func (be *InMemoryBackend) ListProperties(ctx context.Context, accountID string, lvl model.AccessLevel) ([]*account.DataEnvelope, error) {
	acctMap, found := be.properties[accountID]
	if !found {
		return []*account.DataEnvelope{}, nil
	}
	res := make([]*account.DataEnvelope, 0, len(acctMap))
	for _, prop := range acctMap {
		if lvl == 0 || prop.AccessLevel == lvl {
			res = append(res, prop)
		}
	}
	return res, nil
}

func (be *InMemoryBackend) DeleteProperty(ctx context.Context, accountID string, hash string) error {
	acctMap, found := be.properties[accountID]
	if !found {
		return storage.ErrPropertyNotFound
	}

	_, found = acctMap[hash]
	if !found {
		return storage.ErrPropertyNotFound
	}

	delete(acctMap, hash)

	return nil
}

func (be *InMemoryBackend) CreateAccount(ctx context.Context, acct *account.Account) error {
	acct = copyAccount(acct)
	if _, found := be.accounts[acct.ID]; found {
		return storage.ErrAccountExists
	}

	if acct.Email != "" {
		if _, found := be.accountsByEmail[acct.Email]; found {
			return storage.ErrAccountExists
		}
		be.accountsByEmail[acct.Email] = acct
	}

	be.accounts[acct.ID] = acct

	return nil
}

func (be *InMemoryBackend) ListAccounts(ctx context.Context, parentAccountID, stateFilter string) ([]*account.Account, error) {
	res := make([]*account.Account, 0, len(be.accounts))
	for _, acct := range be.accounts {
		if parentAccountID != "" && acct.ParentAccount != parentAccountID {
			continue
		}
		res = append(res, copyAccount(acct))
	}
	return res, nil
}

func (be *InMemoryBackend) ListDIDDocuments(ctx context.Context) ([]*model.DIDDocument, error) {
	res := make([]*model.DIDDocument, len(be.dids))
	i := 0
	for _, did := range be.dids {
		res[i] = did
		i++
	}
	return res, nil
}

func (be *InMemoryBackend) CreateRecoveryCode(ctx context.Context, c *account.RecoveryCode) error {
	be.recoveryCodes[c.Code] = c
	return nil
}

func (be *InMemoryBackend) GetRecoveryCode(ctx context.Context, code string) (*account.RecoveryCode, error) {
	c, found := be.recoveryCodes[code]
	if !found {
		return nil, storage.ErrRecoveryCodeNotFound
	} else {
		return c, nil
	}
}

func (be *InMemoryBackend) DeleteRecoveryCode(ctx context.Context, code string) error {
	if _, found := be.recoveryCodes[code]; !found {
		return storage.ErrRecoveryCodeNotFound
	} else {
		delete(be.recoveryCodes, code)
		return nil
	}
}

func (be *InMemoryBackend) CreateDIDDocument(ctx context.Context, ddoc *model.DIDDocument) error {
	be.dids[ddoc.ID] = ddoc
	return nil
}

func (be *InMemoryBackend) GetDIDDocument(ctx context.Context, iid string) (*model.DIDDocument, error) {
	res, found := be.dids[iid]
	if !found {
		return nil, storage.ErrDIDNotFound
	}
	return res, nil
}

func (be *InMemoryBackend) IsNew() bool {
	return true
}

func copyAccount(acct *account.Account) *account.Account {
	var cpy account.Account
	if err := utils.MarshalToType(acct, &cpy, true); err != nil {
		panic(err)
	}
	return &cpy
}

func CreateIdentityBackend(params storage.Parameters, resolver cmdbase.ParameterResolver) (storage.IdentityBackend, error) {
	return &InMemoryBackend{
		accounts:            make(map[string]*account.Account),
		accountsByEmail:     make(map[string]*account.Account),
		accessKeys:          make(map[string]*model.AccessKey),
		accessKeysByAccount: make(map[string]map[string]*model.AccessKey),
		identities:          make(map[string]map[string]*account.DataEnvelope),
		lockers:             make(map[string]map[string]*account.DataEnvelope),
		properties:          make(map[string]map[string]*account.DataEnvelope),
		dids:                make(map[string]*model.DIDDocument),
		recoveryCodes:       make(map[string]*account.RecoveryCode),
	}, nil
}
