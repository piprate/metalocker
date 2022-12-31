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

package api_test

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/gin-gonic/gin"
	"github.com/piprate/json-gold/ld"
	"github.com/piprate/metalocker/contexts"
	"github.com/piprate/metalocker/ledger/local"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	. "github.com/piprate/metalocker/node/api"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/piprate/metalocker/storage"
	"github.com/piprate/metalocker/storage/memory"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	RegisterRequestManaged = `{
  "account": {
    "id": "did:piprate:oD2BFmMTsRSaiBpWuXfWugcoN9xFtApwLeKvv8eAL6X",
    "type": "Account",
    "email": "test@example.com",
    "encryptedPassword": "8p9eHFEa7HMgnMxtaGDHY3iz2S1dGMXraKp2N1Fj6BiQ",
    "name": "John Doe",
    "level": 2
  },
  "registrationCode": "BLAH"
}`

	RegisterRequestHosted = `{
  "account": {
    "id": "did:piprate:4wfrNwcZhXm5Py9fXVHpNtJhoDGbQGoznV7huqwrMccF",
    "type": "Account",
    "email": "test@example.com",
    "encryptedPassword": "8p9eHFEa7HMgnMxtaGDHY3iz2S1dGMXraKp2N1Fj6BiQ",
    "name": "John Doe",
    "level": 3,
    "masterKeyParams": "xx",
    "hostedSecretStore": {
      "level": 3,
      "masterKeyParams": "xx",
      "encryptedPayloadKey": "xx",
      "encryptedPayload": "xx"
	},
    "managedSecretStore": {
      "level": 2,
      "masterKeyParams": "xx",
      "encryptedPayloadKey": "xx",
      "encryptedPayload": "xx"
    }
  },
  "registrationCode": "BLAH"
}`

	RegistrationCode = "BLAH"
)

var (
	secondLevelRecoveryKey []byte
)

func init() {
	_ = contexts.PreloadContextsIntoMemory()
	testbase.SetupLogFormat()

	gin.SetMode(gin.ReleaseMode)

	secondLevelRecoveryKey, _, _ = ed25519.GenerateKey(strings.NewReader("000000000000000000000000Steward1"))
}

func createTempLedger(t *testing.T) (model.Ledger, string) {
	t.Helper()

	dir, err := os.MkdirTemp(".", "tempdir_")
	require.NoError(t, err)
	dbFilepath := filepath.Join(dir, "ledger.bolt")
	ledgerAPI, err := local.NewBoltLedger(dbFilepath, nil, 10, 0)
	require.NoError(t, err)

	return ledgerAPI, dir
}

func invokeHandler(body string, registrationCode string, slrKey []byte, identityBackend storage.IdentityBackend,
	ledger model.Ledger, hashFunctionError error) *httptest.ResponseRecorder {

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request, _ = http.NewRequest(http.MethodPost, "/v1/register", bytes.NewReader([]byte(body)))

	handlerFunc := RegisterHandler(
		[]string{registrationCode},
		testbase.TestVaultName,
		slrKey,
		func() []byte {
			val := base58.Decode("BQFWpk8z1sW3yjGz2ZqeibjFEYiDWW3ZsYiMEJRyG7y6")
			return val
		},
		func(passwd string) (string, error) {
			return "$hashed$password", hashFunctionError
		},
		func() time.Time {
			return time.Unix(1000, 0)
		},
		nil,
		identityBackend,
		ledger,
	)

	handlerFunc(c)

	return rec
}

func checkResponseBody(t *testing.T, expectedBody string, actualBody []byte) {
	t.Helper()

	var expected any
	var actual any
	err := jsonw.Unmarshal([]byte(expectedBody), &expected)
	if err != nil {
		// Error when unmarshalling expected data. It will be treated as string.
		expected = expectedBody
	}
	err = jsonw.Unmarshal(actualBody, &actual)
	if err != nil {
		// Error when unmarshalling actual data. It will be treated as string.
		actual = string(actualBody)
	}

	if !assert.True(t, ld.DeepCompare(expected, actual, false), "Wrong body") {
		_, _ = os.Stdout.WriteString("==== ACTUAL ====\n")
		b, _ := jsonw.MarshalIndent(actual, "", "  ")
		_, _ = os.Stdout.Write(b)
		_, _ = os.Stdout.WriteString("\n")
		_, _ = os.Stdout.WriteString("==== EXPECTED ====\n")
		b, _ = jsonw.MarshalIndent(expected, "", "  ")
		_, _ = os.Stdout.Write(b)
		t.FailNow()
	}
}

func TestRegisterHandler_EmptyBody(t *testing.T) {
	identityBackend, _ := memory.CreateIdentityBackend(nil, nil)
	ledgerAPI, dir := createTempLedger(t)
	defer func() { _ = os.RemoveAll(dir) }()

	rec := invokeHandler("", "", secondLevelRecoveryKey, identityBackend, ledgerAPI, nil)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	checkResponseBody(t, `{
  "message": "Bad register request"
}`, rec.Body.Bytes())
}

func TestRegisterHandler_BadRequestFormat(t *testing.T) {
	identityBackend, _ := memory.CreateIdentityBackend(nil, nil)
	ledgerAPI, dir := createTempLedger(t)
	defer func() { _ = os.RemoveAll(dir) }()

	rec := invokeHandler("not json string", "", secondLevelRecoveryKey, identityBackend, ledgerAPI, nil)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	checkResponseBody(t, `{
  "message": "Bad register request"
}`, rec.Body.Bytes())
}

func TestRegisterHandler_JSONWithWrongFields(t *testing.T) {
	identityBackend, _ := memory.CreateIdentityBackend(nil, nil)
	ledgerAPI, dir := createTempLedger(t)
	defer func() { _ = os.RemoveAll(dir) }()

	rec := invokeHandler(`{"some_field": 123}`, "", secondLevelRecoveryKey, identityBackend, ledgerAPI, nil)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	checkResponseBody(t, `{
  "message": "Bad register request"
}`, rec.Body.Bytes())
}

func TestRegisterHandler_BadRegistrationCode(t *testing.T) {
	identityBackend, _ := memory.CreateIdentityBackend(nil, nil)
	ledgerAPI, dir := createTempLedger(t)
	defer func() { _ = os.RemoveAll(dir) }()

	rec := invokeHandler(RegisterRequestHosted, "REGCODE", secondLevelRecoveryKey, identityBackend,
		ledgerAPI, nil)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	checkResponseBody(t, `{
  "message": "Bad registration code"
}`, rec.Body.Bytes())
}

func TestRegisterHandler_AccountExists(t *testing.T) {
	identityBackend, _ := memory.CreateIdentityBackend(nil, nil)
	err := identityBackend.CreateAccount(&account.Account{ID: "did:piprate:4wfrNwcZhXm5Py9fXVHpNtJhoDGbQGoznV7huqwrMccF"})
	require.NoError(t, err)

	ledgerAPI, dir := createTempLedger(t)
	defer func() { _ = os.RemoveAll(dir) }()

	rec := invokeHandler(RegisterRequestHosted, RegistrationCode, secondLevelRecoveryKey, identityBackend, ledgerAPI, nil)

	require.Equal(t, http.StatusConflict, rec.Code)
	checkResponseBody(t, `{
  "message": "Account already exists"
}`, rec.Body.Bytes())
}

func TestRegisterHandler_HashFunctionError(t *testing.T) {
	identityBackend, _ := memory.CreateIdentityBackend(nil, nil)

	ledgerAPI, dir := createTempLedger(t)
	defer func() { _ = os.RemoveAll(dir) }()

	rec := invokeHandler(RegisterRequestHosted, RegistrationCode, secondLevelRecoveryKey, identityBackend, ledgerAPI,
		fmt.Errorf("error in hash function"))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	checkResponseBody(t, `{
  "message": "Account creation failed"
}`, rec.Body.Bytes())
}

func TestRegisterHandler_Managed(t *testing.T) {
	identityBackend, _ := memory.CreateIdentityBackend(nil, nil)

	ledgerAPI, dir := createTempLedger(t)
	defer func() { _ = os.RemoveAll(dir) }()

	rec := invokeHandler(RegisterRequestManaged, RegistrationCode, secondLevelRecoveryKey, identityBackend, ledgerAPI, nil)

	require.Equal(t, http.StatusOK, rec.Code)

	var actual map[string]any
	err := jsonw.Unmarshal(rec.Body.Bytes(), &actual)
	require.NoError(t, err)
	assert.Equal(t, "ok", actual["status"])
	assert.Equal(t, "once escape trash actor sunset noble aim screen bring leaf train uncover item organ expect head oval swing report auto arena foil milk ripple",
		actual["recoveryPhrase"])
	assert.NotEmpty(t, actual["secondLevelRecoveryCode"])
}

func TestRegisterHandler_Hosted(t *testing.T) {
	identityBackend, _ := memory.CreateIdentityBackend(nil, nil)

	ledgerAPI, dir := createTempLedger(t)
	defer func() { _ = os.RemoveAll(dir) }()

	rec := invokeHandler(RegisterRequestHosted, RegistrationCode, secondLevelRecoveryKey, identityBackend, ledgerAPI, nil)

	require.Equal(t, http.StatusOK, rec.Code)

	var actual map[string]any
	err := jsonw.Unmarshal(rec.Body.Bytes(), &actual)
	require.NoError(t, err)
	assert.Equal(t, "ok", actual["status"])
}
