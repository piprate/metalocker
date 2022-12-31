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

package account_test

import (
	"testing"

	"github.com/piprate/json-gold/ld"
	"github.com/piprate/metalocker/model"
	. "github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAccount_Hosted(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	acctTemplate := &Account{
		Email:        "test@example.com",
		Name:         "Test User",
		AccessLevel:  model.AccessLevelHosted,
		DefaultVault: testbase.TestVaultName,
	}
	passPhrase := "pass123"
	resp, err := GenerateAccount(
		acctTemplate,
		WithPassphraseAuth(passPhrase))
	require.NoError(t, err)
	assert.NotEmpty(t, resp.RecoveryPhrase)
	assert.Empty(t, resp.SecondLevelRecoveryCode)
	assert.Equal(t, "test@example.com", resp.Account.Email)
	assert.Equal(t, "Test User", resp.Account.Name)
	assert.Equal(t, model.AccessLevelHosted, resp.Account.AccessLevel)
	assert.Equal(t, testbase.TestVaultName, resp.Account.DefaultVault)

	require.Equal(t, 1, len(resp.RootIdentities))
	require.Equal(t, 1, len(resp.EncryptedIdentities))
	require.Equal(t, 2, len(resp.EncryptedLockers))
}

func TestGenerateAccount_Managed(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	acctTemplate := &Account{
		Email:        "test@example.com",
		Name:         "Test User",
		AccessLevel:  model.AccessLevelManaged,
		DefaultVault: testbase.TestVaultName,
	}
	passPhrase := "pass123"
	hashedPassword := HashUserPassword(passPhrase)
	resp, err := GenerateAccount(
		acctTemplate,
		WithHashedPassphraseAuth(hashedPassword))
	require.NoError(t, err)
	assert.NotEmpty(t, resp.RecoveryPhrase)
	assert.Empty(t, resp.SecondLevelRecoveryCode)
	assert.Equal(t, "test@example.com", resp.Account.Email)
	assert.Equal(t, "Test User", resp.Account.Name)
	assert.Equal(t, model.AccessLevelManaged, resp.Account.AccessLevel)
	assert.Equal(t, testbase.TestVaultName, resp.Account.DefaultVault)

	require.Equal(t, 1, len(resp.RootIdentities))
	require.Equal(t, 1, len(resp.EncryptedIdentities))
	require.Equal(t, 1, len(resp.EncryptedLockers))
}

func TestAccount_Copy(t *testing.T) {
	var acct Account
	err := jsonw.Unmarshal([]byte(
		`
{
  "id": "did:piprate:DzSq2WeNPCX3SBAQzJZXyAvzar8zVHpUT59LxS8nLAr1",
  "type": "Account",
  "version": 3,
  "email": "test@example.com",
  "encryptedPassword": "gm7FhMFxFD01wGd8dE1RqUAxx7noD8LvPQyBzK+27LA=",
  "state": "active",
  "registeredAt": "2020-07-01T12:52:50.490787Z",
  "name": "Test User",
  "level": 3,
  "recoveryPublicKey": "w04j6YFaAw91CexI8qogpU23AvIcBJqk6Krqmf2+rs8=",
  "managedSecretStore": {
    "level": 2,
    "masterKeyParams": "10bfDiqKxUzUipCUOIxOhrVU36roURczngzOTO3e6PuExWV355+d/aDs2SfV6+dOikTDpRTcATfg9TaE3NpO3wAIAAAAAAAACAAAAAAAAAABAAAAAAAAAA==",
    "encryptedPayloadKey": "LiQ/Lo1hrNKPtiT5eplxAmL247amq4wPHRWZx/oGS8r/G1ghqUZyWsa9/cI2+kb5bMxVf9NmAt9EJLvbDk/bv9hx66A0FxeO",
    "encryptedPayload": "HdzstvB6Cx4YGhrjR7wiTQpTio2vMZ542X2/H0xbiSPNOBqc6npPurObQaNOBm7H7yLYITEQv0H9JmPV1A+oavYODJ93Hq8i79l48ecJWXh3isL/QoVlMlUf0W9bD2A50c3hWfthrnroNwkdPXjKdqHO93R/9uAS4hpvih71Kn+7NVQzbphOeSRPiyXW41+P0DgbW98yPHLPt8GsW/TdBwpsJNnjkHfyqWXt+Q1O6uoBJaaZrhyZAWX7y+fy5I7iyjpPgCrtQFjztpaVWgZ4PgbCMwaH2w=="
  },
  "hostedSecretStore": {
    "level": 3,
    "masterKeyParams": "YBAccuLGznuUffMOkGT1fc6hR9iFjbsuxgOxC0XaZaATxXf0zksWi8zar8i54e6M+OpTxUJAZ93fwJ5XJzA99wAIAAAAAAAACAAAAAAAAAABAAAAAAAAAA==",
    "encryptedPayloadKey": "4w/08+lWgAIFHv6ZqqNCutweVo1cNTDjKibeWtPHMRVwq0ttnfbLiTWWnB8XbZR01aoES7U3pfeOKYcTn0QAPz0ib6DmrIzR",
    "encryptedPayload": "sKPTxR3Kkkzy2s7XJqr3KhE9q29HS40OFHKxngyzpExn/r80/8+xJOtQHXP+DXXF3wIcOpZ7ePb7oGNL84oAf9xkV2CLlakRjp0XM+1CIVS3aoZ/vZ1SfvLOvH9T36tJJft6Pwr7mQ6s//3pT3c797EdV2Na9AglJA7OcpOQUcWvu9E/eOMk7mxPLmtbzB2hdEBoYC/oZxfoWL30ZYXqcgqbdJdI0xJa69reqP2s9veS0UBkECmFhJWBxCM8XpetWPhELr6sHArquvlUFiksokGt91zTekAgtSocRwQeoU54SCWlbO9t/XW0nxemdo6oMNKJUSb5ZZ8erB+slXdC9KrINb3xBa7W0eL4kSojWED2xGbvhEs1OVpu21ZUvTtJMygN8XOkPw=="
  }
}
	`), &acct)
	require.NoError(t, err)

	acctCopy := acct.Copy()

	assert.True(t, ld.DeepCompare(&acct, acctCopy, true))

	// ensure Copy() creates a new copy of the struct

	acct.Name = "Updated"
	assert.NotEqual(t, "Updated", acctCopy.Name)

	acct.HostedSecretStore.AccessLevel = model.AccessLevelRestricted
	assert.Equal(t, model.AccessLevelHosted, acctCopy.HostedSecretStore.AccessLevel)
}
