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
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/knadh/koanf"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/sdk/cmdbase"
	"github.com/piprate/metalocker/storage"
	"github.com/piprate/restgate"
	"github.com/rs/zerolog/log"
)

type apiKeyStruct struct {
	Key    any `json:"key"`
	UserID any `json:"userID"`
	Secret any `json:"secret"`
}

func NewStaticAPIKeyAuthenticationHandler(ctx context.Context, cfg *koanf.Koanf, key string, mainAuthFunc gin.HandlerFunc,
	resolver cmdbase.ParameterResolver, identityBackend storage.IdentityBackend) (gin.HandlerFunc, error) {

	var keyDefs []*apiKeyStruct
	if err := cfg.Unmarshal(key, &keyDefs); err != nil {
		log.Err(err).Msg("Failed to read API keys")
		return nil, ErrBadConfiguration
	}

	keys := make([]string, 0)
	secrets := make([]string, 0)
	userKeyMap := make(map[string]string)

	for _, k := range keyDefs {
		accountID, err := resolver.ResolveString(k.UserID)
		if err != nil {
			log.Err(err).Msg("Failed to resolve API key's userID")
			return nil, ErrBadConfiguration
		}
		acct, err := identityBackend.GetAccount(ctx, accountID)
		if err != nil {
			if !errors.Is(err, storage.ErrAccountNotFound) {
				log.Err(err).Msg("Error when retrieving account")
				return nil, ErrBadConfiguration
			}

			log.Error().Str("id", accountID).Msg("Account doesn't exist for API key")
			continue
		}

		if acct.State != account.StateActive {
			log.Warn().Str("id", accountID).Msg("Skipping API key for non-active account")
			continue
		}

		// account ID can be either an email of the actual account ID
		accountID = acct.ID

		key, err := resolver.ResolveString(k.Key)
		if err != nil {
			log.Err(err).Msg("Failed to resolve API key")
			return nil, ErrBadConfiguration
		}

		secret, err := resolver.ResolveString(k.Secret)
		if err != nil {
			log.Err(err).Msg("Failed to resolve API secret")
			return nil, ErrBadConfiguration
		}

		keys = append(keys, key)
		secrets = append(secrets, secret)
		userKeyMap[key] = accountID
	}

	// Initialize Restgate
	rg := restgate.New("X-Auth-Key", "X-Auth-Secret", restgate.Static, restgate.Config{
		Key:    keys,
		Secret: secrets,
		Context: func(r *http.Request, authenticatedKey string) {
			r.Header.Set("X-Restgate-Key", authenticatedKey)
		},
		Logger: &log.Logger,
	})

	log.Info().Int("keyCount", len(keys)).Msg("Initialised Restgate API Key middleware")

	// Create Gin middleware - integrate Restgate with Gin
	return func(c *gin.Context) {
		authKey := c.GetHeader("X-Auth-Key")
		if authKey != "" {
			nextCalled := false
			nextAdapter := func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				userID := userKeyMap[r.Header.Get("X-Restgate-Key")]
				c.Set(UserIDKey, userID)
				c.Next()
			}
			rg.ServeHTTP(c.Writer, c.Request, nextAdapter)
			if !nextCalled {
				c.AbortWithStatus(http.StatusUnauthorized)
			}
		} else {
			mainAuthFunc(c)
		}
	}, nil
}
