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
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/utils/jsonw"
)

func (h *Handler) GetAccountListHandler(c *gin.Context) {
	state := c.Query("state")

	accounts, err := h.identityBackend.ListAccounts(c, "", state)
	if err != nil {
		log := apibase.CtxLogger(c)
		log.Err(err).Msg("Error when reading account list")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	apibase.JSON(c, http.StatusOK, accounts)
}

func (h *Handler) GetAccountHandler(c *gin.Context) {
	accountID := c.Params.ByName("id")

	acct, err := h.identityBackend.GetAccount(c, accountID)
	if err != nil {
		log := apibase.CtxLogger(c)
		log.Err(err).Msg("Error when retrieving account details")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	apibase.JSON(c, http.StatusOK, acct)
}

func (h *Handler) PostAccountHandler(c *gin.Context) {
	log := apibase.CtxLogger(c)

	buf, _ := c.GetRawData()

	var acct account.Account
	err := jsonw.Unmarshal(buf, &acct)
	if err != nil {
		log.Err(err).Str("body", string(buf)).Msg("Bad account update request")
		_ = c.AbortWithError(http.StatusBadRequest, errors.New("bad account"))
		return
	}

	log.Debug().Str("body", string(buf)).Msg("Account update submitted")

	if err := acct.Validate(); err != nil {
		log.Err(err).Str("body", string(buf)).Msg("Account validation failed")
		apibase.AbortWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.identityBackend.CreateAccount(c, &acct)
	if err != nil {
		log.Err(err).Msg("Error updating account")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

type AccountAdminPatch struct {
	State string `json:"state,omitempty"`
}

func (h *Handler) PatchAccountHandler(c *gin.Context) {
	log := apibase.CtxLogger(c)

	id := c.Params.ByName("id")

	buf, _ := c.GetRawData()

	var patch AccountAdminPatch
	err := jsonw.Unmarshal(buf, &patch)
	if err != nil {
		log.Err(err).Str("body", string(buf)).Msg("Bad account update request")
		apibase.AbortWithError(c, http.StatusBadRequest, "bad account")
		return
	}

	log.Debug().Str("body", string(buf)).Msg("Account patch submitted")

	acct, err := h.identityBackend.GetAccount(c, id)
	if err != nil {
		log.Err(err).Msg("Error retrieving account")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	switch patch.State {
	case account.StateDeleted:
		acct.State = account.StateDeleted
	case account.StateSuspended:
		if acct.State == account.StateDeleted {
			apibase.AbortWithError(c, http.StatusBadRequest, "bad account state transition")
			return
		}
		acct.State = account.StateSuspended
	case account.StateActive:
		if acct.State == account.StateDeleted {
			apibase.AbortWithError(c, http.StatusBadRequest, "bad account state transition")
			return
		}
		acct.State = account.StateActive
	}

	err = h.identityBackend.UpdateAccount(c, acct)
	if err != nil {
		log.Err(err).Msg("Error updating account")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}
