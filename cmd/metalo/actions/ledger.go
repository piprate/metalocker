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
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func ExportLedger(c *cli.Context) error {
	if c.Args().Len() != 1 {
		fmt.Print("Please specify the path to file or folder.\n\n")
		return cli.Exit("please specify the path to file or folder", InvalidParameter)
	}

	dw, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	err = operations.ExportLedger(dw.Services().Ledger(), dw.Services().OffChainStorage(), c.Args().Get(0))
	if err != nil {
		log.Err(err).Msg("Ledger export failed")
		return cli.Exit(err, OperationFailed)
	}
	return err
}

func ImportLedger(c *cli.Context) error {
	if c.Args().Len() != 1 {
		fmt.Print("Please specify the path to file or folder.\n\n")
		return cli.Exit("please specify the path to file or folder", InvalidParameter)
	}

	importOperations := c.Bool("import-operations")
	waitForConfirmation := c.Bool("wait")

	dw, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	ns, err := dw.Services().NotificationService()
	if err != nil {
		return err
	}

	err = operations.ImportLedger(dw.Services().Ledger(), dw.Services().OffChainStorage(), ns, c.Args().Get(0), importOperations, waitForConfirmation)
	if err != nil {
		log.Err(err).Msg("Ledger import failed")
		return cli.Exit(err, OperationFailed)
	}
	return nil
}
