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
	"strings"
	"testing"

	"github.com/btcsuite/btcd/btcutil/base58"
	. "github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/utils/fingerprint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAssetID(t *testing.T) {
	method := "did:test:"
	id := NewAssetID(method)
	assert.NotEmpty(t, id)
	idBytes := base58.Decode(id[strings.LastIndex(id, ":")+1:])
	assert.Equal(t, 32, len(idBytes))
}

func TestBuildDigitalAssetIDFromReader(t *testing.T) {
	id, err := BuildDigitalAssetIDFromReader(strings.NewReader("test"), fingerprint.AlgoSha256, "")
	require.NoError(t, err)
	assert.NotEmpty(t, id)
}

func BenchmarkBuildDigitalAssetIDWithFingerprint(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = BuildDigitalAssetIDWithFingerprint([]byte("test string"), "example")
	}
}

func TestVerifyDigitalAssetID(t *testing.T) {
	data := []byte("test string")

	// default DID method

	id, err := BuildDigitalAssetID(data, fingerprint.AlgoSha256, "")
	require.NoError(t, err)

	result, err := VerifyDigitalAssetID(id, fingerprint.AlgoSha256, data)
	assert.NoError(t, err)
	assert.True(t, result)

	// custom DID method

	id, err = BuildDigitalAssetID(data, fingerprint.AlgoSha256, "example")
	require.NoError(t, err)

	result, err = VerifyDigitalAssetID(id, fingerprint.AlgoSha256, data)
	assert.NoError(t, err)
	assert.True(t, result)

	// bad ID

	result, err = VerifyDigitalAssetID("did:example:123", fingerprint.AlgoSha256, data)
	assert.NoError(t, err)
	assert.False(t, result)

	// invalid Algo

	_, err = VerifyDigitalAssetID("did:example:123", "bad_algo", data)
	assert.Error(t, err)

	// invalid DID

	_, err = VerifyDigitalAssetID("bad:example:123", fingerprint.AlgoSha256, data)
	assert.Error(t, err)
}

func TestUnwrapDigitalAssetID(t *testing.T) {
	assert.Equal(t, "abc", UnwrapDigitalAssetID("did:example:abc"))
	assert.Equal(t, "some_id", UnwrapDigitalAssetID("some_id"))
}
