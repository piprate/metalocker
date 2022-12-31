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
	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	UserIDKey        = "userID"
	ClientSecretKey  = "clientSecret"
	ContextLoggerKey = "logger"
)

func GetUserID(c *gin.Context) string {
	return c.GetString(UserIDKey)
}

func GetManagedKey(c *gin.Context) *model.AESKey {
	val, exists := c.Get(ClientSecretKey)
	if exists {
		return val.(*model.AESKey)
	} else {
		return nil
	}
}

func ContextLoggerHandler(c *gin.Context) {
	logger := log.Logger.With().
		Str("ip", c.ClientIP()).
		Str("uid", GetUserID(c)).
		Logger()

	c.Set(ContextLoggerKey, &logger)
	c.Next()
}

func CtxLogger(c *gin.Context) *zerolog.Logger {
	if res, ok := c.Get(ContextLoggerKey); ok {
		return res.(*zerolog.Logger)
	} else {
		return &log.Logger
	}
}
