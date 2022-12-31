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

package bolt

import (
	"github.com/piprate/metalocker/utils"
	"go.etcd.io/bbolt"
)

const (

	// root buckets

	LockersKey          = "lockers"
	RecordsKey          = "records"
	RecordLookupKey     = "record_lookup"
	ResourceLookupKey   = "resource_lookup"
	ImpressionLookupKey = "impression_lookup"
	VariantsKey         = "variants"
	AssetLookupKey      = "asset_lookup"
	PropertiesKey       = "properties"
	ControlsKey         = "controls"

	// control variables

	GenesisBlockHashKey = "genesis_block_hash"

	// Properties

	AccessLevelKey = "access_level"
	AccountKey     = "account"
)

var (
	storeBuckets = []string{
		ControlsKey,
	}

	indexBuckets = []string{
		LockersKey, RecordsKey, ResourceLookupKey, ImpressionLookupKey, RecordLookupKey, AssetLookupKey,
		VariantsKey, PropertiesKey,
	}
)

func InstallIndexStoreSchema(bc *utils.BoltClient) error {
	err := bc.DB.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range storeBuckets {
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

func InstallIndexSchema(bc *utils.BoltClient, userID string) error {
	return bc.DB.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(userID))
		if err != nil {
			return err
		}
		for _, bucket := range indexBuckets {
			_, err := b.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return err
			}
		}
		return nil
	})
}
