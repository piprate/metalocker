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

	"github.com/piprate/metalocker/remote/caller"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog/log"
)

func ExportAccounts(ctx context.Context, mlc *caller.MetaLockerHTTPCaller, dest string) error {
	log.Info().Str("dest", dest).Msg("Exporting account database")

	err := os.MkdirAll(dest, 0o700)
	if err != nil {
		return err
	}

	// export accounts

	acctList, err := mlc.AdminGetAccountList()
	if err != nil {
		return err
	}

	b, _ := jsonw.MarshalIndent(acctList, "", "  ")

	if err = os.WriteFile(filepath.Join(dest, "accounts.json"), b, 0o600); err != nil {
		return err
	}

	// export identities

	iidList, err := mlc.ListDIDDocuments(ctx)
	if err != nil {
		return err
	}

	b, _ = jsonw.MarshalIndent(iidList, "", "  ")

	if err = os.WriteFile(filepath.Join(dest, "identities.json"), b, 0o600); err != nil {
		return err
	}

	return nil
}
