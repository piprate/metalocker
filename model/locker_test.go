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
	"strings"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/piprate/metalocker/contexts"
	. "github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testPublicKey  = base58.Decode("FYmoFw55GeQH7SRFa37dkx1d2dZ3zUF8ckg7wmL7ofN4")
	testPrivateKey = base58.Decode("xt19s1sp2UZCGhy9rNyb1FtxdKiDGZZPQ1RLsDSvcomTyZh1EFYHaUoo19qKunQEhkTSzGztovCC3QXma1foGRr")
)

func init() {
	_ = contexts.PreloadContextsIntoMemory()
	testbase.SetupLogFormat()
}

func TestLockerParticipant_Zero(t *testing.T) {
	did, err := GenerateDID(WithSeed("Test0001"))
	require.NoError(t, err)

	lp, err := Us(did, []byte("Seed0001"))()
	require.NoError(t, err)
	require.True(t, lp.IsHydrated())

	lp.Zero()

	assert.False(t, lp.IsHydrated())
}

func TestGenerateNewHDKey(t *testing.T) {
	_, _, err := GenerateNewHDKey(nil)
	require.NoError(t, err)
}

func TestAnonEncrypt(t *testing.T) {
	msg := "test message"

	cypherText := AnonEncrypt([]byte(msg), testPublicKey)

	decryptedMsg, err := AnonDecrypt(cypherText, testPrivateKey)
	require.NoError(t, err)
	assert.Equal(t, msg, string(decryptedMsg))

	_, err = AnonDecrypt([]byte(strings.Repeat("x", len(cypherText))), testPrivateKey)
	require.Error(t, err)
}

func TestGenerateLocker(t *testing.T) {
	did1, err := GenerateDID(WithSeed("Test0001"))
	require.NoError(t, err)
	did2, err := GenerateDID(WithSeed("Test0002"))
	require.NoError(t, err)
	expiryTime, _ := time.Parse(time.RFC3339Nano, "2026-09-25T21:58:23Z")

	locker, err := GenerateLocker(AccessLevelHosted, "Test Locker", &expiryTime, 123,
		Us(did1, []byte("Seed0001")),
		Them(did2, []byte("Seed0002")))
	require.NoError(t, err)

	assert.Equal(t, "Test Locker", locker.Name)
	assert.Equal(t, AccessLevelHosted, locker.AccessLevel)
	assert.Equal(t, &expiryTime, locker.Expires)
	assert.Equal(t, int64(123), locker.FirstBlock)
	assert.NotNil(t, locker.Participants)
	assert.Equal(t, 2, len(locker.Participants))
	assert.Equal(t, did1.ID, locker.Participants[0].ID)
	assert.True(t, locker.Participants[0].Self)
	assert.Equal(t, "xpub661MyMwAqRbcGntiJJdsNs8GWDXgxBqTxfiFMYXYAiY3TLRYJfNYL11LMk1AMthHkbZy51MFBd2hDc1ZWrio3eCgtn1dYa8Spaq9o9mi8gC", locker.Participants[0].RootPublicKey)
	assert.Equal(t, "NYGupZJZfxotkGaLGKpuqOX8K7xVvE7qEDvNgfru4e8=", locker.Participants[0].SharedSecret)

	// confirm the ability to recover the private key
	assert.NotEmpty(t, locker.Participants[0].RootPrivateKeyEnc)
	b, err := base64.StdEncoding.DecodeString(locker.Participants[0].RootPrivateKeyEnc)
	require.NoError(t, err)
	signKey := did1.SignKeyValue()
	keyBytes, err := AnonDecrypt(b, signKey)
	require.NoError(t, err)
	assert.Equal(t, locker.Participants[0].GetRootPrivateKey(), string(keyBytes))

	assert.Equal(t, did2.ID, locker.Participants[1].ID)
	assert.False(t, locker.Participants[1].Self)
	assert.Equal(t, "xpub661MyMwAqRbcGtePb2bDS7vfBsgrs29eSaZEwGFsJA2rZrw2DPQiTGsCQe9jxmppgApdqP5o42LpnecHr4zibu4vTtTUcMt6avzTxuGXDgm", locker.Participants[1].RootPublicKey)
	assert.Equal(t, "b9qxgnnRDWRMcJ6iX1HbHTK5wzFvm7pXFXKpGSOu/sI=", locker.Participants[1].SharedSecret)
	assert.NotEmpty(t, locker.Participants[1].RootPrivateKeyEnc)
}

func TestLocker_Hydrate(t *testing.T) {
	did1, err := GenerateDID(WithSeed("Test0001"))
	require.NoError(t, err)
	did2, err := GenerateDID(WithSeed("Test0002"))
	require.NoError(t, err)
	expiryTime, _ := time.Parse(time.RFC3339Nano, "2026-09-25T21:58:23Z")

	locker, err := GenerateLocker(AccessLevelHosted, "Test Locker", &expiryTime, 123,
		Us(did1, []byte("Seed0001")),
		Them(did2, []byte("Seed0002")))
	require.NoError(t, err)

	assert.True(t, locker.IsHydrated())

	for _, p := range locker.Participants {
		p.Zero()
	}

	assert.False(t, locker.IsHydrated())

	err = locker.Hydrate(did1.SignKeyValue())
	require.NoError(t, err)

	assert.True(t, locker.IsHydrated())
}

func TestLocker_Zero(t *testing.T) {
	did1, err := GenerateDID(WithSeed("Test0001"))
	require.NoError(t, err)
	did2, err := GenerateDID(WithSeed("Test0002"))
	require.NoError(t, err)
	expiryTime, _ := time.Parse(time.RFC3339Nano, "2026-09-25T21:58:23Z")

	locker, err := GenerateLocker(AccessLevelHosted, "Test Locker", &expiryTime, 123,
		Us(did1, []byte("Seed0001")),
		Them(did2, []byte("Seed0002")))
	require.NoError(t, err)

	assert.True(t, locker.IsHydrated())

	locker.Zero()

	assert.False(t, locker.IsHydrated())
	assert.Equal(t, 0, len(locker.Participants))
}

func TestLocker_Us_Them(t *testing.T) {
	did1, err := GenerateDID(WithSeed("Test0001"))
	require.NoError(t, err)
	did2, err := GenerateDID(WithSeed("Test0002"))
	require.NoError(t, err)
	expiryTime, _ := time.Parse(time.RFC3339Nano, "2026-09-25T21:58:23Z")

	// vanilla case, 2 participants

	locker, err := GenerateLocker(AccessLevelHosted, "Test Locker", &expiryTime, 123,
		Us(did1, []byte("Seed0001")),
		Them(did2, []byte("Seed0002")))
	require.NoError(t, err)

	us := locker.Us()
	require.NotEmpty(t, us)
	assert.Equal(t, did1.ID, us.ID)

	them := locker.Them()
	require.NotEmpty(t, them)
	assert.Equal(t, did2.ID, them.ID)

	// uni-locker

	locker, err = GenerateLocker(AccessLevelHosted, "Test Locker", &expiryTime, 123,
		Us(did1, []byte("Seed0001")))
	require.NoError(t, err)

	us = locker.Us()
	require.NotEmpty(t, us)
	assert.Equal(t, did1.ID, us.ID)

	them = locker.Them()
	require.Empty(t, them)

	// multi-self locker (discouraged)

	locker, err = GenerateLocker(AccessLevelHosted, "Test Locker", &expiryTime, 123,
		Us(did1, []byte("Seed0001")),
		Us(did2, []byte("Seed0002")))
	require.NoError(t, err)

	us = locker.Us()
	require.Empty(t, us)

	them = locker.Them()
	require.Empty(t, them)
}
