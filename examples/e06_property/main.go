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
	"os"

	"github.com/piprate/json-gold/ld"
	"github.com/piprate/metalocker/examples"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/remote"
	"github.com/piprate/metalocker/utils"
	"github.com/piprate/metalocker/wallet"
)

// This example shows how a user can create account properties.
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

	// create an account

	passPhrase := "passw0rd!"

	dataWallet, _, err := factory.RegisterAccount(
		&account.Account{
			Email:       "hackers@example.com",
			Name:        "Hackers The Movie",
			AccessLevel: model.AccessLevelHosted,
		},
		passPhrase,
	)
	if err != nil {
		panic(err)
	}

	// the wallet needs to be unlocked before use

	err = dataWallet.Unlock(passPhrase)
	if err != nil {
		panic(err)
	}

	// create an identity

	crash, err := dataWallet.NewIdentity(model.AccessLevelHosted, "Crash Override",
		wallet.WithType(account.IdentityTypePersona))
	if err != nil {
		panic(err)
	}

	// create a new property. Account properties are account-level key/value pairs that allow
	// setting transient (not persisted in the ledger), discoverable values for account
	// management purposes.

	err = dataWallet.SetProperty("my.main.identity", crash.ID(), model.AccessLevelHosted)
	if err != nil {
		panic(err)
	}

	props, err := dataWallet.GetProperties()
	if err != nil {
		panic(err)
	}

	ld.PrintDocument("All Properties", props)
}
