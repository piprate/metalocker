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

import (
	"fmt"

	"github.com/piprate/metalocker/cmd/metalo/operations"
	"github.com/piprate/metalocker/model/account"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func ExportAccounts(c *cli.Context) error {
	if c.Args().Len() != 1 {
		fmt.Print("Please specify the path to file or folder.\n\n")
		return cli.Exit("please specify the path to file or folder", InvalidParameter)
	}

	mlc, err := CreateAdminHTTPCaller(c)
	if err != nil {
		log.Err(err).Msg("Connection to MetaLocker failed")
		return cli.Exit("connection to MetaLocker failed", OperationFailed)
	}

	err = operations.ExportAccounts(mlc, c.Args().Get(0))
	if err != nil {
		log.Err(err).Msg("Accounts export failed")
		return cli.Exit(err, OperationFailed)
	}
	return nil
}

func ImportAccounts(c *cli.Context) error {
	if c.Args().Len() != 1 {
		fmt.Print("Please specify the path to file or folder.\n\n")
		return cli.Exit("please specify the path to file or folder", InvalidParameter)
	}

	mlc, err := CreateAdminHTTPCaller(c)
	if err != nil {
		log.Err(err).Msg("Connection to MetaLocker failed")
		return cli.Exit("connection to MetaLocker failed", OperationFailed)
	}

	err = operations.ImportAccounts(mlc, c.Args().Get(0))
	if err != nil {
		log.Err(err).Msg("Accounts import failed")
		return cli.Exit(err, OperationFailed)
	}

	return nil
}

func UpdateAccountState(c *cli.Context) error {
	if c.Args().Len() != 1 {
		fmt.Print("Please specify the account ID (email or DID) to block.\n\n")
		return cli.Exit("please specify the account ID", InvalidParameter)
	}

	var targetState string

	if c.Bool("lock") {
		targetState = account.StateSuspended
	} else if c.Bool("unlock") {
		targetState = account.StateActive
	} else if c.Bool("delete") {
		targetState = account.StateDeleted
	} else {
		return cli.Exit("new state not provided. Use either --lock or --unlock or --delete", InvalidParameter)
	}

	mlc, err := CreateAdminHTTPCaller(c)
	if err != nil {
		log.Err(err).Msg("Connection to MetaLocker failed")
		return cli.Exit("connection to MetaLocker failed", OperationFailed)
	}

	err = operations.UpdateAccountState(mlc, c.Args().Get(0), targetState)
	if err != nil {
		log.Err(err).Msg("Account state update failed")
		return cli.Exit(err, OperationFailed)
	}
	return err
}
