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
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/piprate/metalocker/examples"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/model/dataset"
	"github.com/piprate/metalocker/model/expiry"
	"github.com/piprate/metalocker/remote"
	"github.com/piprate/metalocker/utils"
)

// This example shows how a user can mark alternative representations of a dataset as 'heads' (similar to
// source code branches). A head has a well-known name that can be used to uniquely identify a specific version
// of the dataset.
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

	// publish the first version of the dataset

	assetID := model.NewAssetID("example")

	lb, err := locker.NewDataSetBuilder(
		ctx,
		dataset.WithVault("local"),
		dataset.WithAsset(assetID))
	if err != nil {
		panic(err)
	}

	_, err = lb.AddMetaResource(map[string]any{
		"type": "Quote",
		"text": "Use The Force, Darth. Let Go.",
	})
	if err != nil {
		panic(err)
	}

	future1 := lb.Submit(expiry.FromNow("1h")) // the lease will expire in 1 hour
	err = future1.Wait(time.Second)
	if err != nil {
		panic(err)
	}

	// point a 'head' for the given asset with a name 'wrong-quotes' to the published dataset

	err = locker.SetAssetHead(ctx, assetID, "wrong-quotes", future1.ID()).Wait(time.Second)
	if err != nil {
		panic(err)
	}

	// create another version of the dataset. This time, we use another way of setting a head by using
	// dataset.SetHeads() option. It saves us from having to perform two operations: publishing
	// the dataset and setting the head.

	lb, err = locker.NewDataSetBuilder(
		ctx,
		dataset.WithVault("local"),
		dataset.WithAsset(assetID),
		dataset.SetHeads("good-quotes"))
	if err != nil {
		panic(err)
	}

	_, err = lb.AddMetaResource(map[string]any{
		"type": "Quote",
		"text": "Use The Force, Luke. Let Go.",
	})
	if err != nil {
		panic(err)
	}

	future3 := lb.Submit(expiry.FromNow("1h")) // the lease will expire in 1 hour
	err = future3.Wait(time.Second)
	if err != nil {
		panic(err)
	}

	printHeads := func(checkpoint string) {
		// head IDs are unique and exist only within a specific locker
		badQuotesHeadID := locker.HeadID(ctx, assetID, "wrong-quotes")
		goodQuotesHeadID := locker.HeadID(ctx, assetID, "good-quotes")

		badQuotesHead, err := dataWallet.DataStore().AssetHead(ctx, badQuotesHeadID)
		if err != nil {
			panic(err)
		}
		goodQuotesHead, err := dataWallet.DataStore().AssetHead(ctx, goodQuotesHeadID)
		if err != nil {
			panic(err)
		}

		println()
		println(strings.ToUpper(checkpoint))
		println("=============")
		fmt.Printf("'bad-quotes'  head record: %s\n", badQuotesHead.ID())
		fmt.Printf("'good-quotes' head record: %s\n", goodQuotesHead.ID())
		println("-------------")
		println()
	}

	printHeads("Checkpoint #1")

	// create a new revision of the last version and update the head

	lb, err = locker.NewDataSetBuilder(
		ctx,
		dataset.WithParent(future3.ID(),
			"",
			dataset.CopyModeNone,
			nil,
			false),
		dataset.WithVault("local"),
		dataset.SetHeads("good-quotes"))
	if err != nil {
		panic(err)
	}

	_, err = lb.AddMetaResource(map[string]any{
		"type": "HTMLQuote",
		"text": "Use The Force, <b>Luke</b>. Let Go.",
	})
	if err != nil {
		panic(err)
	}

	future2 := lb.Submit(expiry.FromNow("1h"))
	err = future2.Wait(time.Second)
	if err != nil {
		panic(err)
	}

	// display 2 revisions

	printHeads("Checkpoint #2")
}
