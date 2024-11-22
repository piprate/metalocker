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

package operations

import (
	"context"
	"os"
	"path/filepath"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/remote/caller"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog/log"
)

func ImportAccounts(ctx context.Context, mlc *caller.MetaLockerHTTPCaller, srcPath string) error {
	log.Info().Str("dest", srcPath).Msg("Importing account database")

	// import accounts

	accountsBody, err := os.ReadFile(filepath.Join(srcPath, "accounts.json"))
	if err != nil {
		return err
	}

	var acctList []*account.Account
	err = jsonw.Unmarshal(accountsBody, &acctList)
	if err != nil {
		return err
	}

	var accountsCount int64

	for _, acct := range acctList {
		err := mlc.AdminStoreAccount(ctx, acct)
		if err != nil {
			log.Err(err).Msg("Error when importing account")
			//return nil, err
		} else {
			accountsCount++
		}
	}

	// import identities

	identitiesBody, err := os.ReadFile(filepath.Join(srcPath, "identities.json"))
	if err != nil {
		return err
	}

	var iidList []*model.DIDDocument
	err = jsonw.Unmarshal(identitiesBody, &iidList)
	if err != nil {
		return err
	}

	var identitiesCount int64

	for _, iid := range iidList {
		err := mlc.AdminStoreIdentity(ctx, iid)
		if err != nil {
			log.Err(err).Msg("Error when importing identity")
			//return nil, err
		} else {
			identitiesCount++
		}
	}

	log.Info().Int64("accounts", accountsCount).Int64("identities", identitiesCount).
		Msg("Imported account database")

	return nil
}
