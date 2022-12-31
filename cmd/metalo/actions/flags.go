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

package actions

import "github.com/urfave/cli/v2"

var (
	BasicFlags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "if true, enable debug mode",
		},
		&cli.StringFlag{
			Name:    "server",
			Value:   DefaultMetaLockerURL,
			Usage:   "MetaLocker url, i.e. " + DefaultMetaLockerURL,
			EnvVars: []string{"METASERVER"},
		},
		&cli.StringFlag{
			Name:    "user",
			Value:   "",
			Usage:   "account email address",
			EnvVars: []string{"METAUSER"},
		},
		&cli.StringFlag{
			Name:    "password",
			Value:   "",
			Usage:   "account password",
			EnvVars: []string{"METAPASS"},
		},
	}

	APIKeyFlags = []cli.Flag{
		&cli.StringFlag{
			Name:    "api-key",
			Value:   "",
			Usage:   "Account API key",
			EnvVars: []string{"APIKEY"},
		},
		&cli.StringFlag{
			Name:    "api-secret",
			Value:   "",
			Usage:   "Account API secret",
			EnvVars: []string{"APISECRET"},
		},
	}

	AdminFlags = []cli.Flag{
		&cli.StringFlag{
			Name:    "admin-key",
			Value:   "",
			Usage:   "(for admin operations) API key",
			EnvVars: []string{"ADMKEY"},
		},
		&cli.StringFlag{
			Name:    "admin-secret",
			Value:   "",
			Usage:   "(for admin operations) API secret",
			EnvVars: []string{"ADMSECRET"},
		},
	}

	StandardFlags []cli.Flag
)

func init() {
	StandardFlags = make([]cli.Flag, len(BasicFlags)+len(AdminFlags)+len(AdminFlags))
	StandardFlags = append(BasicFlags, append(AdminFlags, APIKeyFlags...)...)
}
