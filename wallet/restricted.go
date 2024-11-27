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
	"context"
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

func (r *RestrictedNodeClient) CreateAccessKey(ctx context.Context, key *model.AccessKey) (*model.AccessKey, error) {
	return nil, ErrForbiddenOperation
}

func (r *RestrictedNodeClient) GetAccessKey(ctx context.Context, aid string) (*model.AccessKey, error) {
	return nil, ErrForbiddenOperation
}

func (r *RestrictedNodeClient) DeleteAccessKey(ctx context.Context, keyID string) error {
	return ErrForbiddenOperation
}

func (r *RestrictedNodeClient) ListAccessKeys(ctx context.Context) ([]*model.AccessKey, error) {
	return nil, ErrForbiddenOperation
}

func (r *RestrictedNodeClient) CreateSubAccount(ctx context.Context, acct *account.Account) (*account.Account, error) {
	return nil, ErrForbiddenOperation
}

func (r *RestrictedNodeClient) ListSubAccounts(ctx context.Context, id string) ([]*account.Account, error) {
	return nil, ErrForbiddenOperation
}

func (r *RestrictedNodeClient) CreateAccount(ctx context.Context, acct *account.Account, registrationCode string) error {
	return ErrForbiddenOperation
}

func (r *RestrictedNodeClient) GetOwnAccount(ctx context.Context) (*account.Account, error) {
	return r.nodeClient.GetOwnAccount(ctx)
}

func (r *RestrictedNodeClient) GetAccount(ctx context.Context, id string) (*account.Account, error) {
	return nil, ErrForbiddenOperation
	//acct, err := r.nodeClient.GetAccount(id)
	//if err != nil {
	//	return nil, err
	//}
	//return acct.RestrictedCopy(), nil
}

func (r *RestrictedNodeClient) UpdateAccount(ctx context.Context, acct *account.Account) error {
	return ErrForbiddenOperation
}

func (r *RestrictedNodeClient) PatchAccount(ctx context.Context, email, oldEncryptedPassword, newEncryptedPassword, name, givenName, familyName string) error {
	return ErrForbiddenOperation
}

func (r *RestrictedNodeClient) DeleteAccount(ctx context.Context, id string) error {
	return ErrForbiddenOperation
}

func (r *RestrictedNodeClient) StoreIdentity(ctx context.Context, idy *account.DataEnvelope) error {
	panic("operation not implemented")
}

func (r *RestrictedNodeClient) GetIdentity(ctx context.Context, hash string) (*account.DataEnvelope, error) {
	panic("operation not implemented")
}

func (r *RestrictedNodeClient) ListIdentities(ctx context.Context) ([]*account.DataEnvelope, error) {
	//idyList, err := r.nodeClient.ListIdentities()
	//if err != nil {
	//	return nil, err
	//}
	//var res []*DataEnvelope
	//for _, idy
	panic("operation not implemented")
}

func (r *RestrictedNodeClient) StoreLocker(ctx context.Context, l *account.DataEnvelope) error {
	panic("operation not implemented")
}

func (r *RestrictedNodeClient) GetLocker(ctx context.Context, hash string) (*account.DataEnvelope, error) {
	panic("operation not implemented")
}

func (r *RestrictedNodeClient) ListLockers(ctx context.Context) ([]*account.DataEnvelope, error) {
	panic("operation not implemented")
}

func (r *RestrictedNodeClient) StoreProperty(ctx context.Context, prop *account.DataEnvelope) error {
	panic("operation not implemented")
}

func (r *RestrictedNodeClient) GetProperty(ctx context.Context, hash string) (*account.DataEnvelope, error) {
	panic("operation not implemented")
}

func (r *RestrictedNodeClient) ListProperties(ctx context.Context) ([]*account.DataEnvelope, error) {
	panic("operation not implemented")
}

func (r *RestrictedNodeClient) DeleteProperty(ctx context.Context, hash string) error {
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

func (r *RestrictedNodeClient) NewInstance(ctx context.Context, email, passphrase string, isHash bool) (NodeClient, error) {
	return nil, ErrForbiddenOperation
}

func (r *RestrictedNodeClient) SubAccountInstance(subAccountID string) (NodeClient, error) {
	return nil, ErrForbiddenOperation
}

func (r *RestrictedNodeClient) NotificationService() (notification.Service, error) {
	return r.nodeClient.NotificationService()
}

func (r restrictedDIDProvider) CreateDIDDocument(ctx context.Context, ddoc *model.DIDDocument) error {
	return ErrForbiddenOperation
}

func (r restrictedDIDProvider) GetDIDDocument(ctx context.Context, iid string) (*model.DIDDocument, error) {
	return r.backend.GetDIDDocument(ctx, iid)
}
