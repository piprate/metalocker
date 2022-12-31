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
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type LoggerConfig struct {
	Logger         *zerolog.Logger
	HideUserID     bool
	SkipPath       []string
	SkipPathRegexp *regexp.Regexp
}

// SetRequestLogger initializes the logging middleware.
// This implementation was borrowed from github.com/gin-contrib/logger.
func SetRequestLogger(config ...LoggerConfig) gin.HandlerFunc {
	var newConfig LoggerConfig
	if len(config) > 0 {
		newConfig = config[0]
	}
	var skip map[string]struct{}
	if length := len(newConfig.SkipPath); length > 0 {
		skip = make(map[string]struct{}, length)
		for _, path := range newConfig.SkipPath {
			skip[path] = struct{}{}
		}
	}

	var sublog zerolog.Logger
	if newConfig.Logger == nil {
		sublog = log.Logger
	} else {
		sublog = *newConfig.Logger
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		c.Next()
		track := true

		if _, ok := skip[path]; ok {
			track = false
		}

		if track &&
			newConfig.SkipPathRegexp != nil &&
			newConfig.SkipPathRegexp.MatchString(path) {
			track = false
		}

		if track {
			end := time.Now()
			latency := end.Sub(start)

			msg := "Request"
			if len(c.Errors) > 0 {
				msg = c.Errors.String()
			}

			loggerCtx := sublog.With().
				Int("status", c.Writer.Status()).
				Str("method", c.Request.Method).
				Str("path", path).
				Str("ip", c.ClientIP()).
				Dur("latency", latency).
				Str("user-agent", c.Request.UserAgent())

			if !newConfig.HideUserID {
				loggerCtx = loggerCtx.Str("uid", GetUserID(c))
			}

			dumplogger := loggerCtx.Logger()

			switch {
			case c.Writer.Status() >= http.StatusBadRequest && c.Writer.Status() < http.StatusInternalServerError:
				dumplogger.Warn().Msg(msg)
			case c.Writer.Status() >= http.StatusInternalServerError:
				dumplogger.Error().Msg(msg)
			default:
				dumplogger.Info().Msg(msg)
			}
		}

	}
}
