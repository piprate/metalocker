package onflow_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	. "github.com/piprate/metalocker/ledger/onflow"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/splash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	adminAccountName    = "emulator-metalocker-admin"
	platformAccountName = "emulator-metalocker-platform"
	user1AccountName    = "emulator-user1"
	user2AccountName    = "emulator-user2"
	user3AccountName    = "emulator-user3"

	sampleLease1 = `{
  "datasetType": "graph",
  "expire": "1970-01-01T00:16:40Z",
  "id": "Gm8tPkhqFzDoXNLTaCEfcHxdSKUmaWwRnQfy8V38NUab",
  "impression": {
    "@context": "https://piprate.org/context/piprate.jsonld",
    "asset": "did:piprate:FdywEGiR5nubMFZgbfneUdAx5V858eaiRcBGRGVpj3Y6",
    "generatedAtTime": "1970-01-01T00:16:40Z",
    "id": "FUWbvfpTZbeG8W9w9dCd8R2RxXsx53EyYrGZJe7Qrqi6",
    "proof": {
      "creator": "did:piprate:GmNuLkqi3NY9h5PbeLqHBq",
      "proofValue": "62R8ogF69UcgPyv11ABbcGW7G3zbbNdMtpoYYz5MkDLbMFGfdY6t4TAREdjNU5QHFT6TCeKB6eHSmJJME9jb6nPA",
      "type": "Ed25519Signature2018"
    },
    "resource": {
      "contentType": "File",
      "fingerprint": "3zNymhmMHiZqCD9Nvrg8rI9QZ05cCb7h01KOtCMEH9c=",
      "fingerprintAlgorithm": "fingerprints:sha256",
      "id": "did:piprate:7PDaLZGd459vFXwhKRv8btMppNjKtq7vKkUodF48csAD"
    },
    "type": [
      "Impression",
      "Entity",
      "Bundle"
    ],
    "wasAttributedTo": "did:piprate:GmNuLkqi3NY9h5PbeLqHBq"
  },
  "storage": [
    {
      "asset": "did:piprate:7PDaLZGd459vFXwhKRv8btMppNjKtq7vKkUodF48csAD",
      "id": "did:piprate:7PDaLZGd459vFXwhKRv8btMppNjKtq7vKkUodF48csAD",
      "method": "memory",
      "mimeType": "application/json",
      "size": 301,
      "type": "Resource",
      "vault": "Z2kcCarCE47SDjtWD5ruyijsQyWMF5B1jjk6HHWngoe"
    }
  ],
  "type": "Lease"
}`
)

func NewExtendedKey(t *testing.T) *hdkeychain.ExtendedKey {
	t.Helper()

	seed, err := hdkeychain.GenerateSeed(hdkeychain.RecommendedSeedLen)
	require.NoError(t, err)
	privHD, _, err := model.GenerateNewHDKey(seed)
	require.NoError(t, err)

	return privHD
}

func CreateSampleLease(t *testing.T) *model.Lease {
	t.Helper()

	var lease *model.Lease
	require.NoError(t, jsonw.Decode(strings.NewReader(sampleLease1), &lease))
	return lease
}

func GetFlowBalance(t *testing.T, se *splash.TemplateEngine, address flow.Address) float64 {
	t.Helper()

	v, err := se.NewScript("account_balance_flow").
		Argument(cadence.NewAddress(address)).
		RunReturns(context.Background())
	require.NoError(t, err)

	return splash.ToFloat64(v)
}

func GetRecord(t *testing.T, se *splash.TemplateEngine, id string) *model.Record {
	t.Helper()

	val, err := se.NewScript("metalocker_get_record").
		Argument(cadence.String(id)).
		RunReturns(context.Background())
	require.NoError(t, err)

	res, err := RecordFromCadence(val)
	require.NoError(t, err)

	return res
}

func AssertDataAssetCounter(t *testing.T, se *splash.TemplateEngine, id string, counter uint64) {
	t.Helper()

	val, err := se.NewScript("metalocker_get_data_asset_counter").
		Argument(cadence.String(id)).
		RunReturns(context.Background())
	require.NoError(t, err)

	opt, _ := val.(cadence.Optional)
	if opt.Value == nil {
		assert.Fail(t, fmt.Sprintf("data asset counter is nil"))
	} else {
		assert.Equal(t, counter, uint64(opt.Value.(cadence.UInt64)))
	}
}

func GetAssetHead(t *testing.T, se *splash.TemplateEngine, headID string) *model.Record {
	t.Helper()

	val, err := se.NewScript("metalocker_get_asset_head").
		Argument(cadence.String(headID)).
		RunReturns(context.Background())
	require.NoError(t, err)

	opt, _ := val.(cadence.Optional)
	if opt.Value == nil {
		return nil
	} else {
		rec, err := RecordFromCadence(opt.Value)
		require.NoError(t, err)

		return rec
	}
}
