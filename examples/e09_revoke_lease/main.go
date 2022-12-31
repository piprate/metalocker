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
	"fmt"
	"os"
	"time"

	"github.com/piprate/metalocker/examples"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/model/dataset"
	"github.com/piprate/metalocker/model/expiry"
	"github.com/piprate/metalocker/remote"
	"github.com/piprate/metalocker/utils"
	"github.com/piprate/metalocker/wallet"
)

// This example shows how a user can revoke access to a dataset.
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

	// create account 1: Jack

	passPhrase := "passw0rd!"

	jackWallet, _, err := factory.RegisterAccount(
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

	err = jackWallet.Unlock(passPhrase)
	if err != nil {
		panic(err)
	}

	// create account 2: Jill

	jillWallet, _, err := factory.RegisterAccount(
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

	err = jillWallet.Unlock(passPhrase)
	if err != nil {
		panic(err)
	}

	// create identities for Jack and Jill. An account can have an unlimited number of identities.

	jack, err := jackWallet.NewIdentity(model.AccessLevelHosted, "Jack")
	if err != nil {
		panic(err)
	}

	jill, err := jillWallet.NewIdentity(model.AccessLevelHosted, "Jill")
	if err != nil {
		panic(err)
	}

	// create a locker between Jack and Jill

	lockerForJack, err := jack.NewLocker("Hill", wallet.Participant(jill.DID(), nil))
	if err != nil {
		panic(err)
	}

	lockerForJill, err := jillWallet.AddLocker(lockerForJack.Raw().Perspective(jill.ID()))
	if err != nil {
		panic(err)
	}

	// publish a JSON document in the shared locker

	lb, err := lockerForJack.NewDataSetBuilder(dataset.WithVault("local"))
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

	// Jill can access the published dataset

	_, err = jillWallet.DataStore().Load(future1.ID(), dataset.FromLocker(lockerForJill.ID()))
	if err != nil {
		panic(err)
	}

	// before we revoke the dataset lease, we need to build an index for Jack

	rootIndex, err := jackWallet.CreateRootIndex(examples.DemoIndexStoreName)
	if err != nil {
		panic(err)
	}

	updater, err := jackWallet.IndexUpdater(rootIndex)
	if err != nil {
		panic(err)
	}
	defer updater.Close()

	err = updater.Sync()
	if err != nil {
		panic(err)
	}

	// revoke the lease for the dataset above

	err = jackWallet.DataStore().Revoke(future1.ID()).Wait(time.Second)
	if err != nil {
		panic(err)
	}

	// Jill should still be able to load the dataset, but any attempt to access its data will fail

	ds, err := jillWallet.DataStore().Load(future1.ID(), dataset.FromLocker(lockerForJill.ID()))
	if err != nil {
		panic(err)
	}

	_, err = ds.MetaResource()
	if err == nil {
		panic("revocation didn't work!")
	} else {
		fmt.Printf("\n--== Error message when accessing the dataset ==--: '%s'\n\n", err.Error())
	}

	// to prevent any access completely, including the lease, Jack needs to invoke 'purge data assets' operation.
	// This is a separate step, because it requires physical purging of data assets from underlying storage,
	// which may take time.

	err = jackWallet.DataStore().PurgeDataAssets(future1.ID())
	if err != nil {
		panic(err)
	}

	// and now Jill can't access even the lease itself

	_, err = jillWallet.DataStore().Load(future1.ID(), dataset.FromLocker(lockerForJill.ID()))
	if err == nil {
		panic("data purging didn't work!")
	} else {
		fmt.Printf("\n--== Error message when accessing the purged dataset ==--: '%s'\n", err.Error())
	}
}
