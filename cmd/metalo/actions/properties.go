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
	"os"
	"sort"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
)

func SetProperty(c *cli.Context) error {
	dataWallet, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	if c.Args().Len() != 2 {
		return cli.Exit("please specify property key and value", InvalidParameter)
	}

	return dataWallet.SetProperty(c.Context, c.Args().Get(0), c.Args().Get(1), dataWallet.Account().AccessLevel)
}

func ListProperties(c *cli.Context) error {
	dataWallet, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	props, err := dataWallet.GetProperties(c.Context)
	if err != nil {
		return cli.Exit(err.Error(), OperationFailed)
	}

	keys := make([]string, len(props))
	i := 0
	for key := range props {
		keys[i] = key
		i++
	}

	sort.Strings(keys)

	data := make([][]string, len(keys))
	for i, key := range keys {
		data[i] = []string{key, props[key]}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Key", "Value"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data) // Add Bulk Data
	table.Render()

	return nil
}

func GetProperty(c *cli.Context) error {
	if c.Args().Len() != 1 {
		return cli.Exit("please specify property key", InvalidParameter)
	}

	dw, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	val, err := dw.GetProperty(c.Context, c.Args().Get(0))
	if err != nil {
		return cli.Exit(err.Error(), OperationFailed)
	}

	fmt.Println(val)

	return nil
}

func DeleteProperty(c *cli.Context) error {
	if c.Args().Len() != 1 {
		return cli.Exit("please specify property key", InvalidParameter)
	}

	dw, err := LoadRemoteDataWallet(c, false)
	if err != nil {
		return err
	}

	return dw.DeleteProperty(c.Context, c.Args().Get(0), dw.Account().AccessLevel)
}
