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

package fingerprint_test

import (
	"bytes"
	"testing"

	"github.com/btcsuite/btcd/btcutil/base58"
	. "github.com/piprate/metalocker/utils/fingerprint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSha256Fingerprint(t *testing.T) {
	msg := []byte("test message")
	hash, err := GetSha256Fingerprint(bytes.NewReader(msg))
	require.NoError(t, err)
	assert.Equal(t, "5F5ioPfSDZnsWGQG9zw5WB44TDRoeYy8miszVL74Tcx7", base58.Encode(hash))
}
