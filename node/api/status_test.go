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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	. "github.com/piprate/metalocker/node/api"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/stretchr/testify/assert"
)

func TestGetStatusHandler(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	controls := &ServerControls{
		Status:          "ok",
		MaintenanceMode: false,
		JWTPublicKey:    "public-key",
	}
	handler := GetStatusHandler(controls, env.Ledger)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request, _ = http.NewRequest(http.MethodGet, "/test-url", http.NoBody)

	handler(c)

	var controlsRsp *ServerControls
	readBody(t, rec, &controlsRsp)

	assert.NotEmpty(t, controlsRsp.GenesisBlockHash)
}
