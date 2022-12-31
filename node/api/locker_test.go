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
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountHandler_GetLockerListHandler(t *testing.T) {
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

		handlerBase.GetLockerListHandler(c)

		return rec
	}

	// account not found

	rec := invoke(acct.ID, "did:non-existent-account")
	require.Equal(t, http.StatusNotFound, rec.Code)

	// happy path

	rec = invoke(acct.ID, acct.ID)
	require.Equal(t, http.StatusOK, rec.Code)

	var listRsp []*account.DataEnvelope
	readBody(t, rec, &listRsp)

	require.Equal(t, 2, len(listRsp))
}

func TestPostLockerHandler(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	handlerBase := NewAccountHandler(env.IdentityBackend)

	acct := createTestAccount(t, "test@example.com", model.AccessLevelHosted, "", env)

	invoke := func(authID, accountID string, body io.Reader) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest(http.MethodPost, "/test-url", body)
		c.Set(apibase.UserIDKey, authID)
		c.AddParam("aid", accountID)

		handlerBase.PostLockerHandler(c)

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

	bodyBytes, _ := jsonw.Marshal(&account.DataEnvelope{
		Hash:          "abc",
		AccessLevel:   model.AccessLevelHosted,
		EncryptedBody: "xxx",
	})

	rec = invoke(acct.ID, "did:non-existent-account", bytes.NewReader(bodyBytes))
	require.Equal(t, http.StatusNotFound, rec.Code)

	// happy path

	rec = invoke(acct.ID, acct.ID, bytes.NewReader(bodyBytes))

	require.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "/test-url/abc", rec.Result().Header.Get("Location")) //nolint:bodyclose

	var rsp map[string]string
	readBody(t, rec, &rsp)
	assert.Equal(t, "abc", rsp["hash"])
}

func TestAccountHandler_GetLockerHandler(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	handlerBase := NewAccountHandler(env.IdentityBackend)

	acct := createTestAccount(t, "test@example.com", model.AccessLevelHosted, "", env)

	err := env.IdentityBackend.StoreLocker(acct.ID, &account.DataEnvelope{
		Hash:        "abc",
		AccessLevel: model.AccessLevelHosted,
	})
	require.NoError(t, err)

	invoke := func(authID, accountID, lockerHash string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest(http.MethodGet, "/test-url", http.NoBody)
		c.Set(apibase.UserIDKey, authID)
		c.AddParam("aid", accountID)
		c.AddParam("hash", lockerHash)

		handlerBase.GetLockerHandler(c)

		return rec
	}

	// account not found

	rec := invoke(acct.ID, "did:non-existent-account", "abc")
	require.Equal(t, http.StatusNotFound, rec.Code)

	// locker not found

	rec = invoke(acct.ID, acct.ID, "bad-hash")
	require.Equal(t, http.StatusNotFound, rec.Code)

	// happy path

	rec = invoke(acct.ID, acct.ID, "abc")
	require.Equal(t, http.StatusOK, rec.Code)

	var listRsp *account.DataEnvelope
	readBody(t, rec, &listRsp)

	require.NotEmpty(t, listRsp.Hash)
}
