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

package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/apibase"
)

func (h *LedgerHandler) GetDataAssetStateHandler(c *gin.Context) {
	id := c.Params.ByName("id")

	state, err := h.ledger.GetDataAssetState(id)
	if err != nil {
		if errors.Is(err, model.ErrDataAssetNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		} else {
			log := apibase.CtxLogger(c)
			log.Err(err).Msg("Error when retrieving data asset state")
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	apibase.JSON(c, http.StatusOK, map[string]model.DataAssetState{"state": state})
}
