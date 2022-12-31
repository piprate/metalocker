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
		CreateAccount(account *account.Account, registrationCode string) error
		GetOwnAccount() (*account.Account, error)

		GetAccount(id string) (*account.Account, error)
		UpdateAccount(account *account.Account) error
		PatchAccount(email, oldEncryptedPassword, newEncryptedPassword, name, givenName, familyName string) error
		DeleteAccount(id string) error

		CreateSubAccount(account *account.Account) (*account.Account, error)
		ListSubAccounts(id string) ([]*account.Account, error)

		CreateAccessKey(key *model.AccessKey) (*model.AccessKey, error)
		GetAccessKey(keyID string) (*model.AccessKey, error)
		DeleteAccessKey(keyID string) error
		ListAccessKeys() ([]*model.AccessKey, error)

		StoreIdentity(idy *account.DataEnvelope) error
		GetIdentity(hash string) (*account.DataEnvelope, error)
		ListIdentities() ([]*account.DataEnvelope, error)

		StoreLocker(l *account.DataEnvelope) error
		GetLocker(hash string) (*account.DataEnvelope, error)
		ListLockers() ([]*account.DataEnvelope, error)
		ListLockerHashes() ([]string, error)

		StoreProperty(prop *account.DataEnvelope) error
		GetProperty(hash string) (*account.DataEnvelope, error)
		ListProperties() ([]*account.DataEnvelope, error)
		DeleteProperty(hash string) error
	}
)

func SaveNewAccount(resp *account.GenerationResponse, nodeClient NodeClient, registrationCode string, hashFunction account.PasswordHashFunction) error {
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
		subAcct, err := nodeClient.CreateSubAccount(resp.Account)
		if err != nil {
			return err
		}
		resp.Account = subAcct
	} else {
		if err := nodeClient.CreateAccount(resp.Account, registrationCode); err != nil {
			return err
		}
		// we assume the node client will be authenticated at this point. This may not be the case
		// for remote clients.
	}

	for _, e := range resp.EncryptedIdentities {
		if err := nodeClient.StoreIdentity(e); err != nil {
			return err
		}
	}

	for _, e := range resp.RootIdentities {
		dDoc, err := model.SimpleDIDDocument(e.DID, e.Created)
		if err != nil {
			return err
		}

		if err = nodeClient.DIDProvider().CreateDIDDocument(dDoc); err != nil {
			return err
		}
	}

	for _, e := range resp.EncryptedLockers {
		if err := nodeClient.StoreLocker(e); err != nil {
			return err
		}
	}

	return nil
}
