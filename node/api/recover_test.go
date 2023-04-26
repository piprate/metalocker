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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	. "github.com/piprate/metalocker/node/api"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRecoveryCodeHandler(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	acct := createTestAccount(t, "test@example.com", model.AccessLevelHosted, "", env)

	invoke := func(accountID string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest(http.MethodGet,
			fmt.Sprintf("/recover?email=%s", accountID), http.NoBody)

		GetRecoveryCodeHandler(env.IdentityBackend)(c)

		return rec
	}

	// account not found

	rec := invoke("did:non-existent-account")
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// happy path

	rec = invoke(acct.ID)
	require.Equal(t, http.StatusOK, rec.Code)

	var rsp *GetRecoveryCodeResponse
	readBody(t, rec, &rsp)

	assert.NotEmpty(t, rsp.Code)

	c, err := env.IdentityBackend.GetRecoveryCode(rsp.Code)
	require.NoError(t, err)
	assert.Equal(t, acct.ID, c.UserID)
}

func TestRecoverAccountHandler(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	dw, recDetails, err := env.Factory.RegisterAccount(
		&account.Account{
			Email:        "test@example.com",
			Name:         "John Doe",
			AccessLevel:  model.AccessLevelHosted,
			DefaultVault: testbase.TestVaultName,
		},
		account.WithPassphraseAuth("pass"))
	require.NoError(t, err)

	acct := dw.Account()

	invoke := func(body io.Reader) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest(http.MethodPost,
			"test-url", body)

		RecoverAccountHandler(env.IdentityBackend)(c)

		return rec
	}

	// empty body

	rec := invoke(bytes.NewReader(nil))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// bad body

	rec = invoke(bytes.NewReader([]byte("bad body")))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// empty form

	rec = invoke(bytes.NewReader([]byte("{}")))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// recovery code should be deleted after the first attempt, even if it fails

	rc, err := account.NewRecoveryCode("test@example.com", 5*60)
	require.NoError(t, err)

	err = env.IdentityBackend.CreateRecoveryCode(rc)
	require.NoError(t, err)

	_, err = env.IdentityBackend.GetRecoveryCode(rc.Code)
	require.NoError(t, err)

	reqBytes, _ := jsonw.Marshal(&account.RecoveryRequest{
		RecoveryCode: rc.Code,
	})

	rec = invoke(bytes.NewReader(reqBytes))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// confirm the code is deleted
	_, err = env.IdentityBackend.GetRecoveryCode(rc.Code)
	require.Error(t, err)

	// happy path

	rc, err = account.NewRecoveryCode("test@example.com", 5*60)
	require.NoError(t, err)

	err = env.IdentityBackend.CreateRecoveryCode(rc)
	require.NoError(t, err)

	cryptoKey, _, privKey, err := account.GenerateKeysFromRecoveryPhrase(recDetails.RecoveryPhrase)
	require.NoError(t, err)

	newPassphrase := "pass2"

	req := account.BuildRecoveryRequest(rc.UserID, rc.Code, privKey, newPassphrase, nil)
	reqBytes, _ = jsonw.Marshal(req)

	rec = invoke(bytes.NewReader(reqBytes))
	require.Equal(t, http.StatusOK, rec.Code)

	var rsp *AccountRecoveryResponse
	readBody(t, rec, &rsp)

	require.NotEmpty(t, rsp.Account)
	assert.Equal(t, account.StateRecovery, rsp.Account.State)
	assert.Equal(t, acct.ID, rsp.Account.ID)

	// confirm the code is deleted
	_, err = env.IdentityBackend.GetRecoveryCode(rc.Code)
	require.Error(t, err)

	dwToRecover, err := env.Factory.CreateDataWallet(rsp.Account)
	require.NoError(t, err)

	_, err = dwToRecover.Recover(cryptoKey, newPassphrase)
	require.NoError(t, err)

	// try recovering an account without recovery key (it would be irrecoverable)

	acct.RecoveryPublicKey = ""
	err = env.IdentityBackend.UpdateAccount(acct)
	require.NoError(t, err)

	rc, err = account.NewRecoveryCode("test@example.com", 5*60)
	require.NoError(t, err)

	err = env.IdentityBackend.CreateRecoveryCode(rc)
	require.NoError(t, err)

	_, _, privKey, err = account.GenerateKeysFromRecoveryPhrase(recDetails.RecoveryPhrase)
	require.NoError(t, err)

	req = account.BuildRecoveryRequest(rc.UserID, rc.Code, privKey, newPassphrase, nil)
	reqBytes, _ = jsonw.Marshal(req)

	rec = invoke(bytes.NewReader(reqBytes))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	var msg apibase.Response
	readBody(t, rec, &msg)

	assert.Equal(t, "Account doesn't support recovery", msg.Message)
}

func TestRecoverAccountHandler_ManagedWorkflow(t *testing.T) {
	// this test is for a request that contains the account's managed crypto key.
	// This enables server side recovery for managed accounts for clients
	// that don't have access to advanced cryptography.
	// In the end of this test the managed account should be in 'active' state.

	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	invoke := func(body io.Reader) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest(http.MethodPost,
			"test-url", body)

		RecoverAccountHandler(env.IdentityBackend)(c)

		return rec
	}

	// should fail for hosted accounts

	_, recDetails, err := env.Factory.RegisterAccount(
		&account.Account{
			Email:        "hosted@example.com",
			Name:         "John Doe",
			AccessLevel:  model.AccessLevelHosted,
			DefaultVault: testbase.TestVaultName,
		},
		account.WithPassphraseAuth("pass"))
	require.NoError(t, err)

	rc, err := account.NewRecoveryCode("hosted@example.com", 5*60)
	require.NoError(t, err)

	err = env.IdentityBackend.CreateRecoveryCode(rc)
	require.NoError(t, err)

	cryptoKey, _, privKey, err := account.GenerateKeysFromRecoveryPhrase(recDetails.RecoveryPhrase)
	require.NoError(t, err)

	newPassphrase := "pass2"

	req := account.BuildRecoveryRequest(rc.UserID, rc.Code, privKey, newPassphrase, cryptoKey)
	reqBytes, _ := jsonw.Marshal(req)

	rec := invoke(bytes.NewReader(reqBytes))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// happy path

	dw, recDetails, err := env.Factory.RegisterAccount(
		&account.Account{
			Email:        "managed@example.com",
			Name:         "John Doe",
			AccessLevel:  model.AccessLevelManaged,
			DefaultVault: testbase.TestVaultName,
		},
		account.WithPassphraseAuth("pass"))
	require.NoError(t, err)

	acct := dw.Account()

	rc, err = account.NewRecoveryCode("managed@example.com", 5*60)
	require.NoError(t, err)

	err = env.IdentityBackend.CreateRecoveryCode(rc)
	require.NoError(t, err)

	cryptoKey, _, privKey, err = account.GenerateKeysFromRecoveryPhrase(recDetails.RecoveryPhrase)
	require.NoError(t, err)

	req = account.BuildRecoveryRequest(rc.UserID, rc.Code, privKey, newPassphrase, cryptoKey)
	reqBytes, _ = jsonw.Marshal(req)

	rec = invoke(bytes.NewReader(reqBytes))
	require.Equal(t, http.StatusOK, rec.Code)

	var rsp *AccountRecoveryResponse
	readBody(t, rec, &rsp)

	require.NotEmpty(t, rsp.Account)
	assert.Equal(t, account.StateActive, rsp.Account.State)
	assert.Equal(t, acct.ID, rsp.Account.ID)
}
