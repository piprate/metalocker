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
	"fmt"
	"os"
	"path/filepath"

	"github.com/piprate/metalocker/wallet"
	"github.com/rs/zerolog/log"
)

func StoreDataSets(lib wallet.DataStore, path, metaType, vaultName, lockerID, provPath, provMapping string,
	durationString string, waitForConfirmation bool) ([]string, error) {

	dsFolders, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	recIds := make([]string, 0)
	for _, f := range dsFolders {
		if f.IsDir() {
			log.Debug().Str("path", f.Name()).Msg("Importing dataset")
			dsPath := filepath.Join(path, f.Name())
			recID, impID, err := StoreDataSet(lib, dsPath, metaType, vaultName, lockerID, provPath,
				provMapping, "", durationString, waitForConfirmation)
			if err != nil {
				return nil, err
			}
			recIds = append(recIds, fmt.Sprintf("%s,%s", recID, impID))

			//nextRecordIndex += 1
		} else {
			log.Warn().Str("path", f.Name()).Msg("Can't process files as datasets")
		}
	}

	return recIds, nil
}
