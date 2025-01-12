package onflow

import (
	"github.com/piprate/metalocker/utils"
	"go.etcd.io/bbolt"
)

const (

	// root buckets

	BlocksKey       = "blocks"
	BlockRecordsKey = "block_records"
	RecordsKey      = "records"
	ControlsKey     = "controls"

	// control variables

	GenesisBlockHeightKey = "genesis_block_height"
	TopFlowBlockHeightKey = "top_flow_block_height"
	TopBlockNumberKey     = "top_block_number"
)

var (
	buckets = []string{BlocksKey, BlockRecordsKey, RecordsKey, ControlsKey}
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
