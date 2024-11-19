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

package remote

import (
	"context"
	"errors"
	"time"

	"github.com/muesli/cache2go"
	"github.com/piprate/metalocker/index"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/remote/caller"
	"github.com/piprate/metalocker/wallet"
	"github.com/rs/zerolog/log"
)

type IndexClientSourceFn func(userID string, httpCaller *caller.MetaLockerHTTPCaller) (index.Client, error)

func WithIndexClient(indexClient index.Client) IndexClientSourceFn {
	return func(userID string, httpCaller *caller.MetaLockerHTTPCaller) (index.Client, error) {
		gb, err := httpCaller.GetGenesisBlock()
		if err != nil {
			return nil, err
		}

		if err = indexClient.Bind(gb.Hash); err != nil {
			return nil, err
		}

		return indexClient, nil
	}
}

type Factory struct {
	url                  string
	httpCaller           *caller.MetaLockerHTTPCaller
	indexClient          index.Client
	indexClientSourceFn  IndexClientSourceFn
	accountCache         *cache2go.CacheTable
	accountCacheLifeSpan time.Duration
}

var _ wallet.Factory = (*Factory)(nil)

const factoryUserAgent = "Wallet Factory"

// NewWalletFactory creates a new data wallet factory. This factory can instantiate data wallets that
// interact with MetaLocker node using API.
func NewWalletFactory(url string, indexClientSourceFn IndexClientSourceFn, accountCacheLifeSpan time.Duration) (*Factory, error) {
	if url == "" {
		return nil, errors.New("url is empty. Can't start wallet factory")
	}

	httpCaller, err := caller.NewMetaLockerHTTPCaller(url, factoryUserAgent)
	if err != nil {
		return nil, err
	}

	// initialise context forwarding
	httpCaller.InitContextForwarding()

	rf := &Factory{
		url:                 url,
		httpCaller:          httpCaller,
		indexClientSourceFn: indexClientSourceFn,
	}

	if accountCacheLifeSpan != 0 {
		rf.accountCacheLifeSpan = accountCacheLifeSpan
		rf.accountCache = cache2go.Cache("accountCache")
	}

	return rf, nil
}

func (rf *Factory) GetTopBlock() (int64, error) {
	controls, err := rf.httpCaller.GetServerControls()
	if err != nil {
		return 0, err
	}

	return controls.TopBlock, nil
}

func (rf *Factory) RegisterAccount(ctx context.Context, acctTemplate *account.Account, passwd string, opts ...account.Option) (wallet.DataWallet, *wallet.RecoveryDetails, error) {

	httpCaller, err := caller.NewMetaLockerHTTPCaller(rf.url, factoryUserAgent)
	if err != nil {
		return nil, nil, err
	}

	controls, err := httpCaller.GetServerControls()
	if err != nil {
		return nil, nil, err
	}

	opts = append([]account.Option{account.WithFirstBlock(controls.TopBlock)}, opts...)
	opts = append(opts, account.WithPassphraseAuth(passwd))

	resp, err := account.GenerateAccount(acctTemplate, opts...)
	if err != nil {
		return nil, nil, err
	}

	recDetails := &wallet.RecoveryDetails{
		RecoveryPhrase:          resp.RecoveryPhrase,
		SecondLevelRecoveryCode: resp.SecondLevelRecoveryCode,
	}

	if err = httpCaller.CreateAccount(ctx, resp.Account, resp.RegistrationCode); err != nil {
		log.Err(err).Msg("Registration failed")
		return nil, nil, err
	}

	if err = httpCaller.LoginWithCredentials(resp.Account.Email, passwd); err != nil {
		log.Err(err).Msg("Login failed")
		return nil, recDetails, err
	}

	acct, err := httpCaller.GetOwnAccount(ctx)
	if err != nil {
		return nil, recDetails, err
	}

	// save root identities and lockers

	for _, e := range resp.EncryptedIdentities {
		if err := httpCaller.StoreIdentity(ctx, e); err != nil {
			return nil, recDetails, err
		}
	}

	for _, e := range resp.RootIdentities {
		dDoc, err := model.SimpleDIDDocument(e.DID, e.Created)
		if err != nil {
			return nil, recDetails, err
		}

		if err = httpCaller.CreateDIDDocument(ctx, dDoc); err != nil {
			return nil, recDetails, err
		}
	}

	for _, e := range resp.EncryptedLockers {
		if err := httpCaller.StoreLocker(ctx, e); err != nil {
			return nil, recDetails, err
		}
	}

	if rf.indexClient == nil {
		rf.indexClient, err = rf.indexClientSourceFn(acct.ID, httpCaller)
		if err != nil {
			return nil, recDetails, err
		}
	}

	dw, err := wallet.NewLocalDataWallet(acct, httpCaller, nil, rf.indexClient)
	if err != nil {
		return nil, recDetails, err
	}

	return dw, recDetails, nil
}

func (rf *Factory) GetWalletWithAccessKey(ctx context.Context, apiKey, apiSecret string) (wallet.DataWallet, error) {
	dw, err := rf.loadRemoteWallet(ctx, func(mlc *caller.MetaLockerHTTPCaller) error {
		return mlc.LoginWithAccessKeys(ctx, apiKey, apiSecret)
	})
	if err != nil {
		return nil, err
	}

	if err = dw.UnlockWithAccessKey(ctx, apiKey, apiSecret); err != nil {
		_ = dw.Close()
		return nil, err
	}

	return dw, nil
}

func (rf *Factory) GetWalletWithCredentials(ctx context.Context, userID, secret string) (wallet.DataWallet, error) {
	dw, err := rf.loadRemoteWallet(ctx, func(mlc *caller.MetaLockerHTTPCaller) error {
		return mlc.LoginWithCredentials(userID, secret)
	})
	if err != nil {
		return nil, err
	}

	if err = dw.Unlock(ctx, secret); err != nil {
		_ = dw.Close()
		return nil, err
	}

	return dw, nil
}

func (rf *Factory) GetWalletWithTokenAndKey(ctx context.Context, jwtToken string, managedKey *model.AESKey) (wallet.DataWallet, error) {
	w, err := rf.loadRemoteWallet(ctx, func(mlc *caller.MetaLockerHTTPCaller) error {
		return mlc.LoginWithJWT(jwtToken)
	})
	if err != nil {
		return nil, err
	}

	if err = w.UnlockAsManaged(ctx, managedKey); err != nil {
		return nil, err
	}

	return w, nil
}

func (rf *Factory) getAccount(id string) *account.Account {
	if rf.accountCache != nil {
		cachedItem, err := rf.accountCache.Value(id)
		if err == nil {
			return cachedItem.Data().(*account.Account)
		}
	}

	return nil
}

func (rf *Factory) setAccount(acct *account.Account) {
	if rf.accountCache != nil {
		rf.accountCache.Add(acct.ID, rf.accountCacheLifeSpan, acct)
	}
}

func (rf *Factory) loadRemoteWallet(ctx context.Context, authFn func(mlc *caller.MetaLockerHTTPCaller) error) (wallet.DataWallet, error) {
	log.Debug().Str("url", rf.url).Msg("Initialising remote Data Wallet")

	httpCaller, err := caller.NewMetaLockerHTTPCaller(rf.url, factoryUserAgent)
	if err != nil {
		return nil, err
	}

	if err = authFn(httpCaller); err != nil {
		return nil, err
	}

	acct := rf.getAccount(httpCaller.AuthenticatedAccountID())

	if acct == nil {
		acct, err = httpCaller.GetOwnAccount(ctx)
		if err != nil {
			return nil, err
		}

		rf.setAccount(acct)
	}

	if rf.indexClient == nil {
		rf.indexClient, err = rf.indexClientSourceFn(acct.ID, httpCaller)
		if err != nil {
			return nil, err
		}
	}

	return wallet.NewLocalDataWallet(acct, httpCaller, nil, rf.indexClient)
}
