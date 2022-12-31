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

	"github.com/piprate/metalocker/contexts"
	. "github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/piprate/metalocker/utils/fingerprint"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ldDebugMode = false
)

func init() {
	_ = contexts.PreloadContextsIntoMemory()
	SetDebugMode(ldDebugMode)
}

func TestGenerateNewSemanticAsset(t *testing.T) {
	a, err := GenerateNewSemanticAsset(true, false, "", "")
	require.NoError(t, err)
	assert.Equal(t, "Asset", a.Type)
	assert.Equal(t, true, a.Serial)
	assert.Equal(t, false, a.IsIdentity)
}

func TestGenerateValueAsset(t *testing.T) {
	a, err := GenerateValueAsset("did:piprate:CEVWF3yAURaH12EmT3zZ6p6KrSKeGV3Mw9JFXfaXotKH", map[string]any{
		"z": "did:piprate:AfMyiFjarZMexhYZzfxEmhbuWK2AejYhzY2mAnU3eJzU",
	}, map[string]any{
		"x": "a",
		"y": 120,
	}, "")
	require.NoError(t, err)

	expectedAsset := `
{
  "id": "did:piprate:BqSoFAnFWM1wkvmUAfeJXjDnYdWjbG21iZqi3Sshu4qz",
  "type": [
    "ValueAsset",
    "Entity"
  ],
  "wasGeneratedBy": {
    "algorithm": "did:piprate:CEVWF3yAURaH12EmT3zZ6p6KrSKeGV3Mw9JFXfaXotKH",
    "qualifiedUsage": [
      {
        "entity": {
          "id": "did:piprate:AfMyiFjarZMexhYZzfxEmhbuWK2AejYhzY2mAnU3eJzU",
          "type": "Entity"
        },
        "hadRole": {
          "label": "z",
          "type": "Role"
        },
        "type": "Usage"
      },
      {
        "entity": {
          "type": "Entity",
          "value": "a"
        },
        "hadRole": {
          "label": "x",
          "type": "Role"
        },
        "type": "Usage"
      },
      {
        "entity": {
          "type": "Entity",
          "value": 120
        },
        "hadRole": {
          "label": "y",
          "type": "Role"
        },
        "type": "Usage"
      }
    ],
    "type": "Activity"
  }
}
`

	testbase.AssertEqualUnorderedJSON(t, expectedAsset, a)
}

func TestGenerateNewSemanticDigitalAsset(t *testing.T) {
	a, err := GenerateNewSemanticDigitalAsset([]byte("test string"), fingerprint.AlgoSha256, "")
	require.NoError(t, err)
	assert.Equal(t, "Asset", a.Type)
	assert.Equal(t, false, a.Serial)
	assert.Equal(t, false, a.IsIdentity)
	assert.Equal(t, true, a.IsDigital)
	assert.Equal(t, "FMoKikNGjhRqLJmuZre1uP7rHWqQpTWHCd21vnPmcCyQ", a.Fingerprint)
	assert.Equal(t, "fingerprints:sha256", a.FingerprintAlgorithm)
	assert.Equal(t, "did:piprate:4HKDX4TmoV2Ynp5JMSWKSWHHWBjMkyZg99jjBqLcGr2U", a.ID)
}

func BenchmarkGenerateNewDigitalAsset(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = GenerateNewSemanticDigitalAsset([]byte("test string"), fingerprint.AlgoSha256, "example")
	}
}

func TestSemanticAsset_MerkleSetID(t *testing.T) {
	var a SemanticAsset
	err := jsonw.Unmarshal([]byte(
		`
{
	"type": "Asset",
    "isSerial": false,
	"id": "something",
	"nonce": "5kh9Dsn1Lkik25PCASVXRXsDzFPLVe6pJeg9F2kwauuw"
}
	`), &a)
	require.NoError(t, err)
	err = a.MerkleSetID("")
	require.NoError(t, err)
	assert.Equal(t, "did:piprate:Dkse2MMM23g1cpGBK5QkvnmjG4j5MpsEbBgMGg3TYLwf", a.ID)
}

func TestSemanticAsset_MerkleSetID_CustomDIDMethod(t *testing.T) {
	var a SemanticAsset
	err := jsonw.Unmarshal([]byte(
		`
{
	"type": "Asset",
    "isSerial": false,
	"id": "something",
	"nonce": "5kh9Dsn1Lkik25PCASVXRXsDzFPLVe6pJeg9F2kwauuw"
}
	`), &a)
	require.NoError(t, err)
	err = a.MerkleSetID("test")
	require.NoError(t, err)
	assert.Equal(t, "did:test:Dkse2MMM23g1cpGBK5QkvnmjG4j5MpsEbBgMGg3TYLwf", a.ID)
}

func TestSemanticAsset_MerkleVerify(t *testing.T) {
	var a SemanticAsset
	err := jsonw.Unmarshal([]byte(
		`
{
	"id": "did:piprate:Dkse2MMM23g1cpGBK5QkvnmjG4j5MpsEbBgMGg3TYLwf",
	"nonce": "5kh9Dsn1Lkik25PCASVXRXsDzFPLVe6pJeg9F2kwauuw",
	"type": "Asset",
	"isSerial": false
}
	`), &a)
	require.NoError(t, err)
	ver, err := a.MerkleVerify()
	require.NoError(t, err)
	assert.True(t, ver)
}

func TestVerifySemanticDigitalAssetID(t *testing.T) {
	data := []byte("test string")

	// default DID method

	a, err := GenerateNewSemanticDigitalAsset(data, fingerprint.AlgoSha256, "")
	require.NoError(t, err)

	result, err := VerifySemanticDigitalAssetID(a.ID, fingerprint.AlgoSha256, data)
	assert.NoError(t, err)
	assert.True(t, result)

	// custom DID method

	a, err = GenerateNewSemanticDigitalAsset(data, fingerprint.AlgoSha256, "example")
	require.NoError(t, err)

	result, err = VerifySemanticDigitalAssetID(a.ID, fingerprint.AlgoSha256, data)
	assert.NoError(t, err)
	assert.True(t, result)

	// bad ID

	result, err = VerifySemanticDigitalAssetID("did:example:123", fingerprint.AlgoSha256, data)
	assert.NoError(t, err)
	assert.False(t, result)
}
