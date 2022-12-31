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

package testbase

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/piprate/json-gold/ld"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/vaults"
	"github.com/piprate/metalocker/vaults/memory"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

const (
	TestVaultID           = "Z2kcCarCE47SDjtWD5ruyijsQyWMF5B1jjk6HHWngoe"
	TestVaultName         = "memory"
	TestOffChainStorageID = "28HADuAnDmLghgFTk7wppqmSrPMGV1LkQeJBNN9U3U8r"
	IndexStoreID          = "did:piprate:8mMVQNNz3E94JBtWvibyy745eZYBK3CnoLdFaQY33jd5"
	IndexStoreName        = "test_store"
	testManagedAccount    = `
{
  "Account": {
    "id": "did:piprate:F5zishjvbo9tXMNuxthopX",
    "type": "Account",
    "version": 3,
    "email": "test@example.com",
    "encryptedPassword": "gm7FhMFxFD01wGd8dE1RqUAxx7noD8LvPQyBzK+27LA=",
    "state": "active",
    "registeredAt": "2020-12-03T12:33:03.962338Z",
    "name": "John Doe",
    "level": 2,
    "recoveryPublicKey": "wF8KVgjvsgnGoXlpzvvevZM6y8D8pOMy2VSTK48+Xys=",
    "defaultVault": "memory",
    "managedSecretStore": {
      "level": 2,
      "masterKeyParams": "cARTgm9z/+5JVMYq+9+mK9DIFUNr1PZG1hlSl7V80OEzgbHJIZX+COF1pbgjyka5pvMr/hg1WiJXb9Offmk35wAIAAAAAAAACAAAAAAAAAABAAAAAAAAAA==",
      "encryptedPayloadKey": "nviwwR5m+8Fklk76u5tgqtYXqbnW2FC5MAjHoSyVmVRqN9lm2YNtgvr4wZkehdJPu4nc7GnH7EbIQbgcKilIlUMDJIcQSPhW",
      "encryptedPayload": "CRLCC3iLm92JZDRdDFdxE07wXI4SwjVYZ2/5ld62Wxm27KXAyxJ72J0DW8jzXWRd4EAGzjD0oPpb++KwttXjhl2YE3EaD52Agp1mElMzh/v6erMDQeSccOCC1bUp2V/hWa1WnfjNHTm0KDIA2yIQumb/RWwrbyUAZP2LkRIx1gESn2hdjOdxz4uwfI31zmwisdch4UMF9qU8eIOEwNoTXwdTR8aqiRzrtK1d6iwjkaXmBSw52M6DmTUAMMj0NKZ5srsx0KqZK2zuixmVGgDLNSiVrlliEyDHUHPFKHz7zjYluQBSJ6QovNtOPOjNTwVlxHCIX2Jq6y3Whjh5TpoB4RQt9YBYTY/Wzd+lRkJmk3acUvmuJCydON4="
    }
  },
  "RecoveryPhrase": "snake blood recipe scrub napkin salad limit opera bonus episode bean never flock crisp spare inflict blue fresh snake direct nuclear replace trade only",
  "SecondLevelRecoveryCode": "",
  "RootIdentities": [
    {
      "did": {
        "id": "did:piprate:F5zishjvbo9tXMNuxthopX",
        "verKey": "8gHq63gcCazb9kugavwxAwx1tgC996CpdF2pq1m5X8UM",
        "signKey": "xt19s1sp2UZCGhy9rNyb4NP5pL6cmk6JE3sTcKEYfHHeQnMUQ6oMw2anbhMbziJAkR5JC9cpXXvFJ4uuWTUUzdB"
      },
      "created": "2020-12-03T12:33:03.962338Z",
      "name": "Root Identity",
      "type": "Root",
      "level": 2
    }
  ],
  "EncryptedIdentities": [
    {
      "hash": "3c82acfcfbae79a94aa6d682c4b4cd61074b51d421bc687f2118fc7f8a25d97a",
      "lvl": 2,
      "data": "RQ1unPum0xvg05/bsSBNhzhlbEUfNuc37lFVeXUaNnkFlKt0XcqPQAAPQnt8FQLcz60SSam4CpKvbZ2syHJ94hfLj5+KZWaYgt8M+YTJl3pWa+/UWbDm9XzCKflQyqxac1eR2YlsihplrnWNtpHbI9FHJUX+iQ42/CzRJ2jWO6zFdILsAgQJNCNeC+lkRj3X1jW9i1ijTftQnlYKpeWvm6M5m8tyZekkNt/fob93OwI87RyuL6VTyPv1K3HdOqXfPYpvZ6y3qe1nP7K5kUREBpVKo55v422vzDnu3qlJEO/8siDE/6T62kkDCm40m/JVAUqMxKo8hu/XPV02/WYSeRdE1Jut+Rwed+wAwDQXbsXqLIucVTE0e4Kh2vqXUmIXrN0KRjBDIRf6isMlPVdWBiQdEK1sYG56TiY4Qm6iYEM6cQ=="
    }
  ],
  "EncryptedLockers": [
    {
      "hash": "8804e17c3ffbfe25862ebdaeab7cfae3947a35e92ffa05c4cdfc604694850d4a",
      "lvl": 2,
      "data": "2ERsprwtva79ItCyez/JfO7pWvV2FLSM9JJfdzM0tlGX9QMB9Ag4FG6r27ooEKOJ7kJvyzOI0ZI33xrU6H/JFjQU8DkkWmcz0RmiV+V5hOwzbJn/zBCjmHYxYDdpQ0T+AToCqr3/F4gsiX4IylRkbP5Qw+GDaQGDSimr7qG/6U6yP0lcSy3YdzxMf/63HBpNX13SXVtP37B/OjyTy1goK+P8crulqv7vcqDvPK/Gv6iizuccVAQSTb019NkIAoI1pzwXHIR2Ef0a3awHOufOI4SCLul94ePugpLG2zyM33JDWDYUzLSDF/sfK4kKyHY4YP0TdG0qM7yDGszbzDguHVgV1n5V+JvUe8A4rCNdN/oO2GLtGHUXggolvR6AijBMH5VsOek6ua6nF0XXUNWHyV2KKNZUBdq2ejUxSkG6gvJ1bzFTgN1592fZ1ysKuYUJWZi4t7PJH6RVDV1+rDuvbZrpES86vOTuMhdMSHmw/uawEMFeU2x0XssNtK9gSiinPqXxDHnYLw940K7yFJ/NEh0A2BzF5SM0nydIIAXG55W75BdJNvKT01W3KO8M5Lfs2LfAAXHvQfFMeyTxwDuIHVxd8Ixp0p+oYfc65diSLP1mB0hudVAYF0ChwlQ1x/5ksdqZzyyMg7/r5NycYfRL6XCX8rt5lm80waq728lYWxzmkHj0KmP0SZ5PLaQpAalaCoKUDORBdarV+XooqoebWdpb/x3XpI2uJJqfpvW1d59ZyL863BKfKM4YMVWaozkClymUZerPsTuxVHI5QKtKBoXyZyxRu3Xwv3E4QQFD3kupEPUaYWFri57xwilkEhfIGWn6kOZhZCb8HyG1dnqFsSJDeMAfrjzYQ4RgTetGwiuUcarRgnSbjKTZ"
    }
  ]
}
`
	TestAccountClientSecretString = "4c6KgcjSnKSgTbW/lUgb1cx8SujjkP+1C4lNlrm9mQU=" //nolint:gosec
	TestAccountPassword           = "pass123"
	TestLockerID                  = "5bYpBwLVgXpnwgX1XGaNf6JX87KaEHMfd8fqGt4Jo6aB"
)

var (
	TestAccountRootDID        *model.DID
	TestAccountClientSecret   *model.AESKey
	TestAccountID             string
	TestAccountRecoveryPhrase string
)

func init() {
	TestAccountRootDID, _ = model.GenerateDID(model.WithSeed("MetaLockerTest01"))
	TestAccountClientSecretBytes, _ := base64.StdEncoding.DecodeString(TestAccountClientSecretString)
	TestAccountClientSecret = model.NewAESKey(TestAccountClientSecretBytes)

	respBytes := []byte(testManagedAccount)
	var resp *account.GenerationResponse
	_ = jsonw.Unmarshal(respBytes, &resp)
	TestAccountID = resp.Account.ID
	TestAccountRecoveryPhrase = resp.RecoveryPhrase
}

func NewInMemoryVault(t *testing.T, vaultID, vaultName string, sse, cas bool, verifier model.AccessVerifier) (vaults.Vault, *vaults.Config) {
	t.Helper()

	vaultCfg := &vaults.Config{
		ID:   vaultID,
		Name: vaultName,
		Type: "memory",
		SSE:  sse,
		CAS:  cas,
	}

	vault, err := memory.CreateVault(vaultCfg, nil, verifier)
	require.NoError(t, err)

	return vault, vaultCfg
}

func refreshTestAccount(t *testing.T) { //nolint:unused,thelper
	acctTemplate := &account.Account{
		Type:         account.Type,
		Email:        "test@example.com",
		Name:         "John Doe",
		AccessLevel:  model.AccessLevelManaged,
		DefaultVault: TestVaultName,
	}
	hashedPassword := account.HashUserPassword(TestAccountPassword)
	resp, err := account.GenerateAccount(
		acctTemplate, account.WithHashedPassphraseAuth(hashedPassword),
		account.WithRootIdentity(TestAccountRootDID))
	require.NoError(t, err)

	managedKey, err := resp.Account.ExtractManagedKey(hashedPassword)
	require.NoError(t, err)

	ld.PrintDocument("Test Account", resp)
	ld.PrintDocument("Client Secret", managedKey.Base64())
}

func fastHash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	if err != nil {
		return "", err
	} else {
		return string(hashedPassword), nil
	}
}

func TestDID(t *testing.T) *model.DID {
	did, err := model.GenerateDID(model.WithSeed("Test0001"))
	require.NoError(t, err)

	return did
}

func TestUniLocker(t *testing.T) *model.Locker {
	expiryTime, _ := time.Parse(time.RFC3339Nano, "2026-09-25T21:58:23Z")
	locker, err := model.GenerateLocker(
		model.AccessLevelManaged, "Test Locker", &expiryTime, 1,
		model.Us(TestDID(t), []byte("000000000000000000000000Seed0001")))
	require.NoError(t, err)
	locker.ID = TestLockerID

	return locker
}

func TestBlobManager(t *testing.T, cas bool, verifier model.AccessVerifier) *vaults.LocalBlobManager {
	vault, vaultCfg := NewInMemoryVault(t, TestVaultID, TestVaultName, true, cas, verifier)

	bm := vaults.NewLocalBlobManager()
	bm.AddVault(vault, vaultCfg)

	return bm
}
