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
	"io"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
)

type (
	// NodeClient is an interface to a MetaLocker node that data wallets require to perform data management operations.
	NodeClient interface {
		io.Closer
		AccountBackend
		Services

		NewInstance(email, passphrase string, isHash bool) (NodeClient, error)
		SubAccountInstance(subAccountID string) (NodeClient, error)
	}

	AccountBackend interface {
		CreateAccount(ctx context.Context, account *account.Account, registrationCode string) error
		GetOwnAccount(ctx context.Context) (*account.Account, error)

		GetAccount(ctx context.Context, id string) (*account.Account, error)
		UpdateAccount(ctx context.Context, account *account.Account) error
		PatchAccount(ctx context.Context, email, oldEncryptedPassword, newEncryptedPassword, name, givenName, familyName string) error
		DeleteAccount(ctx context.Context, id string) error

		CreateSubAccount(ctx context.Context, account *account.Account) (*account.Account, error)
		ListSubAccounts(ctx context.Context, id string) ([]*account.Account, error)

		CreateAccessKey(ctx context.Context, key *model.AccessKey) (*model.AccessKey, error)
		GetAccessKey(ctx context.Context, keyID string) (*model.AccessKey, error)
		DeleteAccessKey(ctx context.Context, keyID string) error
		ListAccessKeys(ctx context.Context) ([]*model.AccessKey, error)

		StoreIdentity(ctx context.Context, idy *account.DataEnvelope) error
		GetIdentity(ctx context.Context, hash string) (*account.DataEnvelope, error)
		ListIdentities(ctx context.Context) ([]*account.DataEnvelope, error)

		StoreLocker(ctx context.Context, l *account.DataEnvelope) error
		GetLocker(ctx context.Context, hash string) (*account.DataEnvelope, error)
		ListLockers(ctx context.Context) ([]*account.DataEnvelope, error)
		ListLockerHashes(ctx context.Context) ([]string, error)

		StoreProperty(ctx context.Context, prop *account.DataEnvelope) error
		GetProperty(ctx context.Context, hash string) (*account.DataEnvelope, error)
		ListProperties(ctx context.Context) ([]*account.DataEnvelope, error)
		DeleteProperty(ctx context.Context, hash string) error
	}
)

func SaveNewAccount(ctx context.Context, resp *account.GenerationResponse, nodeClient NodeClient, registrationCode string, hashFunction account.PasswordHashFunction) error {
	if resp.Account.EncryptedPassword != "" {
		if err := account.ReHashPassphrase(resp.Account, hashFunction); err != nil {
			return err
		}
	}

	if err := resp.Account.Validate(); err != nil {
		return err
	}

	if resp.Account.ParentAccount != "" {
		// sub-account
		subAcct, err := nodeClient.CreateSubAccount(ctx, resp.Account)
		if err != nil {
			return err
		}
		resp.Account = subAcct
	} else {
		if err := nodeClient.CreateAccount(ctx, resp.Account, registrationCode); err != nil {
			return err
		}
		// we assume the node client will be authenticated at this point. This may not be the case
		// for remote clients.
	}

	for _, e := range resp.EncryptedIdentities {
		if err := nodeClient.StoreIdentity(ctx, e); err != nil {
			return err
		}
	}

	for _, e := range resp.RootIdentities {
		dDoc, err := model.SimpleDIDDocument(e.DID, e.Created)
		if err != nil {
			return err
		}

		if err = nodeClient.DIDProvider().CreateDIDDocument(ctx, dDoc); err != nil {
			return err
		}
	}

	for _, e := range resp.EncryptedLockers {
		if err := nodeClient.StoreLocker(ctx, e); err != nil {
			return err
		}
	}

	return nil
}
