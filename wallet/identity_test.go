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
	"errors"
	"testing"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/piprate/metalocker/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdentityWrapper_NewLocker_Hosted(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	dw := testHostedAccount(t, env, true)

	ctx := env.Ctx

	idy, err := dw.NewIdentity(ctx, model.AccessLevelHosted, "John XXX")
	require.NoError(t, err)

	locker, err := idy.NewLocker(ctx, idy.Name())
	require.NoError(t, err)
	require.NotEmpty(t, locker)
	assert.Equal(t, "John XXX", locker.Name())
	assert.True(t, locker.IsUniLocker())

	retrievedLocker, err := dw.GetLocker(ctx, locker.ID())
	require.NoError(t, err)
	require.NotEmpty(t, retrievedLocker)

	retrievedLocker, err = dw.GetLocker(ctx, "non-existent-locker")
	require.True(t, errors.Is(err, storage.ErrLockerNotFound))
	require.Empty(t, retrievedLocker)
}

func TestIdentityWrapper_NewLocker_Managed(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	dw := testHostedAccount(t, env, true)

	ctx := env.Ctx

	idy, err := dw.NewIdentity(ctx, model.AccessLevelManaged, "John XXX")
	require.NoError(t, err)

	locker, err := idy.NewLocker(ctx, idy.Name())
	require.NoError(t, err)
	assert.NotEmpty(t, locker)

	assert.Equal(t, "John XXX", locker.Name())
	assert.True(t, locker.IsUniLocker())
}

func TestLocalDataWallet_NewIdentity_Hosted(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	dw := testHostedAccount(t, env, false)

	ctx := env.Ctx

	// this call should fail because the wallet is locked
	_, err := dw.NewIdentity(ctx, model.AccessLevelHosted, "John XXX")
	require.Error(t, err)

	err = dw.Unlock(ctx, TestPassphrase)
	require.NoError(t, err)

	idy, err := dw.NewIdentity(ctx, model.AccessLevelHosted, "John XXX")
	require.NoError(t, err)

	err = dw.Lock()
	require.NoError(t, err)

	err = dw.Unlock(ctx, TestPassphrase)
	require.NoError(t, err)

	text := []byte("plain text")
	identityForSignature, err := dw.GetIdentity(ctx, idy.ID())
	require.NoError(t, err)
	require.NotNil(t, identityForSignature)
	_ = identityForSignature.DID().Sign(text)

	idList, err := dw.GetIdentities(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(idList))
	assert.Equal(t, "John XXX", idList[idy.ID()].Name())
	assert.Equal(t, account.CurrentAccountVersion, dw.Account().Version)
}

func TestLocalDataWallet_NewIdentity_ManagedToHosted(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	dw := testHostedAccount(t, env, false)

	ctx := env.Ctx

	hashedPassphrase := account.HashUserPassword(TestPassphrase)
	managedKey, err := dw.Account().ExtractManagedKey(hashedPassphrase)
	require.NoError(t, err)

	err = dw.UnlockAsManaged(ctx, managedKey)
	require.NoError(t, err)

	idy, err := dw.NewIdentity(ctx, model.AccessLevelManaged, "John XXX")
	require.NoError(t, err)

	require.NoError(t, dw.Lock())

	err = dw.Unlock(ctx, TestPassphrase)
	require.NoError(t, err)

	idList, err := dw.GetIdentities(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(idList))
	assert.Equal(t, "John XXX", idList[idy.ID()].Name())
	assert.Equal(t, account.CurrentAccountVersion, dw.Account().Version)
}

func TestLocalDataWallet_NewIdentity_ManagedToManaged(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	dw := testHostedAccount(t, env, false)

	ctx := env.Ctx

	hashedPassphrase := account.HashUserPassword(TestPassphrase)
	managedKey, err := dw.Account().ExtractManagedKey(hashedPassphrase)
	require.NoError(t, err)

	err = dw.UnlockAsManaged(ctx, managedKey)
	require.NoError(t, err)

	idy, err := dw.NewIdentity(ctx, model.AccessLevelManaged, "John XXX")
	require.NoError(t, err)

	require.NoError(t, dw.Lock())

	err = dw.UnlockAsManaged(ctx, managedKey)
	require.NoError(t, err)

	idList, err := dw.GetIdentities(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(idList))
	assert.Equal(t, "John XXX", idList[idy.ID()].Name())
	assert.Equal(t, account.CurrentAccountVersion, dw.Account().Version)
}
