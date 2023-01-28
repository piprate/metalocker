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
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/knadh/koanf"
	"github.com/piprate/metalocker/utils/grayzero"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func SetupLogging(cfg *koanf.Koanf, prodMode bool) (io.Closer, error) {
	if prodMode {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		gin.SetMode(gin.ReleaseMode)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		gin.SetMode(gin.DebugMode)
	}

	var logWriter io.WriteCloser
	var err error
	if cfg.Exists("logging") {
		consoleFormat := cfg.String("logging.consoleFormat")
		switch consoleFormat {
		case "pretty":
			logWriter = grayzero.NopWriteCloser(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp})
		case "json":
			logWriter = grayzero.NopWriteCloser(os.Stdout)
		}
		graylogURL := cfg.String("logging.graylogURL")
		if graylogURL != "" {
			serviceName := cfg.String("logging.serviceName")
			if serviceName == "" {
				serviceName = "metalocker"
			}
			logWriter, err = grayzero.NewGraylogWriter(
				graylogURL,
				logWriter,
				serviceName,
				cfg.String("logging.instance"),
			)
			if err != nil {
				return nil, err
			}
		}
	} else {
		// default log configuration
		logWriter = grayzero.NopWriteCloser(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp})
	}

	log.Logger = log.Output(logWriter)

	return logWriter, nil
}
