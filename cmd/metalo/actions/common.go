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
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/piprate/metalocker/cmd"
	"github.com/piprate/metalocker/index"
	"github.com/piprate/metalocker/index/bolt"
	"github.com/piprate/metalocker/remote"
	"github.com/piprate/metalocker/remote/caller"
	"github.com/piprate/metalocker/wallet"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

const (
	InvalidParameter     = 1
	OperationFailed      = 2
	AuthenticationFailed = 3

	LocalIndexStoreID   = "did:piprate:B7hqrwxrCnsWqtBnbYnhABwzktnT5oo1ZWbtby3My886"
	LocalIndexStoreName = "local"

	DefaultMetaLockerURL = "https+insecure://127.0.0.1:4000"

	MetaloVersion = "0.0.1"
)

func WithPersonalIndexStore() remote.IndexClientSourceFn {
	return func(userID string, mlc *caller.MetaLockerHTTPCaller) (index.Client, error) {
		gb, err := mlc.GetGenesisBlock()
		if err != nil {
			return nil, err
		}

		// create a local index store (used in command line tools
		walletFilePath, err := cmd.GetWalletPath(mlc.ConnectionURL(), gb.Hash, userID)
		if err != nil {
			return nil, err
		}
		return index.NewLocalIndexClient([]*index.StoreConfig{
			{
				ID:   LocalIndexStoreID,
				Name: LocalIndexStoreName,
				Type: bolt.Type,
				Params: map[string]any{
					bolt.ParameterFilePath: walletFilePath,
				},
			},
		}, nil, gb.Hash)
	}
}

func LoadRemoteDataWallet(c *cli.Context, syncIndexOnStart bool) (wallet.DataWallet, error) {
	url := c.String("server")

	factory, err := remote.NewWalletFactory(url, WithPersonalIndexStore(), 0)
	if err != nil {
		return nil, err
	}

	var dw wallet.DataWallet
	if c.String("api-key") != "" {
		apiKey := c.String("api-key")
		apiSecret := ReadCredential(c.String("api-secret"), "Enter API Secret: ", true)
		dw, err = factory.GetWalletWithAccessKey(c.Context, apiKey, apiSecret)
	} else {
		user := ReadCredential(c.String("user"), "Enter account email: ", false)
		password := ReadCredential(c.String("password"), "Enter password: ", true)

		// save the password value read from console for ChangePassphrase call
		_ = c.Set("password", password)

		dw, err = factory.GetWalletWithCredentials(c.Context, user, password)
	}
	if err != nil {
		if errors.Is(err, caller.ErrLoginFailed) {
			return nil, cli.Exit(err, AuthenticationFailed)
		} else {
			return nil, cli.Exit(err, OperationFailed)
		}
	}

	if syncIndexOnStart {
		ix, err := dw.RootIndex(c.Context)
		if err != nil {
			if errors.Is(err, index.ErrIndexNotFound) {
				ix, err = dw.CreateRootIndex(c.Context, LocalIndexStoreName)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}

		ixUpdater, err := dw.IndexUpdater(c.Context, ix)
		if err != nil {
			return nil, err
		}

		if err = ixUpdater.Sync(); err != nil {
			return nil, err
		}
	}

	return dw, nil
}

func CreateAdminHTTPCaller(c *cli.Context) (*caller.MetaLockerHTTPCaller, error) {
	mlc := CreateHTTPCaller(c)

	key := ReadCredential(c.String("admin-key"), "Enter Admin API key: ", true)
	secret := ReadCredential(c.String("admin-secret"), "Enter Admin API Secret: ", true)

	err := mlc.LoginWithAdminKeys(key, secret)
	if err != nil {
		return nil, err
	}

	return mlc, nil
}

func CreateHTTPCaller(c *cli.Context) *caller.MetaLockerHTTPCaller {
	url := c.String("server")

	userAgent := fmt.Sprintf("MetaLocker CLI (Metalo) v.%s", MetaloVersion)
	mlc, err := caller.NewMetaLockerHTTPCaller(url, userAgent)
	if err != nil {
		return nil
	}

	// initialise context forwarding
	mlc.InitContextForwarding()

	return mlc
}

func ReadCredential(val, prompt string, mask bool) string {
	if val != "" {
		return val
	}

	fmt.Print(prompt)

	if mask {
		byteVal, err := term.ReadPassword(syscall.Stdin)
		if err != nil {
			panic("error when reading password")
		}
		val = string(byteVal)
	} else {
		reader := bufio.NewReader(os.Stdin)

		val, _ = reader.ReadString('\n')
	}

	fmt.Println()

	return strings.TrimSpace(val)
}
