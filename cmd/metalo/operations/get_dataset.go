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
	"path/filepath"

	"github.com/piprate/json-gold/ld"
	"github.com/piprate/metalocker/cmd/metalo/datatypes"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/wallet"
)

func GetDataSet(dw wallet.DataWallet, rid, dest string, exportMetaData bool) error {

	ds, err := dw.DataStore().Load(rid)
	if err != nil {
		return err
	}

	fl, err := datatypes.NewRenderer(ds)
	if err != nil {
		return err
	}

	lease := ds.Lease()

	if dest != "" {
		err = fl.ExportToDisk(dest, exportMetaData)
		if err != nil {
			return err
		}

		if exportMetaData {
			// save impression
			b, err := jsonw.MarshalIndent(lease.Impression, "", "  ")
			if err != nil {
				return err
			}
			if err = os.WriteFile(filepath.Join(dest, datatypes.DefaultImpressionFile), b, 0o600); err != nil {
				return err
			}
			// save operation
			b, err = jsonw.MarshalIndent(lease, "", "  ")
			if err != nil {
				return err
			}
			if err = os.WriteFile(filepath.Join(dest, datatypes.DefaultOperationFile), b, 0o600); err != nil {
				return err
			}
		}
	} else {
		if exportMetaData {
			ld.PrintDocument("", lease)
		} else {
			if err = fl.Print(); err != nil {
				return err
			}
		}
	}

	return nil
}
