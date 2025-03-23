package onflow_test

import (
	"encoding/base64"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/piprate/metalocker/ledger/onflow"
	"github.com/piprate/metalocker/ledger/onflow/emulator"
	"github.com/piprate/metalocker/model"
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

const testRecordTemplate = `
{
 "id": "47Gv8nkNgEr95gRZd98c4NAhi4wszmUBk6tytk4TtVCo",
 "routingKey": "db5ryf4F6JQjD3aRqMF13RdueHkeHU9WcAx9dEVfy1mq",
 "keyIndex": 4,
 "operationType": 1,
 "address": "CW1bDDn1Pp3W486rVxkrHKLUTirqytPMjBfFfXrqihU8",
 "ac": "5z//FjtYNAChMv+8dpMqIiM4uKkGv0BXWx2yu6avIgE=",
 "rc": "pVbAvIkXtnzmF7z3U6n4GVQwgjjhQgvw/bomh1Fm72o=",
 "rcType": 1,
 "dataAssets": [
   "Gp7ZX5S5NECqoMz3u7micC7TuhvWB9z8ErAkympVgt4A"
 ],
 "signature": "AN1rKvtoGVPioTYojunoBTWWx6MTdb8VHB8vN8takPu4fpe5K7LFowgqW8meuLwniFosM2xCBbHkdhqUoTovr4xDCzKF2JNZb"
}
`

func TestMetaLocker_submitRecord_SequentialRecords(t *testing.T) {
	conn, err := splash.NewInMemoryTestConnector(".", true)
	require.NoError(t, err)

	emulator.ConfigureInMemoryEmulator(t, conn, adminAccountName, "10.0")

	se, err := NewTemplateEngine(conn)
	require.NoError(t, err)

	var rec *model.Record
	require.NoError(t, jsonw.Decode(strings.NewReader(testRecordTemplate), &rec))

	nodeAcct := conn.Account(platformAccountName)

	emulator.FundAccountWithFlow(t, se, nodeAcct.Address, "10.0")

	_ = se.NewTransaction("metalocker_submit_record").
		SignProposeAndPayAs(nodeAcct.Name).
		Argument(RecordToCadence(rec, se.ContractAddress("MetaLocker"))).
		Test(t).
		RequireSuccess().
		PrintEvents().
		AssertEventCount(8).
		AssertPartialEvent(splash.NewTestEvent("A.179b6b1cb6755e31.MetaLocker.BlockOpened", map[string]interface{}{
			"chain": "0x179b6b1cb6755e31",
			"id":    "2",
		})).
		AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.MetaLocker.RecordPublished", map[string]interface{}{
			"blockID": "2",
			"id":      rec.ID,
		}))

	res := GetRecord(t, se, rec.ID)

	assert.Equal(t, model.StatusPublished, res.Status)

	AssertDataAssetCounter(t, se, rec.DataAssets[0], 1)
	AssertDataAssetCounter(t, se, rec.OperationAddress, 1)
	AssertNoDataAssetCounter(t, se, "non-existent-asset-id")

	rec2 := rec.Copy()
	rec2.ID = "second-record"

	_ = se.NewTransaction("metalocker_submit_record").
		SignProposeAndPayAs(nodeAcct.Name).
		Argument(RecordToCadence(rec2, se.ContractAddress("MetaLocker"))).
		Test(t).
		RequireSuccess().
		AssertEventCount(8).
		PrintEvents().
		AssertPartialEvent(splash.NewTestEvent("A.179b6b1cb6755e31.MetaLocker.BlockOpened", map[string]interface{}{
			"chain": "0x179b6b1cb6755e31",
			"id":    "3",
		})).
		AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.MetaLocker.RecordPublished", map[string]interface{}{
			"blockID": "3",
			"id":      "second-record",
		}))

	AssertDataAssetCounter(t, se, rec.DataAssets[0], 2)
	AssertDataAssetCounter(t, se, rec.OperationAddress, 2)
}

func TestMetaLocker_submitRecord_CollatedRecords(t *testing.T) {
	conn, err := splash.NewInMemoryTestConnector(".", true)
	require.NoError(t, err)

	emulator.ConfigureInMemoryEmulator(t, conn, adminAccountName, "10.0")

	se, err := NewTemplateEngine(conn)
	require.NoError(t, err)

	var rec *model.Record
	require.NoError(t, jsonw.Decode(strings.NewReader(testRecordTemplate), &rec))

	nodeAcct := conn.Account(platformAccountName)
	adminAcct := conn.Account(adminAccountName)

	emulator.FundAccountWithFlow(t, se, nodeAcct.Address, "10.0")

	conn.DisableAutoMine()

	_ = se.NewTransaction("metalocker_submit_record").
		SignProposeAndPayAs(nodeAcct.Name).
		Argument(RecordToCadence(rec, se.ContractAddress("MetaLocker"))).
		Test(t).
		RequireSuccess().
		AssertEventCount(0)

	rec2 := rec.Copy()
	rec2.ID = "second-record"

	_ = se.NewTransaction("metalocker_submit_record").
		SignProposeAndPayAs(adminAcct.Name).
		Argument(RecordToCadence(rec2, se.ContractAddress("MetaLocker"))).
		Test(t).
		RequireSuccess().
		AssertEventCount(0)

	_, resList, err := conn.ExecuteAndCommitBlock(t)
	require.NoError(t, err)
	require.Len(t, resList, 2)

	resList[0].RequireSuccess().
		AssertEventCount(8).
		AssertPartialEvent(splash.NewTestEvent("A.179b6b1cb6755e31.MetaLocker.BlockOpened", map[string]interface{}{
			"chain": "0x179b6b1cb6755e31",
			"id":    "2",
		})).
		AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.MetaLocker.RecordPublished", map[string]interface{}{
			"blockID": "2",
			"id":      rec.ID,
		}))

	resList[1].RequireSuccess().
		PrintEvents().
		AssertEventCount(6).
		AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.MetaLocker.RecordPublished", map[string]interface{}{
			"blockID": "2",
			"id":      "second-record",
		}))

	res := GetRecord(t, se, rec.ID)
	assert.Equal(t, model.StatusPublished, res.Status)

	AssertDataAssetCounter(t, se, rec.DataAssets[0], 2)
	AssertDataAssetCounter(t, se, rec.OperationAddress, 2)
}

func TestMetaLocker_submitRecord_LeaseRevocation(t *testing.T) {
	conn, err := splash.NewInMemoryTestConnector(".", true)
	require.NoError(t, err)

	emulator.ConfigureInMemoryEmulator(t, conn, adminAccountName, "10.0")

	se, err := NewTemplateEngine(conn)
	require.NoError(t, err)

	privateHDKey := NewExtendedKey(t)
	lease := CreateSampleLease(t)
	leaseAddress := "CW1bDDn1Pp3W486rVxkrHKLUTirqytPMjBfFfXrqihU8"
	keyIndex := model.RandomKeyIndex()
	recPrivKey, err := privateHDKey.Derive(keyIndex)
	require.NoError(t, err)
	rec, err := model.BuildLeaseRecord(keyIndex, recPrivKey, lease, leaseAddress, false)
	require.NoError(t, err)

	nodeAcct := conn.Account(platformAccountName)

	emulator.FundAccountWithFlow(t, se, nodeAcct.Address, "10.0")

	_ = se.NewTransaction("metalocker_submit_record").
		SignProposeAndPayAs(nodeAcct.Name).
		Argument(RecordToCadence(rec, se.ContractAddress("MetaLocker"))).
		Test(t).
		RequireSuccess()

	res := GetRecord(t, se, rec.ID)

	assert.Equal(t, model.StatusPublished, res.Status)

	pk, err := privateHDKey.ECPrivKey()
	require.NoError(t, err)

	routingKey, _ := model.BuildRoutingKey(pk.PubKey())

	subjPrivKey, err := privateHDKey.Derive(rec.KeyIndex)
	require.NoError(t, err)

	goodACInput := model.BuildAuthorisingCommitmentInput(subjPrivKey, rec.OperationAddress)
	badACInput := model.BuildAuthorisingCommitmentInput(subjPrivKey, "another-op-address")

	revocationRec := &model.Record{
		RoutingKey:    routingKey,
		KeyIndex:      12345,
		Operation:     model.OpTypeLeaseRevocation,
		SubjectRecord: rec.ID,
		RevocationProof: []string{
			base64.StdEncoding.EncodeToString(badACInput),
		},
	}
	require.NoError(t, revocationRec.Seal(pk))
	require.NoError(t, revocationRec.Validate())

	_ = se.NewTransaction("metalocker_submit_record").
		SignProposeAndPayAs(nodeAcct.Name).
		Argument(RecordToCadence(revocationRec, se.ContractAddress("MetaLocker"))).
		Test(t).
		AssertFailure("invalid revocation proof")

	revocationRec.RevocationProof[0] = base64.StdEncoding.EncodeToString(goodACInput)

	_ = se.NewTransaction("metalocker_submit_record").
		SignProposeAndPayAs(nodeAcct.Name).
		Argument(RecordToCadence(revocationRec, se.ContractAddress("MetaLocker"))).
		Test(t).
		RequireSuccess().
		AssertEventCount(9).
		PrintEvents().
		AssertPartialEvent(splash.NewTestEvent("A.179b6b1cb6755e31.MetaLocker.BlockOpened", map[string]interface{}{
			"chain": "0x179b6b1cb6755e31",
			"id":    "3",
		})).
		AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.MetaLocker.RecordPublished", map[string]interface{}{
			"blockID": "3",
			"id":      revocationRec.ID,
		})).
		AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.MetaLocker.RecordRevoked", map[string]interface{}{
			"blockID": "3",
			"id":      rec.ID,
		}))

	res = GetRecord(t, se, rec.ID)
	assert.Equal(t, model.StatusRevoked, res.Status)

	AssertDataAssetCounter(t, se, rec.DataAssets[0], 0)
	AssertDataAssetCounter(t, se, rec.OperationAddress, 0)
}

func TestMetaLocker_submitRecord_AssetHead(t *testing.T) {
	conn, err := splash.NewInMemoryTestConnector(".", true)
	require.NoError(t, err)

	emulator.ConfigureInMemoryEmulator(t, conn, adminAccountName, "10.0")

	se, err := NewTemplateEngine(conn)
	require.NoError(t, err)

	privateHDKey := NewExtendedKey(t)
	lease := CreateSampleLease(t)
	leaseAddress := "CW1bDDn1Pp3W486rVxkrHKLUTirqytPMjBfFfXrqihU8"

	keyIndex := model.RandomKeyIndex()
	recPrivKey, err := privateHDKey.Derive(keyIndex)
	require.NoError(t, err)

	rec, err := model.BuildLeaseRecord(keyIndex, recPrivKey, lease, leaseAddress, false)
	require.NoError(t, err)

	nodeAcct := conn.Account(platformAccountName)

	emulator.FundAccountWithFlow(t, se, nodeAcct.Address, "10.0")

	_ = se.NewTransaction("metalocker_submit_record").
		SignProposeAndPayAs(nodeAcct.Name).
		Argument(RecordToCadence(rec, se.ContractAddress("MetaLocker"))).
		Test(t).
		RequireSuccess()

	res := GetRecord(t, se, rec.ID)

	assert.Equal(t, model.StatusPublished, res.Status)

	pk, err := privateHDKey.ECPrivKey()
	require.NoError(t, err)

	//recPrivKey, err := privateHDKey.Derive(rec.KeyIndex)
	//require.NoError(t, err)

	headName := "testHead"
	senderID := "did:piprate:sender-id"
	lockerID := "testLockerID"
	assetID := "did:piprate:FdywEGiR5nubMFZgbfneUdAx5V858eaiRcBGRGVpj3Y6"
	sharedSecret := "NYGupZJZfxotkGaLGKpuqOX8K7xVvE7qEDvNgfru4e8=" //nolint:gosec

	assetHeadRec1, err := model.BuildAssetHeadRecord(senderID, privateHDKey, lockerID, sharedSecret, assetID, headName, rec.ID, nil)
	require.NoError(t, err)

	assert.Empty(t, GetAssetHead(t, se, assetHeadRec1.HeadID))

	// try updating a head that doesn't exist yet by setting SubjectRecord. Should fail.

	assetHeadRec1.SubjectRecord = "wrong-previous-head-record-id"
	assetHeadRec1.RevocationProof = []string{"dummy-proof"}
	require.NoError(t, assetHeadRec1.Seal(pk))
	require.NoError(t, assetHeadRec1.Validate())

	_ = se.NewTransaction("metalocker_submit_record").
		SignProposeAndPayAs(nodeAcct.Name).
		Argument(RecordToCadence(assetHeadRec1, se.ContractAddress("MetaLocker"))).
		Test(t).
		AssertFailure("invalid reference to previous head")

	// set a new asset head. Should succeed.

	assetHeadRec1.SubjectRecord = ""
	assetHeadRec1.RevocationProof = nil
	require.NoError(t, assetHeadRec1.Seal(pk))
	require.NoError(t, assetHeadRec1.Validate())

	_ = se.NewTransaction("metalocker_submit_record").
		SignProposeAndPayAs(nodeAcct.Name).
		Argument(RecordToCadence(assetHeadRec1, se.ContractAddress("MetaLocker"))).
		Test(t).
		RequireSuccess().
		AssertEventCount(8).
		//PrintEvents().
		AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.MetaLocker.RecordPublished", map[string]interface{}{
			"blockID": "3",
			"id":      assetHeadRec1.ID,
		}))

	headRec1 := GetAssetHead(t, se, assetHeadRec1.HeadID)
	assert.NotEmpty(t, headRec1)

	// build a record to update the asset head

	assetHeadRec2, err := model.BuildAssetHeadRecord(senderID, privateHDKey, lockerID, sharedSecret, assetID, headName, rec.ID, assetHeadRec1)
	require.NoError(t, err)

	// try updating the head with invalid proof. Should fail.

	goodProof := assetHeadRec2.RevocationProof[0]

	badACInput := model.BuildAuthorisingCommitmentInput(recPrivKey, "another-op-address")
	assetHeadRec2.RevocationProof[0] = base64.StdEncoding.EncodeToString(badACInput)
	require.NoError(t, assetHeadRec2.Seal(pk))
	require.NoError(t, assetHeadRec2.Validate())

	_ = se.NewTransaction("metalocker_submit_record").
		SignProposeAndPayAs(nodeAcct.Name).
		Argument(RecordToCadence(assetHeadRec2, se.ContractAddress("MetaLocker"))).
		Test(t).
		AssertFailure("asset head revocation failed: invalid proof")

	// update the head with valid proof

	assetHeadRec2.RevocationProof[0] = goodProof
	require.NoError(t, assetHeadRec2.Seal(pk))
	require.NoError(t, assetHeadRec2.Validate())

	_ = se.NewTransaction("metalocker_submit_record").
		SignProposeAndPayAs(nodeAcct.Name).
		Argument(RecordToCadence(assetHeadRec2, se.ContractAddress("MetaLocker"))).
		Test(t).
		RequireSuccess().
		AssertEventCount(9).
		//PrintEvents().
		AssertPartialEvent(splash.NewTestEvent("A.179b6b1cb6755e31.MetaLocker.BlockOpened", map[string]interface{}{
			"chain": "0x179b6b1cb6755e31",
			"id":    "4",
		})).
		AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.MetaLocker.RecordPublished", map[string]interface{}{
			"blockID": "4",
			"id":      assetHeadRec2.ID,
		})).
		AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.MetaLocker.RecordRevoked", map[string]interface{}{
			"blockID": "4",
			"id":      assetHeadRec1.ID,
		}))

	res = GetRecord(t, se, assetHeadRec1.ID)
	assert.Equal(t, model.StatusRevoked, res.Status)

	headRec2 := GetAssetHead(t, se, assetHeadRec2.HeadID)
	assert.NotEmpty(t, headRec2)
	assert.NotEqual(t, headRec1, headRec2)
}
