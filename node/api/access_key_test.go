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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model"
	. "github.com/piprate/metalocker/node/api"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountHandler_GetAccessKeyListHandler(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	handlerBase := NewAccountHandler(env.IdentityBackend)

	acct := createTestAccount(t, "test@example.com", model.AccessLevelHosted, "", env)

	// create an access key

	dw := env.CreateDataWallet(t, acct)

	err := dw.Unlock("pass")
	require.NoError(t, err)

	ak, err := dw.CreateAccessKey(acct.AccessLevel, time.Hour)
	require.NoError(t, err)

	err = env.IdentityBackend.StoreAccessKey(ak)
	require.NoError(t, err)

	invoke := func(authID, accountID string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest(http.MethodGet, "/test-url", http.NoBody)
		c.Set(apibase.UserIDKey, authID)
		c.AddParam("aid", accountID)

		handlerBase.GetAccessKeyListHandler(c)

		return rec
	}

	// account not found

	rec := invoke(acct.ID, "did:non-existent-account")
	require.Equal(t, http.StatusNotFound, rec.Code)

	// happy path

	rec = invoke(acct.ID, acct.ID)
	require.Equal(t, http.StatusOK, rec.Code)

	var listRsp []*model.AccessKey
	readBody(t, rec, &listRsp)

	require.Equal(t, 1, len(listRsp))
	assert.Equal(t, ak.ID, listRsp[0].ID)
}

func TestAccountHandler_PostAccessKeyHandler(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	handlerBase := NewAccountHandler(env.IdentityBackend)

	acct := createTestAccount(t, "test@example.com", model.AccessLevelHosted, "", env)

	// create an access key

	dw := env.CreateDataWallet(t, acct)

	err := dw.Unlock("pass")
	require.NoError(t, err)

	ak, err := dw.CreateAccessKey(acct.AccessLevel, time.Hour)
	require.NoError(t, err)

	invoke := func(authID, accountID string, body io.Reader) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest(http.MethodPost, "/test-url", body)
		c.Set(apibase.UserIDKey, authID)
		c.AddParam("aid", accountID)

		handlerBase.PostAccessKeyHandler(c)

		return rec
	}

	// empty body

	rec := invoke(acct.ID, acct.ID, http.NoBody)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// bad body

	rec = invoke(acct.ID, acct.ID, bytes.NewReader([]byte("bad body")))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// bad envelope

	rec = invoke(acct.ID, acct.ID, bytes.NewReader([]byte("{}")))
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// account not found

	bodyBytes, _ := jsonw.Marshal(ak)

	rec = invoke(acct.ID, "did:non-existent-account", bytes.NewReader(bodyBytes))
	require.Equal(t, http.StatusNotFound, rec.Code)

	// happy path

	rec = invoke(acct.ID, acct.ID, bytes.NewReader(bodyBytes))

	require.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "/test-url/"+ak.ID, rec.Result().Header.Get("Location")) //nolint:bodyclose

	var rsp map[string]any
	readBody(t, rec, &rsp)
	assert.Equal(t, ak.ID, rsp["id"])
}

func TestAccountHandler_GetAccessKeyHandler(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	handlerBase := NewAccountHandler(env.IdentityBackend)

	acct := createTestAccount(t, "test@example.com", model.AccessLevelHosted, "", env)

	// create an access key

	dw := env.CreateDataWallet(t, acct)

	err := dw.Unlock("pass")
	require.NoError(t, err)

	ak, err := dw.CreateAccessKey(acct.AccessLevel, time.Hour)
	require.NoError(t, err)

	err = env.IdentityBackend.StoreAccessKey(ak)
	require.NoError(t, err)

	invoke := func(authID, accountID, keyID string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest(http.MethodGet, "/test-url", http.NoBody)
		c.Set(apibase.UserIDKey, authID)
		c.AddParam("aid", accountID)
		c.AddParam("id", keyID)

		handlerBase.GetAccessKeyHandler(c)

		return rec
	}

	// account not found

	rec := invoke(acct.ID, "did:non-existent-account", "abc")
	require.Equal(t, http.StatusNotFound, rec.Code)

	// key not found

	rec = invoke(acct.ID, acct.ID, "bad-id")
	require.Equal(t, http.StatusNotFound, rec.Code)

	// happy path

	rec = invoke(acct.ID, acct.ID, ak.ID)
	require.Equal(t, http.StatusOK, rec.Code)

	var getRsp *model.AccessKey
	readBody(t, rec, &getRsp)

	require.Equal(t, ak.ID, getRsp.ID)
}

func TestAccountHandler_DeleteAccessKeyHandler(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	handlerBase := NewAccountHandler(env.IdentityBackend)

	acct := createTestAccount(t, "test@example.com", model.AccessLevelHosted, "", env)
	anotherAcct := createTestAccount(t, "another@example.com", model.AccessLevelManaged, "", env)

	// create an access key

	dw := env.CreateDataWallet(t, acct)

	err := dw.Unlock("pass")
	require.NoError(t, err)

	ak, err := dw.CreateAccessKey(acct.AccessLevel, time.Hour)
	require.NoError(t, err)

	err = env.IdentityBackend.StoreAccessKey(ak)
	require.NoError(t, err)

	invoke := func(authID, accountID, keyID string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest(http.MethodDelete, "/test-url", http.NoBody)
		c.Set(apibase.UserIDKey, authID)
		c.AddParam("aid", accountID)
		c.AddParam("id", keyID)

		handlerBase.DeleteAccessKeyHandler(c)

		return rec
	}

	// account not found

	rec := invoke(acct.ID, "did:non-existent-account", "abc")
	require.Equal(t, http.StatusNotFound, rec.Code)

	// key not found

	rec = invoke(acct.ID, acct.ID, "bad-id")
	require.Equal(t, http.StatusNotFound, rec.Code)

	// delete somebody else's key

	rec = invoke(anotherAcct.ID, anotherAcct.ID, ak.ID)
	require.Equal(t, http.StatusNotFound, rec.Code)

	// happy path

	rec = invoke(acct.ID, acct.ID, ak.ID)

	//require.Equal(t, http.StatusNoContent, rec.Code)
	require.True(t, rec.Code >= 200)

	_, err = env.IdentityBackend.GetAccessKey(ak.ID)
	require.Error(t, err)
}
