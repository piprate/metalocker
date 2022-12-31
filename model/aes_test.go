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

	. "github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveEncryptionKey(t *testing.T) {

	// derive key from two secrets

	secret1 := utils.DoubleSha256([]byte("test"))
	secret2 := Hash("server secret", []byte("test secret"))
	key := DeriveEncryptionKey(secret1, secret2)
	assert.Equal(t, "2oSM48CslzFH5w1C6cZIZ4HD9/s+biEEgxGaiOMN264=", key.Base64())

	// try using the key

	msg := []byte("plain test")
	cypherTest, err := EncryptAESCGM(msg, key)
	require.NoError(t, err)
	plainTest, err := DecryptAESCGM(cypherTest, key)
	require.NoError(t, err)
	assert.Equal(t, msg, plainTest)
}

func TestAESKey_Zero(t *testing.T) {
	key := NewEncryptionKey()
	key.Zero()
	for _, b := range key.Bytes() {
		if b != 0 {
			assert.Fail(t, "Non-zero bytes after invoking Zero() on AES key")
		}
	}
}

func TestNewAESKey(t *testing.T) {
	key := NewEncryptionKey()

	newKey := NewAESKey(key.Bytes())
	assert.EqualValues(t, key, newKey)

	// ensure the new key doesn't use the same underlying slice
	key[1] = 0x34
	assert.NotEqualValues(t, key, newKey)
}
