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

package model_test

import (
	"testing"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/piprate/json-gold/ld"
	. "github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecord_Seal(t *testing.T) {

	extKey, err := hdkeychain.NewKeyFromString("xprv9u3yvfDwuqTbFNWRupD1QXsfh8Toh8NuKGrX5P8pu8CWJ2w915spiQFZS4SThkHWwV5yu1wJMsmrhYPEytb5bZZ4Mdut9txywRTa5c1TAzC")
	require.NoError(t, err)
	pk, err := extKey.ECPrivKey()
	require.NoError(t, err)

	var rec Record
	err = jsonw.Unmarshal([]byte(
		`
{
	"routingKey": "261s6AJdF5QWTqinNdCo3xqBzMmBzRSwFT6caPXKwgAjR",
    "keyIndex": 4,
	"operationType": 1,
	"address": "DrnjUqnLW2gsU7R3fVtW65fuHP18tDFupfZPQfQ2GZrA",
	"ac": "7BQRPBcvUgCeiFkxK7SE5F1fjKixsDcfYwyQa8mcQpqK",
	"acType": 0,
	"rc": "5cinaxTkPgjwMnEF1eGWE8VN2BMcud2tAxMvwimT4XsE",
	"rcType": 0,
	"ic": "",
	"icType": 0
}
	`), &rec)
	require.NoError(t, err)

	err = rec.Seal(pk)
	require.NoError(t, err)

	assert.Equal(t, "5a2WLDE9WfArGkayJ18qy5iK4TBr6onUDUCioiyQyYyY", rec.ID)
	assert.Equal(t, "AN1rKvt7GdPDDakTqstPMuj1cFRUXfLskr566MwshaDxzhyxfDLr66zmPtrmizN99nbqhJTrSkyAdg2swjLwr5Tzx9Kf12N1X", rec.Signature)
}

func TestRecord_Seal_WithFlags(t *testing.T) {

	extKey, err := hdkeychain.NewKeyFromString("xprv9u3yvfDwuqTbFNWRupD1QXsfh8Toh8NuKGrX5P8pu8CWJ2w915spiQFZS4SThkHWwV5yu1wJMsmrhYPEytb5bZZ4Mdut9txywRTa5c1TAzC")
	require.NoError(t, err)
	pk, err := extKey.ECPrivKey()
	require.NoError(t, err)

	var rec Record
	err = jsonw.Unmarshal([]byte(
		`
{
	"routingKey": "261s6AJdF5QWTqinNdCo3xqBzMmBzRSwFT6caPXKwgAjR",
    "keyIndex": 4,
	"operationType": 1,
	"flags": 1,
	"address": "DrnjUqnLW2gsU7R3fVtW65fuHP18tDFupfZPQfQ2GZrA",
	"ac": "7BQRPBcvUgCeiFkxK7SE5F1fjKixsDcfYwyQa8mcQpqK",
	"acType": 0,
	"rc": "5cinaxTkPgjwMnEF1eGWE8VN2BMcud2tAxMvwimT4XsE",
	"rcType": 0,
	"ic": "",
	"icType": 0
}
	`), &rec)
	require.NoError(t, err)

	err = rec.Seal(pk)
	require.NoError(t, err)

	assert.Equal(t, "Azo4T2Y11hKyY6ZaBEP3sk5EZQmaS1GS9vXpDDLunwfy", rec.ID)
	assert.Equal(t, "381yXZK43jv691RDJw9Xtj3yce1HJWLw1zCLYFzh2p3kmUoDgSN7uBQGYuHXXcSA7e9SVwXuwjFPN8AV2KhFQ2cLUexZbHYH", rec.Signature)
}

func TestRecord_Verify(t *testing.T) {
	extKey, err := hdkeychain.NewKeyFromString("xprv9u3yvfDwuqTbFNWRupD1QXsfh8Toh8NuKGrX5P8pu8CWJ2w915spiQFZS4SThkHWwV5yu1wJMsmrhYPEytb5bZZ4Mdut9txywRTa5c1TAzC")
	require.NoError(t, err)
	key, err := extKey.ECPubKey()
	require.NoError(t, err)

	var rec Record
	err = jsonw.Unmarshal([]byte(
		`
{
	"id": "5a2WLDE9WfArGkayJ18qy5iK4TBr6onUDUCioiyQyYyY",
	"routingKey": "261s6AJdF5QWTqinNdCo3xqBzMmBzRSwFT6caPXKwgAjR",
    "keyIndex": 4,
	"operationType": 1,
	"address": "DrnjUqnLW2gsU7R3fVtW65fuHP18tDFupfZPQfQ2GZrA",
	"ac": "7BQRPBcvUgCeiFkxK7SE5F1fjKixsDcfYwyQa8mcQpqK",
	"acType": 0,
	"rc": "5cinaxTkPgjwMnEF1eGWE8VN2BMcud2tAxMvwimT4XsE",
	"rcType": 0,
	"ic": "",
	"icType": 0,
	"signature": "AN1rKvt7GdPDDakTqstPMuj1cFRUXfLskr566MwshaDxzhyxfDLr66zmPtrmizN99nbqhJTrSkyAdg2swjLwr5Tzx9Kf12N1X"
}
	`), &rec)
	require.NoError(t, err)

	valid, err := rec.Verify(key)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestRecord_Copy(t *testing.T) {
	var rec Record
	err := jsonw.Unmarshal([]byte(
		`
{
	"id": "3nK9xMuk1oKDuzXUizEGeZMhyzXSYduushH37w1yi4V9",
	"routingKey": "261s6AJdF5QWTqinNdCo3xqBzMmBzRSwFT6caPXKwgAjR",
    "keyIndex": 4,
	"operationType": 1,
	"address": "DrnjUqnLW2gsU7R3fVtW65fuHP18tDFupfZPQfQ2GZrA",
	"ac": "7BQRPBcvUgCeiFkxK7SE5F1fjKixsDcfYwyQa8mcQpqK",
	"acType": 0,
	"rc": "5cinaxTkPgjwMnEF1eGWE8VN2BMcud2tAxMvwimT4XsE",
	"rcType": 0,
	"ic": "",
	"icType": 0,
	"cc": "",
	"ccType": 0,
	"subjectRecord": "89nZtFNmUDbWPHB7ms3edTxsZtXUibYVMVQpZcAE2jSY",
	"revocationProof": [
		"wetscNY/fVsh3D8lUw33yJz0EkmJKWsOhHIyfj3ywCk="
	],
	"signature": "AN1rKvt7fQJLS7aRRDDpMBfKhS8m8qmfQdU1NsK47cSYTu3oiv4hHVzdqXtevyPinkrJmDwrhgxnXUNMj24bA8RUy3JSF7XXu"
}
	`), &rec)
	require.NoError(t, err)

	recCopy := rec.Copy()
	assert.True(t, ld.DeepCompare(&rec, recCopy, true))

	rec.KeyIndex = 123
	assert.NotEqual(t, rec.KeyIndex, recCopy.KeyIndex)
}

func TestRandomKeyIndex(t *testing.T) {
	idx := RandomKeyIndex()
	assert.True(t, idx < uint32(0x80000000))
}
