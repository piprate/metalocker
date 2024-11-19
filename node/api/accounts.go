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
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/storage"
	"github.com/piprate/metalocker/utils/jsonw"
	"golang.org/x/crypto/bcrypt"
)

type (
	AccountHandler struct {
		identityBackend storage.IdentityBackend
	}
)

func NewAccountHandler(identityBackend storage.IdentityBackend) *AccountHandler {
	return &AccountHandler{
		identityBackend: identityBackend,
	}
}

func InitAccountRoutes(rg *gin.RouterGroup, identityBackend storage.IdentityBackend) {

	h := NewAccountHandler(identityBackend)

	rg.GET("/account", h.GetOwnAccountHandler)
	rg.PUT("/account", h.PutAccountHandler)
	rg.PATCH("/account", h.PatchAccountHandler)

	rg.POST("/account", h.PostSubAccountHandler)
	rg.GET("/account/:aid", h.GetAccountHandler)
	rg.DELETE("/account/:aid", h.DeleteAccountHandler)
	rg.GET("/account/:aid/children", h.GetSubAccountListHandler)
	rg.GET("/account/:aid/access-key", h.GetAccessKeyListHandler)
	rg.POST("/account/:aid/access-key", h.PostAccessKeyHandler)
	rg.GET("/account/:aid/access-key/:id", h.GetAccessKeyHandler)
	rg.DELETE("/account/:aid/access-key/:id", h.DeleteAccessKeyHandler)

	rg.GET("/account/:aid/identity", h.GetIdentityListHandler)
	rg.POST("/account/:aid/identity", h.PostIdentityHandler)
	rg.GET("/account/:aid/identity/:hash", h.GetIdentityHandler)

	rg.GET("/account/:aid/locker", h.GetLockerListHandler)
	rg.POST("/account/:aid/locker", h.PostLockerHandler)
	rg.GET("/account/:aid/locker/:hash", h.GetLockerHandler)

	rg.GET("/account/:aid/property", h.GetPropertyListHandler)
	rg.POST("/account/:aid/property", h.PostPropertyHandler)
	rg.GET("/account/:aid/property/:hash", h.GetPropertyHandler)
	rg.DELETE("/account/:aid/property/:hash", h.DeletePropertyHandler)
}

func (h *AccountHandler) GetOwnAccountHandler(c *gin.Context) {
	acct, err := h.identityBackend.GetAccount(c, apibase.GetUserID(c))
	if err != nil {
		log := apibase.CtxLogger(c)
		log.Err(err).Str("id", apibase.GetUserID(c)).Msg("Error when retrieving account details")
		apibase.AbortWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	acct.EncryptedPassword = ""

	apibase.JSON(c, http.StatusOK, acct)
}

func (h *AccountHandler) PutAccountHandler(c *gin.Context) {
	log := apibase.CtxLogger(c)

	buf, _ := c.GetRawData()
	if len(buf) == 0 {
		apibase.AbortWithError(c, http.StatusBadRequest, "bad request")
		return
	}

	var acct account.Account
	err := jsonw.Unmarshal(buf, &acct)
	if err != nil {
		log.Err(err).Str("body", string(buf)).Msg("Bad account update request")
		apibase.AbortWithError(c, http.StatusBadRequest, "bad account")
		return
	}

	log.Debug().Str("body", string(buf)).Msg("Account update submitted")

	if err := acct.Validate(); err != nil {
		log.Err(err).Str("body", string(buf)).Msg("Account validation failed")
		apibase.AbortWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	oldAccountRecord, err := h.identityBackend.GetAccount(c, apibase.GetUserID(c))
	if err != nil {
		log.Err(err).Msg("Error updating account")
		apibase.AbortWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	if (oldAccountRecord.State == account.StateRecovery && acct.State != account.StateActive) ||
		(oldAccountRecord.State != account.StateRecovery && oldAccountRecord.State != acct.State) {
		// we only allow account state change from 'recovery' to 'active' as a one-off update during
		// the account recovery process
		apibase.AbortWithError(c, http.StatusBadRequest,
			fmt.Sprintf("account state change not allowed: old=%s, new=%s", oldAccountRecord.State,
				acct.State))
		return
	}

	if acct.EncryptedPassword != "" {

		// update password

		// We assume that the account and its secrets were updated holistically by the invoker.
		// Hence, we just need to re-hash the password

		err = account.ReHashPassphrase(&acct, nil)
		if err != nil {
			log.Err(err).Msg("Error when hashing password")
			apibase.AbortWithError(c, http.StatusInternalServerError, "Bad register request")
			return
		}
	} else {

		// keep existing password

		acct.EncryptedPassword = oldAccountRecord.EncryptedPassword
	}

	err = h.identityBackend.UpdateAccount(c, &acct)
	if err != nil {
		log.Err(err).Msg("Error updating account")
		apibase.AbortWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusOK)
}

type AccountPatch struct {
	Email                string `json:"email,omitempty"`
	OldEncryptedPassword string `json:"oldEncryptedPassword,omitempty"`
	NewEncryptedPassword string `json:"newEncryptedPassword,omitempty"`
	Name                 string `json:"name,omitempty"`
	GivenName            string `json:"givenName,omitempty"`
	FamilyName           string `json:"familyName,omitempty"`
}

func (h *AccountHandler) PatchAccountHandler(c *gin.Context) {
	log := apibase.CtxLogger(c)

	buf, _ := c.GetRawData()

	var patch AccountPatch
	err := jsonw.Unmarshal(buf, &patch)
	if err != nil {
		log.Err(err).Str("body", string(buf)).Msg("Bad account update request")
		apibase.AbortWithError(c, http.StatusBadRequest, "bad account")
		return
	}

	acct, err := h.identityBackend.GetAccount(c, apibase.GetUserID(c))
	if err != nil {
		log.Err(err).Msg("Error retrieving account")
		apibase.AbortWithError(c, http.StatusInternalServerError, "Error retrieving account")
		return
	}

	if patch.Email != "" {
		acct.Email = strings.ToLower(patch.Email)
	}

	if patch.NewEncryptedPassword != "" {

		if acct.AccessLevel != model.AccessLevelManaged {
			apibase.AbortWithError(c, http.StatusBadRequest,
				"passphrase change for hosted accounts via PATCH not allowed")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(acct.EncryptedPassword), []byte(patch.OldEncryptedPassword)); err != nil {
			apibase.AbortWithError(c, http.StatusUnauthorized,
				"Old passphrase is invalid")
			return
		}

		// update password

		acct, err = account.ChangePassphrase(acct, patch.OldEncryptedPassword, patch.NewEncryptedPassword, true)
		if err != nil {
			log.Err(err).Msg("Error when changing password")
			apibase.AbortWithError(c, http.StatusBadRequest, "Bad account update request")
		}

		err = account.ReHashPassphrase(acct, nil)
		if err != nil {
			log.Err(err).Msg("Error when hashing password")
			apibase.AbortWithError(c, http.StatusBadRequest, "Bad account update request")
			return
		}
	}

	if patch.Name != "" {
		acct.Name = patch.Name
	}

	if patch.GivenName != "" {
		acct.GivenName = patch.GivenName
	}

	if patch.FamilyName != "" {
		acct.FamilyName = patch.FamilyName
	}

	err = h.identityBackend.UpdateAccount(c, acct)
	if err != nil {
		if errors.Is(err, storage.ErrAccountExists) {
			apibase.AbortWithError(c, http.StatusBadRequest, "Error updating account")
		} else {
			log.Err(err).Msg("Error updating account")
			apibase.AbortWithError(c, http.StatusInternalServerError, "Error updating account")
		}
		return
	}

	c.Status(http.StatusOK)
}

func (h *AccountHandler) DeleteAccountHandler(c *gin.Context) {
	masterAccountID := apibase.GetUserID(c)
	accountID := c.Params.ByName("aid")

	if !hasAccountPermissions(c, h.identityBackend, masterAccountID, accountID) {
		return
	}

	acct, err := h.identityBackend.GetAccount(c, accountID)
	if err != nil {
		log := apibase.CtxLogger(c)
		log.Err(err).Msg("Error retrieving account")
		apibase.AbortWithError(c, http.StatusInternalServerError, "Error retrieving account")
		return
	}

	if acct.AccessLevel == model.AccessLevelManaged {
		apibase.AbortWithError(c, http.StatusForbidden, "can't delete managed account")
		return
	}

	err = h.identityBackend.DeleteAccount(c, accountID)
	if err != nil {
		apibase.AbortWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusOK)
}

func (h *AccountHandler) GetSubAccountListHandler(c *gin.Context) {
	state := c.Query("state")

	masterAccountID := apibase.GetUserID(c)
	accountID := c.Params.ByName("aid")

	if !hasAccountPermissions(c, h.identityBackend, masterAccountID, accountID) {
		return
	}

	accounts, err := h.identityBackend.ListAccounts(c, accountID, state)
	if err != nil {
		log := apibase.CtxLogger(c)
		log.Err(err).Msg("Error when reading sub-account list")
		apibase.AbortWithError(c, http.StatusInternalServerError, err.Error())
		return
	}
	apibase.JSON(c, http.StatusOK, accounts)
}

func (h *AccountHandler) PostSubAccountHandler(c *gin.Context) {
	log := apibase.CtxLogger(c)

	buf, _ := c.GetRawData()

	var acct account.Account
	err := jsonw.Unmarshal(buf, &acct)
	if err != nil {
		log.Err(err).Str("body", string(buf)).Msg("Bad sub-account creation request")
		apibase.AbortWithError(c, http.StatusBadRequest, "bad sub-account")
		return
	}

	if acct.ParentAccount == "" {
		apibase.AbortWithError(c, http.StatusBadRequest, "empty parent account")
		return
	}

	acct.State = account.StateActive

	if err := acct.Validate(); err != nil {
		log.Err(err).Str("body", string(buf)).Msg("Account validation failed")
		apibase.AbortWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	masterAccountID := apibase.GetUserID(c)

	if !hasAccountPermissions(c, h.identityBackend, masterAccountID, acct.ParentAccount) {
		log.Err(err).Msg("Attempted to create sub-account for invalid parent account")
		return
	}

	// set registration time
	registeredAt := time.Now()
	acct.RegisteredAt = &registeredAt

	err = h.identityBackend.CreateAccount(c, &acct)

	if err != nil {
		log.Err(err).Msg("Error when registering account")
		if errors.Is(err, storage.ErrAccountExists) {
			apibase.AbortWithError(c, http.StatusConflict, "Account already exists")
		} else {
			apibase.AbortWithError(c, http.StatusBadRequest, "Account creation failed")
		}
		return
	}

	apibase.JSON(c, http.StatusCreated, acct)
}

func hasAccountPermissions(c *gin.Context, backend storage.IdentityBackend, masterAccountID, accountID string) bool {
	hasAccess, err := backend.HasAccountAccess(c, masterAccountID, accountID)
	if err != nil {
		log := apibase.CtxLogger(c)
		if errors.Is(err, storage.ErrAccountNotFound) {
			apibase.AbortWithError(c, http.StatusNotFound, "account not found")
		} else {
			log.Err(err).Msg("Error when checking account access permissions")
			apibase.AbortWithError(c, http.StatusInternalServerError, err.Error())
		}

		return false
	}

	if !hasAccess {
		log := apibase.CtxLogger(c)
		log.Warn().Str("id", accountID).Msg("Attempted unauthorised account access")
		// we return 404 as a security mitigation measure, to avoid
		// leaking the information that the desired account exists but the caller
		// doesn't have permissions
		apibase.AbortWithError(c, http.StatusNotFound, "account not found")
		return false
	}

	return true
}

func (h *AccountHandler) GetAccountHandler(c *gin.Context) {
	masterAccountID := apibase.GetUserID(c)
	accountID := c.Params.ByName("aid")

	if !hasAccountPermissions(c, h.identityBackend, masterAccountID, accountID) {
		return
	}

	acct, err := h.identityBackend.GetAccount(c, accountID)
	if err != nil {
		log := apibase.CtxLogger(c)
		log.Err(err).Msg("Error when retrieving account details")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	acct.EncryptedPassword = ""

	apibase.JSON(c, http.StatusOK, acct)
}
