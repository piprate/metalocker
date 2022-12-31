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

	"github.com/piprate/metalocker/index"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/services/notification"
	"github.com/piprate/metalocker/storage"
)

type (
	LocalNodeClient struct {
		accountID           string
		identityBackend     storage.IdentityBackend
		ledger              model.Ledger
		offChainStorage     model.OffChainStorage
		blobManager         model.BlobManager
		notificationService notification.Service
	}
)

var _ NodeClient = (*LocalNodeClient)(nil)

func NewLocalNodeClient(accountID string, identityBackend storage.IdentityBackend, ledger model.Ledger, offChainStorage model.OffChainStorage, blobManager model.BlobManager, notificationService notification.Service) *LocalNodeClient {
	return &LocalNodeClient{
		accountID:           accountID,
		identityBackend:     identityBackend,
		ledger:              ledger,
		offChainStorage:     offChainStorage,
		blobManager:         blobManager,
		notificationService: notificationService,
	}
}

func (lnc *LocalNodeClient) CreateAccount(acct *account.Account, registrationCode string) error {
	err := lnc.identityBackend.CreateAccount(acct)
	if err != nil {
		return err
	}
	lnc.accountID = acct.ID
	return nil
}

func (lnc *LocalNodeClient) CreateSubAccount(acct *account.Account) (*account.Account, error) {
	err := lnc.identityBackend.CreateAccount(acct)
	if err != nil {
		return nil, err
	}
	return acct, nil
}

func (lnc *LocalNodeClient) GetOwnAccount() (*account.Account, error) {
	return lnc.identityBackend.GetAccount(lnc.accountID)
}

func (lnc *LocalNodeClient) GetAccount(id string) (*account.Account, error) {
	return lnc.identityBackend.GetAccount(id)
}

func (lnc *LocalNodeClient) ListSubAccounts(id string) ([]*account.Account, error) {
	return lnc.identityBackend.ListAccounts(id, "")
}

func (lnc *LocalNodeClient) DeleteAccount(id string) error {
	acct, err := lnc.identityBackend.GetAccount(id)
	if err != nil {
		return err
	}
	if acct.ParentAccount != lnc.accountID {
		return errors.New("sub-account not found")
	}
	return lnc.identityBackend.DeleteAccount(id)
}

func (lnc *LocalNodeClient) CreateAccessKey(key *model.AccessKey) (*model.AccessKey, error) {
	err := lnc.identityBackend.StoreAccessKey(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (lnc *LocalNodeClient) DeleteAccessKey(keyID string) error {
	return lnc.identityBackend.DeleteAccessKey(keyID)
}

func (lnc *LocalNodeClient) GetAccessKey(keyID string) (*model.AccessKey, error) {
	return lnc.identityBackend.GetAccessKey(keyID)
}

func (lnc *LocalNodeClient) ListAccessKeys() ([]*model.AccessKey, error) {
	return lnc.identityBackend.ListAccessKeys(lnc.accountID)
}

func (lnc *LocalNodeClient) CreateDIDDocument(ddoc *model.DIDDocument) error {
	return lnc.identityBackend.CreateDIDDocument(ddoc)
}

func (lnc *LocalNodeClient) GetDIDDocument(iid string) (*model.DIDDocument, error) {
	return lnc.identityBackend.GetDIDDocument(iid)
}

func (lnc *LocalNodeClient) UpdateAccount(acct *account.Account) error {
	return lnc.identityBackend.UpdateAccount(acct)
}

func (lnc *LocalNodeClient) PatchAccount(email, oldEncryptedPassword, newEncryptedPassword, name, givenName, familyName string) error {
	acct, err := lnc.identityBackend.GetAccount(lnc.accountID)
	if err != nil {
		return err
	}
	if email != "" {
		acct.Email = email
	}
	if newEncryptedPassword != "" {
		acct.EncryptedPassword = newEncryptedPassword
	}
	if name != "" {
		acct.Name = name
	}
	if givenName != "" {
		acct.GivenName = givenName
	}
	if familyName != "" {
		acct.FamilyName = familyName
	}
	return lnc.identityBackend.UpdateAccount(acct)
}

func (lnc *LocalNodeClient) StoreIdentity(idy *account.DataEnvelope) error {
	return lnc.identityBackend.StoreIdentity(lnc.accountID, idy)
}

func (lnc *LocalNodeClient) GetIdentity(hash string) (*account.DataEnvelope, error) {
	return lnc.identityBackend.GetIdentity(lnc.accountID, hash)
}

func (lnc *LocalNodeClient) ListIdentities() ([]*account.DataEnvelope, error) {
	return lnc.identityBackend.ListIdentities(lnc.accountID, 0)
}

func (lnc *LocalNodeClient) StoreLocker(l *account.DataEnvelope) error {
	return lnc.identityBackend.StoreLocker(lnc.accountID, l)
}

func (lnc *LocalNodeClient) GetLocker(hash string) (*account.DataEnvelope, error) {
	return lnc.identityBackend.GetLocker(lnc.accountID, hash)
}

func (lnc *LocalNodeClient) ListLockers() ([]*account.DataEnvelope, error) {
	return lnc.identityBackend.ListLockers(lnc.accountID, 0)
}

func (lnc *LocalNodeClient) ListLockerHashes() ([]string, error) {
	return lnc.identityBackend.ListLockerHashes(lnc.accountID, 0)
}

func (lnc *LocalNodeClient) StoreProperty(prop *account.DataEnvelope) error {
	return lnc.identityBackend.StoreProperty(lnc.accountID, prop)
}

func (lnc *LocalNodeClient) GetProperty(hash string) (*account.DataEnvelope, error) {
	return lnc.identityBackend.GetProperty(lnc.accountID, hash)
}

func (lnc *LocalNodeClient) ListProperties() ([]*account.DataEnvelope, error) {
	return lnc.identityBackend.ListProperties(lnc.accountID, 0)
}

func (lnc *LocalNodeClient) DeleteProperty(hash string) error {
	return lnc.identityBackend.DeleteProperty(lnc.accountID, hash)
}

func (lnc *LocalNodeClient) DIDProvider() model.DIDProvider {
	return lnc.identityBackend
}

func (lnc *LocalNodeClient) OffChainStorage() model.OffChainStorage {
	return lnc.offChainStorage
}

func (lnc *LocalNodeClient) Ledger() model.Ledger {
	return lnc.ledger
}

func (lnc *LocalNodeClient) BlobManager() model.BlobManager {
	return lnc.blobManager
}

func (lnc *LocalNodeClient) NotificationService() (notification.Service, error) {
	return lnc.notificationService, nil
}

func (lnc *LocalNodeClient) NewInstance(email, passphrase string, isHash bool) (NodeClient, error) {
	return lnc, nil
}

func (lnc *LocalNodeClient) SubAccountInstance(subAccountID string) (NodeClient, error) {
	return NewLocalNodeClient(subAccountID, lnc.identityBackend, lnc.ledger, lnc.offChainStorage,
		lnc.blobManager, lnc.notificationService), nil
}

func (lnc *LocalNodeClient) Close() error {
	return lnc.identityBackend.Close()
}

type LocalFactory struct {
	ledger              model.Ledger
	offChainStorage     model.OffChainStorage
	blobManager         model.BlobManager
	identityBackend     storage.IdentityBackend
	notificationService notification.Service
	indexClient         index.Client
	hashFunction        account.PasswordHashFunction
}

var _ Factory = (*LocalFactory)(nil)

func NewLocalFactory(ledger model.Ledger, offChainStorage model.OffChainStorage, blobManager model.BlobManager,
	identityBackend storage.IdentityBackend, notificationService notification.Service, indexClient index.Client, hashFunction account.PasswordHashFunction) (*LocalFactory, error) {

	lf := &LocalFactory{
		ledger:              ledger,
		offChainStorage:     offChainStorage,
		blobManager:         blobManager,
		identityBackend:     identityBackend,
		notificationService: notificationService,
		indexClient:         indexClient,
		hashFunction:        hashFunction,
	}

	return lf, nil
}

func (lf *LocalFactory) RegisterAccount(acctTemplate *account.Account, opts ...account.Option) (DataWallet, *RecoveryDetails, error) {
	tb, err := lf.ledger.GetTopBlock()
	if err != nil {
		return nil, nil, err
	}
	opts = append([]account.Option{account.WithFirstBlock(tb.Number)}, opts...)

	resp, err := account.GenerateAccount(acctTemplate, opts...)
	if err != nil {
		return nil, nil, err
	}

	recDetails := &RecoveryDetails{
		RecoveryPhrase:          resp.RecoveryPhrase,
		SecondLevelRecoveryCode: resp.SecondLevelRecoveryCode,
	}

	nodeClient := NewLocalNodeClient(resp.Account.ID, lf.identityBackend, lf.ledger, lf.offChainStorage,
		lf.blobManager, lf.notificationService)

	err = SaveNewAccount(resp, nodeClient, "", lf.hashFunction)
	if err != nil {
		return nil, nil, err
	}

	dw, err := NewLocalDataWallet(resp.Account, nodeClient, nil, lf.indexClient)
	if err != nil {
		return nil, recDetails, err
	}

	return dw, recDetails, nil
}

func (lf *LocalFactory) SaveAccount(acct *account.Account) (DataWallet, error) {
	if acct.EncryptedPassword != "" {
		if err := account.ReHashPassphrase(acct, lf.hashFunction); err != nil {
			return nil, err
		}
	}

	if err := acct.Validate(); err != nil {
		return nil, err
	}

	nodeClient := NewLocalNodeClient(acct.ID, lf.identityBackend, lf.ledger, lf.offChainStorage,
		lf.blobManager, lf.notificationService)

	if acct.ParentAccount != "" {
		// sub-account
		subAcct, err := nodeClient.CreateSubAccount(acct)
		if err != nil {
			return nil, err
		}
		acct = subAcct
	} else {
		if err := nodeClient.CreateAccount(acct, ""); err != nil {
			return nil, err
		}
	}

	return NewLocalDataWallet(acct, nodeClient, nil, lf.indexClient)
}

func (lf *LocalFactory) CreateDataWallet(acct *account.Account) (DataWallet, error) {
	return NewLocalDataWallet(
		acct,
		NewLocalNodeClient(
			acct.ID, lf.identityBackend, lf.ledger, lf.offChainStorage, lf.blobManager, lf.notificationService,
		),
		nil,
		lf.indexClient)
}

func (lf *LocalFactory) GetWalletWithAccessKey(apiKey, apiSecret string) (DataWallet, error) {

	ak, err := lf.identityBackend.GetAccessKey(apiKey)
	if err != nil {
		return nil, err
	}

	localBackend := NewLocalNodeClient(ak.AccountID, lf.identityBackend, lf.ledger, lf.offChainStorage,
		lf.blobManager, lf.notificationService)

	acct, err := localBackend.GetAccount(ak.AccountID)
	if err != nil {
		return nil, err
	}

	dw, err := NewLocalDataWallet(acct, localBackend, nil, lf.indexClient)
	if err != nil {
		return nil, err
	}

	if err = dw.UnlockWithAccessKey(apiKey, apiSecret); err != nil {
		_ = dw.Close()
		return nil, err
	}

	return dw, nil
}
