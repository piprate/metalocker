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
	"testing"
	"time"

	"github.com/piprate/metalocker/index"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/dataset"
	"github.com/piprate/metalocker/model/expiry"
	"github.com/piprate/metalocker/sdk/testbase"
	. "github.com/piprate/metalocker/wallet"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexUpdater(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer env.Close()

	dw := env.CreateTestManagedAccount(t)

	log.Warn().Msg("~~~~~ BEGIN TEST ~~~~~")

	lockers, err := dw.GetLockers()
	require.NoError(t, err)
	require.True(t, len(lockers) > 0)

	locker := lockers[0] // one of the root lockers

	rootIndex, err := dw.CreateRootIndex(testbase.IndexStoreName)
	require.NoError(t, err)

	updater, err := dw.IndexUpdater(rootIndex)
	require.NoError(t, err)
	err = updater.StartSyncOnEvents(env.NS, true, 0)
	require.NoError(t, err)
	defer updater.Close()

	lb, err := dw.DataStore().NewDataSetBuilder(locker.ID, dataset.WithVault(testbase.TestVaultName))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test1",
		"type": "TestDataset1",
	})
	require.NoError(t, err)

	f := lb.Submit(expiry.FromNow("1h"))

	err = f.Wait(time.Second * 5)
	require.NoError(t, err)

	sleepTime := time.Millisecond * 10
	var totalWaitingTime time.Duration
	for totalWaitingTime < time.Second {
		rec, err := rootIndex.GetRecord(f.ID())
		require.NoError(t, err)

		if rec != nil {
			return
		}

		time.Sleep(sleepTime)

		totalWaitingTime += sleepTime
	}

	assert.Fail(t, "timeout when waiting for an index record")
}

func TestIndexUpdater_PublicRecord(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer env.Close()

	dw := env.CreateTestManagedAccount(t)

	log.Warn().Msg("~~~~~ BEGIN TEST ~~~~~")

	lockers, err := dw.GetLockers()
	require.NoError(t, err)
	require.True(t, len(lockers) > 0)

	locker := lockers[0] // one of the root lockers

	rootIndex, err := dw.CreateRootIndex(testbase.IndexStoreName)
	require.NoError(t, err)

	updater, err := dw.IndexUpdater(rootIndex)
	require.NoError(t, err)
	err = updater.StartSyncOnEvents(env.NS, true, 0)
	require.NoError(t, err)
	defer updater.Close()

	lb, err := dw.DataStore().NewDataSetBuilder(locker.ID,
		dataset.WithVault(testbase.TestVaultName),
		dataset.AsCleartext()) // create a cleartext record
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test1",
		"type": "TestDataset1",
	})
	require.NoError(t, err)

	f := lb.Submit(expiry.FromNow("1h"))

	err = f.Wait(time.Second * 5)
	require.NoError(t, err)

	sleepTime := time.Millisecond * 10
	var totalWaitingTime time.Duration
	for totalWaitingTime < time.Second {
		rec, err := rootIndex.GetRecord(f.ID())
		require.NoError(t, err)

		if rec != nil {
			return
		}

		time.Sleep(sleepTime)

		totalWaitingTime += sleepTime
	}

	assert.Fail(t, "timeout when waiting for an index record")
}

func TestUpdater_MultipleIndexes(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer env.Close()

	// set up wallet 1

	dataWallet1 := env.CreateCustomAccount(t, "test1@example.com", "John Doe 1", model.AccessLevelManaged)

	idy1, err := dataWallet1.NewIdentity(model.AccessLevelManaged, "")
	require.NoError(t, err)

	uniLocker, err := idy1.NewLocker("UniLocker")
	require.NoError(t, err)

	// set up wallet 2

	dataWallet2 := env.CreateCustomAccount(t, "test2@example.com", "John Doe 2", model.AccessLevelManaged)

	idy2, err := dataWallet2.NewIdentity(model.AccessLevelManaged, "")
	require.NoError(t, err)

	// set up a shared locker

	sharedLocker, err := idy1.NewLocker("Test Locker", Participant(idy2.DID(), nil))
	require.NoError(t, err)

	_, err = dataWallet2.AddLocker(sharedLocker.Raw().Perspective(idy2.ID()))
	require.NoError(t, err)

	lb, err := uniLocker.NewDataSetBuilder(dataset.WithVault(testbase.TestVaultName))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test1",
		"type": "TestDataset1",
	})
	require.NoError(t, err)

	f := lb.Submit(expiry.FromNow("1h"))
	require.NoError(t, f.Error()) // don't wait for this record

	rid1 := f.ID()

	lb, err = sharedLocker.NewDataSetBuilder(dataset.WithVault(testbase.TestVaultName))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test2",
		"type": "TestDataset2",
	})
	require.NoError(t, err)

	f = lb.Submit(expiry.FromNow("1h"))
	err = f.Wait(time.Second * 2)
	require.NoError(t, err)

	rid2 := f.ID()

	_, err = env.IndexStore.CreateIndex(dataWallet1.ID(), index.TypeRoot, model.AccessLevelManaged)
	require.NoError(t, err)

	index1, err := env.IndexStore.RootIndex(dataWallet1.ID(), model.AccessLevelManaged)
	require.NoError(t, err)

	iw1, _ := index1.Writer()
	err = iw1.AddLockerState(dataWallet1.ID(), uniLocker.ID(), uniLocker.Raw().FirstBlock)
	require.NoError(t, err)
	err = iw1.AddLockerState(dataWallet1.ID(), sharedLocker.ID(), sharedLocker.Raw().FirstBlock)
	require.NoError(t, err)

	_, err = env.IndexStore.CreateIndex(dataWallet2.ID(), index.TypeRoot, model.AccessLevelManaged)
	require.NoError(t, err)

	index2, err := env.IndexStore.RootIndex(dataWallet2.ID(), model.AccessLevelManaged)
	require.NoError(t, err)

	iw2, _ := index2.Writer()
	err = iw2.AddLockerState(dataWallet2.ID(), sharedLocker.ID(), sharedLocker.Raw().FirstBlock)
	require.NoError(t, err)

	updater := NewIndexUpdater(env.Ledger)

	err = updater.AddIndexes(dataWallet1, index1)
	require.NoError(t, err)
	err = updater.AddIndexes(dataWallet2, index2)
	require.NoError(t, err)

	err = updater.Sync()
	require.NoError(t, err)

	// check if records were added into relevant Indexes

	resRec1, err := index1.GetRecord(rid1)
	require.NoError(t, err)
	require.NotEmpty(t, resRec1)
	assert.Equal(t, uniLocker.ID(), resRec1.LockerID)

	resRec2, err := index1.GetRecord(rid2)
	require.NoError(t, err)
	require.NotEmpty(t, resRec2)
	assert.Equal(t, sharedLocker.ID(), resRec2.LockerID)

	resRec2, err = index2.GetRecord(rid2)
	require.NoError(t, err)
	require.NotEmpty(t, resRec2)
	assert.Equal(t, sharedLocker.ID(), resRec2.LockerID)

	// check that Index statistics were updated

	topBlock, _ := env.Ledger.GetTopBlock()
	states, _ := iw1.LockerStates()
	for _, ls := range states {
		assert.Equal(t, topBlock.Number, ls.TopBlock)
	}

	states, _ = iw2.LockerStates()
	for _, ls := range states {
		assert.Equal(t, topBlock.Number, ls.TopBlock)
	}
}

func TestIndexUpdater_StaggeredLockers(t *testing.T) {

	// This test creates an updater for an index with 1 locker state.
	// Then it runs a sync, adds the second locker state, and runs the sync
	// again. The sync should pick up the second locker.

	env := testbase.SetUpTestEnvironment(t)
	defer env.Close()

	dw := env.CreateCustomAccount(t, "test1@example.com", "John Doe", model.AccessLevelManaged)

	log.Warn().Msg("~~~~~ BEGIN TEST ~~~~~")

	lockers, err := dw.GetLockers()
	require.NoError(t, err)
	require.Equal(t, 1, len(lockers))

	locker := lockers[0]

	rootIndex, err := dw.CreateRootIndex(testbase.IndexStoreName)
	require.NoError(t, err)

	iw, _ := rootIndex.Writer()
	states, err := iw.LockerStates()
	require.NoError(t, err)
	require.Equal(t, 1, len(states))

	updater, err := dw.IndexUpdater(rootIndex)
	require.NoError(t, err)
	defer updater.Close()

	lb, err := dw.DataStore().NewDataSetBuilder(locker.ID, dataset.WithVault(testbase.TestVaultName))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test1",
		"type": "TestDataset1",
	})
	require.NoError(t, err)

	f := lb.Submit(expiry.FromNow("1h"))

	err = f.Wait(time.Second * 5)
	require.NoError(t, err)

	err = updater.Sync()
	require.NoError(t, err)

	rec, err := rootIndex.GetRecord(f.ID())
	require.NoError(t, err)
	require.NotEmpty(t, rec)

	idy2, err := dw.NewIdentity(model.AccessLevelManaged, "Second Identity")
	require.NoError(t, err)

	locker2, err := idy2.NewLocker("Second Locker")
	require.NoError(t, err)

	lb, err = locker2.NewDataSetBuilder(dataset.WithVault(testbase.TestVaultName))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test1",
		"type": "TestDataset1",
	})
	require.NoError(t, err)

	f = lb.Submit(expiry.FromNow("1h"))

	err = f.Wait(time.Second * 5)
	require.NoError(t, err)

	err = updater.Sync()
	require.NoError(t, err)

	states, err = iw.LockerStates()
	require.NoError(t, err)
	require.Equal(t, 2, len(states))

	rec, err = rootIndex.GetRecord(f.ID())
	require.NoError(t, err)
	require.NotEmpty(t, rec)
}

func TestIndexUpdater_SubAccounts(t *testing.T) {

	// This test creates an updater for an index with 1 locker state.
	// Then it creates a sub-account with one locker and publishes a dataset
	// into it. The sync should pick up both the locker from the sub-account
	// and the dataset in it.

	env := testbase.SetUpTestEnvironment(t)
	defer env.Close()

	dw := env.CreateTestManagedAccount(t)

	log.Warn().Msg("~~~~~ BEGIN TEST ~~~~~")

	lockers, err := dw.GetLockers()
	require.NoError(t, err)
	require.Equal(t, 1, len(lockers))

	locker := lockers[0]

	rootIndex, err := dw.CreateRootIndex(testbase.IndexStoreName)
	require.NoError(t, err)

	iw, _ := rootIndex.Writer()
	states, err := iw.LockerStates()
	require.NoError(t, err)
	require.Equal(t, 1, len(states))

	updater, err := dw.IndexUpdater(rootIndex)
	require.NoError(t, err)
	defer updater.Close()

	lb, err := dw.DataStore().NewDataSetBuilder(locker.ID, dataset.WithVault(testbase.TestVaultName))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test1",
		"type": "TestDataset1",
	})
	require.NoError(t, err)

	f := lb.Submit(expiry.FromNow("1h"))

	err = f.Wait(time.Second * 5)
	require.NoError(t, err)

	err = updater.Sync()
	require.NoError(t, err)

	rec, err := rootIndex.GetRecord(f.ID())
	require.NoError(t, err)
	require.NotEmpty(t, rec)

	subDW, err := dw.CreateSubAccount(model.AccessLevelManaged, "Sub-Account")
	require.NoError(t, err)

	idy1, err := subDW.NewIdentity(model.AccessLevelManaged, "Second Identity")
	require.NoError(t, err)

	secondLocker, err := idy1.NewLocker("Second Locker")
	require.NoError(t, err)

	lb, err = secondLocker.NewDataSetBuilder(dataset.WithVault(testbase.TestVaultName))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test2",
		"type": "TestDataset2",
	})
	require.NoError(t, err)

	f = lb.Submit(expiry.FromNow("1h"))

	err = f.Wait(time.Second * 5)
	require.NoError(t, err)

	_, err = iw.LockerStates()
	require.NoError(t, err)

	err = updater.Sync()
	require.NoError(t, err)

	states, err = iw.LockerStates()
	require.NoError(t, err)
	require.Equal(t, 3, len(states))

	rec, err = rootIndex.GetRecord(f.ID())
	require.NoError(t, err)
	require.NotEmpty(t, rec)

	// create another IndexUpdater instance and ensure it picks up
	// the previously saved locker states from the sub-account

	secondUpdater, err := dw.IndexUpdater(rootIndex)
	require.NoError(t, err)
	defer secondUpdater.Close()

	lb, err = secondLocker.NewDataSetBuilder(dataset.WithVault(testbase.TestVaultName))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test3",
		"type": "TestDataset3",
	})
	require.NoError(t, err)

	f = lb.Submit(expiry.FromNow("1h"))

	err = f.Wait(time.Second * 5)
	require.NoError(t, err)

	err = secondUpdater.Sync()
	require.NoError(t, err)

	rec, err = rootIndex.GetRecord(f.ID())
	require.NoError(t, err)
	require.NotEmpty(t, rec)
}

func TestForceSyncRootIndex_LockerStateExists(t *testing.T) {

	// When ForceSyncRootIndex is invoked, it will add locker states
	// to the index that are most likely included in existing AccountUpdate
	// messages. This means that when we process these messages,
	// we may try adding a locker state that already exists. That should be fine,
	// and the Sync() process should complete without errors.

	env := testbase.SetUpTestEnvironment(t)
	defer env.Close()

	dw := env.CreateTestManagedAccount(t)

	log.Warn().Msg("~~~~~ BEGIN TEST ~~~~~")

	rootIndex, err := dw.CreateRootIndex(testbase.IndexStoreName)
	require.NoError(t, err)

	updater, err := dw.IndexUpdater(rootIndex)
	require.NoError(t, err)
	defer updater.Close()

	err = updater.Sync()
	require.NoError(t, err)

	idy1, err := dw.NewIdentity(model.AccessLevelManaged, "Second Identity")
	require.NoError(t, err)

	_, err = idy1.NewLocker("New Locker")
	require.NoError(t, err)

	err = ForceSyncRootIndex(dw)
	require.NoError(t, err)

	err = updater.Sync()
	require.NoError(t, err)
}

func TestIndexUpdater_RemoveIndex(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer env.Close()

	dw := env.CreateTestManagedAccount(t)

	log.Warn().Msg("~~~~~ BEGIN TEST ~~~~~")

	lockers, err := dw.GetLockers()
	require.NoError(t, err)
	require.True(t, len(lockers) > 0)

	locker := lockers[0] // one of the root lockers

	rootIndex, err := dw.CreateRootIndex(testbase.IndexStoreName)
	require.NoError(t, err)

	updater, err := dw.IndexUpdater(rootIndex)
	require.NoError(t, err)
	err = updater.StartSyncOnEvents(env.NS, true, 0)
	require.NoError(t, err)
	defer updater.Close()

	lb, err := dw.DataStore().NewDataSetBuilder(locker.ID, dataset.WithVault(testbase.TestVaultName))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test1",
		"type": "TestDataset1",
	})
	require.NoError(t, err)

	f := lb.Submit(expiry.FromNow("1h"))

	err = f.Wait(time.Second * 5)
	require.NoError(t, err)

	sleepTime := time.Millisecond * 10
	var totalWaitingTime time.Duration
	received := false
	for totalWaitingTime < time.Second {
		rec, err := rootIndex.GetRecord(f.ID())
		require.NoError(t, err)

		if rec != nil {
			received = true
			break
		}

		time.Sleep(sleepTime)

		totalWaitingTime += sleepTime
	}

	assert.True(t, received, "timeout when waiting for an index record")

	err = updater.RemoveIndex(rootIndex.ID())
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test2",
		"type": "TestDataset2",
	})
	require.NoError(t, err)

	f = lb.Submit(expiry.FromNow("1h"))

	err = f.Wait(time.Second * 5)
	require.NoError(t, err)

	totalWaitingTime = 0
	received = false
	for totalWaitingTime < time.Second {
		rec, err := rootIndex.GetRecord(f.ID())
		require.NoError(t, err)

		if rec != nil {
			received = true
			break
		}

		time.Sleep(sleepTime)

		totalWaitingTime += sleepTime
	}

	assert.False(t, received, "received a record in the removed index")
}
