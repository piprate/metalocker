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
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/dataset"
	"github.com/piprate/metalocker/services/notification"
	"github.com/piprate/metalocker/utils"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog/log"
)

func ImportLedger(ledger model.Ledger, offChainStorage model.OffChainStorage, ns notification.Service, dest string, importOperations, waitForConfirmation bool) error {
	tbBytes, err := os.ReadFile(path.Join(dest, "_top.json"))
	if err != nil {
		return err
	}

	var tb model.Block
	err = jsonw.Unmarshal(tbBytes, &tb)
	if err != nil {
		return err
	}

	var i int64
	for i = 0; i <= tb.Number; i++ {
		blockPath := path.Join(dest, utils.Int64ToString(i))

		if importOperations {
			if err := filepath.Walk(path.Join(blockPath, "operations"), func(filePath string, f os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				log.Info().Str("file", filePath).Msg("Saving operation")

				if strings.HasSuffix(filePath, "operations") {
					return nil
				}

				opBytes, err := os.ReadFile(filePath)
				if err != nil {
					return err
				}

				opid, err := offChainStorage.SendOperation(opBytes)
				if err != nil {
					return err
				}

				log.Info().Str("file", filePath).Str("id", opid).Msg("Saved operation")

				return nil
			}); err != nil {
				return err
			}
		}

		recsBytes, err := os.ReadFile(path.Join(blockPath, "_records.json"))
		if err != nil {
			return err
		}

		var recs []*model.Record
		err = jsonw.Unmarshal(recsBytes, &recs)
		if err != nil {
			return err
		}

		var r *model.Record
		for _, r = range recs {
			log.Info().Str("rid", r.ID).Msg("Importing record")

			if err = ledger.SubmitRecord(r); err != nil {
				return err
			}
		}

		if r != nil && waitForConfirmation {
			// wait for the last record in the block

			if _, err = dataset.WaitForConfirmation(ledger, nil, time.Second, 60*time.Second, r.ID); err != nil {
				return err
			}
		}
	}

	return nil
}
