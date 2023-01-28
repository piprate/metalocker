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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/knadh/koanf"
	"github.com/piprate/metalocker/sdk/cmdbase"
	"github.com/piprate/restgate"
	"github.com/rs/zerolog/log"
)

type (
	adminKeyStruct struct {
		Key    any `koanf:"apiKey" json:"apiKey"`
		Secret any `koanf:"apiSecret" json:"apiSecret"`
	}
)

func NewAdminAuthenticationHandler(cfg *koanf.Koanf, name string, resolver cmdbase.ParameterResolver) (gin.HandlerFunc, error) {
	var adminKey adminKeyStruct
	if err := cfg.Unmarshal(name, &adminKey); err != nil {
		log.Err(err).Msg("Failed to read admin API key")
		return nil, ErrBadConfiguration
	}

	apiKey, err := resolver.ResolveString(adminKey.Key)
	if err != nil {
		log.Error().Msg("Error reading administration API key")
		return nil, ErrBadConfiguration
	}

	if apiKey == "" {
		log.Error().Msg("Administration API key not defined")
		return nil, ErrBadConfiguration
	}

	apiSecret, err := resolver.ResolveString(adminKey.Secret)
	if err != nil {
		log.Error().Msg("Error reading administration API secret")
		return nil, ErrBadConfiguration
	}

	if apiSecret == "" {
		log.Error().Msg("Administration API secret not defined")
		return nil, ErrBadConfiguration
	}

	adminAuth := restgate.New("X-Auth-Key", "X-Auth-Secret", restgate.Static, restgate.Config{
		Key: []string{
			apiKey,
		},
		Secret: []string{
			apiSecret,
		},
		Logger: &log.Logger,
	})

	// Create Gin middleware - integrate Restgate with Gin
	return func(c *gin.Context) {
		nextCalled := false
		nextAdapter := func(http.ResponseWriter, *http.Request) {
			nextCalled = true
			c.Next()
		}
		adminAuth.ServeHTTP(c.Writer, c.Request, nextAdapter)
		if !nextCalled {
			c.AbortWithStatus(401)
		}
	}, nil
}
