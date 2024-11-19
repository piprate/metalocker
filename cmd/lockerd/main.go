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
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/piprate/metalocker/cmd"
	"github.com/piprate/metalocker/node"
	"github.com/piprate/metalocker/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

const (
	Version = "1.0.0"
)

func main() {
	app := cli.NewApp()
	app.Name = "lockerd"
	app.Usage = "MetaLocker Server"
	app.Version = Version

	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "if true, enable debug mode",
		},
		&cli.StringFlag{
			Name:  "config",
			Value: "config",
			Usage: "config name (will use $HOME/.metalocker/{name}.[json|yaml|...] config file)",
		},
	}
	app.Before = func(c *cli.Context) error {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp})

		return nil
	}

	app.Commands = []*cli.Command{
		{
			Name:   "init",
			Usage:  "initialize a new server configuration",
			Action: InitialiseCommand,
		},
	}

	app.Action = RunServer

	if err := app.Run(os.Args); err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func InitialiseCommand(c *cli.Context) error {
	configName := c.String("config")
	configDir := utils.AbsPathify(cmd.GetMetaLockerConfigDir())

	if err := node.SafeWriteConfigToFile(configDir, configName); err != nil {
		return cli.Exit(err, 1)
	}

	return nil
}

var cfg = koanf.New(".")

func RunServer(c *cli.Context) error {

	// read configuration

	configName := c.String("config")
	configDir := utils.AbsPathify(cmd.GetMetaLockerConfigDir())

	err := cfg.Load(
		file.Provider(
			filepath.Join(configDir, fmt.Sprintf("%s.yaml", configName)),
		),
		yaml.Parser(),
	)
	if err != nil {
		return err
	}

	// start MetaLocker server

	srv := node.NewMetaLockerServer(configDir)
	if err = srv.InitServices(cfg, c.Bool("debug")); err != nil {
		return err
	}

	if err = srv.InitAuthentication(c.Context, cfg); err != nil {
		return err
	}

	if err = srv.InitStandardRoutes(cfg); err != nil {
		return err
	}

	if err = srv.Run(cfg); err != nil {
		return err
	}

	return nil
}
