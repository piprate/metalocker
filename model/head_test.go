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
	"encoding/base64"
	"testing"
	"time"

	"github.com/piprate/metalocker/contexts"
	. "github.com/piprate/metalocker/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadID(t *testing.T) {
	_ = contexts.PreloadContextsIntoMemory()

	did1, err := GenerateDID(WithSeed("Test0001"))
	require.NoError(t, err)
	did2, err := GenerateDID(WithSeed("Test0002"))
	require.NoError(t, err)

	expiryTime, _ := time.Parse(time.RFC3339Nano, "2026-09-25T21:58:23Z")
	locker, err := GenerateLocker(AccessLevelHosted, "Test Locker", &expiryTime, 123,
		Us(did1, []byte("Seed0001")),
		Them(did2, []byte("Seed0002")))
	require.NoError(t, err)

	// fix locker ID
	locker.ID = "G7ShEXRnKS3pyuzbpnrhVwmRcq3TP3TL8eWJktUv3upU"

	assetID := "did:piprate:Fw4CEkwm3n3gMcRGtC9r2aDR7iuTLkciJkkmMPi6b7Km"

	headID := HeadID(assetID, locker.ID, locker.Us(), "main")
	assert.Equal(t, "2BfG8PvqKFSKpGWLyQfnmsqE3m3YAfQaDacCMw58t41n", headID)
}

func TestPackHeadBody(t *testing.T) {
	did1, err := GenerateDID(WithSeed("Test0001"))
	require.NoError(t, err)
	did2, err := GenerateDID(WithSeed("Test0002"))
	require.NoError(t, err)

	expiryTime, _ := time.Parse(time.RFC3339Nano, "2026-09-25T21:58:23Z")
	locker, err := GenerateLocker(AccessLevelHosted, "Test Locker", &expiryTime, 123,
		Us(did1, []byte("Seed0001")),
		Them(did2, []byte("Seed0002")))
	require.NoError(t, err)

	// fix locker ID
	locker.ID = "G7ShEXRnKS3pyuzbpnrhVwmRcq3TP3TL8eWJktUv3upU"

	assetID := "did:piprate:Fw4CEkwm3n3gMcRGtC9r2aDR7iuTLkciJkkmMPi6b7Km"

	packedBody := PackHeadBody(assetID, locker.ID, locker.Us().ID, "main", "record_123")

	assert.Equal(t, "ZGlkOnBpcHJhdGU6Rnc0Q0Vrd20zbjNnTWNSR3RDOXIyYURSN2l1VExrY2lKa2ttTVBpNmI3S218RzdTaEVYUm5LUzNweXV6YnBucmhWd21SY3EzVFAzVEw4ZVdKa3RVdjN1cFV8ZGlkOnBpcHJhdGU6R21OdUxrcWkzTlk5aDVQYmVMcUhCcXxtYWlufHJlY29yZF8xMjM=",
		base64.StdEncoding.EncodeToString(packedBody))

	unpackedAssetID, lockerID, participantID, headName, recordID := UnpackHeadBody(packedBody)
	assert.Equal(t, assetID, unpackedAssetID)
	assert.Equal(t, locker.ID, lockerID)
	assert.Equal(t, locker.Us().ID, participantID)
	assert.Equal(t, "main", headName)
	assert.Equal(t, "record_123", recordID)
}
