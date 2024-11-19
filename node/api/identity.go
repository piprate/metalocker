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
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/storage"
	"github.com/piprate/metalocker/utils"
	"github.com/piprate/metalocker/utils/jsonw"
)

func (h *AccountHandler) GetIdentityListHandler(c *gin.Context) {
	accessLevel := utils.StringToInt(c.Query("level"))

	masterAccountID := apibase.GetUserID(c)
	accountID := c.Params.ByName("aid")

	if !hasAccountPermissions(c, h.identityBackend, masterAccountID, accountID) {
		return
	}

	idyList, err := h.identityBackend.ListIdentities(c, accountID, model.AccessLevel(accessLevel))
	if err != nil {
		log := apibase.CtxLogger(c)
		log.Err(err).Msg("Error when reading identity list")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	apibase.JSON(c, http.StatusOK, idyList)
}

func (h *AccountHandler) PostIdentityHandler(c *gin.Context) {

	masterAccountID := apibase.GetUserID(c)
	accountID := c.Params.ByName("aid")

	if !hasAccountPermissions(c, h.identityBackend, masterAccountID, accountID) {
		return
	}

	buf, _ := c.GetRawData()

	var idyEnv account.DataEnvelope
	err := jsonw.Unmarshal(buf, &idyEnv)
	if err != nil {
		log := apibase.CtxLogger(c)
		log.Err(err).Str("body", string(buf)).Msg("Bad identity creation request")
		apibase.AbortWithError(c, http.StatusBadRequest, "bad identity")
		return
	}

	if err := idyEnv.Validate(); err != nil {
		log := apibase.CtxLogger(c)
		log.Err(err).Str("body", string(buf)).Msg("Identity envelope validation failed")
		apibase.AbortWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.identityBackend.StoreIdentity(c, accountID, &idyEnv)

	if err != nil {
		apibase.AbortWithError(c, http.StatusInternalServerError, "Identity creation failed")
		return
	}

	c.Writer.Header().Add("Location", fmt.Sprintf("%s/%s", c.Request.URL.RequestURI(), idyEnv.Hash))

	var result struct {
		Hash string `json:"hash"`
	}

	result.Hash = idyEnv.Hash

	apibase.JSON(c, http.StatusCreated, &result)
}

func (h *AccountHandler) GetIdentityHandler(c *gin.Context) {

	masterAccountID := apibase.GetUserID(c)
	accountID := c.Params.ByName("aid")

	if !hasAccountPermissions(c, h.identityBackend, masterAccountID, accountID) {
		return
	}

	hash := c.Params.ByName("hash")

	idyEnv, err := h.identityBackend.GetIdentity(c, accountID, hash)
	if err != nil {
		if errors.Is(err, storage.ErrIdentityNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
		} else {
			log := apibase.CtxLogger(c)
			log.Err(err).Msg("Error when retrieving identity envelope")
			apibase.AbortWithError(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	apibase.JSON(c, http.StatusOK, idyEnv)
}
