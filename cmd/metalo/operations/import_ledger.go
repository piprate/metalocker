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
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/knadh/koanf"
	"github.com/piprate/metalocker/ledger"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/dataset"
	"github.com/piprate/metalocker/sdk/cmdbase"
	"github.com/piprate/metalocker/utils"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/vaults"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func ImportLedger(ctx context.Context, importDirPath, configFilePath string, importOperations, preserveBlocks, waitForConfirmation bool) error { //nolint:gocyclo
	// read configuration
	cfg, err := readConfigFile(configFilePath)
	if err != nil {
		return err
	}

	resolver, err := cmdbase.ConfigureParameterResolver(cfg, filepath.Dir(configFilePath))
	if err != nil {
		return err
	}

	destLedger, err := initLedgerInstance(ctx, cfg, resolver)
	if err != nil {
		return err
	}

	var offChainVault vaults.Vault
	if importOperations {
		offChainVault, err = initOffChainVault(cfg, resolver)
		if err != nil {
			return err
		}
	}

	tbBytes, err := os.ReadFile(path.Join(importDirPath, "_top.json"))
	if err != nil {
		return err
	}

	var tb model.Block
	err = jsonw.Unmarshal(tbBytes, &tb)
	if err != nil {
		return err
	}

	var topDestBlockNumber uint64
	if preserveBlocks {
		top, err := destLedger.GetTopBlock(ctx)
		if err != nil {
			return err
		}
		if top != nil {
			topDestBlockNumber = top.Number
		}
	}
	var blockNumber uint64
	var blockOneRecords []*model.Record
	importedRecords := make(map[string]bool)
	for ; blockNumber <= tb.Number; blockNumber++ {
		blockPath := path.Join(importDirPath, utils.Uint64ToString(blockNumber))

		if importOperations {
			if err = filepath.Walk(path.Join(blockPath, "operations"), func(filePath string, f os.FileInfo, err error) error {
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

				opRec, err := offChainVault.CreateBlob(ctx, bytes.NewReader(opBytes))
				if err != nil {
					return err
				}

				log.Info().Str("file", filePath).Str("id", opRec.ID).Msg("Saved operation")

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

		log.Info().Uint64("number", blockNumber).Int("record_count", len(recs)).Msg("Importing block")

		// remove duplicate records

		cleanRecs := make([]*model.Record, 0, len(recs))
		for _, rec := range recs {
			if _, duplicate := importedRecords[rec.ID]; duplicate {
				log.Warn().Uint64("number", blockNumber).Str("rid", rec.ID).Msg("Duplicate record found, skipping...")
				continue
			} else {
				cleanRecs = append(cleanRecs, rec)
				importedRecords[rec.ID] = true
			}
		}
		recs = cleanRecs

		if preserveBlocks {
			if blockNumber <= topDestBlockNumber {
				// incremental mode. Check that source and destination blocks are identical
				destRecList, err := destLedger.GetBlockRecords(ctx, blockNumber)
				if err != nil {
					return err
				}
				if len(destRecList) != len(recs) {
					if blockNumber == 1 {
						blockOneRecords = recs
						log.Warn().Int("record_count", len(recs)).Msg("Found records in block 1. Adding to block 2")
						continue
					} else if blockNumber == 2 && len(destRecList) > len(recs) {
						log.Warn().Int("src_record_count", len(recs)).Int("dest_record_count", len(destRecList)).Msg("Extra records found in block 2. Assuming they were moved from block 1")
						continue
					} else {
						return fmt.Errorf("block already exists: %d. Got %d record(s) in destination ledger, got %d source record(s)", blockNumber, len(destRecList), len(recs))
					}
				}
			} else {
				if blockNumber == 2 && len(blockOneRecords) > 0 {
					log.Warn().Int("b2_record_count", len(recs)).Int("b1_record_count", len(blockOneRecords)).Msg("Importing block 2 with extra records from block 1")
					recs = append(recs, blockOneRecords...)
					blockOneRecords = nil
				}
				if err = destLedger.ImportBlock(ctx, blockNumber, recs); err != nil {
					return err
				}
			}
		} else {
			var r *model.Record
			for _, r = range recs {
				log.Info().Str("rid", r.ID).Msg("Importing record")

				if r.Status == model.StatusRevoked {
					r.Status = model.StatusPublished
				}

				if err = destLedger.SubmitRecord(ctx, r); err != nil {
					return err
				}
			}

			if r != nil && waitForConfirmation {
				// wait for the last record in the block

				if _, err = dataset.WaitForConfirmation(ctx, destLedger, nil, time.Second, 60*time.Second, r.ID); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func initLedgerInstance(ctx context.Context, cfg *koanf.Koanf, resolver cmdbase.ParameterResolver) (model.Ledger, error) {
	if cfg.Exists("ledger") {
		var ledgerCfg *ledger.Config

		err := cfg.Unmarshal("ledger", &ledgerCfg)
		if err != nil {
			log.Err(err).Msg("Failed to read ledger configuration")
			return nil, cli.Exit(err, 1)
		}

		identityBackend, err := ledger.CreateLedgerConnector(ctx, ledgerCfg, nil, resolver)
		if err != nil {
			log.Err(err).Msg("Failed to create ledger instance")
			return nil, cli.Exit(err, 1)
		}

		return identityBackend, nil
	} else {
		return nil, cli.Exit("ledger not defined", 1)
	}
}

func initOffChainVault(cfg *koanf.Koanf, resolver cmdbase.ParameterResolver) (vaults.Vault, error) {
	var vaultCfg vaults.Config
	err := cfg.Unmarshal("offChainStore", &vaultCfg)
	if err != nil {
		log.Err(err).Msg("Failed to read vault configuration")
		return nil, cli.Exit(err, 1)
	}

	offchainAPI, err := vaults.CreateVault(&vaultCfg, resolver, nil)
	if err != nil {
		log.Err(err).Msg("Failed to create an offchain vault")
		os.Exit(1)
	}

	if !offchainAPI.CAS() {
		log.Err(err).Msg("Offchain operation vault should be a content-addressable storage")
		return nil, cli.Exit(err, 1)
	}

	return offchainAPI, nil
}
