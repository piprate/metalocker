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

package apibase

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/storage"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/utils/security"
	"github.com/rs/zerolog/log"
)

type AccessKeyPersister interface {
	GetAccessKey(ctx context.Context, keyID string) (*model.AccessKey, error)
}

func AccessKeyMiddleware(accessKeyStorage AccessKeyPersister, next gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.Request
		keyID, sig, err := model.ExtractSignature(req.Header)
		if err == nil {
			key, err := accessKeyStorage.GetAccessKey(c, keyID)
			if err != nil {
				if errors.Is(err, storage.ErrAccessKeyNotFound) {
					c.AbortWithStatus(http.StatusUnauthorized)
				} else {
					log.Err(err).Msg("Error reading access key")
					c.AbortWithStatus(http.StatusInternalServerError)
				}
				return
			}

			var bodyHash []byte
			if c.Request.Header.Get(model.AccessKeyHeaderBodyHash) != "" {
				bodyBytes, err := io.ReadAll(req.Body)
				if err != nil {
					log.Err(err).Msg("Error reading request body")
					c.AbortWithStatus(http.StatusBadRequest)
					return
				}
				_ = req.Body.Close() //  must close
				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				bodyHash = model.HashRequestBody(bodyBytes)
			}

			reqTime := time.Now()
			url := c.Request.URL.RequestURI()

			encryptedHMACKey, err := base64.StdEncoding.DecodeString(key.Secret)
			if err != nil {
				log.Err(err).Msg("Error decoding access key secret")
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			valid, err := model.ValidateRequest(c.Request.Header, sig, encryptedHMACKey, reqTime, url, bodyHash)
			if err != nil {
				log.Warn().AnErr("err", err).Msg("Error validating request signature")
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			if !valid {
				log.Warn().Str("url", url).Msg("Invalid request signature")
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			c.Set(UserIDKey, key.AccountID)
			c.Next()
		} else if errors.Is(err, model.ErrAuthorizationNotFound) {
			next(c)
		} else {
			log.Warn().AnErr("err", err).Msg("Error processing Authorization header")
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

type SignatureValidationRequest struct {
	URL       string      `json:"url"`
	KeyID     string      `json:"key"`
	Signature string      `json:"sig"`
	Header    http.Header `json:"hdr"`
	Timestamp int64       `json:"ts"`
	BodyHash  string      `json:"hash,omitempty"`
}

type SignatureValidationResponse struct {
	Account string `json:"acct"`
}

// DelegatedAccessKeyMiddleware call an external signature validation service to confirm
// if the signature is valid. This is useful when a service doesn't have access to identity backend.
func DelegatedAccessKeyMiddleware(validatorURL string, next gin.HandlerFunc) (gin.HandlerFunc, error) {

	validatorURL, httpClient, _, err := security.CreateHTTPClient(validatorURL)
	if err != nil {
		return nil, err
	}

	return func(c *gin.Context) {
		req := c.Request
		keyID, sig, err := model.ExtractSignature(req.Header)
		if err == nil {
			filteredHeader := http.Header{}
			for k, vals := range req.Header {
				if strings.HasPrefix(k, "X-Meta-") {
					for _, v := range vals {
						filteredHeader.Add(k, v)
					}
				}
			}

			reqBody := SignatureValidationRequest{
				URL:       c.Request.URL.RequestURI(),
				KeyID:     keyID,
				Signature: sig,
				Header:    filteredHeader,
				Timestamp: time.Now().UTC().Unix(),
			}

			if c.Request.Header.Get(model.AccessKeyHeaderBodyHash) != "" {
				bodyBytes, err := io.ReadAll(req.Body)
				if err != nil {
					log.Err(err).Msg("Error reading request body")
					c.AbortWithStatus(http.StatusBadRequest)
					return
				}
				_ = req.Body.Close() //  must close
				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				reqBody.BodyHash = base64.StdEncoding.EncodeToString(model.HashRequestBody(bodyBytes))
			}

			vReqBytes, err := jsonw.Marshal(reqBody)
			if err != nil {
				log.Err(err).Msg("Error marshalling validation request")
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			vReq, err := http.NewRequestWithContext(c, http.MethodPost, validatorURL, bytes.NewReader(vReqBytes))
			if err != nil {
				log.Err(err).Msg("Error building validation request")
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			res, err := httpClient.Do(vReq)
			if err != nil {
				log.Err(err).Msg("Error sending validation request")
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			defer res.Body.Close()

			switch res.StatusCode {
			case http.StatusOK:
				decoder := json.NewDecoder(res.Body)
				var rspStruct SignatureValidationResponse
				err := decoder.Decode(&rspStruct)
				if err != nil {
					log.Err(err).Msg("Error parsing validation response")
					c.AbortWithStatus(http.StatusInternalServerError)
					return
				}

				c.Set(UserIDKey, rspStruct.Account)
				c.Next()
			case http.StatusUnauthorized:
				log.Warn().Str("url", reqBody.URL).Msg("Invalid request signature")
				c.AbortWithStatus(http.StatusUnauthorized)
			default:
				msg := ParseResponseMessage(res)
				log.Error().Str("url", reqBody.URL).Str("msg", msg).Msg("Validation call failed")
				c.AbortWithStatus(http.StatusUnauthorized)
			}
		} else if errors.Is(err, model.ErrAuthorizationNotFound) {
			next(c)
		} else {
			log.Warn().AnErr("err", err).Msg("Error processing Authorization header")
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}, nil
}
