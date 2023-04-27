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
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/storage"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog/log"
)

type GetRecoveryCodeResponse struct {
	Code string `json:"code"`
}

type AccountRecoveryResponse struct {
	Account *account.Account `json:"account"`
}

func GetRecoveryCodeHandler(identityBackend storage.IdentityBackend) gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.Query("email")

		if strings.Contains(email, "@") {
			// if the username is an email, transform to lower case
			email = strings.ToLower(email)
		}

		acct, err := identityBackend.GetAccount(email)
		if err != nil {
			log.Err(err).Msg("Error when retrieving account details")
			if errors.Is(err, storage.ErrAccountNotFound) {
				apibase.AbortWithError(c, http.StatusBadRequest, "bad request")
			} else {
				apibase.AbortWithError(c, http.StatusInternalServerError, "internal error")
			}
			return
		}

		if acct.RecoveryPublicKey == "" {
			apibase.AbortWithError(c, http.StatusUnauthorized, "Account doesn't support recovery")
			return
		}

		rc, err := account.NewRecoveryCode(email, 5*60)
		if err != nil {
			log.Err(err).Msg("Failed to generate recovery code")
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if err = identityBackend.CreateRecoveryCode(rc); err != nil {
			log.Err(err).Msg("Failed to persist recovery code")
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		rsp := &GetRecoveryCodeResponse{
			Code: rc.Code,
		}

		apibase.JSON(c, http.StatusOK, rsp)
	}
}

func RecoverAccountHandler(identityBackend storage.IdentityBackend) gin.HandlerFunc {
	return func(c *gin.Context) {
		buf, _ := c.GetRawData()

		var req account.RecoveryRequest
		err := jsonw.Unmarshal(buf, &req)
		if err != nil {
			log.Err(err).Str("body", string(buf)).Msg("Bad recover request body")
			apibase.AbortWithError(c, http.StatusBadRequest, "Bad register request")
			return
		}

		if strings.Contains(req.UserID, "@") {
			// if the user id is an email, transform to lower case
			req.UserID = strings.ToLower(req.UserID)
		}

		rc, err := identityBackend.GetRecoveryCode(req.RecoveryCode)
		if err != nil {
			if errors.Is(err, storage.ErrRecoveryCodeNotFound) {
				_ = c.AbortWithError(http.StatusBadRequest, err)
			} else {
				log.Err(err).Msg("Error when retrieving recovery code details")
				_ = c.AbortWithError(http.StatusInternalServerError, err)
			}
			return
		}

		if err = identityBackend.DeleteRecoveryCode(req.RecoveryCode); err != nil {
			log.Err(err).Msg("Error when deleting recovery code")
		}

		now := time.Now()
		if now.After(*rc.ExpiresAt) {
			log.Error().Str("code", req.RecoveryCode).Msg("Recovery code expired")
			apibase.AbortWithError(c, http.StatusUnauthorized, "Recovery code expired")
			return
		}

		acct, err := identityBackend.GetAccount(req.UserID)
		if err != nil {
			if errors.Is(err, storage.ErrAccountNotFound) {
				_ = c.AbortWithError(http.StatusBadRequest, err)
			} else {
				log.Err(err).Msg("Error when retrieving account details")
				_ = c.AbortWithError(http.StatusInternalServerError, err)
			}
			return
		}

		if acct.RecoveryPublicKey == "" {
			apibase.AbortWithError(c, http.StatusBadRequest, "Account doesn't support recovery")
			return
		}

		recPubKey, err := base64.StdEncoding.DecodeString(acct.RecoveryPublicKey)
		if err != nil {
			log.Err(err).Msg("Error when decoding recovery public key")
			apibase.AbortWithError(c, http.StatusBadRequest, "Bad recovery key")
			return
		}
		if !req.Valid(recPubKey) {
			log.Error().Msg("Account recovery signature incorrect")
			apibase.AbortWithError(c, http.StatusUnauthorized, "signature verification failed")
			return
		}

		if req.ManagedCryptoKey != "" {
			// perform full managed account recovery. The account will return to 'active' state

			if acct.AccessLevel != model.AccessLevelManaged {
				log.Error().Msg("Can't recover non-managed account using managed workflow")
				apibase.AbortWithError(c, http.StatusBadRequest, "Can't use managed crypto key for non-managed account")
				return
			}

			managedCryptoKeyBytes, err := base64.StdEncoding.DecodeString(req.ManagedCryptoKey)
			if err != nil {
				log.Err(err).Msg("Error when decoding managed crypto key")
				apibase.AbortWithError(c, http.StatusBadRequest, "Bad managed crypto key")
				return
			}
			managedCryptoKey := model.NewAESKey(managedCryptoKeyBytes)
			acct, err = account.RecoverManaged(acct, managedCryptoKey, req.EncryptedPassword)
			if err != nil {
				log.Err(err).Msg("Error when recovering managed account")
				apibase.AbortWithError(c, http.StatusInternalServerError, "Error when recovering managed account")
				return
			}
		} else {
			// We update the password to enable the user to log in and update the account properly,
			// including internal secrets. Until then, the recorded password will be out of sync with the secrets.
			// This is ok because we consider the password to be irretrievably lost.
			acct.EncryptedPassword = req.EncryptedPassword

			acct.State = account.StateRecovery
		}

		err = account.ReHashPassphrase(acct, nil)
		if err != nil {
			log.Err(err).Msg("Error when hashing password")
			apibase.AbortWithError(c, http.StatusBadRequest, "Bad account recovery request")
			return
		}

		err = identityBackend.UpdateAccount(acct)
		if err != nil {
			log.Err(err).Msg("Error updating account")
			apibase.AbortWithError(c, http.StatusInternalServerError, err.Error())
			return
		}

		// return account data

		acct.EncryptedPassword = ""

		rsp := &AccountRecoveryResponse{
			Account: acct,
		}
		apibase.JSON(c, http.StatusOK, rsp)
	}
}
