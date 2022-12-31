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
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/utils/measure"
	"github.com/piprate/metalocker/vaults"
)

func PostPurge(vaultAPI vaults.Vault) func(c *gin.Context) {
	return func(c *gin.Context) {
		defer measure.ExecTime("api.PostPurge")()

		log := apibase.CtxLogger(c)

		defer c.Request.Body.Close()

		var res model.StoredResource
		err := apibase.BindJSON(c, &res)
		if err != nil {
			apibase.AbortWithError(c, http.StatusBadRequest, "Bad request body")
			return
		}

		log.Debug().Str("vault", vaultAPI.Name()).Str("id", res.StorageID()).Msg("Purging blob")

		// purge blob

		err = vaultAPI.PurgeBlob(res.StorageID(), res.Params)
		if err != nil {
			if errors.Is(err, model.ErrBlobNotFound) {
				c.Status(http.StatusNotFound)
			} else {
				log.Err(err).Msg("Error purging blob")
				apibase.AbortWithInternalServerError(c, err)
			}
			return
		}

		c.Status(http.StatusOK)
	}
}
