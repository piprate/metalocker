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

package vaultapi

import (
	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/utils/measure"
	"github.com/piprate/metalocker/vaults"

	"net/http"
)

func PostStoreRaw(vaultAPI vaults.Vault) func(c *gin.Context) {
	return func(c *gin.Context) {
		defer measure.ExecTime("api.PostStoreRaw")()

		log := apibase.CtxLogger(c)

		defer c.Request.Body.Close()

		log.Debug().Str("vault", vaultAPI.Name()).Msg("Uploading raw blob")

		// persist blob
		res, err := vaultAPI.CreateBlob(c.Request.Body)
		if err != nil {
			apibase.AbortWithInternalServerError(c, err)
			return
		}

		c.JSON(http.StatusOK, res)
	}
}
