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
	"os"
	"sort"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/piprate/json-gold/ld"
	"github.com/urfave/cli/v2"
)

func GenerateAccessKey(c *cli.Context) error {
	dataWallet, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	ak, err := dataWallet.CreateAccessKey(dataWallet.Account().AccessLevel, time.Hour*24*365)
	if err != nil {
		return err
	}

	id, secrets := ak.ClientKeys()

	res := map[string]any{
		"id":     id,
		"secret": secrets,
	}

	ld.PrintDocument("", res)

	return nil
}

func ListAccessKeys(c *cli.Context) error {
	dataWallet, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	accessKeys, err := dataWallet.AccessKeys()
	if err != nil {
		return cli.Exit(err.Error(), OperationFailed)
	}

	sort.Slice(accessKeys, func(i, j int) bool {
		return accessKeys[i].ID < accessKeys[j].ID
	})

	data := make([][]string, 0)
	for _, key := range accessKeys {
		data = append(data, []string{key.ID})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data) // Add Bulk Data
	table.Render()

	return nil
}

func GetAccessKey(c *cli.Context) error {
	if c.Args().Len() != 1 {
		return cli.Exit("please specify key ID", InvalidParameter)
	}

	dw, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	ak, err := dw.GetAccessKey(c.Args().Get(0))
	if err != nil {
		return cli.Exit(err.Error(), OperationFailed)
	}

	ld.PrintDocument("", ak)

	return nil
}

func DeleteAccessKey(c *cli.Context) error {
	if c.Args().Len() != 1 {
		return cli.Exit("please specify key ID", InvalidParameter)
	}

	dw, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	return dw.RevokeAccessKey(c.Args().Get(0))
}
