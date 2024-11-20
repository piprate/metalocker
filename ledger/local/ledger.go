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
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/claudiu/gocron"
	"github.com/piprate/metalocker/ledger"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/cmdbase"
	"github.com/piprate/metalocker/services/notification"
	"github.com/piprate/metalocker/utils"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/utils/measure"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.etcd.io/bbolt"
)

func init() {
	ledger.Register("local", CreateLedgerConnector)
}

type BoltLedger struct {
	client      *utils.BoltClient
	instantMode bool
	scheduler   *gocron.Scheduler
	ns          notification.Service

	records chan model.Record
	ticks   chan bool
	control chan string

	maxRecordsPerBlock int

	pendingRecords []string
}

var _ model.Ledger = (*BoltLedger)(nil)

type LocalBlock struct {
	Number     int64  `json:"number"`
	Hash       string `json:"hash"`
	ParentHash string `json:"parentHash,omitempty"`
	Nonce      string `json:"nonce,omitempty"`
}

func (b *LocalBlock) Seal() error {

	buf := bytes.NewBuffer(nil)
	parentIDVal, _ := base64.StdEncoding.DecodeString(b.ParentHash)
	buf.Write(parentIDVal)
	nonceVal, _ := base64.StdEncoding.DecodeString(b.Nonce)
	buf.Write(nonceVal)
	buf.WriteString(strconv.Itoa(int(b.Number)))

	hash := model.Hash("block construction", buf.Bytes())

	b.Hash = base64.StdEncoding.EncodeToString(hash)

	return nil
}

func NewBoltLedger(ctx context.Context, dbFilepath string, ns notification.Service, maxRecordsPerBlock int, blockCheckInterval uint64) (*BoltLedger, error) {
	log.Info().Str("db", dbFilepath).Uint64("interval", blockCheckInterval).Msg("Initialising local ledger")

	// open Bolt DB

	bc, err := utils.NewBoltClient(utils.AbsPathify(dbFilepath), InstallLedgerSchema)
	if err != nil {
		return nil, err
	}

	bl := &BoltLedger{
		client:             bc,
		ns:                 ns,
		maxRecordsPerBlock: maxRecordsPerBlock,
		records:            make(chan model.Record, 100),
		ticks:              make(chan bool),
		control:            make(chan string),
		pendingRecords:     make([]string, 0),
	}

	// generate genesis block, if it doesn't exist
	if _, err = bl.GetGenesisBlock(ctx); err != nil {
		if errors.Is(err, model.ErrBlockNotFound) {
			if err = generateNewBlock(ctx, bl, ""); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// ensure there's a current block session which is ready to accept records

	curSessionID := bl.CurrentBlockSession()
	if curSessionID == "" {

		// open new session

		curSessionID, err = bl.OpenNewBlockSession()
		if err != nil {
			return nil, err
		}
		log.Warn().Str("id", curSessionID).Msg("No current block session defined. Creating new session")
	} else {

		// load existing unconfirmed records

		if err = bl.loadUnconfirmedRecords(curSessionID); err != nil {
			return nil, err
		}

		if len(bl.pendingRecords) > 0 {
			log.Info().Int("count", len(bl.pendingRecords)).Msg("Loaded unconfirmed ledger records")
		}

	}

	bl.startLoop(blockCheckInterval) //nolint:contextcheck

	return bl, nil
}

func (bl *BoltLedger) SubmitRecord(ctx context.Context, r *model.Record) error {

	if err := r.Validate(); err != nil {
		return err
	}

	bl.records <- *r

	return nil
}

func (bl *BoltLedger) GetRecord(ctx context.Context, rid string) (*model.Record, error) {

	var recordBytes []byte
	var r model.Record
	found := false
	err := bl.client.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(RecordsKey))
		if b == nil {
			return fmt.Errorf("bucket %s not found", RecordsKey)
		}

		recordBytes = b.Get([]byte(rid))

		if recordBytes != nil {

			found = true

			if err := jsonw.Unmarshal(recordBytes, &r); err != nil {
				return err
			}

			b = tx.Bucket([]byte(RecordStatesKey))
			if b == nil {
				return fmt.Errorf("bucket %s not found", RecordStatesKey)
			}

			rsBytes := b.Get([]byte(rid))

			if rsBytes != nil {
				var rs model.RecordState
				err := jsonw.Unmarshal(rsBytes, &rs)
				if err != nil {
					return err
				}
				r.Status = rs.Status
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if found {
		return &r, nil
	} else {
		return nil, model.ErrRecordNotFound
	}
}

func (bl *BoltLedger) GetRecordState(ctx context.Context, rid string) (*model.RecordState, error) {
	b, err := bl.client.FetchBytes(RecordStatesKey, rid)
	if err != nil {
		return nil, err
	}
	if b != nil {
		var rs model.RecordState
		err = jsonw.Unmarshal(b, &rs)
		if err != nil {
			return nil, err
		}
		return &rs, nil
	} else {
		// we don't know if the requested record doesn't exist or haven't yet reached the ledger
		return nil, nil
	}
}

func (bl *BoltLedger) GetBlock(ctx context.Context, bn int64) (*model.Block, error) {
	blockKey := utils.Int64ToString(bn)
	b, err := bl.client.FetchBytes(BlocksKey, blockKey)
	if err != nil {
		return nil, err
	}

	if b == nil {
		return nil, model.ErrBlockNotFound
	}

	var bb model.Block
	err = jsonw.Unmarshal(b, &bb)
	if err != nil {
		return nil, err
	}
	return &bb, nil
}

func (bl *BoltLedger) GetBlockRecords(ctx context.Context, bn int64) ([][]string, error) {
	blockKey := utils.Int64ToString(bn)
	resMap := make(map[int][]string)
	err := bl.client.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BlockCompositionsKey))
		if b == nil {
			return fmt.Errorf("bucket %s not found", BlockCompositionsKey)
		}
		b = b.Bucket([]byte(blockKey))
		if b == nil {
			log.Warn().Int64("number", bn).Msg("Block composition not found")
			return nil
		}

		var idx int
		err := b.ForEach(func(k, v []byte) error {
			recID, routingKey, keyIndex := unpack(v)

			idx, _ = strconv.Atoi(string(k))

			resMap[idx] = []string{recID, routingKey, keyIndex}
			return nil
		})

		return err
	})
	if err != nil {
		return nil, err
	}

	res := make([][]string, len(resMap))
	for idx, rec := range resMap {
		res[idx] = rec
	}

	return res, nil
}

func (bl *BoltLedger) GetGenesisBlock(ctx context.Context) (*model.Block, error) {
	return bl.GetBlock(ctx, 0)
}

func (bl *BoltLedger) GetTopBlock(ctx context.Context) (*model.Block, error) {
	var blockBytes []byte
	err := bl.client.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(ControlsKey))
		if b == nil {
			return fmt.Errorf("bucket %s not found", ControlsKey)
		}

		topNumberBytes := b.Get([]byte(TopBlockNumberKey))

		if len(topNumberBytes) > 0 {
			b := tx.Bucket([]byte(BlocksKey))
			if b == nil {
				return fmt.Errorf("bucket %s not found", BlocksKey)
			}

			blockBytes = b.Get(topNumberBytes)

			return nil
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if blockBytes == nil {
		return nil, model.ErrBlockNotFound
	}

	var bb model.Block
	err = jsonw.Unmarshal(blockBytes, &bb)
	if err != nil {
		return nil, err
	}
	return &bb, nil
}

func (bl *BoltLedger) GetChain(ctx context.Context, startNumber int64, depth int) ([]*model.Block, error) {
	result := make([]*model.Block, 0)
	var i int64 = 0
	err := bl.client.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BlocksKey))
		if b == nil {
			return fmt.Errorf("bucket %s not found", BlocksKey)
		}

		for ; i < int64(depth); i++ {

			val := b.Get([]byte(utils.Int64ToString(startNumber + i)))

			if val == nil {
				break
			}

			var bb model.Block
			err := jsonw.Unmarshal(val, &bb)
			if err != nil {
				return err
			}

			result = append(result, &bb)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (bl *BoltLedger) GetDataAssetState(ctx context.Context, id string) (model.DataAssetState, error) {

	res := model.DataAssetStateKeep
	err := bl.client.DB.View(func(tx *bbolt.Tx) error {
		dasb := tx.Bucket([]byte(DataAssetStatesKey))
		cntBytes := dasb.Get([]byte(id))
		if cntBytes != nil {
			if utils.BytesToUint32(cntBytes) == 0 {
				res = model.DataAssetStateRemove
			}
		} else {
			res = model.DataAssetStateNotFound
		}

		return nil
	})
	if err != nil {
		return model.DataAssetStateKeep, err
	}

	return res, nil
}

func (bl *BoltLedger) GetAssetHead(ctx context.Context, headID string) (*model.Record, error) {
	var recordBytes []byte
	var r model.Record
	found := false
	err := bl.client.DB.View(func(tx *bbolt.Tx) error {
		hb := tx.Bucket([]byte(HeadsKey))
		if hb == nil {
			return fmt.Errorf("bucket %s not found", HeadsKey)
		}

		headBytes := hb.Get([]byte(headID))
		if headBytes == nil {
			return model.ErrAssetHeadNotFound
		}

		rid := string(headBytes)

		b := tx.Bucket([]byte(RecordsKey))
		if b == nil {
			return fmt.Errorf("bucket %s not found", RecordsKey)
		}

		recordBytes = b.Get([]byte(rid))

		if recordBytes != nil {

			found = true

			if err := jsonw.Unmarshal(recordBytes, &r); err != nil {
				return err
			}

			b = tx.Bucket([]byte(RecordStatesKey))
			if b == nil {
				return fmt.Errorf("bucket %s not found", RecordStatesKey)
			}

			rsBytes := b.Get([]byte(rid))

			if rsBytes != nil {
				var rs model.RecordState
				err := jsonw.Unmarshal(rsBytes, &rs)
				if err != nil {
					return err
				}
				r.Status = rs.Status
			}
		} else {
			return model.ErrAssetHeadNotFound
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if found {
		return &r, nil
	} else {
		return nil, model.ErrRecordNotFound
	}
}

func (bl *BoltLedger) Close() error {
	if bl.scheduler != nil {
		log.Info().Msg("Stopping block scheduler")
		bl.scheduler.Clear()
	}

	if bl.control != nil {
		bl.control <- "stop"
	}

	if bl.client != nil {
		log.Info().Msg("Closing local Bolt ledger")
		err := bl.client.Close()
		return err
	}

	return nil
}

func (bl *BoltLedger) CurrentBlockSession() string {
	v, err := bl.client.FetchString(ControlsKey, CurrentSessionIDKey)
	if err != nil {
		log.Err(err).Msg("Error when reading the current block session ID")
	}
	return v
}

func (bl *BoltLedger) OpenNewBlockSession() (string, error) {
	curSessionID := fmt.Sprintf("block%d", time.Now().Unix())

	// create new nested bucket

	if err := bl.client.DB.Update(func(tx *bbolt.Tx) error {
		urBucket := tx.Bucket([]byte(UnconfirmedRecordsKey))

		_, err := urBucket.CreateBucketIfNotExists([]byte(curSessionID))
		if err != nil {
			return err
		}

		if err := bl.client.UpdateInline(tx, ControlsKey, CurrentSessionIDKey, []byte(curSessionID)); err != nil {
			log.Err(err).Msg("Error when writing the current block session ID")
			return err
		}

		return nil
	}); err != nil {
		return "", err
	}

	bl.pendingRecords = make([]string, 0)

	return curSessionID, nil
}

func (bl *BoltLedger) updateRecordState(tx *bbolt.Tx, rid string, status model.RecordStatus, blockNumber int64) error {
	b := tx.Bucket([]byte(RecordStatesKey))
	if b == nil {
		return fmt.Errorf("bucket %s not found", RecordStatesKey)
	}

	val := b.Get([]byte(rid))

	if val == nil {
		return fmt.Errorf("record state not found for %s", rid)
	}

	var rs model.RecordState
	err := jsonw.Unmarshal(val, &rs)
	if err != nil {
		return err
	}

	if rs.Status == model.StatusRevoked {
		return errors.New("cannot change status of revoked record")
	}

	rs.Status = status
	if blockNumber != -1 {
		rs.BlockNumber = blockNumber
	}

	err = b.Put([]byte(rid), rs.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func (bl *BoltLedger) startLoop(blockCheckInterval uint64) {
	go func() {
		defer func() {
			bl.control = nil
		}()

		ctx := context.Background()
		sampledLog := log.Sample(&zerolog.BasicSampler{N: 10})
	STOP:
		for {
			select {
			case r := <-bl.records:

				if err := bl.SaveRecord(&r); err != nil {
					log.Err(err).Str("rid", r.ID).Msg("Failed to save ledger record")
				}
				for len(bl.records) > 0 && len(bl.pendingRecords) < bl.maxRecordsPerBlock {
					r = <-bl.records
					if err := bl.SaveRecord(&r); err != nil {
						log.Err(err).Str("rid", r.ID).Msg("Failed to save ledger record")
					}
				}

				if bl.instantMode || len(bl.pendingRecords) >= bl.maxRecordsPerBlock {
					if err := generateNewBlock(ctx, bl, ""); err != nil {
						log.Err(err).Msg("Failed to generate new block")
					}
				}
			case <-bl.ticks:
				sampledLog.Debug().Msg("Block generation check")
				if len(bl.pendingRecords) > 0 {
					if err := generateNewBlock(ctx, bl, ""); err != nil {
						log.Err(err).Msg("Failed to generate new block")
					}
				}
			case <-bl.control:
				log.Info().Msg("Shutting down main loop for local ledger")
				break STOP
			}
		}
	}()

	if blockCheckInterval > 0 {
		log.Warn().Uint64("interval", blockCheckInterval).Msg("Local Ledger block check interval")
		bl.scheduler = gocron.NewScheduler()
		bl.scheduler.Every(blockCheckInterval).Seconds().Do(blockCheck, bl)
		bl.scheduler.Start()
	} else {
		bl.instantMode = true
	}
}

func pack(rec *model.Record) []byte {
	return []byte(fmt.Sprintf("%s,%s,%d", rec.ID, rec.RoutingKey, rec.KeyIndex))
}

func unpack(b []byte) (string, string, string) {
	parts := strings.Split(string(b), ",")
	return parts[0], parts[1], parts[2]
}

//nolint:gocyclo
func (bl *BoltLedger) SubmitNewBlock(block *model.Block, records []*model.Record) error {
	defer measure.ExecTime("local.SubmitNewBlock")()

	bb, err := jsonw.Marshal(block)
	if err != nil {
		return err
	}

	log.Info().Int64("number", block.Number).Msg("----==== NEW BLOCK ====----")

	blockKey := utils.Int64ToString(block.Number)

	if err := bl.client.DB.Update(func(tx *bbolt.Tx) error {
		if err = bl.client.UpdateInline(tx, BlocksKey, blockKey, bb); err != nil {
			return err
		}

		compositionBucket := tx.Bucket([]byte(BlockCompositionsKey))

		blockCompBucket, err := compositionBucket.CreateBucketIfNotExists([]byte(blockKey))
		if err != nil {
			return err
		}

		idx := 0
		for _, rec := range records {

			switch rec.Operation {
			case model.OpTypeLease:
				// update data asset counters
				dasb := tx.Bucket([]byte(DataAssetStatesKey))
				for _, id := range append(rec.DataAssets, rec.OperationAddress) {
					var counter uint32
					cntBytes := dasb.Get([]byte(id))
					if cntBytes != nil {
						counter = utils.BytesToUint32(cntBytes)
					}
					counter++

					if err = dasb.Put([]byte(id), utils.Uint32ToBytes(counter)); err != nil {
						return err
					}
				}
				if err = bl.updateRecordState(tx, rec.ID, model.StatusPublished, block.Number); err != nil {
					return err
				}
			case model.OpTypeLeaseRevocation:
				rb := tx.Bucket([]byte(RecordsKey))
				if blockCompBucket == nil {
					return fmt.Errorf("bucket %s not found", RecordsKey)
				}

				subjBytes := rb.Get([]byte(rec.SubjectRecord))

				if subjBytes == nil {
					return fmt.Errorf("subject record not found for revocation record %s", rec.ID)
				}

				var subj model.Record
				err = jsonw.Unmarshal(subjBytes, &subj)
				if err != nil {
					return err
				}

				// apply revocations
				if err = bl.updateRecordState(tx, rec.SubjectRecord, model.StatusRevoked, -1); err != nil {
					if err := bl.updateRecordState(tx, rec.ID, model.StatusFailed, -1); err != nil {
						log.Err(err).Str("rid", rec.ID).Msg("Error when setting record status as failed")
					}
					continue
				}

				// update data asset counters

				dasb := tx.Bucket([]byte(DataAssetStatesKey))
				for _, id := range append(subj.DataAssets, subj.OperationAddress) {
					var counter uint32
					cntBytes := dasb.Get([]byte(id))
					if cntBytes != nil {
						counter = utils.BytesToUint32(cntBytes)
					}

					if counter > 0 {
						// counter may go negative if the same record is revoked multiple times.
						// For now, we allow multiple revocations.
						counter--
					}

					if err = dasb.Put([]byte(id), utils.Uint32ToBytes(counter)); err != nil {
						return err
					}
				}
				if err = bl.updateRecordState(tx, rec.ID, model.StatusPublished, block.Number); err != nil {
					return err
				}
			case model.OpTypeAssetHead:
				hb := tx.Bucket([]byte(HeadsKey))

				prevHeadRecordID := string(hb.Get([]byte(rec.HeadID)))

				if err = hb.Put([]byte(rec.HeadID), []byte(rec.ID)); err != nil {
					return err
				}

				if prevHeadRecordID != "" {
					if err = bl.updateRecordState(tx, prevHeadRecordID, model.StatusRevoked, -1); err != nil {
						return err
					}
				}

				if err = bl.updateRecordState(tx, rec.ID, model.StatusPublished, block.Number); err != nil {
					return err
				}
			default:
				if err = bl.updateRecordState(tx, rec.ID, model.StatusFailed, block.Number); err != nil {
					return err
				}
			}

			err = blockCompBucket.Put([]byte(strconv.Itoa(idx)), pack(rec))
			if err != nil {
				return err
			}

			idx++
		}

		// Consider this block published

		if err := bl.client.UpdateInline(tx, ControlsKey, TopBlockNumberKey, []byte(blockKey)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	if bl.ns != nil {
		_ = bl.ns.Publish(&model.NewBlockMessage{
			Type:   model.MessageTypeNewBlockNotification,
			Number: block.Number,
		}, false, false, model.NTopicNewBlock)
	}

	return nil
}

func (bl *BoltLedger) loadUnconfirmedRecords(sessionID string) error {
	return bl.client.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(UnconfirmedRecordsKey))
		if b == nil {
			return fmt.Errorf("bucket %s not found", UnconfirmedRecordsKey)
		}
		b = b.Bucket([]byte(sessionID))

		recMap := make(map[int]string)
		err := b.ForEach(func(k, v []byte) error {
			idx, _ := strconv.Atoi(string(k))
			recMap[idx] = string(v)
			return nil
		})

		bl.pendingRecords = make([]string, len(recMap))
		for idx, id := range recMap {
			bl.pendingRecords[idx] = id
		}

		return err
	})
}

func (bl *BoltLedger) SaveRecord(r *model.Record) error {

	if err := bl.client.DB.Update(func(tx *bbolt.Tx) error {

		// save the record

		if err := bl.client.UpdateInline(tx, RecordsKey, r.ID, r.Bytes()); err != nil {
			return err
		}

		// save record state

		rs := model.RecordState{
			Status: model.StatusPending,
		}
		if err := bl.client.UpdateInline(tx, RecordStatesKey, r.ID, rs.Bytes()); err != nil {
			return err
		}

		// add the record into the list of unconfirmed records for the current session

		urBucket := tx.Bucket([]byte(UnconfirmedRecordsKey))

		// get current block session

		b := tx.Bucket([]byte(ControlsKey))
		if b == nil {
			return fmt.Errorf("bucket %s not found", ControlsKey)
		}

		cs := b.Get([]byte(CurrentSessionIDKey))

		b = urBucket.Bucket(cs)
		if b == nil {
			if len(cs) == 0 {
				return errors.New("no active block session found")
			} else {
				return fmt.Errorf("current block session key not found for %s", cs)
			}
		}
		err := b.Put([]byte(strconv.Itoa(len(bl.pendingRecords))), []byte(r.ID))
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	bl.pendingRecords = append(bl.pendingRecords, r.ID)

	return nil
}

func blockCheck(bl *BoltLedger) {
	bl.ticks <- true
}

func generateNonce(seed string, size int) ([]byte, error) {
	var randSeed io.Reader
	if seed != "" {
		randSeed = strings.NewReader(seed)
	} else {
		randSeed = rand.Reader
	}

	nonce := make([]byte, size)
	_, err := io.ReadFull(randSeed, nonce)
	if err != nil {
		return nil, err
	}
	return nonce, nil
}

func generateNewBlock(ctx context.Context, bl *BoltLedger, seed string) error {

	records := make([]*model.Record, 0)
	for _, id := range bl.pendingRecords {
		r, err := bl.GetRecord(ctx, id)
		if err != nil {
			return err
		}

		// validate record

		if r.Operation == model.OpTypeLeaseRevocation {
			subj, err := bl.GetRecord(ctx, r.SubjectRecord)
			if err != nil {
				return err
			}

			if len(r.RevocationProof) != 1 {
				return errors.New("revocation failed: bad revocation proof format")
			}
			proof, _ := base64.StdEncoding.DecodeString(r.RevocationProof[0])
			subjAC := sha256.Sum256(proof)

			if base64.StdEncoding.EncodeToString(subjAC[:]) != subj.AuthorisingCommitment {
				return errors.New("revocation failed: bad revocation proof")
			}
		}

		records = append(records, r)
	}

	// find the top block

	prevBlockHash := ""
	prevBlock, err := bl.GetTopBlock(ctx)
	if err != nil && !errors.Is(err, model.ErrBlockNotFound) {
		return err
	}

	// generate new block

	var number int64 = 0
	if prevBlock != nil {
		prevBlockHash = prevBlock.Hash
		number = prevBlock.Number + 1
	}

	nonce, err := generateNonce(seed, 32)
	if err != nil {
		return err
	}

	newLocalBlock := &LocalBlock{
		Number:     number,
		ParentHash: prevBlockHash,
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
	}

	if err := newLocalBlock.Seal(); err != nil {
		return err
	}

	newBlock := &model.Block{
		Number:     newLocalBlock.Number,
		Hash:       newLocalBlock.Hash,
		ParentHash: newLocalBlock.ParentHash,
	}

	err = bl.SubmitNewBlock(newBlock, records)
	if err != nil {
		return err
	}

	if _, err := bl.OpenNewBlockSession(); err != nil {
		return err
	}

	return nil
}

func CreateLedgerConnector(ctx context.Context, params ledger.Parameters, ns notification.Service, resolver cmdbase.ParameterResolver) (model.Ledger, error) {
	dbFilepath, ok := params["dbFile"].(string)
	if !ok {
		return nil, errors.New("parameter not found: dbfile. Can't start ledger connector")
	}
	blockCheckInterval, _ := params["interval"].(float64)

	bl, err := NewBoltLedger(ctx, dbFilepath, ns, 1000, uint64(blockCheckInterval))
	return bl, err
}
