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

package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/apibase"
)

func (h *Handler) GetIdentityListHandler(c *gin.Context) {
	accounts, err := h.identityBackend.ListDIDDocuments(c)
	if err != nil {
		log := apibase.CtxLogger(c)
		log.Err(err).Msg("Error when reading identity list")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	apibase.JSON(c, http.StatusOK, accounts)
}

func (h *Handler) PostIdentityHandler(c *gin.Context) {
	log := apibase.CtxLogger(c)

	var ddoc model.DIDDocument
	err := apibase.BindJSON(c, &ddoc)
	if err != nil {
		apibase.AbortWithError(c, http.StatusBadRequest, "Bad request body")
		return
	}

	err = h.identityBackend.CreateDIDDocument(c, &ddoc)
	if err != nil {
		log.Err(err).Msg("Error when saving identity")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	log.Info().Str("did", ddoc.ID).Msg("Identity published")

	c.Status(http.StatusOK)
}
