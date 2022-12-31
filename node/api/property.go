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

func (h *AccountHandler) GetPropertyListHandler(c *gin.Context) {
	accessLevel := utils.StringToInt(c.Query("level"))

	masterAccountID := apibase.GetUserID(c)
	accountID := c.Params.ByName("aid")

	if !hasAccountPermissions(c, h.identityBackend, masterAccountID, accountID) {
		return
	}

	propList, err := h.identityBackend.ListProperties(accountID, model.AccessLevel(accessLevel))
	if err != nil {
		log := apibase.CtxLogger(c)
		log.Err(err).Msg("Error when reading property list")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	apibase.JSON(c, http.StatusOK, propList)
}

func (h *AccountHandler) PostPropertyHandler(c *gin.Context) {

	masterAccountID := apibase.GetUserID(c)
	accountID := c.Params.ByName("aid")

	if !hasAccountPermissions(c, h.identityBackend, masterAccountID, accountID) {
		return
	}

	buf, _ := c.GetRawData()

	var propEnv account.DataEnvelope
	err := jsonw.Unmarshal(buf, &propEnv)
	if err != nil {
		log := apibase.CtxLogger(c)
		log.Err(err).Str("body", string(buf)).Msg("Bad property creation request")
		apibase.AbortWithError(c, http.StatusBadRequest, "bad property envelope")
		return
	}

	if err := propEnv.Validate(); err != nil {
		log := apibase.CtxLogger(c)
		log.Err(err).Str("body", string(buf)).Msg("Property envelope validation failed")
		apibase.AbortWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.identityBackend.StoreProperty(accountID, &propEnv)

	if err != nil {
		apibase.AbortWithError(c, http.StatusInternalServerError, "Property creation failed")
		return
	}

	c.Writer.Header().Add("Location", fmt.Sprintf("%s/%s", c.Request.URL.RequestURI(), propEnv.Hash))

	var result struct {
		Hash string `json:"hash"`
	}

	result.Hash = propEnv.Hash

	apibase.JSON(c, http.StatusCreated, &result)
}

func (h *AccountHandler) GetPropertyHandler(c *gin.Context) {

	masterAccountID := apibase.GetUserID(c)
	accountID := c.Params.ByName("aid")

	if !hasAccountPermissions(c, h.identityBackend, masterAccountID, accountID) {
		return
	}

	hash := c.Params.ByName("hash")

	propEnv, err := h.identityBackend.GetProperty(accountID, hash)
	if err != nil {
		if errors.Is(err, storage.ErrPropertyNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
		} else {
			log := apibase.CtxLogger(c)
			log.Err(err).Msg("Error when retrieving property envelope")
			apibase.AbortWithError(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	apibase.JSON(c, http.StatusOK, propEnv)
}

func (h *AccountHandler) DeletePropertyHandler(c *gin.Context) {
	masterAccountID := apibase.GetUserID(c)
	accountID := c.Params.ByName("aid")

	if !hasAccountPermissions(c, h.identityBackend, masterAccountID, accountID) {
		return
	}

	hash := c.Params.ByName("hash")

	err := h.identityBackend.DeleteProperty(accountID, hash)
	if err != nil {
		if errors.Is(err, storage.ErrPropertyNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
		} else {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}
