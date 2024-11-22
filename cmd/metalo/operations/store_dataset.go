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
	"fmt"
	"os"
	"time"

	"github.com/piprate/metalocker/cmd/metalo/datatypes"
	"github.com/piprate/metalocker/model/dataset"
	"github.com/piprate/metalocker/model/expiry"
	"github.com/piprate/metalocker/utils"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/wallet"
)

func StoreDataSet(ctx context.Context, lib wallet.DataStore, path, metaType, vaultName, lockerID, provPath, provMapping, parentRecordID string,
	durationString string, waitForConfirmation bool) (string, string, error) {

	provMap, err := utils.BuildMapFromString(provMapping)
	if err != nil {
		return "", "", err
	}

	var builder dataset.Builder
	if parentRecordID != "" {
		builder, err = lib.NewDataSetBuilder(ctx,
			lockerID,
			dataset.WithVault(vaultName),
			dataset.WithParent(
				parentRecordID,
				"",
				dataset.CopyModeNone,
				nil,
				false),
		)
	} else {
		builder, err = lib.NewDataSetBuilder(ctx, lockerID, dataset.WithVault(vaultName))
	}

	if err != nil {
		return "", "", err
	}

	// read provenance definition, if specified

	if _, err := os.Stat(provPath); err != nil {
		if !os.IsNotExist(err) {
			return "", "", err
		}
	} else {
		provTemplate, err := os.ReadFile(provPath)
		if err != nil {
			return "", "", err
		}
		provBytes, err := utils.SubstituteEntities(provTemplate, provMap, nil)
		if err != nil {
			return "", "", err
		}

		var prov any
		if err = jsonw.Unmarshal(provBytes, &prov); err != nil {
			return "", "", fmt.Errorf("error reading provenance definition: %w", err)
		}

		if err = builder.AddProvenance("", prov, true); err != nil {
			return "", "", err
		}
	}

	fw, err := datatypes.NewUploader(metaType, path, vaultName, provMap)
	if err != nil {
		return "", "", err
	}

	err = fw.Write(builder)
	if err != nil {
		return "", "", err
	}

	f := builder.Submit(expiry.FromNow(durationString))

	if waitForConfirmation {
		err = f.Wait(60 * time.Second)
	} else {
		err = f.Error()
	}
	if err != nil {
		return "", "", err
	}

	opRec := f.Lease()

	return f.ID(), opRec.Impression.ID, nil
}
