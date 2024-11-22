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
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/contexts"
	"github.com/piprate/metalocker/model"
	. "github.com/piprate/metalocker/node/api"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/piprate/metalocker/storage"
	"github.com/piprate/metalocker/storage/memory"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = contexts.PreloadContextsIntoMemory()
	testbase.SetupLogFormat()

	gin.SetMode(gin.ReleaseMode)
}

func invokeValidateRequestSignatureHandler(t *testing.T, body *apibase.SignatureValidationRequest, identityBackend storage.IdentityBackend) *httptest.ResponseRecorder {
	t.Helper()

	bodyBytes, err := jsonw.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request, _ = http.NewRequest(http.MethodPost, "/v1/request-validator", bytes.NewReader(bodyBytes))

	ValidateRequestSignatureHandler(identityBackend)(c)

	return rec
}

func TestValidateRequestSignatureHandler_HappyPath(t *testing.T) {
	identityBackend, _ := memory.CreateIdentityBackend(nil, nil)

	key := &model.AccessKey{
		ID:          "HYJWc4FMvHt56o3K9RHCpt",
		AccountID:   "did:piprate:test-account",
		AccessLevel: model.AccessLevelManaged,
		Secret:      "Kv2w9cS9rtMU88cGLwdqo3duNvvUhZC1Pk/KGodKVDS0O+5T5aD079m7FMNdra4BfX+UN0Ld2GfdrgVznu3KxjnSUWbm7pCLALYopnSZYfvEH0XFGwIIQ/qS0Ug=",
		Type:        model.AccessKeyType,
	}

	body := []byte("test")
	bodyHash := model.HashRequestBody(body)

	err := identityBackend.StoreAccessKey(context.Background(), key)
	require.NoError(t, err)

	req := &apibase.SignatureValidationRequest{
		URL:       "/v1/account",
		KeyID:     "HYJWc4FMvHt56o3K9RHCpt",
		Signature: "Ea58gE9YL3Oo1hLyoi6BPHNWRPC6YP1nDo/3qRLjXas=",
		Header: http.Header{
			"X-Meta-Body-Hash": []string{
				"IvZOgH9mrFaxDYqwGF5P4dpS+yZ5RuHeUfvTJHjXEAk=",
			},
			"X-Meta-Client-Key": []string{
				"cCx+SYim/H/L26YXlWWz4InWWz7hbu0whvB/LSiBjiA=",
			},
			"X-Meta-Date": []string{
				"19700426",
			},
		},
		Timestamp: 10020000,
		BodyHash:  base64.StdEncoding.EncodeToString(bodyHash),
	}

	rec := invokeValidateRequestSignatureHandler(t, req, identityBackend)

	require.Equal(t, http.StatusOK, rec.Code)
	checkResponseBody(t, `{
  "acct": "did:piprate:test-account"
}`, rec.Body.Bytes())
}

func TestValidateRequestSignatureHandler_ExpiredSignature(t *testing.T) {
	identityBackend, _ := memory.CreateIdentityBackend(nil, nil)

	key := &model.AccessKey{
		ID:          "HYJWc4FMvHt56o3K9RHCpt",
		AccountID:   "did:piprate:test-account",
		AccessLevel: model.AccessLevelManaged,
		Secret:      "Kv2w9cS9rtMU88cGLwdqo3duNvvUhZC1Pk/KGodKVDS0O+5T5aD079m7FMNdra4BfX+UN0Ld2GfdrgVznu3KxjnSUWbm7pCLALYopnSZYfvEH0XFGwIIQ/qS0Ug=",
		Type:        model.AccessKeyType,
	}

	body := []byte("test")
	bodyHash := model.HashRequestBody(body)

	err := identityBackend.StoreAccessKey(context.Background(), key)
	require.NoError(t, err)

	req := &apibase.SignatureValidationRequest{
		URL:       "/v1/account",
		KeyID:     "HYJWc4FMvHt56o3K9RHCpt",
		Signature: "Ea58gE9YL3Oo1hLyoi6BPHNWRPC6YP1nDo/3qRLjXas=",
		Header: http.Header{
			"X-Meta-Body-Hash": []string{
				"IvZOgH9mrFaxDYqwGF5P4dpS+yZ5RuHeUfvTJHjXEAk=",
			},
			"X-Meta-Client-Key": []string{
				"cCx+SYim/H/L26YXlWWz4InWWz7hbu0whvB/LSiBjiA=",
			},
			"X-Meta-Date": []string{
				"19700426",
			},
		},
		Timestamp: 20002000,
		BodyHash:  base64.StdEncoding.EncodeToString(bodyHash),
	}

	rec := invokeValidateRequestSignatureHandler(t, req, identityBackend)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	checkResponseBody(t, `{
  "message": "Bad signature"
}`, rec.Body.Bytes())
}

func TestValidateRequestSignatureHandler_BadKeyID(t *testing.T) {
	identityBackend, _ := memory.CreateIdentityBackend(nil, nil)

	body := []byte("test")
	bodyHash := model.HashRequestBody(body)

	req := &apibase.SignatureValidationRequest{
		URL:       "/v1/account",
		KeyID:     "bad-key-id",
		Signature: "Ea58gE9YL3Oo1hLyoi6BPHNWRPC6YP1nDo/3qRLjXas=",
		Header:    http.Header{},
		Timestamp: 20002000,
		BodyHash:  base64.StdEncoding.EncodeToString(bodyHash),
	}

	rec := invokeValidateRequestSignatureHandler(t, req, identityBackend)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	checkResponseBody(t, `{
  "message": "Bad signature"
}`, rec.Body.Bytes())
}
