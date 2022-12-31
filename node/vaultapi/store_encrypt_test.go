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

package vaultapi_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/piprate/json-gold/ld"
	"github.com/piprate/metalocker/contexts"
	. "github.com/piprate/metalocker/node/vaultapi"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/vaults"
	_ "github.com/piprate/metalocker/vaults/fs"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustRemoveAll(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		panic(err)
	}
}

func init() {
	_ = contexts.PreloadContextsIntoMemory()
	testbase.SetupLogFormat()
}

func TestPostStoreEncrypt(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	dir, err := os.MkdirTemp(".", "tempdir_")
	require.NoError(t, err)

	defer mustRemoveAll(dir)

	fvConfig := &vaults.Config{
		ID:   "Z2kcCarCE47SDjtWD5ruyijsQyWMF5B1jjk6HHWngoe",
		Name: "local",
		Type: "fs",
		SSE:  false,
		Params: map[string]any{
			"root_dir": dir,
		},
	}
	fileVault, err := vaults.CreateVault(fvConfig, nil, nil)
	require.NoError(t, err)

	fvConfigSSE := &vaults.Config{
		ID:   "Z2kcCarCE47SDjtWD5ruyijsQyWMF5B1jjk6HHWngoe",
		Name: "local",
		Type: "fs",
		SSE:  true,
		Params: map[string]any{
			"root_dir": dir,
		},
	}
	fileVaultSSE, err := vaults.CreateVault(fvConfigSSE, nil, nil)
	require.NoError(t, err)

	tests := []struct {
		vaultAPI    vaults.Vault
		requestBody io.Reader
		code        int
		body        string
	}{
		// #1 successful call (non-SSE)
		{
			fileVault,
			strings.NewReader("test blob"),
			http.StatusOK,
			`
{
  "asset": "did:piprate:CqpkA3UfeCsRUSAexhhV67UJ6ZviPsWZrZ1iq7ngxkmL",
  "method": "fs",
  "mimeType": "text/plain; charset=utf-8",
  "size": 9,
  "type": "Resource",
  "vault": "Z2kcCarCE47SDjtWD5ruyijsQyWMF5B1jjk6HHWngoe"
}`,
		},
		// #2 successful call (SSE)
		{
			fileVaultSSE,
			strings.NewReader("test blob"),
			http.StatusOK,
			`
{
  "asset": "did:piprate:CqpkA3UfeCsRUSAexhhV67UJ6ZviPsWZrZ1iq7ngxkmL",
  "method": "fs",
  "mimeType": "text/plain; charset=utf-8",
  "size": 9,
  "type": "Resource",
  "vault": "Z2kcCarCE47SDjtWD5ruyijsQyWMF5B1jjk6HHWngoe"
}`,
		},
	}
	for i, test := range tests {
		log.Info().Int("number", i+1).Msg("~~~ Running PostStoreEncrypt test")

		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest(http.MethodPost, "/v1/vault/Z2kcCarCE47SDjtWD5ruyijsQyWMF5B1jjk6HHWngoe/encrypt", test.requestBody)

		handlerFunc := PostStoreEncrypt(test.vaultAPI)

		handlerFunc(c)
		if rec.Code != test.code {
			t.Fatalf("%d: wrong code, got %d want %d", i, rec.Code, test.code)
		}

		var expected map[string]any
		var actual map[string]any
		err := jsonw.Unmarshal([]byte(test.body), &expected)
		require.NoError(t, err)

		err = jsonw.Unmarshal(rec.Body.Bytes(), &actual)
		require.NoError(t, err)

		// delete randomly generated values
		delete(actual, "encryptionKey")
		delete(actual, "id")
		delete(actual, "params")

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
}
