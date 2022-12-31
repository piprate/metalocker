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

package local

import (
	"github.com/piprate/metalocker/utils"
	"go.etcd.io/bbolt"
)

const (

	// root buckets

	BlocksKey             = "blocks"
	BlockCompositionsKey  = "block_composition"
	RecordsKey            = "records"
	RecordStatesKey       = "record_states"
	ControlsKey           = "controls"
	UnconfirmedRecordsKey = "unconfirmed_records"
	DataAssetStatesKey    = "data_asset_states"
	HeadsKey              = "heads"

	// control variables

	TopBlockNumberKey   = "top_block_number"
	CurrentSessionIDKey = "current_session_id"
)

var (
	buckets = []string{BlocksKey, BlockCompositionsKey, RecordsKey, RecordStatesKey, ControlsKey,
		UnconfirmedRecordsKey, DataAssetStatesKey, HeadsKey}
)

func InstallLedgerSchema(bc *utils.BoltClient) error {
	err := bc.DB.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range buckets {
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
