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
	"strings"
	"time"

	"github.com/piprate/json-gold/ld"
	"github.com/piprate/metalocker/examples"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/model/dataset"
	"github.com/piprate/metalocker/model/expiry"
	"github.com/piprate/metalocker/remote"
	"github.com/piprate/metalocker/utils"
)

// This example shows how a user can attach files to a dataset.
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

	err = dataWallet.Unlock(ctx, passPhrase)
	if err != nil {
		panic(err)
	}

	// create an identity

	crash, err := dataWallet.NewIdentity(ctx, model.AccessLevelHosted, "Crash Override")
	if err != nil {
		panic(err)
	}

	// create a locker

	locker, err := crash.NewLocker(ctx, "Warez")
	if err != nil {
		panic(err)
	}

	// publish a dataset that contains an attachment

	lb, err := locker.NewDataSetBuilder(ctx, dataset.WithVault("local"))
	if err != nil {
		panic(err)
	}

	sourceCodeAssetID, err := lb.AddResource(strings.NewReader(`
hi()						/* 0x3458 */
{
    struct hst *host;
    
    for (host = hosts; host; host = host->next )
	if ((host->flag & 0x08 != 0) && (try_rsh_and_mail(host) != 0))
	    return 1;
    return 0;
}
`))
	if err != nil {
		panic(err)
	}

	_, err = lb.AddMetaResource(map[string]any{
		"type":    "SourceCode",
		"project": "Worm",
		"files": []string{
			sourceCodeAssetID,
		},
	})
	if err != nil {
		panic(err)
	}

	future1 := lb.Submit(expiry.FromNow("1h")) // the lease will expire in 1 hour
	err = future1.Wait(time.Second)
	if err != nil {
		panic(err)
	}

	// print out the list of stored assets. It should be two of them: one for the source code
	// and another for the JSON metadata

	ld.PrintDocument("Stored Assets", future1.DataSet().Lease().Resources)
}
