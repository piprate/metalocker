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

package main

import (
	"os"
	"time"

	"github.com/piprate/metalocker/cmd"
	"github.com/piprate/metalocker/cmd/metalo/actions"
	"github.com/piprate/metalocker/contexts"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "metalo"
	app.Usage = "A CLI tool for MetaLocker"
	app.Version = actions.MetaloVersion

	app.Flags = actions.StandardFlags

	app.Before = func(c *cli.Context) error {
		cmd.SetConfigDirName(".metalo")

		if c.Bool("debug") {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.Stamp})

		// pre-cache JSON-LD context to speed up processing
		_ = contexts.PreloadContextsIntoMemory()

		return nil
	}

	app.After = func(c *cli.Context) error {
		return nil
	}

	app.Commands = actions.StandardSet

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("CLI command failed")
	}
}
