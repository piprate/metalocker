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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/storage"
)

func ValidateRequestSignatureHandler(identityBackend storage.IdentityBackend) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := apibase.CtxLogger(c)

		var req apibase.SignatureValidationRequest
		if err := c.BindJSON(&req); err != nil {
			log.Err(err).Msg("Error reading request body")
			apibase.AbortWithError(c, http.StatusBadRequest, "Bad request body")
			return
		}

		var bodyHash []byte
		var err error
		if req.BodyHash != "" {
			bodyHash, err = base64.StdEncoding.DecodeString(req.BodyHash)
			if err != nil {
				log.Err(err).Msg("Error decoding body hash")
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
		}

		ts := time.Unix(req.Timestamp, 0)

		key, err := identityBackend.GetAccessKey(c, req.KeyID)
		if err != nil {
			if errors.Is(err, storage.ErrAccessKeyNotFound) {
				log.Warn().Str("url", req.URL).Str("key", req.KeyID).Msg("Signature validation failed: key not found")
				apibase.AbortWithError(c, http.StatusUnauthorized, "Bad signature")
			} else {
				log.Err(err).Msg("Error reading access key")
				c.AbortWithStatus(http.StatusUnauthorized)
			}
			return
		}

		encryptedHMACKey, err := base64.StdEncoding.DecodeString(key.Secret)
		if err != nil {
			log.Err(err).Msg("Error decoding access key secret")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		valid, err := model.ValidateRequest(req.Header, req.Signature, encryptedHMACKey, ts, req.URL, bodyHash)
		if err != nil {
			log.Warn().AnErr("err", err).Msg("Error validating request signature")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if !valid {
			log.Warn().Str("url", req.URL).Msg("Signature validation failed")
			apibase.AbortWithError(c, http.StatusUnauthorized, "Bad signature")
			return
		}

		apibase.JSON(c, http.StatusOK, apibase.SignatureValidationResponse{
			Account: key.AccountID,
		})
	}
}
