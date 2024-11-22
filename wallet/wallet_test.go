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

package wallet_test

import (
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/piprate/metalocker/storage"
	. "github.com/piprate/metalocker/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocalDataWallet_Hosted(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	acctTemplate := &account.Account{
		Email:       "test@example.com",
		Name:        "John Doe",
		AccessLevel: model.AccessLevelHosted,
	}
	passPhrase := TestPassphrase
	resp, err := account.GenerateAccount(acctTemplate, account.WithPassphraseAuth(passPhrase))
	require.NoError(t, err)
	assert.NotEqual(t, "", resp.RecoveryPhrase)

	dw := env.CreateDataWallet(t, resp.Account)

	assert.Equal(t, "test@example.com", dw.Account().Email)
}

func TestNewLocalDataWallet_Managed(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	acctTemplate := &account.Account{
		Name:        "Test Managed User",
		AccessLevel: model.AccessLevelManaged,
	}
	hashedPassphrase := account.HashUserPassword(TestPassphrase)
	secondLevelRecoveryKey := base58.Decode("8GhJ8gQ59gkGSooCYmcRVqtdcNNWvNyoVUYLX496r5Hg")
	resp, err := account.GenerateAccount(acctTemplate, account.WithHashedPassphraseAuth(hashedPassphrase), account.WithSLRK(secondLevelRecoveryKey))
	require.NoError(t, err)
	assert.NotEmpty(t, resp.RecoveryPhrase)
	assert.NotEmpty(t, resp.SecondLevelRecoveryCode)

	acct := resp.Account
	managedKey, err := acct.ExtractManagedKey(hashedPassphrase)
	require.NoError(t, err)

	dw := env.CreateDataWallet(t, acct)

	err = dw.UnlockAsManaged(env.Ctx, managedKey)
	require.NoError(t, err)

	err = dw.Lock()
	require.NoError(t, err)
}

func TestLocalDataWallet_Unlock(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	dw := testHostedAccount(t, env, false)

	err := dw.Unlock(env.Ctx, "wrong password")
	assert.Error(t, err)

	err = dw.Unlock(env.Ctx, TestPassphrase)
	assert.NoError(t, err)
}

func TestLocalDataWallet_UnlockAsManaged_Hosted(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	dw := testHostedAccount(t, env, false)

	wrongKey := model.NewEncryptionKey()
	err := dw.UnlockAsManaged(env.Ctx, wrongKey)
	assert.Error(t, err)

	hashedPassphrase := account.HashUserPassword(TestPassphrase)

	managedKey, err := dw.Account().ExtractManagedKey(hashedPassphrase)
	require.NoError(t, err)

	err = dw.UnlockAsManaged(env.Ctx, managedKey)
	assert.NoError(t, err)
}

func TestLocalDataWallet_UnlockAsManaged_Managed(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	dw := testManagedAccount(t, env, false)

	wrongKey := model.NewEncryptionKey()
	err := dw.UnlockAsManaged(env.Ctx, wrongKey)
	assert.Error(t, err)

	hashedPassphrase := account.HashUserPassword(TestPassphrase)

	managedKey, err := dw.Account().ExtractManagedKey(hashedPassphrase)
	require.NoError(t, err)

	err = dw.UnlockAsManaged(env.Ctx, managedKey)
	assert.NoError(t, err)
}

func TestLocalDataWallet_EncryptionKey(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	dw := testHostedAccount(t, env, true)

	// check key derivation is deterministic for the same tag

	key1, err := dw.EncryptionKey("test", model.AccessLevelHosted)
	assert.NoError(t, err)

	key2, err := dw.EncryptionKey("test", model.AccessLevelHosted)
	assert.NoError(t, err)

	assert.True(t, base64.StdEncoding.EncodeToString(key1[:]) == base64.StdEncoding.EncodeToString(key2[:]))

	// ... and other tags produce a different key

	key3, err := dw.EncryptionKey("another_test", model.AccessLevelHosted)
	assert.NoError(t, err)

	assert.False(t, base64.StdEncoding.EncodeToString(key1[:]) == base64.StdEncoding.EncodeToString(key3[:]))
}

const (
	TestPassphrase = "pass123"
)

func testHostedAccount(t *testing.T, env *testbase.TestMetaLockerEnvironment, unlock bool) DataWallet { //nolint:thelper
	return testAccount(t, env, model.AccessLevelHosted, unlock)
}

func testManagedAccount(t *testing.T, env *testbase.TestMetaLockerEnvironment, unlock bool) DataWallet { //nolint:thelper
	return testAccount(t, env, model.AccessLevelManaged, unlock)
}

func testAccount(t *testing.T, env *testbase.TestMetaLockerEnvironment, accessLevel model.AccessLevel, unlock bool) DataWallet {
	t.Helper()

	dw, _, err := env.Factory.RegisterAccount(
		env.Ctx,
		&account.Account{
			Email:        "test@example.com",
			Name:         "John Doe",
			AccessLevel:  accessLevel,
			DefaultVault: testbase.TestVaultName,
		},
		account.WithPassphraseAuth(TestPassphrase))
	require.NoError(t, err)

	if unlock {
		err = dw.Unlock(env.Ctx, TestPassphrase)
		require.NoError(t, err)
	}

	return dw
}

func TestLocalDataWallet_UnlockWithAccessKey(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	dw := testHostedAccount(t, env, true)

	key, err := dw.CreateAccessKey(env.Ctx, model.AccessLevelHosted, 1*time.Hour)
	require.NoError(t, err)

	apiKey, apiSecret := key.ClientKeys()

	err = dw.Lock()
	require.NoError(t, err)

	err = dw.UnlockWithAccessKey(env.Ctx, apiKey, apiSecret)
	require.NoError(t, err)

	idyList, err := dw.GetIdentities(env.Ctx)
	require.NoError(t, err)
	assert.NotNil(t, idyList)
}

func TestLocalDataWallet_AddLocker_Hosted(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	dw := testHostedAccount(t, env, true)

	idy, err := dw.NewIdentity(env.Ctx, model.AccessLevelHosted, "John XXX")
	require.NoError(t, err)

	expiryTime := time.Now().AddDate(0, 120, 0).UTC()
	locker, err := model.GenerateLocker(model.AccessLevelHosted, idy.Name(), &expiryTime, 1,
		model.Us(idy.DID(), nil))
	require.NoError(t, err)

	wrapper, err := dw.AddLocker(env.Ctx, locker)
	require.NoError(t, err)
	require.NotEmpty(t, wrapper)
	assert.Equal(t, "John XXX", wrapper.Name())
	assert.True(t, wrapper.IsUniLocker())

	retrievedLocker, err := dw.GetLocker(env.Ctx, locker.ID)
	require.NoError(t, err)
	require.NotEmpty(t, retrievedLocker)

	retrievedLocker, err = dw.GetLocker(env.Ctx, "non-existent-locker")
	require.True(t, errors.Is(err, storage.ErrLockerNotFound))
	require.Empty(t, retrievedLocker)
}

func TestLocalDataWallet_AddLocker_Managed(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	dw := testHostedAccount(t, env, true)

	idy, err := dw.NewIdentity(env.Ctx, model.AccessLevelManaged, "John XXX")
	require.NoError(t, err)

	expiryTime := time.Now().AddDate(0, 120, 0).UTC()
	locker, err := model.GenerateLocker(model.AccessLevelManaged, idy.Name(), &expiryTime, 1,
		model.Us(idy.DID(), nil))
	require.NoError(t, err)

	wrapper, err := dw.AddLocker(env.Ctx, locker)
	require.NoError(t, err)
	assert.NotEmpty(t, wrapper)

	assert.Equal(t, "John XXX", wrapper.Name())
	assert.True(t, wrapper.IsUniLocker())
}

func TestLocalDataWallet_AddLocker_NoFirstBlock(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	dw := testHostedAccount(t, env, true)

	idy, err := dw.NewIdentity(env.Ctx, model.AccessLevelManaged, "John XXX", WithType(account.IdentityTypePairwise))
	require.NoError(t, err)

	expiryTime := time.Now().AddDate(0, 120, 0).UTC()
	locker, err := model.GenerateLocker(model.AccessLevelManaged, idy.Name(), &expiryTime, 0,
		model.Us(idy.DID(), nil))
	require.NoError(t, err)

	l, err := dw.AddLocker(env.Ctx, locker)
	require.NoError(t, err)

	assert.True(t, l.Raw().FirstBlock != 0)
}

func TestLocalDataWallet_AddLocker_ThirdParty(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	dw := testHostedAccount(t, env, true)

	did, err := model.GenerateDID()
	require.NoError(t, err)

	expiryTime := time.Now().AddDate(0, 120, 0).UTC()
	locker, err := model.GenerateLocker(model.AccessLevelHosted, "Test Locker", &expiryTime, 1,
		model.Us(did, nil))
	require.NoError(t, err)

	_, err = dw.AddLocker(env.Ctx, locker.Perspective(""))
	require.NoError(t, err)

	savedLocker, err := dw.GetLocker(env.Ctx, locker.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, savedLocker)
}

func TestLocalDataWallet_SetProperty(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	dw := testHostedAccount(t, env, false)

	hashedPassphrase := account.HashUserPassword(TestPassphrase)
	managedKey, err := dw.Account().ExtractManagedKey(hashedPassphrase)
	require.NoError(t, err)

	// adding new property should fail because the wallet is locked

	err = dw.SetProperty(env.Ctx, "key1", "value1", model.AccessLevelHosted)
	require.Error(t, err)

	err = dw.Unlock(env.Ctx, TestPassphrase)
	require.NoError(t, err)

	// try saving a hosted property

	err = dw.SetProperty(env.Ctx, "key1", "value1", model.AccessLevelHosted)
	require.NoError(t, err)

	val, err := dw.GetProperty(env.Ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", val)

	// try saving a managed property

	err = dw.SetProperty(env.Ctx, "key2", "value2", model.AccessLevelManaged)
	require.NoError(t, err)

	val, err = dw.GetProperty(env.Ctx, "key2")
	require.NoError(t, err)
	assert.Equal(t, "value2", val)

	// try saving a managed property that overrides an existing hosted property

	err = dw.SetProperty(env.Ctx, "key1", "managed_value1", model.AccessLevelManaged)
	require.NoError(t, err)

	val, err = dw.GetProperty(env.Ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", val)

	valMap, err := dw.GetProperties(env.Ctx)
	require.NoError(t, err)
	assert.EqualValues(t, map[string]string{
		"key1": "value1",
		"key2": "value2",
	}, valMap)

	require.NoError(t, dw.Lock())

	err = dw.UnlockAsManaged(env.Ctx, managedKey)
	require.NoError(t, err)

	val, err = dw.GetProperty(env.Ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, "managed_value1", val)

	valMap, err = dw.GetProperties(env.Ctx)
	require.NoError(t, err)
	assert.EqualValues(t, map[string]string{
		"key1": "managed_value1",
		"key2": "value2",
	}, valMap)
}

func TestLocalDataWallet_DeleteProperty(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	dw := testHostedAccount(t, env, true)

	hashedPassphrase := account.HashUserPassword(TestPassphrase)
	managedKey, err := dw.Account().ExtractManagedKey(hashedPassphrase)
	require.NoError(t, err)

	// insert test properties

	err = dw.SetProperty(env.Ctx, "key1", "hosted_value1", model.AccessLevelHosted)
	require.NoError(t, err)

	err = dw.SetProperty(env.Ctx, "key2", "value2", model.AccessLevelManaged)
	require.NoError(t, err)

	err = dw.SetProperty(env.Ctx, "key3", "value3", model.AccessLevelManaged)
	require.NoError(t, err)

	err = dw.SetProperty(env.Ctx, "key1", "managed_value1", model.AccessLevelManaged)
	require.NoError(t, err)

	// try deleting a property with a wrong (non-existent) level

	err = dw.DeleteProperty(env.Ctx, "key2", model.AccessLevelHosted)
	require.Error(t, err)
	require.True(t, errors.Is(err, storage.ErrPropertyNotFound))

	// this call should succeed

	err = dw.DeleteProperty(env.Ctx, "key2", model.AccessLevelManaged)
	require.NoError(t, err)

	// delete a property when a higher level property exists

	err = dw.DeleteProperty(env.Ctx, "key1", model.AccessLevelManaged)
	require.NoError(t, err)

	valMap, err := dw.GetProperties(env.Ctx)
	require.NoError(t, err)
	assert.EqualValues(t, map[string]string{
		"key1": "hosted_value1",
		"key3": "value3",
	}, valMap)

	require.NoError(t, dw.Lock())

	err = dw.UnlockAsManaged(env.Ctx, managedKey)
	require.NoError(t, err)

	valMap, err = dw.GetProperties(env.Ctx)
	require.NoError(t, err)
	assert.EqualValues(t, map[string]string{
		"key3": "value3",
	}, valMap)
}

func TestLocalDataWallet_CreateSubAccount_Hosted(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	// create hosted and managed sub-accounts of a hosted account

	dw, _, err := env.Factory.RegisterAccount(
		env.Ctx,
		&account.Account{
			Email:        "test@example.com",
			Name:         "John Doe",
			AccessLevel:  model.AccessLevelHosted,
			DefaultVault: testbase.TestVaultName,
		},
		account.WithPassphraseAuth(TestPassphrase))
	require.NoError(t, err)

	err = dw.Unlock(env.Ctx, TestPassphrase)
	require.NoError(t, err)

	saWallet, err := dw.CreateSubAccount(env.Ctx, model.AccessLevelHosted, "Sub-Account #1 (hosted)",
		account.WithPassphraseAuth(TestPassphrase))
	require.NoError(t, err)
	subAcct := saWallet.Account()
	assert.Equal(t, dw.ID(), subAcct.ParentAccount)
	assert.Equal(t, dw.ID(), subAcct.MasterAccount)

	saWallet = env.CreateDataWallet(t, subAcct)

	err = saWallet.Unlock(env.Ctx, TestPassphrase)
	require.NoError(t, err)

	// create a managed sub-account when the wallet is unlocked at Hosted level

	saWallet, err = dw.CreateSubAccount(env.Ctx, model.AccessLevelManaged, "Sub-Account #2 (managed)",
		account.WithPassphraseAuth(TestPassphrase))
	require.NoError(t, err)
	subAcct = saWallet.Account()
	assert.Equal(t, dw.ID(), subAcct.ParentAccount)
	assert.Equal(t, dw.ID(), subAcct.MasterAccount)

	saWallet = env.CreateDataWallet(t, subAcct)

	err = saWallet.Unlock(env.Ctx, TestPassphrase)
	require.NoError(t, err)

	// create a managed sub-account when the wallet is unlocked at Managed level

	hashedPassword := account.HashUserPassword(TestPassphrase)
	managedKey, err := dw.Account().ExtractManagedKey(hashedPassword)
	require.NoError(t, err)

	require.NoError(t, dw.Lock())

	err = dw.UnlockAsManaged(env.Ctx, managedKey)
	require.NoError(t, err)

	saWallet, err = dw.CreateSubAccount(env.Ctx, model.AccessLevelManaged, "Sub-Account #3 (managed)",
		account.WithPassphraseAuth(TestPassphrase))
	require.NoError(t, err)
	subAcct = saWallet.Account()
	assert.Equal(t, dw.ID(), subAcct.ParentAccount)
	assert.Equal(t, dw.ID(), subAcct.MasterAccount)

	saWallet = env.CreateDataWallet(t, subAcct)

	err = saWallet.Unlock(env.Ctx, TestPassphrase)
	require.NoError(t, err)
}

func TestLocalDataWallet_CreateSubAccount_Managed(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	// create a managed sub-account of a managed account

	dw, _, err := env.Factory.RegisterAccount(
		env.Ctx,
		&account.Account{
			Email:        "test@example.com",
			Name:         "John Doe",
			AccessLevel:  model.AccessLevelManaged,
			DefaultVault: testbase.TestVaultName,
		},
		account.WithPassphraseAuth(TestPassphrase))
	require.NoError(t, err)

	err = dw.Unlock(env.Ctx, TestPassphrase)
	require.NoError(t, err)

	_, err = dw.CreateSubAccount(env.Ctx, model.AccessLevelHosted, "Sub-Account #1 (hosted)",
		account.WithPassphraseAuth(TestPassphrase))
	require.Error(t, err)

	saWallet, err := dw.CreateSubAccount(env.Ctx, model.AccessLevelManaged, "Sub-Account #2 (managed)",
		account.WithPassphraseAuth(TestPassphrase))
	require.NoError(t, err)
	subAcct := saWallet.Account()
	assert.Equal(t, dw.ID(), subAcct.ParentAccount)
	assert.Equal(t, dw.ID(), subAcct.MasterAccount)

	saWallet = env.CreateDataWallet(t, subAcct)

	err = saWallet.Unlock(env.Ctx, TestPassphrase)
	require.NoError(t, err)

	saWallet, err = dw.GetSubAccountWallet(env.Ctx, subAcct.ID)
	require.NoError(t, err)

	_, err = saWallet.GetIdentities(env.Ctx)
	require.NoError(t, err)
}

func TestLocalDataWallet_CreateRestrictedWallet(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	dw := testHostedAccount(t, env, true)

	_, err := dw.NewIdentity(env.Ctx, model.AccessLevelManaged, "John XXX")
	require.NoError(t, err)

	idyTwo, err := dw.NewIdentity(env.Ctx, model.AccessLevelManaged, "John YYY")
	require.NoError(t, err)

	idList, err := dw.GetIdentities(env.Ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, len(idList))

	_, err = idyTwo.NewLocker(env.Ctx, idyTwo.Name())
	require.NoError(t, err)

	rdw, err := dw.RestrictedWallet([]string{idyTwo.ID()})
	require.NoError(t, err)

	idList, err = rdw.GetIdentities(env.Ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(idList))

	lockerList, err := rdw.GetLockers(env.Ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(lockerList))
}

func TestLocalDataWallet_GetRootIdentity(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	dw := testHostedAccount(t, env, true)

	rootIdy, err := dw.GetRootIdentity(env.Ctx)
	require.NoError(t, err)
	assert.Equal(t, dw.ID(), rootIdy.ID())
}

func TestLocalDataWallet_ChangePassphrase_Hosted(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	passPhrase := TestPassphrase

	dw, _, err := env.Factory.RegisterAccount(
		env.Ctx,
		&account.Account{
			Name:         "Test Hosted User",
			AccessLevel:  model.AccessLevelHosted,
			DefaultVault: testbase.TestVaultName,
		},
		account.WithPassphraseAuth(passPhrase))
	require.NoError(t, err)

	err = dw.Unlock(env.Ctx, passPhrase)
	require.NoError(t, err)

	idy, err := dw.NewIdentity(env.Ctx, model.AccessLevelHosted, "John XXX")
	require.NoError(t, err)

	newPassPhrase := "new123"
	newWallet, err := dw.ChangePassphrase(env.Ctx, passPhrase, newPassPhrase, false)
	require.NoError(t, err)
	assert.NotNil(t, newWallet)

	err = newWallet.Lock()
	require.NoError(t, err)

	err = newWallet.Unlock(env.Ctx, newPassPhrase)
	require.NoError(t, err)

	text := []byte("plain text")
	identityForSignature, err := newWallet.GetIdentity(env.Ctx, idy.ID())
	require.NoError(t, err)
	_ = identityForSignature.DID().Sign(text)

	idList, err := newWallet.GetIdentities(env.Ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(idList))
	assert.Equal(t, "John XXX", idList[idy.ID()].Name())
}

func TestLocalDataWallet_ChangePassphrase_Managed(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	passPhrase := TestPassphrase
	hashedPassphrase := account.HashUserPassword(passPhrase)

	dw, _, err := env.Factory.RegisterAccount(
		env.Ctx,
		&account.Account{
			Name:         "Test Managed User",
			AccessLevel:  model.AccessLevelManaged,
			DefaultVault: testbase.TestVaultName,
		},
		account.WithHashedPassphraseAuth(hashedPassphrase))
	require.NoError(t, err)

	err = dw.Unlock(env.Ctx, passPhrase)
	require.NoError(t, err)

	idy, err := dw.NewIdentity(env.Ctx, model.AccessLevelManaged, "John XXX")
	require.NoError(t, err)

	newPassPhrase := "new123"
	newWallet, err := dw.ChangePassphrase(env.Ctx, passPhrase, newPassPhrase, false)
	require.NoError(t, err)
	assert.NotNil(t, newWallet)

	// test unlocking with password

	err = newWallet.Lock()
	require.NoError(t, err)

	err = newWallet.Unlock(env.Ctx, newPassPhrase)
	require.NoError(t, err)

	text := []byte("plain text")
	identityForSignature, err := newWallet.GetIdentity(env.Ctx, idy.ID())
	require.NoError(t, err)
	_ = identityForSignature.DID().Sign(text)

	idList, err := newWallet.GetIdentities(env.Ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(idList))
	assert.Equal(t, "John XXX", idList[idy.ID()].Name())

	// test unlocking with password hash

	newPassPhraseHash := account.HashUserPassword(newPassPhrase)
	err = newWallet.Lock()
	require.NoError(t, err)

	managedKey, err := newWallet.Account().ExtractManagedKey(newPassPhraseHash)
	require.NoError(t, err)

	err = newWallet.UnlockAsManaged(env.Ctx, managedKey)
	require.NoError(t, err)

	identityForSignature, err = newWallet.GetIdentity(env.Ctx, idy.ID())
	require.NoError(t, err)
	_ = identityForSignature.DID().Sign(text)

	idList, err = newWallet.GetIdentities(env.Ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(idList))
	assert.Equal(t, "John XXX", idList[idy.ID()].Name())
}

func TestLocalDataWallet_ChangePassphrase_ManagedWithHash(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	passPhrase := TestPassphrase
	hashedPassphrase := account.HashUserPassword(passPhrase)

	dw, _, err := env.Factory.RegisterAccount(
		env.Ctx,
		&account.Account{
			Name:         "Test Managed User",
			AccessLevel:  model.AccessLevelManaged,
			DefaultVault: testbase.TestVaultName,
		},
		account.WithHashedPassphraseAuth(hashedPassphrase))
	require.NoError(t, err)

	err = dw.Unlock(env.Ctx, passPhrase)
	require.NoError(t, err)

	idy, err := dw.NewIdentity(env.Ctx, model.AccessLevelManaged, "John XXX")
	require.NoError(t, err)

	newPassPhrase := "new123"
	newPassPhraseHash := account.HashUserPassword(newPassPhrase)
	newWallet, err := dw.ChangePassphrase(env.Ctx, hashedPassphrase, newPassPhraseHash, true)
	require.NoError(t, err)
	assert.NotNil(t, newWallet)

	// test unlocking with password

	err = newWallet.Lock()
	require.NoError(t, err)

	err = newWallet.Unlock(env.Ctx, newPassPhrase)
	require.NoError(t, err)

	text := []byte("plain text")
	identityForSignature, err := newWallet.GetIdentity(env.Ctx, idy.ID())
	require.NoError(t, err)
	_ = identityForSignature.DID().Sign(text)

	idList, err := newWallet.GetIdentities(env.Ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(idList))
	assert.Equal(t, "John XXX", idList[idy.ID()].Name())

	// test unlocking with password hash

	err = newWallet.Lock()
	require.NoError(t, err)

	managedKey, err := newWallet.Account().ExtractManagedKey(newPassPhraseHash)
	require.NoError(t, err)

	err = newWallet.UnlockAsManaged(env.Ctx, managedKey)
	require.NoError(t, err)

	identityForSignature, err = newWallet.GetIdentity(env.Ctx, idy.ID())
	require.NoError(t, err)
	_ = identityForSignature.DID().Sign(text)

	idList, err = newWallet.GetIdentities(env.Ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(idList))
	assert.Equal(t, "John XXX", idList[idy.ID()].Name())
}

func TestLocalDataWallet_Recover_Hosted(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	oldPassPhrase := "passw0rd"

	dw, recDetails, err := env.Factory.RegisterAccount(
		env.Ctx,
		&account.Account{
			Name:         "Test Hosted User",
			AccessLevel:  model.AccessLevelHosted,
			DefaultVault: testbase.TestVaultName,
		},
		account.WithPassphraseAuth(oldPassPhrase))
	require.NoError(t, err)

	recoveryPhrase := recDetails.RecoveryPhrase
	assert.NotEqual(t, "", recoveryPhrase)

	err = dw.Unlock(env.Ctx, oldPassPhrase)
	require.NoError(t, err)

	idy, err := dw.NewIdentity(env.Ctx, model.AccessLevelHosted, "John XXX")
	require.NoError(t, err)

	err = dw.Lock()
	require.NoError(t, err)

	newPassPhrase := "newpass"

	cryptoKey, _, _, err := account.GenerateKeysFromRecoveryPhrase(recoveryPhrase)
	require.NoError(t, err)

	newDataWallet, err := dw.Recover(env.Ctx, cryptoKey, newPassPhrase)
	require.NoError(t, err)

	err = newDataWallet.Lock()
	require.NoError(t, err)

	err = newDataWallet.Unlock(env.Ctx, oldPassPhrase)
	assert.Error(t, err)

	err = newDataWallet.Unlock(env.Ctx, newPassPhrase)
	require.NoError(t, err)

	idList, err := newDataWallet.GetIdentities(env.Ctx)
	require.NoError(t, err)

	assert.Equal(t, 2, len(idList))
	assert.Equal(t, "John XXX", idList[idy.ID()].Name())
}

func TestLocalDataWallet_Recover_Managed(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	oldPassPhrase := TestPassphrase
	hashedPassphrase := account.HashUserPassword(oldPassPhrase)

	dw, recDetails, err := env.Factory.RegisterAccount(
		env.Ctx,
		&account.Account{
			Name:         "Test Managed User",
			AccessLevel:  model.AccessLevelManaged,
			DefaultVault: testbase.TestVaultName,
		},
		account.WithHashedPassphraseAuth(hashedPassphrase))
	require.NoError(t, err)

	recoveryPhrase := recDetails.RecoveryPhrase
	assert.NotEqual(t, "", recoveryPhrase)

	err = dw.Unlock(env.Ctx, oldPassPhrase)
	require.NoError(t, err)

	idy, err := dw.NewIdentity(env.Ctx, model.AccessLevelManaged, "John XXX")
	require.NoError(t, err)

	err = dw.Lock()
	require.NoError(t, err)

	newPassPhrase := "newpass"

	cryptoKey, _, _, err := account.GenerateKeysFromRecoveryPhrase(recoveryPhrase)
	require.NoError(t, err)

	newDataWallet, err := dw.Recover(env.Ctx, cryptoKey, newPassPhrase)
	require.NoError(t, err)

	err = newDataWallet.Lock()
	require.NoError(t, err)

	err = newDataWallet.Unlock(env.Ctx, oldPassPhrase)
	assert.Error(t, err)

	err = newDataWallet.Unlock(env.Ctx, newPassPhrase)
	require.NoError(t, err)

	idList, err := newDataWallet.GetIdentities(env.Ctx)
	require.NoError(t, err)

	assert.Equal(t, 2, len(idList))
	assert.Equal(t, "John XXX", idList[idy.ID()].Name())
}

func TestLocalDataWallet_RecoverWithPayload(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	passPhrase := TestPassphrase

	dw, recDetails, err := env.Factory.RegisterAccount(
		env.Ctx,
		&account.Account{
			Email:        "test@example.com",
			Name:         "John Doe",
			AccessLevel:  model.AccessLevelHosted,
			DefaultVault: testbase.TestVaultName,
		},
		account.WithPassphraseAuth(passPhrase))
	require.NoError(t, err)

	recoveryPhrase := recDetails.RecoveryPhrase

	err = dw.Unlock(env.Ctx, passPhrase)
	require.NoError(t, err)

	_, err = dw.NewIdentity(env.Ctx, model.AccessLevelLocal, "John XXX")
	require.NoError(t, err)

	err = dw.Lock()
	require.NoError(t, err)

	newPassPhrase := "newpass"

	cryptoKey, _, _, err := account.GenerateKeysFromRecoveryPhrase(recoveryPhrase)
	require.NoError(t, err)

	newDataWallet, err := dw.Recover(env.Ctx, cryptoKey, newPassPhrase)
	require.NoError(t, err)

	err = newDataWallet.Lock()
	require.NoError(t, err)

	err = newDataWallet.Unlock(env.Ctx, newPassPhrase)
	require.NoError(t, err)

	idList, err := newDataWallet.GetIdentities(env.Ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(idList))
}
