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
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/storage"
)

func (h *AccountHandler) GetAccessKeyListHandler(c *gin.Context) {

	masterAccountID := apibase.GetUserID(c)
	accountID := c.Params.ByName("aid")

	if !hasAccountPermissions(c, h.identityBackend, masterAccountID, accountID) {
		return
	}

	keys, err := h.identityBackend.ListAccessKeys(accountID)
	if err != nil {
		log := apibase.CtxLogger(c)
		log.Err(err).Msg("Error when reading sub-account list")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	for _, ak := range keys {
		// we never return the key secret
		ak.Secret = ""
	}
	apibase.JSON(c, http.StatusOK, keys)
}

func (h *AccountHandler) GetAccessKeyHandler(c *gin.Context) {
	masterAccountID := apibase.GetUserID(c)
	accountID := c.Params.ByName("aid")

	if !hasAccountPermissions(c, h.identityBackend, masterAccountID, accountID) {
		return
	}

	id := c.Params.ByName("id")

	ak, err := h.identityBackend.GetAccessKey(id)
	if err != nil {
		if errors.Is(err, storage.ErrAccessKeyNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		} else {
			log := apibase.CtxLogger(c)
			log.Err(err).Msg("Error when retrieving access key")
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	if ak.AccountID != accountID {
		log := apibase.CtxLogger(c)
		log.Warn().Msg("Attempting to access an access key for another account")
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// we never return the key secret
	ak.Secret = ""

	apibase.JSON(c, http.StatusOK, ak)
}

func (h *AccountHandler) PostAccessKeyHandler(c *gin.Context) {
	log := apibase.CtxLogger(c)

	var ak model.AccessKey
	err := apibase.BindJSON(c, &ak)
	if err != nil {
		apibase.AbortWithError(c, http.StatusBadRequest, "Bad request body")
		return
	}

	masterAccountID := apibase.GetUserID(c)
	accountID := c.Params.ByName("aid")

	if !hasAccountPermissions(c, h.identityBackend, masterAccountID, accountID) {
		return
	}

	if ak.AccountID != accountID {
		apibase.AbortWithError(c, http.StatusBadRequest, "Key/account mismatch")
		return
	}

	if ak.ID == "" {
		ak.ID = model.GenerateAccessKeyID()
	}

	err = h.identityBackend.StoreAccessKey(&ak)
	if err != nil {
		log.Err(err).Msg("Error when saving new access key")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	log.Info().Str("id", ak.ID).Msg("Access key generated")

	c.Writer.Header().Add("Location", fmt.Sprintf("%s/%s", c.Request.URL.RequestURI(), ak.ID))
	apibase.JSON(c, http.StatusCreated, ak)
}

func (h *AccountHandler) DeleteAccessKeyHandler(c *gin.Context) {
	masterAccountID := apibase.GetUserID(c)
	accountID := c.Params.ByName("aid")

	if !hasAccountPermissions(c, h.identityBackend, masterAccountID, accountID) {
		return
	}

	id := c.Params.ByName("id")

	ak, err := h.identityBackend.GetAccessKey(id)
	if err != nil {
		if errors.Is(err, storage.ErrAccessKeyNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		} else {
			log := apibase.CtxLogger(c)
			log.Err(err).Msg("Error when retrieving access key for deletion")
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	if ak.AccountID != accountID {
		log := apibase.CtxLogger(c)
		log.Warn().Msg("Attempting to delete an access key for another account")
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	err = h.identityBackend.DeleteAccessKey(id)
	if err != nil {
		log := apibase.CtxLogger(c)
		if errors.Is(err, storage.ErrAccessKeyNotFound) {
			log.Err(err).Msg("Attempting to delete a non-existent access key")
			c.AbortWithStatus(http.StatusNotFound)
			return
		} else {
			log.Err(err).Msg("Error when deleting access key")
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	c.Status(http.StatusNoContent)
}
