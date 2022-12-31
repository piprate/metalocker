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
	"github.com/piprate/metalocker/model"
	"github.com/urfave/cli/v2"
)

func GenerateSubAccountAccessKey(c *cli.Context) error {
	if c.Args().Len() != 1 {
		return cli.Exit("please specify sub-account ID", InvalidParameter)
	}

	dw, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	dw, err = dw.GetSubAccountWallet(c.Args().Get(0))
	if err != nil {
		return err
	}

	ak, err := dw.CreateAccessKey(model.AccessLevelHosted, time.Hour*24*365)
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

func ListSubAccountAccessKeys(c *cli.Context) error {
	if c.Args().Len() != 1 {
		return cli.Exit("please specify sub-account ID", InvalidParameter)
	}

	dw, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	dw, err = dw.GetSubAccountWallet(c.Args().Get(0))
	if err != nil {
		return err
	}

	accessKeys, err := dw.AccessKeys()
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

func GetSubAccount(c *cli.Context) error {
	if c.Args().Len() != 1 {
		return cli.Exit("please specify sub-account ID", InvalidParameter)
	}

	dw, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	acct, err := dw.GetSubAccount(c.Args().Get(0))
	if err != nil {
		return cli.Exit(err.Error(), OperationFailed)
	}

	ld.PrintDocument("", acct)

	return nil
}

func GetSubAccountAccessKey(c *cli.Context) error {
	if c.Args().Len() != 2 {
		return cli.Exit("please specify sub-account ID and key ID", InvalidParameter)
	}

	dw, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	dw, err = dw.GetSubAccountWallet(c.Args().Get(0))
	if err != nil {
		return err
	}

	ak, err := dw.GetAccessKey(c.Args().Get(1))
	if err != nil {
		return cli.Exit(err.Error(), OperationFailed)
	}

	ld.PrintDocument("", ak)

	return nil
}

func DeleteSubAccountAccessKey(c *cli.Context) error {
	if c.Args().Len() != 2 {
		return cli.Exit("please specify sub-account ID and key ID", InvalidParameter)
	}

	dw, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	dw, err = dw.GetSubAccountWallet(c.Args().Get(0))
	if err != nil {
		return err
	}

	return dw.RevokeAccessKey(c.Args().Get(1))
}

func CreateSubAccount(c *cli.Context) error {
	dw, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	if c.String("sa-email") != "" || c.String("sa-password") != "" {
		return cli.Exit("email and password options aren't yet supported", OperationFailed)
	}

	subDW, err := dw.CreateSubAccount(model.AccessLevelHosted, c.String("name"))
	if err != nil {
		return cli.Exit(err.Error(), OperationFailed)
	}

	res := map[string]any{
		"account": subDW.Account(),
	}

	if c.Bool("new-key") {
		ak, err := subDW.CreateAccessKey(model.AccessLevelHosted, time.Hour*24*365)
		if err != nil {
			return err
		}

		id, secrets := ak.ClientKeys()

		res["key"] = map[string]any{
			"id":     id,
			"secret": secrets,
		}
	}

	ld.PrintDocument("", res)

	return nil
}

func DeleteSubAccount(c *cli.Context) error {
	return cli.Exit("operation not yet supported", OperationFailed)
}

func ListSubAccounts(c *cli.Context) error {
	dataWallet, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	subAccounts, err := dataWallet.SubAccounts()
	if err != nil {
		return cli.Exit(err.Error(), OperationFailed)
	}

	sort.Slice(subAccounts, func(i, j int) bool {
		return subAccounts[i].RegisteredAt.Before(*subAccounts[j].RegisteredAt)
	})

	tf := "2006-01-02 15:04:05-07:00"
	data := make([][]string, 0)
	for _, acct := range subAccounts {
		data = append(data, []string{acct.ID, acct.Name, acct.RegisteredAt.Format(tf)})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Sub-Account ID", "Name", "Registered At"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data) // Add Bulk Data
	table.Render()

	return nil
}
