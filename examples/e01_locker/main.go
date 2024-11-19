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

	"github.com/piprate/json-gold/ld"
	"github.com/piprate/metalocker/examples"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/model/dataset"
	"github.com/piprate/metalocker/model/expiry"
	"github.com/piprate/metalocker/remote"
	"github.com/piprate/metalocker/utils"
	"github.com/piprate/metalocker/wallet"
)

// This example shows how two users can create a bidirectional locker and send data to each other.
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

	// create account 1: Jack

	passPhrase := "passw0rd!"

	jackWallet, _, err := factory.RegisterAccount(
		ctx,
		&account.Account{
			Email:       "jack@example.com",
			Name:        "Jack",
			AccessLevel: model.AccessLevelHosted,
		},
		passPhrase,
	)
	if err != nil {
		panic(err)
	}

	// the wallet needs to be unlocked before use

	err = jackWallet.Unlock(ctx, passPhrase)
	if err != nil {
		panic(err)
	}

	// create account 2: Jill

	jillWallet, _, err := factory.RegisterAccount(
		ctx,
		&account.Account{
			Email:       "jill@example.com",
			Name:        "Jill",
			AccessLevel: model.AccessLevelHosted,
		},
		passPhrase,
	)
	if err != nil {
		panic(err)
	}

	err = jillWallet.Unlock(ctx, passPhrase)
	if err != nil {
		panic(err)
	}

	// create identities for Jack and Jill. An account can have an unlimited number of identities.

	jack, err := jackWallet.NewIdentity(ctx, model.AccessLevelHosted, "Jack")
	if err != nil {
		panic(err)
	}

	jill, err := jillWallet.NewIdentity(ctx, model.AccessLevelHosted, "Jill")
	if err != nil {
		panic(err)
	}

	// create a locker between Jack and Jill. It involves two steps:
	// 1. Jack creates the locker, stating Jill as the (second) participant.
	// 2. Jill imports this locker into her wallet. In real life Jack would need to share the locker with Jill
	// using a secure communication channel. MetaLocker doesn't prescribe what channel should be used.

	lockerForJack, err := jack.NewLocker(ctx, "Hill", wallet.Participant(jill.DID(), nil))
	if err != nil {
		panic(err)
	}

	lockerForJill, err := jillWallet.AddLocker(ctx, lockerForJack.Raw().Perspective(jill.ID()))
	if err != nil {
		panic(err)
	}

	// send a JSON document from one Jack to Jill

	lb, err := lockerForJack.NewDataSetBuilder(ctx, dataset.WithVault("local"))
	if err != nil {
		panic(err)
	}

	_, err = lb.AddMetaResource(map[string]any{
		"type":     "Fetch",
		"object":   "pail",
		"contents": "water",
	})
	if err != nil {
		panic(err)
	}

	future1 := lb.Submit(expiry.FromNow("1h")) // the lease will expire in 1 hour
	err = future1.Wait(time.Second)
	if err != nil {
		panic(err)
	}

	// send another document back from Jill to Jack

	lb, err = lockerForJill.NewDataSetBuilder(ctx, dataset.WithVault("local"))
	if err != nil {
		panic(err)
	}

	_, err = lb.AddMetaResource(map[string]any{
		"type":      "Tumble",
		"direction": "down the hill",
	})
	if err != nil {
		panic(err)
	}

	future2 := lb.Submit(expiry.FromNow("1h"))
	err = future2.Wait(time.Second)
	if err != nil {
		panic(err)
	}

	// load the first dataset from Jill's wallet

	ds1, err := jillWallet.DataStore().Load(ctx, future1.ID(), dataset.FromLocker(lockerForJill.ID()))
	if err != nil {
		panic(err)
	}

	var dataset1 map[string]any
	_ = ds1.DecodeMetaResource(&dataset1)

	ld.PrintDocument("What Jack did", dataset1)

	// load the second dataset from Jack's wallet

	ds2, err := jackWallet.DataStore().Load(ctx, future2.ID(), dataset.FromLocker(lockerForJack.ID()))
	if err != nil {
		panic(err)
	}

	var dataset2 map[string]any
	_ = ds2.DecodeMetaResource(&dataset2)

	ld.PrintDocument("What Jill did", dataset2)
}
