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

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/utils"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog/log"
)

func ExportLedger(ledger model.Ledger, offChainStorage model.OffChainStorage, basePath string) error {
	gb, err := ledger.GetGenesisBlock()
	if err != nil {
		return err
	}

	tb, err := ledger.GetTopBlock()
	if err != nil {
		return err
	}
	log.Info().Int64("number", tb.Number).Msg("Top block")

	currentBlock := gb.Number
	blockBatchSize := 10
	b := gb

	if err = SaveBlock(ledger, offChainStorage, basePath, b); err != nil {
		return err
	}

	for {
		blocks, err := ledger.GetChain(currentBlock, blockBatchSize)
		if err != nil {
			return err
		}

		if len(blocks) == 1 {
			break
		}

		for _, b = range blocks[1:] {
			if err = SaveBlock(ledger, offChainStorage, basePath, b); err != nil {
				return err
			}
		}

		if b.Number == tb.Number {
			break
		}

		currentBlock = b.Number
	}

	topBlockFilePath := path.Join(basePath, "_top.json")

	bb, err := jsonw.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}

	if err = os.WriteFile(topBlockFilePath, bb, 0o600); err != nil {
		return err
	}

	return nil
}

func SaveBlock(ledger model.Ledger, offChainStorage model.OffChainStorage, basePath string, b *model.Block) error {
	log.Info().Int64("number", b.Number).Msg("Saving block")

	dest := path.Join(basePath, utils.Int64ToString(b.Number))
	err := os.MkdirAll(dest, 0o700)
	if err != nil {
		return err
	}

	opDest := path.Join(dest, "operations")
	err = os.MkdirAll(opDest, 0o700)
	if err != nil {
		return err
	}

	blockFilePath := path.Join(dest, "_block.json")

	bb, err := jsonw.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}

	if err = os.WriteFile(blockFilePath, bb, 0o600); err != nil {
		return err
	}

	recList, err := ledger.GetBlockRecords(b.Number)
	if err != nil {
		return err
	}

	recs := make([]*model.Record, 0)
	for _, r := range recList {
		log.Warn().Str("rid", r[0]).Msg("Saving record")

		rec, err := ledger.GetRecord(r[0])
		if err != nil {
			return err
		}

		recs = append(recs, rec)

		if rec.Operation == model.OpTypeLease {
			opRecBytes, err := offChainStorage.GetOperation(rec.OperationAddress)
			if err != nil {
				return err
			}

			if err = os.WriteFile(path.Join(opDest, rec.OperationAddress), opRecBytes, 0o600); err != nil {
				return err
			}
		}
	}

	recBytes, err := jsonw.MarshalIndent(recs, "", "  ")
	if err != nil {
		return err
	}

	recordsFilePath := path.Join(dest, "_records.json")

	if err = os.WriteFile(recordsFilePath, recBytes, 0o600); err != nil {
		return err
	}

	return nil
}
