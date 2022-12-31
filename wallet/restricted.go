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

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/services/notification"
)

var (
	ErrForbiddenOperation = errors.New("forbidden operation with restricted wallet")
)

type (
	RestrictedNodeClient struct {
		identities []string
		nodeClient NodeClient
	}

	restrictedDIDProvider struct {
		backend model.DIDProvider
	}
)

var _ NodeClient = (*RestrictedNodeClient)(nil)
var _ model.DIDProvider = (*restrictedDIDProvider)(nil)

// NewRestrictedNodeClient is currently not in use, since we moved to encrypted identities/lockers
func NewRestrictedNodeClient(identities []string, nodeClient NodeClient) *RestrictedNodeClient {
	return &RestrictedNodeClient{
		identities: identities,
		nodeClient: nodeClient,
	}
}

func (r *RestrictedNodeClient) Close() error {
	return r.nodeClient.Close()
}

func (r *RestrictedNodeClient) CreateAccessKey(key *model.AccessKey) (*model.AccessKey, error) {
	return nil, ErrForbiddenOperation
}

func (r *RestrictedNodeClient) GetAccessKey(aid string) (*model.AccessKey, error) {
	return nil, ErrForbiddenOperation
}

func (r *RestrictedNodeClient) DeleteAccessKey(keyID string) error {
	return ErrForbiddenOperation
}

func (r *RestrictedNodeClient) ListAccessKeys() ([]*model.AccessKey, error) {
	return nil, ErrForbiddenOperation
}

func (r *RestrictedNodeClient) CreateSubAccount(acct *account.Account) (*account.Account, error) {
	return nil, ErrForbiddenOperation
}

func (r *RestrictedNodeClient) ListSubAccounts(id string) ([]*account.Account, error) {
	return nil, ErrForbiddenOperation
}

func (r *RestrictedNodeClient) CreateAccount(acct *account.Account, registrationCode string) error {
	return ErrForbiddenOperation
}

func (r *RestrictedNodeClient) GetOwnAccount() (*account.Account, error) {
	return r.nodeClient.GetOwnAccount()
}

func (r *RestrictedNodeClient) GetAccount(id string) (*account.Account, error) {
	return nil, ErrForbiddenOperation
	//acct, err := r.nodeClient.GetAccount(id)
	//if err != nil {
	//	return nil, err
	//}
	//return acct.RestrictedCopy(), nil
}

func (r *RestrictedNodeClient) UpdateAccount(acct *account.Account) error {
	return ErrForbiddenOperation
}

func (r *RestrictedNodeClient) PatchAccount(email, oldEncryptedPassword, newEncryptedPassword, name, givenName, familyName string) error {
	return ErrForbiddenOperation
}

func (r *RestrictedNodeClient) DeleteAccount(id string) error {
	return ErrForbiddenOperation
}

func (r *RestrictedNodeClient) StoreIdentity(idy *account.DataEnvelope) error {
	panic("operation not implemented")
}

func (r *RestrictedNodeClient) GetIdentity(hash string) (*account.DataEnvelope, error) {
	panic("operation not implemented")
}

func (r *RestrictedNodeClient) ListIdentities() ([]*account.DataEnvelope, error) {
	//idyList, err := r.nodeClient.ListIdentities()
	//if err != nil {
	//	return nil, err
	//}
	//var res []*DataEnvelope
	//for _, idy
	panic("operation not implemented")
}

func (r *RestrictedNodeClient) StoreLocker(l *account.DataEnvelope) error {
	panic("operation not implemented")
}

func (r *RestrictedNodeClient) GetLocker(hash string) (*account.DataEnvelope, error) {
	panic("operation not implemented")
}

func (r *RestrictedNodeClient) ListLockers() ([]*account.DataEnvelope, error) {
	panic("operation not implemented")
}

func (r *RestrictedNodeClient) ListLockerHashes() ([]string, error) {
	panic("operation not implemented")
}

func (r *RestrictedNodeClient) StoreProperty(prop *account.DataEnvelope) error {
	panic("operation not implemented")
}

func (r *RestrictedNodeClient) GetProperty(hash string) (*account.DataEnvelope, error) {
	panic("operation not implemented")
}

func (r *RestrictedNodeClient) ListProperties() ([]*account.DataEnvelope, error) {
	panic("operation not implemented")
}

func (r *RestrictedNodeClient) DeleteProperty(hash string) error {
	panic("operation not implemented")
}

func (r *RestrictedNodeClient) DIDProvider() model.DIDProvider {
	return &restrictedDIDProvider{
		backend: r.nodeClient.DIDProvider(),
	}
}

func (r *RestrictedNodeClient) OffChainStorage() model.OffChainStorage {
	return r.nodeClient.OffChainStorage()
}

func (r *RestrictedNodeClient) Ledger() model.Ledger {
	return r.nodeClient.Ledger()
}

func (r *RestrictedNodeClient) BlobManager() model.BlobManager {
	return r.nodeClient.BlobManager()
}

func (r *RestrictedNodeClient) NewInstance(email, passphrase string, isHash bool) (NodeClient, error) {
	return nil, ErrForbiddenOperation
}

func (r *RestrictedNodeClient) SubAccountInstance(subAccountID string) (NodeClient, error) {
	return nil, ErrForbiddenOperation
}

func (r *RestrictedNodeClient) NotificationService() (notification.Service, error) {
	return r.nodeClient.NotificationService()
}

func (r restrictedDIDProvider) CreateDIDDocument(ddoc *model.DIDDocument) error {
	return ErrForbiddenOperation
}

func (r restrictedDIDProvider) GetDIDDocument(iid string) (*model.DIDDocument, error) {
	return r.backend.GetDIDDocument(iid)
}
