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

func (h *AccountHandler) GetLockerListHandler(c *gin.Context) {
	accessLevel := utils.StringToInt(c.Query("level"))

	masterAccountID := apibase.GetUserID(c)
	accountID := c.Params.ByName("aid")

	if !hasAccountPermissions(c, h.identityBackend, masterAccountID, accountID) {
		return
	}

	lockerList, err := h.identityBackend.ListLockers(c, accountID, model.AccessLevel(accessLevel))
	if err != nil {
		log := apibase.CtxLogger(c)
		log.Err(err).Msg("Error when reading locker list")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	apibase.JSON(c, http.StatusOK, lockerList)
}

func (h *AccountHandler) PostLockerHandler(c *gin.Context) {

	masterAccountID := apibase.GetUserID(c)
	accountID := c.Params.ByName("aid")

	if !hasAccountPermissions(c, h.identityBackend, masterAccountID, accountID) {
		return
	}

	buf, _ := c.GetRawData()

	var lockerEnvelope account.DataEnvelope
	err := jsonw.Unmarshal(buf, &lockerEnvelope)
	if err != nil {
		log := apibase.CtxLogger(c)
		log.Err(err).Str("body", string(buf)).Msg("Bad locker creation request")
		apibase.AbortWithError(c, http.StatusBadRequest, "bad locker")
		return
	}

	if err := lockerEnvelope.Validate(); err != nil {
		log := apibase.CtxLogger(c)
		log.Err(err).Str("body", string(buf)).Msg("Locker envelope validation failed")
		apibase.AbortWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.identityBackend.StoreLocker(c, accountID, &lockerEnvelope)

	if err != nil {
		apibase.AbortWithError(c, http.StatusInternalServerError, "Locker creation failed")
		return
	}

	c.Writer.Header().Add("Location", fmt.Sprintf("%s/%s", c.Request.URL.RequestURI(), lockerEnvelope.Hash))

	var result struct {
		Hash string `json:"hash"`
	}

	result.Hash = lockerEnvelope.Hash

	apibase.JSON(c, http.StatusCreated, &result)
}

func (h *AccountHandler) GetLockerHandler(c *gin.Context) {

	masterAccountID := apibase.GetUserID(c)
	accountID := c.Params.ByName("aid")

	if !hasAccountPermissions(c, h.identityBackend, masterAccountID, accountID) {
		return
	}

	hash := c.Params.ByName("hash")

	lockerEnvelope, err := h.identityBackend.GetLocker(c, accountID, hash)
	if err != nil {
		if errors.Is(err, storage.ErrLockerNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
		} else {
			log := apibase.CtxLogger(c)
			log.Err(err).Msg("Error when retrieving locker envelope")
			apibase.AbortWithError(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	apibase.JSON(c, http.StatusOK, lockerEnvelope)
}
