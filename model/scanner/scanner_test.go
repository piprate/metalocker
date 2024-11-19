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

package scanner_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/piprate/json-gold/ld"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/dataset"
	"github.com/piprate/metalocker/model/expiry"
	. "github.com/piprate/metalocker/model/scanner"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/wallet"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanner_Scan(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer env.Close()

	// set up Data Wallet 1

	dw1 := env.CreateCustomAccount(t, "test1@example.com", "John Doe 1", model.AccessLevelManaged, model.WithSeed("Acct1"))

	idy1, err := dw1.NewIdentity(env.Ctx, model.AccessLevelManaged, "")
	require.NoError(t, err)

	uniLocker, err := idy1.NewLocker(env.Ctx, "UniLocker")
	require.NoError(t, err)

	// set up Data Wallet 2

	dw2 := env.CreateCustomAccount(t, "test2@example.com", "John Doe 2", model.AccessLevelManaged, model.WithSeed("Acct2"))

	idy2, err := dw2.NewIdentity(env.Ctx, model.AccessLevelManaged, "")
	require.NoError(t, err)

	// set up a shared locker

	sharedLocker1, err := idy1.NewLocker(env.Ctx, "Test Locker", wallet.Participant(idy2.DID(), nil))
	require.NoError(t, err)

	sharedLocker2, err := dw2.AddLocker(env.Ctx, sharedLocker1.Raw().Perspective(idy2.ID()))
	require.NoError(t, err)

	// publish several records

	lb, err := uniLocker.NewDataSetBuilder(env.Ctx, dataset.WithVault(testbase.TestVaultName))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test1",
		"type": "TestDataset1",
	})
	require.NoError(t, err)

	f := lb.Submit(expiry.FromNow("1h"))
	require.NoError(t, f.Error()) // don't wait for this record

	rid1 := f.ID()

	lb, err = sharedLocker1.NewDataSetBuilder(env.Ctx, dataset.WithVault(testbase.TestVaultName))
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

	ledgerScanner := NewScanner(env.Ledger)

	consumer1 := &CheckingConsumer{
		T: t,
		ExpectedMatches: []map[string]any{
			{
				"userID":        dw1.ID(),
				"rid":           rid1,
				"t":             model.OpTypeLease,
				"lockerID":      uniLocker.ID(),
				"participantID": idy1.ID(),
				"acceptedAt":    uniLocker.Raw().AcceptedAtBlock(),
			},
			{
				"userID":        dw1.ID(),
				"rid":           rid2,
				"t":             model.OpTypeLease,
				"lockerID":      sharedLocker1.ID(),
				"participantID": idy1.ID(),
				"acceptedAt":    sharedLocker1.Raw().AcceptedAtBlock(),
			},
		},
	}

	sub1 := NewIndexSubscription(dw1.ID(), consumer1)
	_ = sub1.AddLockers(LockerEntry{Locker: uniLocker.Raw()}, LockerEntry{Locker: sharedLocker1.Raw()})
	_ = ledgerScanner.AddSubscription(sub1)

	consumer2 := &CheckingConsumer{
		T: t,
		ExpectedMatches: []map[string]any{
			{
				"userID":        dw2.ID(),
				"rid":           rid2,
				"t":             model.OpTypeLease,
				"lockerID":      sharedLocker2.ID(),
				"participantID": idy1.ID(),
				"acceptedAt":    sharedLocker2.Raw().AcceptedAtBlock(), // NOTE: this value would differ from sharedLocker1's.
			},
		},
	}

	sub2 := NewIndexSubscription(dw2.ID(), consumer2)
	_ = sub2.AddLockers(LockerEntry{Locker: sharedLocker2.Raw()})
	_ = ledgerScanner.AddSubscription(sub2)

	complete, err := ledgerScanner.Scan()
	require.NoError(t, err)
	assert.True(t, complete)

	assert.True(t, consumer1.AllRecordsMatched(), "Received less matching records than expected")
	assert.True(t, consumer2.AllRecordsMatched(), "Received less matching records than expected")
}

func TestScanner_Scan_TwoIterations(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer env.Close()

	// set up Data Wallet 1

	dw1 := env.CreateCustomAccount(t, "test1@example.com", "John Doe 1", model.AccessLevelManaged, model.WithSeed("Acct1"))

	idy1, err := dw1.NewIdentity(env.Ctx, model.AccessLevelManaged, "")
	require.NoError(t, err)

	uniLocker, err := idy1.NewLocker(env.Ctx, "UniLocker")
	require.NoError(t, err)

	// set up Data Wallet 2

	dw2 := env.CreateCustomAccount(t, "test2@example.com", "John Doe 2", model.AccessLevelManaged, model.WithSeed("Acct2"))

	idy2, err := dw2.NewIdentity(env.Ctx, model.AccessLevelManaged, "")
	require.NoError(t, err)

	// set up a shared locker

	sharedLocker1, err := idy1.NewLocker(env.Ctx, "Test Locker", wallet.Participant(idy2.DID(), nil))
	require.NoError(t, err)

	sharedLocker2, err := dw2.AddLocker(env.Ctx, sharedLocker1.Raw().Perspective(idy2.ID()))
	require.NoError(t, err)

	// publish several records

	lb, err := uniLocker.NewDataSetBuilder(env.Ctx, dataset.WithVault(testbase.TestVaultName))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test1",
		"type": "TestDataset1",
	})
	require.NoError(t, err)

	f := lb.Submit(expiry.FromNow("1h"))
	require.NoError(t, f.Error()) // don't wait for this record

	rid1 := f.ID()

	lb, err = sharedLocker1.NewDataSetBuilder(env.Ctx, dataset.WithVault(testbase.TestVaultName))
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

	ledgerScanner := NewScanner(env.Ledger)

	consumer1 := &CheckingConsumer{
		T: t,
		ExpectedMatches: []map[string]any{
			{
				"userID":        dw1.ID(),
				"rid":           rid1,
				"t":             model.OpTypeLease,
				"lockerID":      uniLocker.ID(),
				"participantID": idy1.ID(),
				"acceptedAt":    uniLocker.Raw().AcceptedAtBlock(),
			},
			{
				"userID":        dw1.ID(),
				"rid":           rid2,
				"t":             model.OpTypeLease,
				"lockerID":      sharedLocker1.ID(),
				"participantID": idy1.ID(),
				"acceptedAt":    sharedLocker1.Raw().AcceptedAtBlock(),
			},
		},
	}

	sub1 := NewIndexSubscription(dw1.ID(), consumer1)
	_ = sub1.AddLockers(LockerEntry{Locker: uniLocker.Raw()}, LockerEntry{Locker: sharedLocker1.Raw()})
	_ = ledgerScanner.AddSubscription(sub1)

	consumer2 := &CheckingConsumer{
		T: t,
		ExpectedMatches: []map[string]any{
			{
				"userID":        dw2.ID(),
				"rid":           rid2,
				"t":             model.OpTypeLease,
				"lockerID":      sharedLocker2.ID(),
				"participantID": idy1.ID(),
				"acceptedAt":    sharedLocker2.Raw().AcceptedAtBlock(), // NOTE: this value would differ from sharedLocker1's.
			},
		},
	}

	sub2 := NewIndexSubscription(dw2.ID(), consumer2)
	_ = sub2.AddLockers(LockerEntry{Locker: sharedLocker2.Raw()})
	_ = ledgerScanner.AddSubscription(sub2)

	complete, err := ledgerScanner.Scan()
	require.NoError(t, err)
	assert.True(t, complete)

	assert.True(t, consumer1.AllRecordsMatched(), "Received less matching records than expected")
	assert.True(t, consumer2.AllRecordsMatched(), "Received less matching records than expected")

	lb, err = sharedLocker1.NewDataSetBuilder(env.Ctx, dataset.WithVault(testbase.TestVaultName))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test3",
		"type": "TestDataset3",
	})
	require.NoError(t, err)

	f = lb.Submit(expiry.FromNow("1h"))
	err = f.Wait(time.Second * 2)
	require.NoError(t, err)

	rid3 := f.ID()

	consumer1.ExpectedMatches = []map[string]any{
		{
			"userID":        dw1.ID(),
			"rid":           rid3,
			"t":             model.OpTypeLease,
			"lockerID":      sharedLocker1.ID(),
			"participantID": idy1.ID(),
			"acceptedAt":    sharedLocker1.Raw().AcceptedAtBlock(),
		},
	}
	consumer1.Counter = 0

	consumer2.ExpectedMatches = []map[string]any{
		{
			"userID":        dw2.ID(),
			"rid":           rid3,
			"t":             model.OpTypeLease,
			"lockerID":      sharedLocker2.ID(),
			"participantID": idy1.ID(),
			"acceptedAt":    sharedLocker2.Raw().AcceptedAtBlock(),
		},
	}
	consumer2.Counter = 0

	complete, err = ledgerScanner.Scan()
	require.NoError(t, err)
	assert.True(t, complete)

	assert.True(t, consumer1.AllRecordsMatched(), "Received less matching records than expected")
	assert.True(t, consumer2.AllRecordsMatched(), "Received less matching records than expected")
}

func TestScanner_RemoveSubscription(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer env.Close()

	// set up Data Wallet 1

	dw1 := env.CreateCustomAccount(t, "test1@example.com", "John Doe 1", model.AccessLevelManaged, model.WithSeed("Acct1"))

	idy1, err := dw1.NewIdentity(env.Ctx, model.AccessLevelManaged, "")
	require.NoError(t, err)

	// set up Data Wallet 2

	dw2 := env.CreateCustomAccount(t, "test2@example.com", "John Doe 2", model.AccessLevelManaged, model.WithSeed("Acct2"))

	idy2, err := dw2.NewIdentity(env.Ctx, model.AccessLevelManaged, "")
	require.NoError(t, err)

	// set up a shared locker

	sharedLocker1, err := idy1.NewLocker(env.Ctx, "Test Locker", wallet.Participant(idy2.DID(), nil))
	require.NoError(t, err)

	sharedLocker2, err := dw2.AddLocker(env.Ctx, sharedLocker1.Raw().Perspective(idy2.ID()))
	require.NoError(t, err)

	// publish 1 record

	lb, err := sharedLocker1.NewDataSetBuilder(env.Ctx, dataset.WithVault(testbase.TestVaultName))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test2",
		"type": "TestDataset2",
	})
	require.NoError(t, err)

	f := lb.Submit(expiry.FromNow("1h"))
	err = f.Wait(time.Second * 2)
	require.NoError(t, err)

	rid2 := f.ID()

	ledgerScanner := NewScanner(env.Ledger)

	consumer1 := &CheckingConsumer{
		T: t,
		ExpectedMatches: []map[string]any{
			{
				"userID":        dw1.ID(),
				"rid":           rid2,
				"t":             model.OpTypeLease,
				"lockerID":      sharedLocker1.ID(),
				"participantID": idy1.ID(),
				"acceptedAt":    sharedLocker1.Raw().AcceptedAtBlock(),
			},
		},
	}

	consumer2 := &CheckingConsumer{
		T: t,
		ExpectedMatches: []map[string]any{
			{
				"userID":        dw2.ID(),
				"rid":           rid2,
				"t":             model.OpTypeLease,
				"lockerID":      sharedLocker2.ID(),
				"participantID": idy1.ID(),
				"acceptedAt":    sharedLocker2.Raw().AcceptedAtBlock(), // NOTE: this value would differ from sharedLocker1's.
			},
		},
	}

	sub1 := NewIndexSubscription(dw1.ID(), consumer1)
	_ = sub1.AddLockers(LockerEntry{Locker: sharedLocker1.Raw()})
	_ = ledgerScanner.AddSubscription(sub1)

	sub2 := NewIndexSubscription(dw2.ID(), consumer2)
	_ = sub2.AddLockers(LockerEntry{Locker: sharedLocker2.Raw()})
	_ = ledgerScanner.AddSubscription(sub2)

	complete, err := ledgerScanner.Scan()
	require.NoError(t, err)
	assert.True(t, complete)

	assert.True(t, consumer1.AllRecordsMatched(), "Received less matching records than expected")
	assert.True(t, consumer2.AllRecordsMatched(), "Received less matching records than expected")

	err = ledgerScanner.RemoveSubscription(dw2.ID())
	require.NoError(t, err)

	lb, err = sharedLocker1.NewDataSetBuilder(env.Ctx, dataset.WithVault(testbase.TestVaultName))
	require.NoError(t, err)

	_, err = lb.AddMetaResource(map[string]any{
		"id":   "test3",
		"type": "TestDataset3",
	})
	require.NoError(t, err)

	f = lb.Submit(expiry.FromNow("1h"))
	err = f.Wait(time.Second * 2)
	require.NoError(t, err)

	rid3 := f.ID()

	consumer1.ExpectedMatches = []map[string]any{
		{
			"userID":        dw1.ID(),
			"rid":           rid3,
			"t":             model.OpTypeLease,
			"lockerID":      sharedLocker1.ID(),
			"participantID": idy1.ID(),
			"acceptedAt":    sharedLocker1.Raw().AcceptedAtBlock(),
		},
	}
	consumer1.Reset()

	consumer2.ExpectedMatches = []map[string]any{}
	consumer2.Reset()

	complete, err = ledgerScanner.Scan()
	require.NoError(t, err)
	assert.True(t, complete)

	assert.True(t, consumer1.AllRecordsMatched(), "Received less matching records than expected")
	assert.True(t, consumer2.AllRecordsMatched(), "Received less matching records than expected")
}

type CheckingConsumer struct {
	T               *testing.T
	ExpectedMatches []map[string]any
	Counter         int
}

func (cc *CheckingConsumer) ConsumeBlock(ctx context.Context, indexID string, partyLookup PartyLookup, n BlockNotification) error {
	for _, dsn := range n.Datasets {
		lockerID, participantID, _, acceptedAtBlock := partyLookup(dsn.KeyID)

		log.Debug().Str("LockerID", lockerID).Str("ParticipantID", participantID).Str("id", dsn.RecordID).Msg("RECORD FOUND")

		if cc.Counter >= len(cc.ExpectedMatches) {
			require.Fail(cc.T, "Too many matching records")
		}

		actual := map[string]any{
			"userID":        indexID,
			"rid":           dsn.RecordID,
			"t":             model.OpTypeLease,
			"lockerID":      lockerID,
			"participantID": participantID,
			"acceptedAt":    acceptedAtBlock,
		}

		if !assert.True(cc.T, ld.DeepCompare(cc.ExpectedMatches[cc.Counter], actual, true)) {
			_, _ = os.Stdout.WriteString("==== ACTUAL ====\n")
			b, _ := jsonw.MarshalIndent(actual, "", "  ")
			_, _ = os.Stdout.Write(b)
			_, _ = os.Stdout.WriteString("\n")
			_, _ = os.Stdout.WriteString("==== EXPECTED ====\n")
			b, _ = jsonw.MarshalIndent(cc.ExpectedMatches[cc.Counter], "", "  ")
			_, _ = os.Stdout.Write(b)
			_, _ = os.Stdout.WriteString("\n")
			cc.T.FailNow()
		}

		cc.Counter++

	}

	return nil
}

func (cc *CheckingConsumer) NotifyScanCompleted(block int64) error {
	return nil
}

func (cc *CheckingConsumer) SetSubscription(sub Subscription) {
}

func (cc *CheckingConsumer) AllRecordsMatched() bool {
	return cc.Counter == len(cc.ExpectedMatches)
}

func (cc *CheckingConsumer) Reset() {
	cc.Counter = 0
}
