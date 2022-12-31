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
	"github.com/piprate/metalocker/storage/memory"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitAccountRoutes(t *testing.T) {
	router := gin.Default()
	identityBackend, _ := memory.CreateIdentityBackend(nil, nil)
	InitAccountRoutes(router.Group("test"), identityBackend)
}

func TestAccountHandler_GetOwnAccountHandler(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	handlerBase := NewAccountHandler(env.IdentityBackend)

	acct := createTestAccount(t, "test@example.com", model.AccessLevelHosted, "", env)

	invoke := func(authID string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest(http.MethodGet, "/test-url", http.NoBody)
		c.Set(apibase.UserIDKey, authID)

		handlerBase.GetOwnAccountHandler(c)

		return rec
	}

	// account not found

	rec := invoke("did:non-existent-account")
	require.Equal(t, http.StatusInternalServerError, rec.Code)

	// happy path

	rec = invoke(acct.ID)
	require.Equal(t, http.StatusOK, rec.Code)

	var acctRsp *account.Account
	readBody(t, rec, &acctRsp)

	assert.Equal(t, "John Doe", acctRsp.Name)
	assert.Empty(t, acctRsp.EncryptedPassword)
}

func TestPutAccountHandler(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	handlerBase := NewAccountHandler(env.IdentityBackend)

	acct := createTestAccount(t, "test@example.com", model.AccessLevelHosted, "", env)

	invoke := func(authID string, body io.Reader) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest(http.MethodPut, "/test-url", body)
		c.Set(apibase.UserIDKey, authID)

		handlerBase.PutAccountHandler(c)

		return rec
	}

	// empty body

	rec := invoke(acct.ID, http.NoBody)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// bad body

	rec = invoke(acct.ID, bytes.NewReader([]byte("bad body")))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// account not found

	acctBytes, _ := jsonw.Marshal(acct)

	rec = invoke("did:non-existent-account", bytes.NewReader(acctBytes))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// happy path

	rec = invoke(acct.ID, bytes.NewReader(acctBytes))
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestPatchAccountHandler(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	handlerBase := NewAccountHandler(env.IdentityBackend)

	hostedAcct := createTestAccount(t, "hosted@example.com", model.AccessLevelHosted, "", env)
	managedAcct := createTestAccount(t, "managed@example.com", model.AccessLevelManaged, "", env)

	invoke := func(authID string, body io.Reader) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest(http.MethodPatch, "/test-url", body)
		c.Set(apibase.UserIDKey, authID)

		handlerBase.PatchAccountHandler(c)

		return rec
	}

	// empty body

	rec := invoke(hostedAcct.ID, http.NoBody)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// bad body

	rec = invoke(hostedAcct.ID, bytes.NewReader([]byte("bad body")))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// account not found

	patch := &AccountPatch{
		Email:                "updated@example.com",
		OldEncryptedPassword: "",
		NewEncryptedPassword: "",
		Name:                 "New Name",
		GivenName:            "New",
		FamilyName:           "Name",
	}
	patchBytes, _ := jsonw.Marshal(patch)

	rec = invoke("did:non-existent-account", bytes.NewReader(patchBytes))
	require.Equal(t, http.StatusInternalServerError, rec.Code)

	// happy path

	rec = invoke(hostedAcct.ID, bytes.NewReader(patchBytes))
	require.Equal(t, http.StatusOK, rec.Code)

	updatedAcct, err := env.IdentityBackend.GetAccount(hostedAcct.ID)
	require.NoError(t, err)

	assert.Equal(t, patch.Email, updatedAcct.Email)
	assert.Equal(t, patch.Name, updatedAcct.Name)
	assert.Equal(t, patch.GivenName, updatedAcct.GivenName)
	assert.Equal(t, patch.FamilyName, updatedAcct.FamilyName)

	// try changing password for hosted account

	patch.OldEncryptedPassword = account.HashUserPassword("pass")
	patch.NewEncryptedPassword = account.HashUserPassword("new_pass")
	patchBytes, _ = jsonw.Marshal(patch)

	rec = invoke(hostedAcct.ID, bytes.NewReader(patchBytes))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// fail, if new email taken

	rec = invoke(managedAcct.ID, bytes.NewReader(patchBytes))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// change password for managed account

	patch.Email = "updated2@example.com"
	patch.OldEncryptedPassword = account.HashUserPassword("pass")
	patch.NewEncryptedPassword = account.HashUserPassword("new_pass")
	patchBytes, _ = jsonw.Marshal(patch)

	rec = invoke(managedAcct.ID, bytes.NewReader(patchBytes))
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAccountHandler_DeleteAccountHandler(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	handlerBase := NewAccountHandler(env.IdentityBackend)

	acct := createTestAccount(t, "test@example.com", model.AccessLevelHosted, "", env)
	anotherAcct := createTestAccount(t, "another@example.com", model.AccessLevelHosted, "", env)

	invoke := func(authID, accountID string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest(http.MethodDelete, "/test-url", http.NoBody)
		c.Set(apibase.UserIDKey, authID)
		c.AddParam("aid", accountID)

		handlerBase.DeleteAccountHandler(c)

		return rec
	}

	// account not found

	rec := invoke(acct.ID, "did:non-existent-account")
	require.Equal(t, http.StatusNotFound, rec.Code)

	// try deleting somebody else's account

	rec = invoke(acct.ID, anotherAcct.ID)
	require.Equal(t, http.StatusNotFound, rec.Code)

	// happy path

	rec = invoke(acct.ID, acct.ID)
	require.Equal(t, http.StatusOK, rec.Code)

	_, err := env.IdentityBackend.GetAccount(acct.ID)
	require.Error(t, err)
}

func TestAccountHandler_GetSubAccountListHandler(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	handlerBase := NewAccountHandler(env.IdentityBackend)

	acct := createTestAccount(t, "test@example.com", model.AccessLevelHosted, "", env)
	subAcct := createTestAccount(t, "sub@example.com", model.AccessLevelManaged, acct.ID, env)

	invoke := func(authID, accountID string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest(http.MethodGet, "/test-url", http.NoBody)
		c.Set(apibase.UserIDKey, authID)
		c.AddParam("aid", accountID)

		handlerBase.GetSubAccountListHandler(c)

		return rec
	}

	// account not found

	rec := invoke("did:non-existent-account", acct.ID)
	require.Equal(t, http.StatusNotFound, rec.Code)

	// happy path

	rec = invoke(acct.ID, acct.ID)
	require.Equal(t, http.StatusOK, rec.Code)

	var acctRsp []*account.Account
	readBody(t, rec, &acctRsp)

	require.Equal(t, 1, len(acctRsp))
	assert.Equal(t, subAcct.ID, acctRsp[0].ID)
}

func TestPostSubAccountHandler(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	handlerBase := NewAccountHandler(env.IdentityBackend)

	parent := createTestAccount(t, "test@example.com", model.AccessLevelHosted, "", env)

	invoke := func(authID string, body io.Reader) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest(http.MethodPost, "/test-url", body)
		c.Set(apibase.UserIDKey, authID)

		handlerBase.PostSubAccountHandler(c)

		return rec
	}

	// empty body

	rec := invoke(parent.ID, http.NoBody)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// bad body

	rec = invoke(parent.ID, bytes.NewReader([]byte("bad body")))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// account not found

	acctTemplate := &account.Account{
		Email:         "sub@example.com",
		Name:          "Sub-account",
		AccessLevel:   model.AccessLevelManaged,
		DefaultVault:  testbase.TestVaultName,
		ParentAccount: parent.ID,
	}
	acctResp, err := account.GenerateAccount(acctTemplate, account.WithPassphraseAuth("pass"))
	require.NoError(t, err)
	acct := acctResp.Account

	acctBytes, _ := jsonw.Marshal(acct)

	rec = invoke("did:non-existent-account", bytes.NewReader(acctBytes))
	require.Equal(t, http.StatusNotFound, rec.Code)

	// happy path

	rec = invoke(parent.ID, bytes.NewReader(acctBytes))
	require.Equal(t, http.StatusCreated, rec.Code)

	// fail if account exists

	rec = invoke(parent.ID, bytes.NewReader(acctBytes))
	require.Equal(t, http.StatusConflict, rec.Code)
}

func TestAccountHandler_GetAccountHandler(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	handlerBase := NewAccountHandler(env.IdentityBackend)

	acct := createTestAccount(t, "test@example.com", model.AccessLevelHosted, "", env)

	invoke := func(authID, accountID string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest(http.MethodGet, "/test-url", http.NoBody)
		c.Set(apibase.UserIDKey, authID)
		c.AddParam("aid", accountID)

		handlerBase.GetAccountHandler(c)

		return rec
	}

	// account not found

	rec := invoke("did:non-existent-account", acct.ID)
	require.Equal(t, http.StatusNotFound, rec.Code)

	// happy path

	rec = invoke(acct.ID, acct.ID)
	require.Equal(t, http.StatusOK, rec.Code)

	var acctRsp *account.Account
	readBody(t, rec, &acctRsp)

	assert.Equal(t, "John Doe", acctRsp.Name)
	assert.Empty(t, acctRsp.EncryptedPassword)
}

func createTestAccount(t *testing.T, email string, accessLevel model.AccessLevel, parent string, env *testbase.TestMetaLockerEnvironment) *account.Account {
	t.Helper()

	dw, _, err := env.Factory.RegisterAccount(
		&account.Account{
			Email:         email,
			Name:          "John Doe",
			AccessLevel:   accessLevel,
			DefaultVault:  testbase.TestVaultName,
			ParentAccount: parent,
		},
		account.WithPassphraseAuth("pass"))
	require.NoError(t, err)

	return dw.Account()
}

func readBody(t *testing.T, rec *httptest.ResponseRecorder, dest any) {
	t.Helper()

	rspBytes, err := io.ReadAll(rec.Result().Body) //nolint:bodyclose
	require.NoError(t, err)

	require.NoError(t, jsonw.Unmarshal(rspBytes, &dest))
}
