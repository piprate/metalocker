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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignableDocument_Context(t *testing.T) {
	sd, err := NewSignableDocument([]byte("{}"))
	require.NoError(t, err)

	ctx := sd.Context()
	assert.Nil(t, ctx)

	sd.SetContext(map[string]any{"test": "it works"})

	ctxMap, isMap := sd.Context().(map[string]any)
	assert.True(t, isMap)
	assert.Equal(t, "it works", ctxMap["test"].(string))
}

func TestSignableDocument_MerkleSetId(t *testing.T) {
	_ = contexts.PreloadContextsIntoMemory()

	sd, _ := NewSignableDocument([]byte(`
{
	"@context": "https://piprate.org/context/piprate.jsonld",
	"type": [
        "Impression",
        "Entity",
        "Bundle"
    ],
	"asset": {
		"id": "did:piprate:EznK1nCNk8FLBJPiYp7w9k6LoNgCaZDFsBb8YEedEAoY",
		"isSerial": false,
		"nonce": "5EFP1D8T2NLYuLkdyF7TPhAWLXG7jtKiNK3qo2ZgA6Xn",
		"type": "Asset"
	},
	"wasAttributedTo": "did:piprate:7ep2P1LqcQshFaJuai3B1d",
	"generatedAtTime": "2018-03-19T01:43:47.652569Z",
	"contentType": "file",
	"fingerprint": "GdswP3Pwfn925ymQUGydGvxFnRWT98s1FaXUoi9R7hZQ",
	"fingerprintAlgorithm": "fingerprints:sha256"
}
`))
	id, err := sd.MerkleSetID("")
	require.NoError(t, err)

	assert.Equal(t, "76b6n2TfzteMJi7FagNTEBGULms42K78PcvygpzofTia", id)
}

func TestSignableDocument_Sign(t *testing.T) {
	_ = contexts.PreloadContextsIntoMemory()

	sd, _ := NewSignableDocument([]byte(`
{
	"@context": "https://piprate.org/context/piprate.jsonld",
	"id": "did:piprate:EznK1nCNk8FLBJPiYp7w9k6LoNgCaZDFsBb8YEedEAoY",
	"isSerial": false,
	"nonce": "5EFP1D8T2NLYuLkdyF7TPhAWLXG7jtKiNK3qo2ZgA6Xn",
	"type": "Asset"
}
`))
	did, err := GenerateDID(WithSeed("Test0001"))
	require.NoError(t, err)

	signKeyVal := did.SignKeyValue()
	proof, err := sd.Sign(did.ID, signKeyVal)
	require.NoError(t, err)

	assert.Equal(t, "4kjjwqSZ2QF2MecCRQWnBm9ZzPjVL6Tnxy18wPAveawXNUtV7eG3szwZhLkv2UKdXotrrjjGijQjc2esCw3NhgWc",
		proof.Value)
}
