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
	"crypto/ed25519"
	"encoding/base64"
	"testing"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/piprate/json-gold/ld"
	"github.com/piprate/metalocker/model"
	. "github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildAccountRecoveryRequest(t *testing.T) {
	userID := "test@example.com"
	recoveryCode := "53QPUKdVDjLEbxZA3BZWT7oLpsD1VjqGA7XnN3T21vcV"
	newPassphrase := "new_password"
	privKeyBytes := base58.Decode("ucHoMKY1EVgGrEMg3aQejMDQvq6hrLcxSZ27eEvK3V3iPv4nxukQ7eLyMK4jGmjkRZpueFmChXNsEV3eawvYbHc")
	var privKey ed25519.PrivateKey = privKeyBytes
	req := BuildRecoveryRequest(userID, recoveryCode, privKey, newPassphrase)

	actualBytes, _ := jsonw.Marshal(req)

	expectedReq := `
{
  "userID": "test@example.com",
  "recoveryCode": "53QPUKdVDjLEbxZA3BZWT7oLpsD1VjqGA7XnN3T21vcV",
  "signature": "D872hDxdZaxqnfinC54V1ngr99avEbDu13vCAFgMhYL5VkUumFXc5NiVxtfFPQujAbsKXrmQ38jJ57qu9nQFbLn",
  "encryptedPassword": "lx0jHs83Jhp/KAeZ7O1Quziebie9rUeMyPUfrwkusfI="
}
`

	testbase.AssertEqualJSON(t, expectedReq, actualBytes)
}

func TestGenerateKeysFromRecoveryPhrase(t *testing.T) {
	recoveryPhrase := "book shed chapter large work worth record robot enough extend gadget major just entry umbrella icon stomach miss maid glance push debate pass first"

	ck, pubk, privk, err := GenerateKeysFromRecoveryPhrase(recoveryPhrase)
	require.NoError(t, err)

	assert.Equal(t, "446ZHDoHFsXFfAPWe3YbAecm4D3B1xty9TNFnhd4U7L8", base58.Encode(ck[:]))
	assert.Equal(t, "7Q5nKCvH3EXo56fGHndqRzqadCs5K2WfoovRjitYubKg", base58.Encode(pubk[:]))
	assert.Equal(t, "ucHoMKY1EVgGrEMg3aQejMDQvq6hrLcxSZ27eEvK3V3iPv4nxukQ7eLyMK4jGmjkRZpueFmChXNsEV3eawvYbHc", base58.Encode(privk[:]))
}

func TestGenerateMasterRecoveryKeyPair(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	ld.PrintDocument("publicKey", base58.Encode(publicKey))
	ld.PrintDocument("privateKey", base58.Encode(privateKey))
}

func TestSecondLevelRecoveryProcedure(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	// Application deployment

	pubKey := base58.Decode("hKszucKiubKFsHatmzF1ytqNC1V3yiAsc37JiFQ3Bd5")
	masterPublicKey := ed25519.PublicKey(pubKey)
	masterPrivateKeyStr := "3idhf7MtKtJWGV2m5CQooaAyNMy9inPW3xarCra3z4hf56xpq95MqHygRoYgFvSxWK22RN4bbLsu71QeQ8DF4yDd"

	// User registration

	acct := &Account{
		Email:       "test@example.com",
		Name:        "Tester",
		AccessLevel: model.AccessLevelManaged,
	}

	genResp, err := GenerateAccount(
		acct,
		WithPassphraseAuth("pass"),
		WithSLRK(masterPublicKey))
	require.NoError(t, err)

	acct = genResp.Account
	slrc := genResp.SecondLevelRecoveryCode

	// Account recovery

	recoveryCode := "53QPUKdVDjLEbxZA3BZWT7oLpsD1VjqGA7XnN3T21vcV"
	newPassphrase := "new_password"

	// administrator enters the master private key
	masterPrivateKeyBytes := base58.Decode(masterPrivateKeyStr)
	masterPrivateKey := ed25519.PrivateKey(masterPrivateKeyBytes)
	// Admin UI decrypts recovery private key from SLRC code
	slrcBytes := base58.Decode(slrc)
	pkBytes, err := model.AnonDecrypt(slrcBytes, masterPrivateKey)
	require.NoError(t, err)
	privKey := ed25519.PrivateKey(pkBytes)
	recPubKey, err := base64.StdEncoding.DecodeString(acct.RecoveryPublicKey)
	require.NoError(t, err)
	assert.Equal(t, ed25519.PublicKey(recPubKey), privKey.Public())

	req := BuildRecoveryRequest(acct.Email, recoveryCode, privKey, newPassphrase)

	require.True(t, req.Valid(recPubKey))

	acct.EncryptedPassword = req.EncryptedPassword

	require.NoError(t, ReHashPassphrase(acct, nil))

	acct.State = StateRecovery

	encryptedKeyBytes, err := base64.StdEncoding.DecodeString(acct.EncryptedRecoverySecret)
	require.NoError(t, err)
	cryptoKeyBytes, err := model.AnonDecrypt(encryptedKeyBytes, privKey)
	require.NoError(t, err)

	recoveredAcct, err := Recover(acct, model.NewAESKey(cryptoKeyBytes), newPassphrase)
	require.NoError(t, err)

	managedKey, err := recoveredAcct.ExtractManagedKey(HashUserPassword(newPassphrase))
	require.NoError(t, err)

	dw := env.CreateDataWallet(t, recoveredAcct)

	err = dw.UnlockAsManaged(managedKey)
	require.NoError(t, err)
}
