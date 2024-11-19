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
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/piprate/metalocker/cmd/metalo/datatypes"
	"github.com/piprate/metalocker/index"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/dataset"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/wallet"
	"github.com/rs/zerolog/log"
)

func ExportWallet(ctx context.Context, dw wallet.DataWallet, destDir, lockerID, participantID string, userFriendly, forceRewrite bool) error {
	rootIndex, err := dw.RootIndex(ctx)
	if err != nil {
		return err
	}

	err = rootIndex.TraverseRecords(lockerID, participantID, func(r *index.RecordState) error {
		if r.Status == model.StatusRevoked {
			// skip revoked records
			return nil
		}

		locker, err := dw.GetLocker(ctx, r.LockerID)
		if err != nil {
			return err
		}

		lockerName := locker.Name()
		if lockerName == "" {
			lockerName = "unnamed"
		}

		lockerDir := fmt.Sprintf("%s.%s", r.LockerID, lockerName)
		participantDir := strings.ReplaceAll(r.ParticipantID, ":", "_")
		recordDir := fmt.Sprintf("%d.%s.%s", r.BlockNumber, r.ID, r.ContentType)
		dest := path.Join(destDir, lockerDir, participantDir, recordDir)

		if !forceRewrite {
			if _, err := os.Stat(dest); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else {
				log.Debug().Str("rid", r.ID).Str("dest", dest).Msg("Record folder exists, skipping...")
				return nil
			}
		} else {
			log.Debug().Msg("forceRewrite is ON")
		}

		err = os.MkdirAll(dest, 0o700)
		if err != nil {
			return err
		}

		log.Debug().Str("rid", r.ID).Str("path", dest).Msg("Exporting dataset")

		ds, err := dw.DataStore().Load(ctx, r.ID, dataset.FromLocker(locker.ID()))
		if err != nil {
			return err
		}

		fl, err := datatypes.NewRenderer(ds)
		if err != nil {
			return err
		}

		err = fl.ExportToDisk(dest, true)
		if err != nil {
			return err
		}

		if userFriendly {
			metaDataPath := filepath.Join(dest, datatypes.DefaultMetadataFile)
			if _, err := os.Stat(metaDataPath); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			}

			// pretty print _metadata.json

			meta, err := os.ReadFile(metaDataPath)
			if err != nil {
				return err
			}

			var v map[string]any
			if err := jsonw.Unmarshal(meta, &v); err != nil {
				return err
			}

			// hack: use standard encoder because jsonw indentation is broken
			meta, err = json.MarshalIndent(v, "", "  ")
			if err != nil {
				return err
			}

			if err = os.WriteFile(metaDataPath, meta, 0o600); err != nil {
				return err
			}
		}

		lease := ds.Lease()

		// save impression
		b, err := jsonw.MarshalIndent(lease.Impression, "", "  ")
		if err != nil {
			return err
		}
		if err = os.WriteFile(filepath.Join(dest, datatypes.DefaultImpressionFile), b, 0o600); err != nil {
			return err
		}

		if !userFriendly {
			// save operation
			b, err = jsonw.MarshalIndent(lease, "", "  ")
			if err != nil {
				return err
			}
			if err = os.WriteFile(filepath.Join(dest, datatypes.DefaultOperationFile), b, 0o600); err != nil {
				return err
			}
		}

		return nil
	}, 0)

	return err
}
