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

package testbase

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/contexts"
	"github.com/piprate/metalocker/index"
	"github.com/piprate/metalocker/index/bolt"
	"github.com/piprate/metalocker/ledger/local"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/node"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/services/notification"
	"github.com/piprate/metalocker/storage"
	"github.com/piprate/metalocker/storage/memory"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/vaults"
	"github.com/piprate/metalocker/wallet"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = contexts.PreloadContextsIntoMemory()
	SetupLogFormat()
}

type TestMetaLockerEnvironment struct {
	IdentityBackend storage.IdentityBackend
	BlobManager     *vaults.LocalBlobManager
	IndexStore      index.Store
	IndexClient     index.Client
	OffChainStorage model.OffChainStorage
	NS              notification.Service
	Ledger          model.Ledger
	Factory         *wallet.LocalFactory
	Router          *gin.Engine
	TempDir         string
	ExtraServices   map[string]any
	Ctx             context.Context
}

func (env *TestMetaLockerEnvironment) Close() error {
	if env.IndexStore != nil {
		if err := env.IndexStore.Close(); err != nil {
			log.Err(err).Msg("Error closing Index Store")
		}
	}

	if env.Ledger != nil {
		if err := env.Ledger.Close(); err != nil {
			log.Err(err).Msg("Error closing test ledger")
		}
	}

	for _, s := range env.ExtraServices {
		if c, isCloser := s.(io.Closer); isCloser {
			if err := c.Close(); err != nil {
				log.Err(err).Msg("Error closing resource")
			}
		}

	}

	return os.RemoveAll(env.TempDir)
}

func (env *TestMetaLockerEnvironment) CreateDataWallet(t *testing.T, acct *account.Account) wallet.DataWallet {
	t.Helper()

	dw, err := env.Factory.CreateDataWallet(acct)
	require.NoError(t, err)

	return dw
}

func (env *TestMetaLockerEnvironment) CreateTestManagedAccount(t *testing.T) wallet.DataWallet {
	t.Helper()

	//refreshTestAccount(t)

	respBytes := []byte(testManagedAccount)
	var resp *account.GenerationResponse
	err := jsonw.Unmarshal(respBytes, &resp)
	require.NoError(t, err)

	nodeClient := wallet.NewLocalNodeClient(resp.Account.ID, env.IdentityBackend, env.Ledger, env.OffChainStorage,
		env.BlobManager, env.NS)

	err = wallet.SaveNewAccount(env.Ctx, resp, nodeClient, "", fastHash)
	require.NoError(t, err)

	dw := env.CreateDataWallet(t, resp.Account)

	err = dw.Unlock(env.Ctx, TestAccountPassword)
	require.NoError(t, err)

	return dw
}

func (env *TestMetaLockerEnvironment) CreateCustomAccount(t *testing.T, email, name string, lvl model.AccessLevel, rootIdentityOptions ...model.DIDOption) wallet.DataWallet {
	t.Helper()

	acctTemplate := &account.Account{
		Email:        email,
		Name:         name,
		AccessLevel:  lvl,
		DefaultVault: TestVaultName,
	}

	rootIdentity, err := model.GenerateDID(rootIdentityOptions...)
	require.NoError(t, err)

	dw, _, err := env.Factory.RegisterAccount(env.Ctx,
		acctTemplate,
		account.WithPassphraseAuth(TestAccountPassword),
		account.WithRootIdentity(rootIdentity))
	require.NoError(t, err)

	err = dw.Unlock(env.Ctx, TestAccountPassword)
	require.NoError(t, err)

	return dw
}

func (env *TestMetaLockerEnvironment) SetUpTestScenario1(t *testing.T) {
	t.Helper()

	dw := env.CreateTestManagedAccount(t)

	did1 := TestDID(t)

	idy1 := &account.Identity{
		DID:         did1,
		AccessLevel: model.AccessLevelManaged,
	}

	unilocker := TestUniLocker(t)

	err := dw.AddIdentity(env.Ctx, idy1)
	require.NoError(t, err)

	_, err = dw.AddLocker(env.Ctx, unilocker)
	require.NoError(t, err)

	env.Router.Use(func(c *gin.Context) {
		c.Set(apibase.UserIDKey, TestAccountID)
		c.Set(apibase.ClientSecretKey, TestAccountClientSecret)
		c.Next()
	})
}

func (env *TestMetaLockerEnvironment) TestScenario1Context() *gin.Context {
	c := &gin.Context{}
	c.Set(apibase.UserIDKey, TestAccountID)
	c.Set(apibase.ClientSecretKey, TestAccountClientSecret)
	return c
}

func SetUpTestEnvironment(t *testing.T) *TestMetaLockerEnvironment {
	t.Helper()

	env := &TestMetaLockerEnvironment{
		ExtraServices: make(map[string]any),
		Ctx:           context.Background(),
	}

	router := gin.Default()
	env.Router = router

	dir, err := os.MkdirTemp(".", "tempdir_")
	require.NoError(t, err)
	env.TempDir = dir

	env.IdentityBackend, _ = memory.CreateIdentityBackend(nil, nil)

	dbFilepath := filepath.Join(dir, "ledger.bolt")

	env.NS = notification.NewLocalNotificationService(100)

	ledgerAPI, err := local.NewBoltLedger(env.Ctx, dbFilepath, env.NS, 10, 0)
	require.NoError(t, err)

	env.BlobManager = TestBlobManager(t, false, ledgerAPI)

	gb, err := ledgerAPI.GetGenesisBlock(env.Ctx)
	require.NoError(t, err)

	env.Ledger = ledgerAPI

	ocVault, _ := NewInMemoryVault(t, TestOffChainStorageID, "offchain", false, true, nil)
	env.OffChainStorage = node.NewOffChainStorageProxy(ocVault)

	env.IndexClient, err = index.NewLocalIndexClient(env.Ctx, []*index.StoreConfig{
		{
			ID:   IndexStoreID,
			Name: IndexStoreName,
			Type: bolt.Type,
			Params: map[string]any{
				bolt.ParameterFilePath: filepath.Join(dir, "index_store.bolt"),
			},
		},
	}, nil, gb.Hash)
	require.NoError(t, err)

	env.IndexStore, err = env.IndexClient.IndexStore(env.Ctx, IndexStoreName)
	require.NoError(t, err)

	// create wallet factory

	env.Factory, err = wallet.NewLocalFactory(env.Ledger, env.OffChainStorage, env.BlobManager, env.IdentityBackend, env.NS, env.IndexClient, fastHash)
	require.NoError(t, err)

	router.Use(apibase.SetRequestLogger())

	return env
}
