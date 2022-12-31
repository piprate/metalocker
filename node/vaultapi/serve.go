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
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/utils/measure"
	"github.com/piprate/metalocker/vaults"
)

func PostServeBlobHandler(vaultAPI vaults.Vault) func(c *gin.Context) {
	return func(c *gin.Context) {
		defer measure.ExecTime("api.PostServeBlobHandler")()

		accessToken := c.GetHeader("X-Vault-Access-Token")

		var res model.StoredResource
		err := apibase.BindJSON(c, &res)
		if err != nil {
			apibase.AbortWithError(c, http.StatusBadRequest, "Bad request body")
			return
		}

		rdr, err := vaultAPI.ServeBlob(res.StorageID(), res.Params, accessToken)
		if err != nil {
			log := apibase.CtxLogger(c)
			if errors.Is(err, model.ErrDataAssetAccessDenied) {
				log.Error().Str("id", res.ID).Interface("params", res.Params).Msg("Access denied")
				apibase.AbortWithError(c, http.StatusUnauthorized, err.Error())
			} else if errors.Is(err, model.ErrBlobNotFound) {
				apibase.AbortWithError(c, http.StatusNotFound, "blob not found")
			} else {
				log.Err(err).Str("id", res.ID).Interface("params", res.Params).Msg("Error serving blob")
				apibase.AbortWithInternalServerError(c, err)
			}
			return
		}

		defer rdr.Close()

		_, err = io.Copy(c.Writer, rdr)
		if err != nil {
			log := apibase.CtxLogger(c)
			log.Err(err).Str("id", res.ID).Interface("params", res.Params).Msg("Error serving blob")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}
}
