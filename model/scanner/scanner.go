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

package scanner

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sort"
	"strconv"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/piprate/metalocker/model"
	"github.com/rs/zerolog/log"
)

var (
	// ErrIndexResultPending is a signal used by ConsumeRecordFn functions to tell the scanner
	// that the given record will be processed outside of the scanner's loop and processing
	// of the underlying index (a set of locker participant keys that belongs to the same entity)
	// should be paused until the processing is completed.
	ErrIndexResultPending = errors.New("index result pending")

	ErrSubscriptionNotFound = errors.New("subscription not found")
)

const (
	ScanStatusActive = 0
	ScanStatusPaused = 1
	ScanStatusError  = 2
)

type (
	ConsumeRecordFn func(blockNumber int64, keyID int, key []byte, idx uint32, r *model.Record) error

	DatasetNotification struct {
		KeyID     int          `json:"kid"`
		RecordID  string       `json:"r"`
		Operation model.OpType `json:"t"`

		// The fields below are passed in order to optimise downstream processing
		Key    []byte        `json:"-"`
		Record *model.Record `json:"-"`
	}

	BlockNotification struct {
		Block    int64                 `json:"b"`
		Datasets []DatasetNotification `json:"ds"`
	}

	Subscription interface {
		io.Closer

		IndexID() string
		LockerConfigs() []*LockerConfig
		ConsumeBlock(ctx context.Context, n BlockNotification) error
		NotifyScanCompleted(topBlock int64) error
		InactiveSince() int64
		Status() int
		SetStatus(status int)
	}

	MutableSubscription interface {
		AddLockers(lockers ...LockerEntry) error
	}

	LockerConfig struct {
		KeyID        int    `json:"key"`
		LastBlock    int64  `json:"last"`
		PublicKeyStr string `json:"pubk"`

		Subscription Subscription `json:"-"`

		publicKey *hdkeychain.ExtendedKey
	}

	subscriptionState struct {
		sub          Subscription
		notification BlockNotification
	}

	Scanner struct {
		ledgerAPI        model.Ledger
		subscriptionList []string
		subscriptions    map[string]Subscription
	}
)

func NewScanner(ledgerAPI model.Ledger) *Scanner {
	return &Scanner{
		ledgerAPI:        ledgerAPI,
		subscriptionList: make([]string, 0),
		subscriptions:    make(map[string]Subscription),
	}
}

func (isu *Scanner) scanLedger(scannerList []*LockerConfig, startBlockNumber int64, endBlockNumber int64, blockBatchSize int) (int64, bool, error) {
	currentBlockNumber := startBlockNumber
	firstBlockIndex := 0
	earlyExit := false

	ctx := context.Background()

AllBlocks:
	for {
		blocks, err := isu.ledgerAPI.GetChain(ctx, currentBlockNumber, blockBatchSize)
		if err != nil {
			return -1, false, err
		}

		if len(blocks) == 0 {
			break
		}

		for _, b := range blocks[firstBlockIndex:] {
			currentBlockNumber = b.Number

			log.Debug().Int64("number", currentBlockNumber).Msg("Processing block")

			states := make(map[string]*subscriptionState)

			records, err := isu.ledgerAPI.GetBlockRecords(ctx, currentBlockNumber)
			if err != nil {
				return -1, false, err
			}

			for _, v := range records {
				rid := v[0]
				routingKey := base58.Decode(v[1])
				idx64, err := strconv.ParseUint(v[2], 10, 32)
				if err != nil {
					return -1, false, err
				}
				idx := uint32(idx64)

				var lr *model.Record

				log.Debug().Str("id", rid).Str("routingKey", v[1]).Msg("Processing record")

				for _, cfg := range scannerList {
					if cfg.Subscription.Status() != ScanStatusActive {
						continue
					}

					k, err := cfg.publicKey.Derive(idx)
					if err != nil {
						return -1, false, err
					}
					recordPubKey, _ := k.ECPubKey()
					indexKey := recordPubKey.SerializeCompressed()

					if !bytes.Equal(indexKey, routingKey) {
						continue
					}

					log.Debug().Int("KeyID", cfg.KeyID).Str("routingKey", v[1]).Uint32("idx", idx).Msg("Record found")

					if lr == nil {
						// lazy-read record
						lr, err = isu.ledgerAPI.GetRecord(ctx, rid)
						if err != nil {
							return -1, false, err
						}
					}

					state, found := states[cfg.Subscription.IndexID()]
					if !found {
						state = &subscriptionState{
							sub: cfg.Subscription,
						}
						state.notification.Block = b.Number
						states[cfg.Subscription.IndexID()] = state
					}
					state.notification.Datasets = append(state.notification.Datasets, DatasetNotification{
						KeyID:     cfg.KeyID,
						RecordID:  rid,
						Operation: lr.Operation,
						Key:       indexKey,
						Record:    lr,
					})
				}
			}

			for _, state := range states {
				if err := state.sub.ConsumeBlock(ctx, state.notification); err != nil {
					if errors.Is(err, ErrIndexResultPending) {
						state.sub.SetStatus(ScanStatusPaused)
					} else {
						log.Err(err).Str("sub", state.sub.IndexID()).Msg("Error consuming a block. Setting status=error")
						state.sub.SetStatus(ScanStatusError)
					}
				}
			}

			reducedScannerList := make([]*LockerConfig, 0, len(scannerList))
			for _, cfg := range scannerList {
				if cfg.Subscription.Status() != ScanStatusError {
					cfg.LastBlock = currentBlockNumber
					if cfg.Subscription.Status() == ScanStatusActive {
						reducedScannerList = append(reducedScannerList, cfg)
					}
				}
			}

			scannerList = reducedScannerList

			if len(reducedScannerList) == 0 {
				earlyExit = true
				break AllBlocks
			}

			if currentBlockNumber == endBlockNumber {
				break AllBlocks
			}
		}

		if len(blocks) < blockBatchSize {
			break
		}

		firstBlockIndex = 1
	}

	return currentBlockNumber, earlyExit, nil
}

func (isu *Scanner) AddSubscription(sub Subscription) error {
	for _, cfg := range sub.LockerConfigs() {
		if err := cfg.Hydrate(); err != nil {
			return err
		}
	}

	_, found := isu.subscriptions[sub.IndexID()]
	if !found {
		// update the ordered list of subscriptions to ensure deterministic iterations in Scan()
		subscriptionList := append(isu.subscriptionList, sub.IndexID())
		sort.Strings(subscriptionList)
		isu.subscriptionList = subscriptionList
	}

	isu.subscriptions[sub.IndexID()] = sub

	return nil
}

func (isu *Scanner) RemoveSubscription(indexID string) error {
	sub, found := isu.subscriptions[indexID]
	if !found {
		return ErrSubscriptionNotFound
	}

	newSubscriptionList := make([]string, len(isu.subscriptionList)-1)
	i := 0
	for _, idx := range isu.subscriptionList {
		if idx != indexID {
			newSubscriptionList[i] = idx
			i++
		}
	}
	isu.subscriptionList = newSubscriptionList

	delete(isu.subscriptions, indexID)

	return sub.Close()
}

func (isu *Scanner) Subscriptions() map[string]Subscription {
	return isu.subscriptions
}

func (isu *Scanner) Scan(ctx context.Context) (bool, error) {
	for {
		complete, restart, err := isu.scanOneRound(ctx)
		if err != nil {
			return false, err
		}

		if !restart {
			hasErrors := false
			for _, sub := range isu.subscriptions {
				if sub.Status() == ScanStatusError {
					hasErrors = true
					break
				}
			}
			if hasErrors {
				return complete, errors.New("one or more subscriptions in error state")
			} else {
				return complete, nil
			}

		}

		log.Debug().Msg("Sync exited early due to an account update message. Restarting...")
	}
}

func (isu *Scanner) scanOneRound(ctx context.Context) (bool, bool, error) {

	blockBatchSize := 10

	complete := true
	blockSeqNoList := make([]int64, 0)
	blockToLockerConfigs := make(map[int64][]*LockerConfig)
	for _, indexID := range isu.subscriptionList {
		sub := isu.subscriptions[indexID]
		if sub.Status() == ScanStatusActive {
			for _, ls := range sub.LockerConfigs() {
				lastBlockNumber := ls.LastBlock
				list, found := blockToLockerConfigs[lastBlockNumber]
				blockToLockerConfigs[lastBlockNumber] = append(list, ls)
				if !found {
					blockSeqNoList = append(blockSeqNoList, lastBlockNumber)
				}
			}
		} else {
			complete = false
		}
	}
	sort.Slice(blockSeqNoList, func(i, j int) bool { return blockSeqNoList[i] < blockSeqNoList[j] })

	accumulatedLockers := make([]*LockerConfig, 0)

	topBlock, err := isu.ledgerAPI.GetTopBlock(ctx)
	if err != nil {
		return false, false, err
	}

	var topBlockNumber int64 = -1
	exitedEarly := false
	for idx, blockNumber := range blockSeqNoList {

		if blockNumber == topBlock.Number {
			log.Debug().Msg("Scanning sequence finished")
			break
		}

		// pick the first block after the last scanned block
		startBlockNumber := blockNumber + 1
		endBlockNumber := int64(-1)
		if idx < len(blockSeqNoList)-1 {
			endBlockNumber = blockSeqNoList[idx+1]
		}

		accumulatedLockers = append(blockToLockerConfigs[blockNumber], accumulatedLockers...)

		if len(accumulatedLockers) == 0 {
			log.Warn().Int("idx", idx).Int64("start", startBlockNumber).Int64("end", endBlockNumber).
				Msg("No lockers to sync in current round")
			continue
		}

		reducedScannerList := make([]*LockerConfig, 0, len(accumulatedLockers))
		for _, cfg := range accumulatedLockers {
			if cfg.Subscription.Status() == ScanStatusActive {
				reducedScannerList = append(reducedScannerList, cfg)
			} else {
				log.Warn().Int("status", cfg.Subscription.Status()).Int("key", cfg.KeyID).Msg("Dropping inactive locker")
			}
		}

		if len(reducedScannerList) == 0 {
			exitedEarly = true
			break
		} else {
			accumulatedLockers = reducedScannerList
		}

		log.Debug().Int("idx", idx).Int64("start", startBlockNumber).Int64("end", endBlockNumber).
			Int("lockerCount", len(accumulatedLockers)).Msg("Initiating new scanning round")

		topBlockNumber, exitedEarly, err = isu.scanLedger(accumulatedLockers, startBlockNumber, endBlockNumber, blockBatchSize)
		if err != nil {
			return false, false, err
		}
		if exitedEarly {
			break
		}
	}

	restart := false

	if topBlockNumber >= 0 {
		for _, indexID := range isu.subscriptionList {
			sub := isu.subscriptions[indexID]
			if err = sub.NotifyScanCompleted(topBlockNumber); err != nil {
				return false, false, err
			}
			if sub.Status() == ScanStatusActive && exitedEarly {
				restart = true
			}
			if sub.Status() != ScanStatusActive {
				complete = false
			}
		}
	}

	return complete, restart, nil
}

func (lc *LockerConfig) Hydrate() error {
	if lc.publicKey == nil {
		pubKey, err := hdkeychain.NewKeyFromString(lc.PublicKeyStr)
		if err != nil {
			return err
		}
		lc.publicKey = pubKey
	}

	return nil
}
