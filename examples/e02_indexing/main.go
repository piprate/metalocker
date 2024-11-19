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
	"context"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/piprate/metalocker/examples"
	"github.com/piprate/metalocker/index"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/model/dataset"
	"github.com/piprate/metalocker/model/expiry"
	"github.com/piprate/metalocker/remote"
	"github.com/piprate/metalocker/utils"
)

// This example shows how a user can build an index and use it to traverse all records in their data wallet.
//
// It runs a fully featured demo server. We will use a remote client to connect to the demo server via HTTPS.
func main() {
	tempDir, err := os.MkdirTemp(".", "tempdir_")
	if err != nil {
		panic(err)
	}
	tempDir = utils.AbsPathify(tempDir)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// create an in-memory MetaLocker node

	srv, url, err := examples.StartDemoServer(tempDir, false)
	if err != nil {
		panic(err)
	}
	defer srv.Close()

	// create remote client

	factory, err := remote.NewWalletFactory(url, examples.WithDemoIndexStore(tempDir), 0)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// create an account

	passPhrase := "passw0rd!"

	dataWallet, _, err := factory.RegisterAccount(
		ctx,
		&account.Account{
			Email:       "jeditemple@example.com",
			Name:        "Jedi Temple",
			AccessLevel: model.AccessLevelHosted,
		},
		passPhrase,
	)
	if err != nil {
		panic(err)
	}

	// the wallet needs to be unlocked before use

	err = dataWallet.Unlock(ctx, passPhrase)
	if err != nil {
		panic(err)
	}

	// create an identity

	kenobi, err := dataWallet.NewIdentity(ctx, model.AccessLevelHosted, "Obi-Wan Kenobi")
	if err != nil {
		panic(err)
	}

	// create a locker

	locker, err := kenobi.NewLocker(ctx, "Quotes")
	if err != nil {
		panic(err)
	}

	// publish several JSON documents

	lb, err := locker.NewDataSetBuilder(ctx, dataset.WithVault("local"))
	if err != nil {
		panic(err)
	}

	_, err = lb.AddMetaResource(map[string]any{
		"type": "TextQuote",
		"text": "Use The Force, Luke. Let Go.",
	})
	if err != nil {
		panic(err)
	}

	future1 := lb.Submit(expiry.FromNow("1h")) // the lease will expire in 1 hour
	err = future1.Wait(time.Second)
	if err != nil {
		panic(err)
	}

	lb, err = locker.NewDataSetBuilder(ctx, dataset.WithVault("local"))
	if err != nil {
		panic(err)
	}

	_, err = lb.AddMetaResource(map[string]any{
		"type": "MarkdownQuote",
		"text": "*The Negotiations Were Short.*",
	})
	if err != nil {
		panic(err)
	}

	future2 := lb.Submit(expiry.FromNow("1h"))
	err = future2.Wait(time.Second)
	if err != nil {
		panic(err)
	}

	// create an index

	rootIndex, err := dataWallet.CreateRootIndex(ctx, examples.DemoIndexStoreName)
	if err != nil {
		panic(err)
	}

	// sync the index with the ledger

	updater, err := dataWallet.IndexUpdater(ctx, rootIndex)
	if err != nil {
		panic(err)
	}
	defer updater.Close()

	err = updater.Sync(ctx)
	if err != nil {
		panic(err)
	}

	// traverse all records

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Content Type", "Status"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")

	err = rootIndex.TraverseRecords(ctx, "", "", func(r *index.RecordState) error {
		table.Append([]string{r.ID, r.ContentType, string(r.Status)})
		return nil
	}, 0)

	table.Render()
}
