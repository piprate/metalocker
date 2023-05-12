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

package wallet

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/piprate/metalocker/index"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/dataset"
	"github.com/piprate/metalocker/model/scanner"
	"github.com/piprate/metalocker/services/notification"
	"github.com/piprate/metalocker/storage"
	"github.com/piprate/metalocker/utils/measure"
	"github.com/rs/zerolog/log"
)

type (
	IndexUpdater struct {
		ledger  model.Ledger
		scanner *scanner.Scanner
		indexes map[string]index.Writer

		syncMutex *sync.Mutex

		syncCh         chan bool
		controlCh      chan string
		eventControlCh chan string
	}
)

func NewIndexUpdater(ledger model.Ledger) *IndexUpdater {

	updater := &IndexUpdater{
		ledger:    ledger,
		scanner:   scanner.NewScanner(ledger),
		indexes:   map[string]index.Writer{},
		syncMutex: &sync.Mutex{},
	}

	return updater
}

func (ixf *IndexUpdater) AddIndexes(dw DataWallet, indexes ...index.Index) error {
	for _, ix := range indexes {
		if !ix.IsWritable() {
			return fmt.Errorf("index isn't writable: %s, can't update", ix.ID())
		}
		iw, _ := ix.Writer()
		ixf.indexes[ix.ID()] = iw

		recordConsumer, err := newConsumer(dw, iw)
		if err != nil {
			return err
		}

		sub := scanner.NewIndexSubscription(iw.ID(), recordConsumer)

		lockerStates, err := iw.LockerStates()
		if err != nil {
			return err
		}
		for _, ls := range lockerStates {

			lockerDW, err := recordConsumer.getDataWallet(ls.AccountID)
			if err != nil {
				return err
			}

			l, err := lockerDW.GetLocker(ls.ID)
			if err != nil {
				if errors.Is(err, storage.ErrLockerNotFound) {
					log.Warn().Str("lid", ls.ID).Msg("Locker not found for index locker state")
				} else {
					return err
				}
			} else {
				err = sub.AddLockers(scanner.LockerEntry{
					Locker:    l.Raw(),
					LastBlock: ls.TopBlock,
				})
				if err != nil {
					return err
				}
			}
		}
		if err = ixf.scanner.AddSubscription(sub); err != nil {
			return err
		}
	}

	return nil
}

func (ixf *IndexUpdater) StartSyncOnEvents(ns notification.Service, syncOnStart bool, forceSyncInterval time.Duration) error {
	if syncOnStart {
		if err := ixf.Sync(); err != nil {
			return err
		}
	}

	ixf.syncCh = make(chan bool)
	ixf.controlCh = make(chan string)
	ixf.eventControlCh = make(chan string)

	stopping := false

	go func() {
		defer func() {
			ixf.controlCh = nil
		}()
	STOP:
		for {
			select {
			case <-ixf.syncCh:
				if stopping {
					log.Warn().Msg("Attempted to sync an index updater in a closing state. Quitting...")
					break STOP
				}
				if err := ixf.Sync(); err != nil {
					log.Err(err).Msg("Error when pulling latest records into the local wallet index")
				}
			case <-ixf.controlCh:
				log.Debug().Msg("Shutting down main loop for index updater")

				break STOP
			}
		}
	}()

	// subscribe to block events

	newBlockCh, err := ns.Subscribe(model.NTopicNewBlock)
	if err != nil {
		return err
	}

	// set up a ticker, so it's never equal to nil. Stop immediately, if the timer isn't on.
	ticker := time.NewTicker(time.Second)

	log.Debug().Dur("int", forceSyncInterval).Msg("Index Feeder sync interval")
	if forceSyncInterval > 0 {
		ticker = time.NewTicker(forceSyncInterval)
	} else {
		ticker.Stop()
	}

	go func() {
		defer func() {
			ixf.eventControlCh = nil
		}()
	STOP:
		for {
			select {
			case msg := <-newBlockCh:
				if msg == nil {
					break
				}

				log.Info().Msg("Sync index after new block notification")

				select {
				case ixf.syncCh <- true:
					// success
				default:
					log.Warn().Msg("No receiver for index sync message (block notification)")
				}
			case <-ticker.C:
				log.Info().Msg("Sync index after sync interval")

				select {
				case ixf.syncCh <- true:
					// success
				default:
					log.Warn().Msg("No receiver for index sync message (timer)")
					ticker.Stop()
				}
			case <-ixf.eventControlCh:
				log.Debug().Msg("Shutting event loop for index updater")

				stopping = true

				_ = ns.Unsubscribe(newBlockCh)
				ticker.Stop()

				// shutdown main loop goroutine
				select {
				case ixf.controlCh <- "stop":
				default:
				}

				break STOP
			}
		}
	}()

	return nil
}

func (ixf *IndexUpdater) StopSyncOnEvents() {
	if ixf.eventControlCh != nil {
		ixf.eventControlCh <- "stop"
	}
}

func (ixf *IndexUpdater) Close() error {
	log.Debug().Msg("Stopping Index Updater")

	ixf.StopSyncOnEvents()

	log.Debug().Msg("Closing Index Updater")
	return nil
}

func (ixf *IndexUpdater) Sync() error {
	ixf.syncMutex.Lock()
	defer ixf.syncMutex.Unlock()

	defer measure.ExecTime("updater.Sync")()

	log.Debug().Msg("Sync indexes in Index Updater")

	_, err := ixf.scanner.Scan()

	return err
}

func (ixf *IndexUpdater) SyncNoWait() {
	select {
	case ixf.syncCh <- true:
	default:
	}
}

type consumer struct {
	index          index.Writer
	accountIndex   AccountIndex
	accountID      string
	dataWallets    map[string]DataWallet
	sub            *scanner.IndexSubscription
	accountUpdates []*AccountUpdate

	ledger          model.Ledger
	offChainStorage model.OffChainStorage
	blobManager     model.BlobManager
}

var _ scanner.IndexBlockConsumer = (*consumer)(nil)

func newConsumer(dw DataWallet, iw index.Writer) (*consumer, error) {
	c := &consumer{
		index:           iw,
		accountID:       dw.ID(),
		dataWallets:     map[string]DataWallet{dw.ID(): dw},
		ledger:          dw.Services().Ledger(),
		offChainStorage: dw.Services().OffChainStorage(),
		blobManager:     dw.Services().BlobManager(),
	}

	if ai, ok := iw.(AccountIndex); ok {
		c.accountIndex = ai
	}

	return c, nil
}

func (c *consumer) ConsumeBlock(ctx context.Context, indexID string, partyLookup scanner.PartyLookup, n scanner.BlockNotification) error {

	var returnError error

	for _, dsn := range n.Datasets {
		lockerID, participantID, sharedSecret, acceptedAtBlock := partyLookup(dsn.KeyID)

		effectiveBlock := n.Block
		if acceptedAtBlock > n.Block {
			effectiveBlock = acceptedAtBlock
		}

		r := dsn.Record
		key := dsn.Key
		iw := c.index

		if r.Operation == model.OpTypeLease {
			if r.Status == model.StatusPublished {

				opRecBytes, err := c.offChainStorage.GetOperation(r.OperationAddress)
				if err != nil {
					log.Error().Str("rid", r.ID).Err(err).Msg("Failed to read ledger operation")
					return err
				}

				if r.Flags&model.RecordFlagPublic == 0 {
					skBytes, err := base64.StdEncoding.DecodeString(sharedSecret)
					if err != nil {
						return err
					}
					sk := model.Hash("Symmetrical key", append(skBytes, key...))

					symKey := &model.AESKey{}
					copy(symKey[:], sk)

					opRecBytes, err = model.DecryptAESCGM(opRecBytes, symKey)
					if err != nil {
						return err
					}
				}

				lease, err := model.NewLease(opRecBytes)
				if err != nil {
					return err
				}

				ds := dataset.NewDataSetImpl(r, lease, n.Block, lockerID, participantID, c.blobManager)

				if lease.Impression.MetaResource.ContentType == AccountUpdateType {
					err = c.processAccountUpdateMessage(ds)
					if err != nil {
						if errors.Is(err, scanner.ErrIndexResultPending) {
							returnError = err
						} else {
							return err
						}
					}
				}

				if err = iw.AddLease(ds, effectiveBlock); err != nil {
					return err
				}
			} else if r.Status == model.StatusRevoked {
				// we want to add revoked leases for the record
				ds := dataset.NewRevokedDataSetImpl(r, n.Block, lockerID, participantID)
				if err := iw.AddLease(ds, effectiveBlock); err != nil {
					return err
				}
			}
		} else if r.Operation == model.OpTypeLeaseRevocation {
			ds := dataset.NewRevokedDataSetImpl(r, n.Block, lockerID, participantID)
			if err := iw.AddLeaseRevocation(ds); err != nil {
				return err
			}
		}
	}

	return returnError
}

func (c *consumer) addLocker(dw DataWallet, lockerID string) error {

	l, err := dw.GetLocker(lockerID)
	if err != nil {
		return err
	}
	if err = c.index.AddLockerState(dw.ID(), l.ID(), l.Raw().FirstBlock); err != nil {
		if errors.Is(err, index.ErrLockerStateExists) {
			log.Warn().Str("lid", lockerID).Msg("Attempted to add a locker state that already exists in the index")
		} else {
			return err
		}
	}

	return c.sub.AddLockers(scanner.LockerEntry{
		Locker:    l.Raw(),
		LastBlock: l.Raw().FirstBlock,
	})
}

func (c *consumer) getDataWallet(accountID string) (DataWallet, error) {
	dw, found := c.dataWallets[accountID]
	if !found {
		var err error
		dw, err = c.dataWallets[c.accountID].GetSubAccountWallet(accountID)
		if err != nil {
			return nil, err
		}
		c.dataWallets[accountID] = dw
	}
	return dw, nil
}

var (
	hostedLevels  = []model.AccessLevel{model.AccessLevelHosted, model.AccessLevelManaged}
	managedLevels = []model.AccessLevel{model.AccessLevelManaged}
)

func (c *consumer) NotifyScanCompleted(topBlock int64) error {
	if topBlock > 0 {
		// top block was updated
		if err := c.index.UpdateTopBlock(topBlock); err != nil {
			return err
		}
	}

	// add new lockers after we invoke UpdateTopBlock, so that we don't set the top block
	// for the unprocessed lockers.

	for _, update := range c.accountUpdates {
		dw, err := c.getDataWallet(update.AccountID)
		if err != nil {
			return err
		}

		for _, lid := range update.LockersOpened {
			err = c.addLocker(dw, lid)
			if err != nil {
				if errors.Is(err, storage.ErrLockerNotFound) {
					log.Warn().Str("lid", lid).Msg("Locker from AccountUpdate message not accessible. Skipping...")
					continue
				} else {
					return err
				}
			}
		}

		for _, subID := range update.SubAccountsAdded {
			log.Debug().Str("subID", subID).Msg("Processing new sub-account")

			subDW, err := c.getDataWallet(subID)
			if err != nil {
				return err
			}

			var levels []model.AccessLevel
			if subDW.Account().AccessLevel >= model.AccessLevelHosted {
				levels = hostedLevels
			} else {
				levels = managedLevels
			}

			for _, lvl := range levels {
				rootLocker, err := subDW.GetRootLocker(lvl)
				if err != nil {
					if errors.Is(err, storage.ErrLockerNotFound) {
						log.Warn().Str("lid", rootLocker.ID()).Msg("Sub-account's root locker message not accessible. Skipping...")
						continue
					} else {
						return err
					}
				}
				err = c.addLocker(subDW, rootLocker.ID())
				if err != nil {
					return err
				}
				log.Debug().Int32("lvl", int32(lvl)).Str("lid", rootLocker.ID()).Msg("Added sub-account's root locker")
			}
		}

		if c.accountIndex != nil {
			if err = ApplyAccountUpdate(c.accountIndex, update, dw); err != nil {
				return err
			}
		}
	}

	if c.sub.Status() != scanner.ScanStatusError {
		c.sub.SetStatus(scanner.ScanStatusActive)
	}

	c.accountUpdates = nil

	return nil
}

func (c *consumer) SetSubscription(sub scanner.Subscription) {
	c.sub = sub.(*scanner.IndexSubscription)
}

func (c *consumer) processAccountUpdateMessage(ds model.DataSet) error {

	var msg AccountUpdate
	if err := ds.DecodeMetaResource(&msg); err != nil {
		return err
	}

	if len(msg.LockersOpened) > 0 || len(msg.IdentitiesAdded) > 0 || len(msg.SubAccountsAdded) > 0 {
		c.accountUpdates = append(c.accountUpdates, &msg)
		return scanner.ErrIndexResultPending
	} else {
		return nil
	}
}

func ForceSyncRootIndex(dw DataWallet) error {
	defer measure.ExecTime("wallet.ForceSyncRootIndex")()

	log.Warn().Msg("Force root index sync: adding missing locker states and syncing")

	ix, err := dw.RootIndex()
	if err != nil {
		return err
	}

	if ix == nil {
		log.Warn().Msg("No root index found")
		return nil
	}

	iw, _ := ix.Writer()

	// find new lockers

	walletLockerStates, err := iw.LockerStates()
	if err != nil {
		return err
	}
	walletLockerMap := make(map[string]index.LockerState)
	for _, ls := range walletLockerStates {
		walletLockerMap[ls.ID] = ls
	}

	lockerList, err := dw.GetLockers()
	if err != nil {
		return err
	}

	for _, l := range lockerList {
		if _, found := walletLockerMap[l.ID]; !found {
			if err = iw.AddLockerState(dw.ID(), l.ID, l.FirstBlock); err != nil {
				return err
			}
		}
	}

	updated, err := dw.IndexUpdater(ix)
	if err != nil {
		return err
	}

	err = updated.Sync()
	if err != nil {
		return err
	}

	return nil
}
