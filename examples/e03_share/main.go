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

// This example shows how a user can create a dataset and then share it with another identity, preserving provenance.
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

	// create a private uni-locker for Jack

	privateLockerForJack, err := jack.NewLocker(ctx, "Private Hill")
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

	// publish a JSON document in Jack's private locker

	lb, err := privateLockerForJack.NewDataSetBuilder(ctx, dataset.WithVault("local"))
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

	// share the JSON document with Jill. This operation will create a full copy of the original dataset
	// in the shared locker and add provenance information that can prove it's an authentic copy
	// of the original record.

	future2 := lockerForJack.Share(ctx, future1.ID(), "local", expiry.FromNow("45min"))
	err = future2.Wait(time.Second)
	if err != nil {
		panic(err)
	}

	// load the dataset from Jill's wallet and show provenance

	ds1, err := jillWallet.DataStore().Load(ctx, future2.ID(), dataset.FromLocker(lockerForJill.ID()))
	if err != nil {
		panic(err)
	}

	ld.PrintDocument("Provenance data for the shared dataset", ds1.Lease().Provenance)
}
