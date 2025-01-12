package onflow_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/piprate/json-gold/ld"
	. "github.com/piprate/metalocker/ledger/onflow"
	"github.com/piprate/metalocker/ledger/onflow/emulator"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/services/notification"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/splash"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp})
}

func NewTestLedger(t *testing.T, workerCount int) (*Ledger, *splash.Connector, string) {
	t.Helper()

	flowConnector, err := NewInMemoryConnectorEmbedded(true)
	require.NoError(t, err)

	emulator.ConfigureInMemoryEmulator(t, flowConnector, adminAccountName, "1000.0")

	dir, err := os.MkdirTemp(".", "tempdir_")
	require.NoError(t, err)

	dbFilepath := filepath.Join(dir, "db.bolt")

	ns := notification.NewLocalNotificationService(workerCount * 10)

	se, err := NewTemplateEngine(flowConnector)
	require.NoError(t, err)

	nodeAcct := flowConnector.Account(platformAccountName)
	emulator.FundAccountWithFlow(t, se, nodeAcct.Address, "10.0")

	emulator.AddLedgerPoolKeys(t, flowConnector, platformAccountName, 0, workerCount)

	keyIndexes := make([]uint32, workerCount)
	for i := 0; i < workerCount; i++ {
		keyIndexes[i] = uint32(i + 1)
	}

	fl, err := NewLedger(context.Background(), flowConnector, "emulator", nodeAcct, keyIndexes, dbFilepath, ns)
	require.NoError(t, err)

	return fl, flowConnector, dir
}

func TestNewLedger(t *testing.T) {
	flowConnector, err := splash.NewInMemoryTestConnector(".", true)
	require.NoError(t, err)

	emulator.ConfigureInMemoryEmulator(t, flowConnector, adminAccountName, "1000.0")

	dir, err := os.MkdirTemp(".", "tempdir_")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(dir) }()

	nodeAcct := flowConnector.Account(platformAccountName)

	ctx := context.Background()
	lr, err := NewLedger(ctx, flowConnector, "emulator", nodeAcct, []uint32{0}, filepath.Join(dir, "db.bolt"), nil)
	require.NoError(t, err)

	assert.NoError(t, lr.Close())
}

func TestLedger_GetGenesisBlock(t *testing.T) {
	fl, _, dir := NewTestLedger(t, 1)
	defer func() {
		_ = fl.Close()
		_ = os.RemoveAll(dir)
	}()

	ctx := context.Background()
	require.NoError(t, fl.Sync(ctx))

	gb, err := fl.GetGenesisBlock(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, gb)

	ld.PrintDocument("GEN", gb)
}

func TestLedger_SubmitRecord_Single(t *testing.T) {
	fl, _, dir := NewTestLedger(t, 1)
	defer func() {
		_ = fl.Close()
		_ = os.RemoveAll(dir)
	}()

	var rec *model.Record
	require.NoError(t, jsonw.Decode(strings.NewReader(testRecordTemplate), &rec))

	ctx := context.Background()

	err := fl.SubmitRecord(ctx, rec)
	require.NoError(t, err)

	require.NoError(t, fl.Sync(ctx))

	loadedRec, err := fl.GetRecord(ctx, rec.ID)
	require.NoError(t, err)

	assert.Equal(t, model.StatusPublished, loadedRec.Status)
	loadedRec.Status = ""

	assert.Equal(t, rec, loadedRec)
}

func TestLedger_SubmitRecord_Multiple_SequentialTX(t *testing.T) {
	fl, conn, dir := NewTestLedger(t, 1)
	defer func() {
		_ = fl.Close()
		_ = os.RemoveAll(dir)
	}()

	var template *model.Record
	require.NoError(t, jsonw.Decode(strings.NewReader(testRecordTemplate), &template))

	conn.DisableAutoMine()

	recordCount := 10

	var wg sync.WaitGroup
	wg.Add(recordCount)

	for i := 0; i < recordCount; i++ {
		go func() {
			defer wg.Done()

			ctx := context.Background()

			rec := template.Copy()
			rec.ID = fmt.Sprintf("record-%d", i+1)

			err := fl.SubmitRecord(ctx, rec)
			require.NoError(t, err)

		}()
	}

	time.Sleep(time.Second * 1)

	_, res, err := conn.ExecuteAndCommitBlock(t)
	require.NoError(t, err)
	require.Len(t, res, 1)

	conn.EnableAutoMine()

	wg.Wait()

	ctx := context.Background()
	for i := 0; i < recordCount; i++ {
		loadedRec, err := fl.GetRecord(ctx, fmt.Sprintf("record-%d", i+1))
		require.NoError(t, err)
		require.NotNil(t, loadedRec)
		assert.Equal(t, model.StatusPublished, loadedRec.Status)
	}
}

func TestLedger_SubmitRecord_Multiple_OneTX(t *testing.T) {
	workerCount := 10

	fl, conn, dir := NewTestLedger(t, workerCount)
	defer func() {
		_ = fl.Close()
		_ = os.RemoveAll(dir)
	}()

	var template *model.Record
	require.NoError(t, jsonw.Decode(strings.NewReader(testRecordTemplate), &template))

	conn.DisableAutoMine()

	recordCount := workerCount

	var wg sync.WaitGroup
	wg.Add(recordCount)

	for i := 0; i < recordCount; i++ {
		go func() {
			defer wg.Done()

			ctx := context.Background()

			rec := template.Copy()
			rec.ID = fmt.Sprintf("record-%d", i+1)

			err := fl.SubmitRecord(ctx, rec)
			require.NoError(t, err)

		}()
	}

	time.Sleep(time.Second * 1)

	_, res, err := conn.ExecuteAndCommitBlock(t)
	require.NoError(t, err)
	require.Len(t, res, recordCount)

	wg.Wait()

	ctx := context.Background()
	for i := 0; i < recordCount; i++ {
		loadedRec, err := fl.GetRecord(ctx, fmt.Sprintf("record-%d", i+1))
		require.NoError(t, err)
		require.NotNil(t, loadedRec)
		assert.Equal(t, model.StatusPublished, loadedRec.Status)
	}
}

func TestLedger_SubmitRecord_Multiple_GroupedTX(t *testing.T) {
	workerCount := 2
	recordCount := 6

	fl, conn, dir := NewTestLedger(t, workerCount)
	defer func() {
		_ = fl.Close()
		_ = os.RemoveAll(dir)
	}()

	var template *model.Record
	require.NoError(t, jsonw.Decode(strings.NewReader(testRecordTemplate), &template))

	conn.DisableAutoMine()

	var wg sync.WaitGroup
	wg.Add(recordCount)

	for i := 0; i < recordCount; i++ {
		go func() {
			defer wg.Done()

			ctx := context.Background()

			rec := template.Copy()
			rec.ID = fmt.Sprintf("record-%d", i+1)

			err := fl.SubmitRecord(ctx, rec)
			require.NoError(t, err)

		}()
	}

	recordsSubmitted := 0

	for recordsSubmitted < recordCount {
		time.Sleep(time.Second * 1)

		_, res, err := conn.ExecuteAndCommitBlock(t)
		require.NoError(t, err)
		require.Len(t, res, workerCount)

		recordsSubmitted += workerCount
	}

	wg.Wait()

	ctx := context.Background()
	for i := 0; i < recordCount; i++ {
		loadedRec, err := fl.GetRecord(ctx, fmt.Sprintf("record-%d", i+1))
		require.NoError(t, err)
		require.NotNil(t, loadedRec)
		assert.Equal(t, model.StatusPublished, loadedRec.Status)
	}
}

func TestLedger_GetBlock(t *testing.T) {
	fl, _, dir := NewTestLedger(t, 1)
	defer func() {
		_ = fl.Close()
		_ = os.RemoveAll(dir)
	}()

	var rec *model.Record
	require.NoError(t, jsonw.Decode(strings.NewReader(testRecordTemplate), &rec))

	ctx := context.Background()

	b, err := fl.GetBlock(ctx, 0)
	require.NoError(t, err)
	require.Equal(t, model.BlockStatusConfirmed, b.Status)

	ld.PrintDocument("B0", b)

	b, err = fl.GetBlock(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, model.BlockStatusOpen, b.Status)

	err = fl.SubmitRecord(ctx, rec)
	require.NoError(t, err)

	b, err = fl.GetBlock(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, model.BlockStatusConfirmed, b.Status)

	b, err = fl.GetBlock(ctx, 2)
	require.NoError(t, err)
	require.Equal(t, model.BlockStatusOpen, b.Status)

	_, err = fl.GetBlock(ctx, 3)
	require.Error(t, err)
}

func TestLedger_GetBlockRecords(t *testing.T) {
	workerCount := 3
	recordCount := 4

	fl, conn, dir := NewTestLedger(t, workerCount)
	defer func() {
		_ = fl.Close()
		_ = os.RemoveAll(dir)
	}()

	var template *model.Record
	require.NoError(t, jsonw.Decode(strings.NewReader(testRecordTemplate), &template))

	conn.DisableAutoMine()

	var wg sync.WaitGroup
	wg.Add(recordCount)

	for i := 0; i < recordCount; i++ {
		go func() {
			defer wg.Done()

			ctx := context.Background()

			rec := template.Copy()
			rec.ID = fmt.Sprintf("record-%d", i+1)

			err := fl.SubmitRecord(ctx, rec)
			require.NoError(t, err)

		}()
	}

	recordsSubmitted := 0

	for recordsSubmitted < recordCount {
		time.Sleep(time.Second * 1)

		_, res, err := conn.ExecuteAndCommitBlock(t)
		require.NoError(t, err)

		recordsSubmitted += len(res)
	}

	wg.Wait()

	ctx := context.Background()

	//require.NoError(t, fl.Sync(ctx))

	blockRecords, err := fl.GetBlockRecords(ctx, 1)
	require.NoError(t, err)

	ld.PrintDocument("BLOCK", blockRecords)
}

func TestLedger_GetRecordState(t *testing.T) {
	fl, _, dir := NewTestLedger(t, 1)
	defer func() {
		_ = fl.Close()
		_ = os.RemoveAll(dir)
	}()

	var rec *model.Record
	require.NoError(t, jsonw.Decode(strings.NewReader(testRecordTemplate), &rec))

	ctx := context.Background()

	err := fl.SubmitRecord(ctx, rec)
	require.NoError(t, err)

	require.NoError(t, fl.Sync(ctx))

	state, err := fl.GetRecordState(ctx, rec.ID)
	require.NoError(t, err)

	assert.Equal(t, model.StatusPublished, state.Status)
	assert.True(t, state.BlockNumber > 0)
}

func TestLedger_GetAssetHead(t *testing.T) {

}
