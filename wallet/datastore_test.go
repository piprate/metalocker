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
	"io"
	"strings"
	"testing"
	"time"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/dataset"
	"github.com/piprate/metalocker/model/expiry"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/piprate/metalocker/utils"
	"github.com/piprate/metalocker/utils/jsonw"
	. "github.com/piprate/metalocker/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalStoreImpl_AssetHead_SetAssetHead(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	ctx := env.Ctx

	// set up wallet 1

	dataWallet1 := env.CreateCustomAccount(t, "test1@example.com", "John Doe 1", model.AccessLevelManaged)

	idy1, err := dataWallet1.NewIdentity(ctx, model.AccessLevelManaged, "")
	require.NoError(t, err)

	// set up wallet 2

	dataWallet2 := env.CreateCustomAccount(t, "test2@example.com", "John Doe 2", model.AccessLevelManaged)

	idy2, err := dataWallet2.NewIdentity(ctx, model.AccessLevelManaged, "")
	require.NoError(t, err)

	// set up a shared locker

	sharedLocker, err := idy1.NewLocker(ctx, "Test Locker", Participant(idy2.DID(), nil))
	require.NoError(t, err)

	_, err = dataWallet2.AddLocker(ctx, sharedLocker.Raw().Perspective(idy2.ID()))
	require.NoError(t, err)

	// create a data set

	assetID := "test1"

	lb, err := sharedLocker.NewDataSetBuilder(ctx, dataset.WithVault(testbase.TestVaultName))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   assetID,
		"type": "TestDataset1",
	})
	require.NoError(t, err)

	f := lb.Submit(expiry.FromNow("1h"))
	require.NoError(t, f.Wait(time.Second*10))

	recordID := f.ID()

	// check the asset head doesn't exist

	headID := sharedLocker.HeadID(ctx, assetID, "test")

	_, err = dataWallet1.DataStore().AssetHead(ctx, headID)
	require.Error(t, err)
	assert.Equal(t, model.ErrAssetHeadNotFound, err)

	// set the asset head to the new data set

	headFuture := sharedLocker.SetAssetHead(ctx, assetID, "test", recordID)
	require.NoError(t, headFuture.Wait(time.Second*10))

	// read the asset head and run checks

	// with FromLocker option

	_, err = dataWallet1.DataStore().AssetHead(ctx, headID, dataset.FromLocker(sharedLocker.ID()))
	require.NoError(t, err)

	// without FromLocker option

	headDataSet, err := dataWallet1.DataStore().AssetHead(ctx, headID)
	require.NoError(t, err)

	var headResource map[string]string
	err = headDataSet.DecodeMetaResource(&headResource)
	require.NoError(t, err)

	assert.Equal(t, "TestDataset1", headResource["type"])

	rs, err := dataWallet1.Services().Ledger().GetRecordState(headFuture.ID())
	require.NoError(t, err)
	assert.Equal(t, model.StatusPublished, rs.Status)

	// add a new revision of the data set

	lb, err = sharedLocker.NewDataSetBuilder(ctx, dataset.WithVault(testbase.TestVaultName))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   assetID,
		"type": "TestDataset2",
	})
	require.NoError(t, err)

	f = lb.Submit(expiry.FromNow("1h"))
	require.NoError(t, f.Wait(time.Second*10))

	recordID = f.ID()

	// update the asset head

	headFuture2 := sharedLocker.SetAssetHead(ctx, assetID, "test", recordID)
	require.NoError(t, headFuture2.Wait(time.Second*10))

	// run checks

	headDataSet2, err := dataWallet1.DataStore().AssetHead(ctx, headID)
	require.NoError(t, err)

	var headResource2 map[string]string
	err = headDataSet2.DecodeMetaResource(&headResource2)
	require.NoError(t, err)

	assert.Equal(t, "TestDataset2", headResource2["type"])

	rs, err = dataWallet1.Services().Ledger().GetRecordState(headFuture.ID())
	require.NoError(t, err)
	assert.Equal(t, model.StatusRevoked, rs.Status)

	// read the head from wallet 2

	headDataSet3, err := dataWallet2.DataStore().AssetHead(ctx, headID)
	require.NoError(t, err)

	var headResource3 map[string]string
	err = headDataSet3.DecodeMetaResource(&headResource3)
	require.NoError(t, err)

	assert.Equal(t, "TestDataset2", headResource3["type"])
}

func TestLocalStoreImpl_Submit_SetHeads(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	ctx := env.Ctx

	dw := env.CreateCustomAccount(t, "test1@example.com", "John Doe 1", model.AccessLevelManaged)

	idy, err := dw.NewIdentity(ctx, model.AccessLevelManaged, "")
	require.NoError(t, err)

	uniLocker, err := idy.NewLocker(ctx, "Test Locker")
	require.NoError(t, err)

	// create a data set

	lb, err := uniLocker.NewDataSetBuilder(ctx, dataset.WithVault(testbase.TestVaultName), dataset.SetHeads("test"))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test1",
		"type": "TestDataset1",
	})
	require.NoError(t, err)

	f := lb.Submit(expiry.FromNow("1h"))
	require.NoError(t, f.Wait(time.Second*10))

	// read the asset head and run checks

	headID := uniLocker.HeadID(ctx, f.DataSet().Impression().Asset, "test")

	headDataSet, err := dw.DataStore().AssetHead(ctx, headID)
	require.NoError(t, err)

	var headResource map[string]string
	err = headDataSet.DecodeMetaResource(&headResource)
	require.NoError(t, err)

	assert.Equal(t, "TestDataset1", headResource["type"])

	headRecord1 := f.Heads()[headID]

	rs, err := dw.Services().Ledger().GetRecordState(headRecord1)
	require.NoError(t, err)
	assert.Equal(t, model.StatusPublished, rs.Status)

	// add a new revision of the data set

	lb, err = uniLocker.NewDataSetBuilder(ctx, dataset.WithVault(testbase.TestVaultName), dataset.SetHeads("test"))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test1",
		"type": "TestDataset2",
	})
	require.NoError(t, err)

	f = lb.Submit(expiry.FromNow("1h"))
	require.NoError(t, f.Wait(time.Second*10))

	// run checks

	headDataSet2, err := dw.DataStore().AssetHead(ctx, headID)
	require.NoError(t, err)

	var headResource2 map[string]string
	err = headDataSet2.DecodeMetaResource(&headResource2)
	require.NoError(t, err)

	assert.Equal(t, "TestDataset2", headResource2["type"])

	rs, err = dw.Services().Ledger().GetRecordState(headRecord1)
	require.NoError(t, err)
	assert.Equal(t, model.StatusRevoked, rs.Status)
}

func TestLocalStoreImpl_Submit_Public(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	ctx := env.Ctx

	// register a vault with SSE turned off
	vaultName := "non_sse_vault_id"
	vault, vaultCfg := testbase.NewInMemoryVault(t, "non_sse_vault_id", vaultName, false, false, env.Ledger)

	env.BlobManager.AddVault(vault, vaultCfg)

	dw := env.CreateCustomAccount(t, "test1@example.com", "John Doe 1", model.AccessLevelManaged)

	idy, err := dw.NewIdentity(ctx, model.AccessLevelManaged, "")
	require.NoError(t, err)

	uniLocker, err := idy.NewLocker(ctx, "Test Locker")
	require.NoError(t, err)

	// create a data set with a plain text attachment

	lb, err := uniLocker.NewDataSetBuilder(ctx, dataset.WithVault(vaultName), dataset.AsCleartext())
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test1",
		"type": "TestDataset1",
	})
	require.NoError(t, err)

	attachmentID, err := lb.AddResource(strings.NewReader("attachment"))
	require.NoError(t, err)

	f := lb.Submit(expiry.FromNow("1h"))
	require.NoError(t, f.Wait(time.Second*10))

	ds, err := dw.DataStore().Load(ctx, f.ID())
	require.NoError(t, err)

	var resource map[string]string
	err = ds.DecodeMetaResource(&resource)
	require.NoError(t, err)
	assert.Equal(t, "TestDataset1", resource["type"])

	// check the lease is not encrypted

	opBytes, err := env.OffChainStorage.GetOperation(ds.Record().OperationAddress)
	require.NoError(t, err)

	var lease model.Lease
	err = jsonw.Unmarshal(opBytes, &lease)
	require.NoError(t, err)

	accessToken := lease.GenerateAccessToken(ds.ID())

	// check the meta resource is not encrypted

	rdr, err := vault.ServeBlob(lease.MetaResource().ID, nil, accessToken)
	require.NoError(t, err)
	metaBytes, err := io.ReadAll(rdr)
	require.NoError(t, err)
	err = jsonw.Unmarshal(metaBytes, &resource)
	require.NoError(t, err)

	// check the attachment is not encrypted

	rdr, err = ds.Resource(attachmentID)
	require.NoError(t, err)
	attBytes, err := io.ReadAll(rdr)
	require.NoError(t, err)
	assert.Equal(t, "attachment", string(attBytes))

	rdr, err = vault.ServeBlob(lease.Resource(attachmentID).ID, nil, accessToken)
	require.NoError(t, err)
	attBytes, err = io.ReadAll(rdr)
	require.NoError(t, err)
	assert.Equal(t, "attachment", string(attBytes))
}

func TestLocalStoreImpl_Load(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	ctx := env.Ctx

	// set up a wallet

	dataWallet := env.CreateCustomAccount(t, "test1@example.com", "John Doe 1", model.AccessLevelManaged)

	idy, err := dataWallet.NewIdentity(ctx, model.AccessLevelManaged, "")
	require.NoError(t, err)

	// set up a uni-locker

	uniLocker, err := idy.NewLocker(ctx, "Uni-locker")
	require.NoError(t, err)

	// create a data set

	lb, err := uniLocker.NewDataSetBuilder(ctx, dataset.WithVault(testbase.TestVaultName))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test1",
		"type": "TestDataset1",
	})
	require.NoError(t, err)

	f := lb.Submit(expiry.FromNow("1h"))
	require.NoError(t, f.Wait(time.Second*10))

	recordID := f.ID()

	// load with FromLocker option

	_, err = dataWallet.DataStore().Load(ctx, recordID, dataset.FromLocker(uniLocker.ID()))
	require.NoError(t, err)

	// load with FromLocker option (wrong locker)

	_, err = dataWallet.DataStore().Load(ctx, recordID, dataset.FromLocker("8XPEnfnmyhb8ZBmU7Ln7QCrR8jZusJvvoZdn9qKCUpND"))
	require.NoError(t, err)

	// load without FromLocker option

	_, err = dataWallet.DataStore().Load(ctx, recordID)
	require.NoError(t, err)
}

func TestLocalStoreImpl_Load_PublicFromForeignLocker(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	ctx := env.Ctx

	// register a vault with SSE turned off
	vaultName := "non_sse_vault_id"
	vault, vaultCfg := testbase.NewInMemoryVault(t, "non_sse_vault_id", vaultName, false, false, env.Ledger)

	env.BlobManager.AddVault(vault, vaultCfg)

	dw := env.CreateCustomAccount(t, "test1@example.com", "John Doe 1", model.AccessLevelManaged)

	idy, err := dw.NewIdentity(ctx, model.AccessLevelManaged, "")
	require.NoError(t, err)

	uniLocker, err := idy.NewLocker(ctx, "Test Locker")
	require.NoError(t, err)

	// create a data set with a plain text attachment

	lb, err := uniLocker.NewDataSetBuilder(ctx, dataset.WithVault(vaultName), dataset.AsCleartext())
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test1",
		"type": "TestDataset1",
	})
	require.NoError(t, err)

	f := lb.Submit(expiry.FromNow("1h"))
	require.NoError(t, f.Wait(time.Second*10))

	readerWallet := env.CreateCustomAccount(t, "test2@example.com", "John Reader", model.AccessLevelManaged)

	ds, err := readerWallet.DataStore().Load(ctx, f.ID())
	require.NoError(t, err)

	var resource map[string]string
	err = ds.DecodeMetaResource(&resource)
	require.NoError(t, err)
	assert.Equal(t, "TestDataset1", resource["type"])
}

func TestLocalStoreImpl_Share(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	ctx := env.Ctx

	// set up wallet 1

	dataWallet1 := env.CreateCustomAccount(t, "test1@example.com", "John Doe 1", model.AccessLevelManaged)

	idy1, err := dataWallet1.NewIdentity(ctx, model.AccessLevelManaged, "")
	require.NoError(t, err)

	// set up wallet 2

	dataWallet2 := env.CreateCustomAccount(t, "test2@example.com", "John Doe 2", model.AccessLevelManaged)

	idy2, err := dataWallet2.NewIdentity(ctx, model.AccessLevelManaged, "")
	require.NoError(t, err)

	// set up a uni-locker

	uniLocker, err := idy1.NewLocker(ctx, "Uni-locker")
	require.NoError(t, err)

	// set up a shared locker

	sharedLocker, err := idy1.NewLocker(ctx, "Test Locker", Participant(idy2.DID(), nil))
	require.NoError(t, err)

	_, err = dataWallet2.AddLocker(ctx, sharedLocker.Raw().Perspective(idy2.ID()))
	require.NoError(t, err)

	rootIndex1, err := dataWallet1.CreateRootIndex(ctx, testbase.IndexStoreName)
	require.NoError(t, err)

	updater1, err := dataWallet1.IndexUpdater(ctx, rootIndex1)
	require.NoError(t, err)
	defer updater1.Close()

	rootIndex2, err := dataWallet2.CreateRootIndex(ctx, testbase.IndexStoreName)
	require.NoError(t, err)

	updater2, err := dataWallet2.IndexUpdater(ctx, rootIndex2)
	require.NoError(t, err)
	defer updater2.Close()

	// create a data set

	lb, err := uniLocker.NewDataSetBuilder(ctx, dataset.WithVault(testbase.TestVaultName))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test1",
		"type": "TestDataset1",
	})
	require.NoError(t, err)

	f := lb.Submit(expiry.FromNow("1h"))
	require.NoError(t, f.Wait(time.Second*10))

	recordID := f.ID()

	checkAccess := func(dw DataWallet, id string, canAccess bool) {
		ds, err := dw.DataStore().Load(ctx, id)
		if canAccess {
			require.NoError(t, err, "no access")
		} else if err != nil {
			// operation has been purged
			return
		}

		rdr, err := ds.MetaResource()
		if canAccess {
			require.NoError(t, err)

			_, err = io.ReadAll(rdr)
			require.NoError(t, err)
		} else {
			require.Error(t, err, "has access")
		}
	}

	// check both sides can have access to the dataset

	checkAccess(dataWallet1, recordID, true)
	checkAccess(dataWallet2, recordID, false)

	require.NoError(t, updater1.Sync())
	require.NoError(t, updater2.Sync())

	// try sharing a non-existent record

	f = sharedLocker.Share(ctx, "bad-id", testbase.TestVaultName, expiry.Months(12))
	require.Error(t, f.Error())

	f = sharedLocker.Share(ctx, recordID, testbase.TestVaultName, expiry.Months(12))

	require.NoError(t, f.Wait(time.Minute))

	sharedRecordID := f.ID()

	require.NoError(t, updater1.Sync())
	require.NoError(t, updater2.Sync())

	checkAccess(dataWallet1, sharedRecordID, true)
	checkAccess(dataWallet2, recordID, false)
	checkAccess(dataWallet2, sharedRecordID, true)
}

func TestLocalStoreImpl_Revoke(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	ctx := env.Ctx

	// set up wallet 1

	dataWallet1 := env.CreateCustomAccount(t, "test1@example.com", "John Doe 1", model.AccessLevelManaged)

	idy1, err := dataWallet1.NewIdentity(ctx, model.AccessLevelManaged, "")
	require.NoError(t, err)

	// set up wallet 2

	dataWallet2 := env.CreateCustomAccount(t, "test2@example.com", "John Doe 2", model.AccessLevelManaged)

	idy2, err := dataWallet2.NewIdentity(ctx, model.AccessLevelManaged, "")
	require.NoError(t, err)

	// set up a shared locker

	sharedLocker, err := idy1.NewLocker(ctx, "Test Locker", Participant(idy2.DID(), nil))
	require.NoError(t, err)

	_, err = dataWallet2.AddLocker(ctx, sharedLocker.Raw().Perspective(idy2.ID()))
	require.NoError(t, err)

	rootIndex1, err := dataWallet1.CreateRootIndex(ctx, testbase.IndexStoreName)
	require.NoError(t, err)

	updater, err := dataWallet1.IndexUpdater(ctx, rootIndex1)
	require.NoError(t, err)
	defer updater.Close()

	// create a data set

	lb, err := sharedLocker.NewDataSetBuilder(ctx, dataset.WithVault(testbase.TestVaultName))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test1",
		"type": "TestDataset1",
	})
	require.NoError(t, err)

	f := lb.Submit(expiry.FromNow("1h"))
	require.NoError(t, f.Wait(time.Second*10))

	recordID := f.ID()

	checkAccess := func(dw DataWallet, canAccess bool) {
		ds, err := dw.DataStore().Load(ctx, recordID)
		if canAccess {
			require.NoError(t, err)
		} else if err != nil {
			// operation has been purged
			return
		}

		rdr, err := ds.MetaResource()
		if canAccess {
			require.NoError(t, err)

			_, err = io.ReadAll(rdr)
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}

	// check both sides can have access to the dataset

	checkAccess(dataWallet1, true)
	checkAccess(dataWallet2, true)

	require.NoError(t, updater.Sync())

	// revocation of somebody else's lease should fail
	f = dataWallet2.DataStore().Revoke(ctx, recordID)
	require.Error(t, f.Wait(time.Second*10))

	// revocation of non-existent lease should fail
	badRecord, _ := utils.RandomID(32)
	f = dataWallet1.DataStore().Revoke(ctx, badRecord)
	require.Error(t, f.Wait(time.Second*10))

	// should succeed
	f = dataWallet1.DataStore().Revoke(ctx, recordID)
	require.NoError(t, f.Wait(time.Second*10))

	// check both side don't have access anymore
	checkAccess(dataWallet1, false)
	checkAccess(dataWallet2, false)

	require.NoError(t, updater.Sync())

	// check access again after index refresh
	checkAccess(dataWallet1, false)
	checkAccess(dataWallet2, false)

	// try revoking again
	f = dataWallet1.DataStore().Revoke(ctx, recordID)
	require.Error(t, f.Wait(time.Second*10))

	err = dataWallet1.DataStore().PurgeDataAssets(ctx, recordID)
	require.NoError(t, err)

	checkAccess(dataWallet1, false)
	checkAccess(dataWallet2, false)
}
