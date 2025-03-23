package onflow

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alitto/pond/v2"
	"github.com/gammazero/deque"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/piprate/metalocker/ledger"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/cmdbase"
	"github.com/piprate/metalocker/services/notification"
	"github.com/piprate/metalocker/utils"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/utils/measure"
	"github.com/piprate/splash"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	saveScannerProgressIntervalBlocks = 1000
	maxBlocks                         = 500
	emulatorTimeoutSeconds            = 5
)

func init() {
	ledger.Register("flow", CreateLedgerConnector)
}

type Ledger struct {
	network       string
	conn          *splash.Connector
	scriptsEngine *splash.TemplateEngine
	nodeAccount   *accounts.Account
	ns            notification.Service

	cacheDBClient *utils.BoltClient

	privateKey       crypto.PrivateKey
	dequeMutex       *sync.Mutex
	keyDeque         deque.Deque[uint32]
	recordWorkerPool pond.ResultPool[model.RecordStatus]

	latestBlockHeight uint64
	latestBlockNumber uint64
	eventTypes        []string

	syncMutex *sync.Mutex

	controlCh chan string
	stopping  bool
}

var _ model.Ledger = (*Ledger)(nil)

func (l *Ledger) Close() error {
	log.Debug().Msg("Stopping Flow Ledger connector")

	l.StopContinuousSync()

	log.Debug().Msg("Closing Flow Ledger connector")

	return nil
}

func (l *Ledger) StartLedgerEvents(ctx context.Context, syncOnStart bool, syncInterval time.Duration) error {
	if syncOnStart {
		if err := l.Sync(ctx); err != nil {
			return err
		}
	}

	if l.conn.IsInMemoryEmulator() {
		return l.startPeriodicSync(ctx, syncInterval)
	} else {
		return l.startBlockSubscription(ctx)
	}
}

func (l *Ledger) startPeriodicSync(ctx context.Context, syncInterval time.Duration) error {
	if syncInterval == 0 {
		return nil
	}

	l.controlCh = make(chan string, 5)

	log.Debug().Dur("int", syncInterval).Msg("Setting Flow Scanner's sync interval")
	ticker := time.NewTicker(syncInterval)

	sampledLog := log.Sample(&zerolog.Rarely)

	go func() {
		defer func() {
			l.controlCh = nil
		}()
	STOP:
		for {
			select {
			case <-ticker.C:
				sampledLog.Debug().Msg("Flow Sync after sync interval")

				if l.stopping {
					log.Warn().Msg("Attempted to sync Flow Scanner in a closing state. Quitting...")
					ticker.Stop()
					break STOP
				}
				if err := l.Sync(ctx); err != nil {
					log.Err(err).Msg("Failed to sync with Flow")
				}
			case <-l.controlCh:
				log.Debug().Msg("Shutting event loop for Flow Scanner")
				ticker.Stop()
				break STOP
			}
		}
	}()

	return nil
}

func (l *Ledger) startBlockSubscription(ctx context.Context) error {
	l.controlCh = make(chan string, 5)

	sampledLog := log.Sample(&zerolog.Rarely)

	flowClient := l.conn.GRPCClient

	eventFilter := flow.EventFilter{}

	blockEvents, errChan, err := flowClient.SubscribeEventsByBlockHeight(ctx, l.latestBlockHeight, eventFilter)
	if err != nil {
		return err
	}

	reconnect := func(height uint64) {
		sampledLog.Warn().Uint64("height", height).Msg("Reconnecting...")
		blockEvents, errChan, err = flowClient.SubscribeEventsByBlockHeight(ctx, height, eventFilter)
		if err != nil {
			sampledLog.Err(err).Msg("Failed to reconnect")
		}
	}

	go func() {
		defer func() {
			l.controlCh = nil
		}()

		lastSavedHeight := l.latestBlockHeight
	STOP:
		for {
			select {
			case <-l.controlCh:
				log.Debug().Msg("Shutting event loop for Flow Ledger connector")
				break STOP
			case <-ctx.Done():
				return
			case eventData, ok := <-blockEvents:
				if !ok {
					if ctx.Err() != nil {
						return // graceful shutdown
					}
					// unexpected close
					reconnect(l.latestBlockHeight + 1)
					continue
				}

				sampledLog.Debug().Uint64("height", eventData.Height).Msg("New Flow block found")

				var eventsFound bool
				if len(eventData.Events) > 0 {
					if eventsFound, err = l.processEvents(ctx, []flow.BlockEvents{eventData}); err != nil {
						log.Err(err).Uint64("height", eventData.Height).Msg("Failed to process Flow events from block")
					}
				}

				if eventsFound || (eventData.Height-lastSavedHeight > saveScannerProgressIntervalBlocks) {
					if err = l.cacheDBClient.Update(ControlsKey, TopFlowBlockHeightKey, []byte(utils.Uint64ToString(eventData.Height))); err != nil {
						log.Err(err).Uint64("height", eventData.Height).Msg("Failed to update latest block height")
					}
					lastSavedHeight = eventData.Height
				}

				l.latestBlockHeight = eventData.Height

			case err, ok := <-errChan:
				if !ok {
					if ctx.Err() != nil {
						return // graceful shutdown
					}
					// unexpected close
					reconnect(l.latestBlockHeight + 1)
					continue
				}

				log.Err(err).Msg("Error scanning Flow blockchain")
				reconnect(l.latestBlockHeight + 1)
				continue
			}
		}
	}()

	return nil
}

func (l *Ledger) StopContinuousSync() {
	if l.controlCh != nil {
		l.stopping = true
		l.controlCh <- "stop"
	}
}

func (l *Ledger) Sync(ctx context.Context) error {
	l.syncMutex.Lock()
	defer l.syncMutex.Unlock()

	block, err := l.conn.Services.GetBlock(ctx, flowkit.LatestBlockQuery)
	if err != nil {
		return err
	}
	latestBlockHeight := block.Height

	if latestBlockHeight > l.latestBlockHeight {
		return l.sync(ctx, l.latestBlockHeight+1, latestBlockHeight)
	}

	return nil
}

func (l *Ledger) sync(ctx context.Context, startHeight uint64, endHeight uint64) error {
	log.Info().Uint64("startHeight", startHeight).Uint64("endHeight", endHeight).Msg("Start scanning Flow block range")

	for currentStartHeight := startHeight; currentStartHeight <= endHeight && !l.stopping; {
		currentEndHeight := currentStartHeight + maxBlocks
		if currentEndHeight > endHeight {
			currentEndHeight = endHeight
		}

		log.Info().Uint64("currentStartHeight", currentStartHeight).
			Uint64("currentEndHeight", currentEndHeight).Msg("Scanning block sub-range")

		blockEvents, err := l.conn.Services.GetEvents(ctx, l.eventTypes, currentStartHeight, currentEndHeight, &flowkit.EventWorker{
			Count:           100,
			BlocksPerWorker: 1,
		})
		if err != nil {
			return err
		}

		if l.stopping {
			return errors.New("scanner interrupted")
		}

		if _, err = l.processEvents(ctx, blockEvents); err != nil {
			return err
		}

		l.latestBlockHeight = currentEndHeight

		if err := l.cacheDBClient.Update(ControlsKey, TopFlowBlockHeightKey, []byte(utils.Uint64ToString(currentEndHeight))); err != nil {
			return err
		}
		currentStartHeight += maxBlocks
	}

	return nil
}

func sortEvents(events []flow.BlockEvents) []flow.Event {
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].Height < events[j].Height
	})

	res := make([]flow.Event, 0)

	var prevHeight uint64 = 0
	var buf []flow.Event
	sortEventsFn := func(i, j int) bool {
		if buf[i].TransactionIndex != buf[j].TransactionIndex {
			return buf[i].TransactionIndex < buf[j].TransactionIndex
		} else {
			return buf[i].EventIndex < buf[j].EventIndex
		}
	}
	for _, eventBlock := range events {
		if eventBlock.Height != prevHeight {
			sort.Slice(buf, sortEventsFn)
			res = append(res, buf...)
			buf = nil
			prevHeight = eventBlock.Height
		}
		buf = append(buf, eventBlock.Events...)
	}

	if len(buf) > 0 {
		sort.Slice(buf, sortEventsFn)
		res = append(res, buf...)
	}

	return res
}

func (l *Ledger) processEvents(ctx context.Context, blockEvents []flow.BlockEvents) (bool, error) {
	//ld.PrintDocument("blockEvents", blockEvents)
	eventsFound := false

	events := sortEvents(blockEvents)

	var blockNumber uint64
	var blockRecords []*model.Record

	for _, e := range events {
		log.Debug().Str("id", e.Value.EventType.QualifiedIdentifier).Msg("NEW EVENT")
		switch e.Value.EventType.QualifiedIdentifier {
		case "MetaLocker.BlockOpened":
			if len(blockRecords) > 0 {
				if err := l.saveBlockRecords(blockNumber, blockRecords); err != nil {
					return true, err
				}
				blockRecords = nil
			}

			eventsFound = true
			b := parseEventBlockOpened(&e)
			log.Debug().Uint64("number", b.Number).Msg("NEW OPEN BLOCK RECEIVED")
			if err := l.addBlock(b, false); err != nil {
				return true, err
			}
			blockNumber = b.Number
		case "MetaLocker.BlockConfirmed":
			eventsFound = true
			b := parseEventBlockConfirmed(&e)
			log.Debug().Uint64("number", b.Number).Msg("NEW CONFIRMED BLOCK RECEIVED")
			if err := l.addBlock(b, true); err != nil {
				return true, err
			}

		case "MetaLocker.RecordPublished":
			rid := string(e.Value.SearchFieldByName("id").(cadence.String))
			rec, err := l.GetRecord(ctx, rid)
			if err != nil {
				return true, err
			}

			// the record may be revoked, if it's a rescan. Reset status to Published.
			rec.Status = model.StatusPublished

			if err := l.addRecord(rec); err != nil {
				return true, err
			}
			blockRecords = append(blockRecords, rec)
		case "MetaLocker.RecordRevoked":
			rid := string(e.Value.SearchFieldByName("id").(cadence.String))
			rec, err := l.GetRecord(ctx, rid)
			if err != nil {
				return true, err
			}

			rec.Status = model.StatusRevoked

			if err := l.addRecord(rec); err != nil {
				return true, err
			}
		}
	}
	if len(blockRecords) > 0 {
		if err := l.saveBlockRecords(blockNumber, blockRecords); err != nil {
			return true, err
		}
	}

	return eventsFound, nil
}

func parseEventBlockOpened(e *flow.Event) *model.Block {
	fieldMap := e.Value.FieldsMappedByName()
	//chain := fieldMap["chain"].(cadence.Address).String()
	id := uint64(fieldMap["id"].(cadence.UInt64))
	hexParentHash := string(fieldMap["parentHash"].(cadence.String))
	parentHash, _ := hex.DecodeString(hexParentHash)

	return &model.Block{
		Number:     id,
		ParentHash: base64.StdEncoding.EncodeToString(parentHash),
		Status:     model.BlockStatusOpen,
	}
}

func parseEventBlockConfirmed(e *flow.Event) *model.Block {
	fieldMap := e.Value.FieldsMappedByName()
	//chain := fieldMap["chain"].(cadence.Address).String()
	id := uint64(fieldMap["id"].(cadence.UInt64))
	hexParentHash := string(fieldMap["parentHash"].(cadence.String))
	parentHash, _ := hex.DecodeString(hexParentHash)
	hexHash := string(fieldMap["hash"].(cadence.String))
	hash, _ := hex.DecodeString(hexHash)

	return &model.Block{
		Number:     id,
		Hash:       base64.StdEncoding.EncodeToString(hash),
		ParentHash: base64.StdEncoding.EncodeToString(parentHash),
		Status:     model.BlockStatusConfirmed,
	}
}

func (l *Ledger) addBlock(b *model.Block, onConfirmation bool) error {
	bb, err := jsonw.Marshal(b)
	if err != nil {
		return err
	}

	key := utils.Uint64ToString(b.Number)
	if err = l.cacheDBClient.Update(BlocksKey, key, bb); err != nil {
		return err
	}

	if l.ns != nil && !onConfirmation {
		_ = l.ns.Publish(&model.NewBlockMessage{
			Type:   model.MessageTypeNewBlockNotification,
			Number: b.Number,
		}, false, false, model.NTopicNewBlock)
	}

	return nil
}

func (l *Ledger) loadBlock(bn uint64) (*model.Block, error) {
	defer measure.ExecTime("onflow.loadBlock")()

	b, err := l.cacheDBClient.FetchBytes(BlocksKey, utils.Uint64ToString(bn))
	if err != nil {
		return nil, err
	}
	if b != nil {
		var bb model.Block
		err = jsonw.Unmarshal(b, &bb)
		if err != nil {
			return nil, err
		}
		return &bb, nil
	} else {
		return nil, nil
	}
}

func (l *Ledger) saveBlockRecords(blockNumber uint64, records []*model.Record) error {
	body := make([][]string, len(records))

	for i, rec := range records {
		body[i] = []string{
			rec.ID, rec.RoutingKey, strconv.FormatUint(uint64(rec.KeyIndex), 10),
		}
	}

	bb, err := jsonw.Marshal(body)
	if err != nil {
		return err
	}

	log.Warn().Uint64("blockNumber", blockNumber).Int("len", len(records)).Msg("IN saveBlockRecords")

	return l.cacheDBClient.Update(BlockRecordsKey, utils.Uint64ToString(blockNumber), bb)
}

func (l *Ledger) loadBlockRecords(blockNumber uint64) ([][]string, error) {
	defer measure.ExecTime("onflow.loadBlockRecords")()

	b, err := l.cacheDBClient.FetchBytes(BlockRecordsKey, utils.Uint64ToString(blockNumber))
	if err != nil {
		return nil, err
	}
	if b != nil {
		var bb [][]string
		err = jsonw.Unmarshal(b, &bb)
		if err != nil {
			return nil, err
		}
		return bb, nil
	} else {
		return nil, model.ErrBlockNotFound
	}
}

func (l *Ledger) addRecord(rec *model.Record) error {
	bb, err := jsonw.Marshal(rec)
	if err != nil {
		return err
	}

	if err = l.cacheDBClient.Update(RecordsKey, rec.ID, bb); err != nil {
		return err
	}

	return nil
}

func (l *Ledger) loadRecord(rid string) (*model.Record, error) {
	defer measure.ExecTime("onflow.loadRecord")()

	b, err := l.cacheDBClient.FetchBytes(RecordsKey, rid)
	if err != nil {
		return nil, err
	}
	if b != nil {
		var rec model.Record
		err = jsonw.Unmarshal(b, &rec)
		if err != nil {
			return nil, err
		}
		return &rec, nil
	} else {
		return nil, nil
	}
}

func (l *Ledger) runTx(ctx context.Context, txBuilder *splash.FlowTransactionBuilder) error {
	var heightBeforeTx uint64

	if l.conn.IsInMemoryEmulator() {
		block, err := l.conn.Services.GetBlock(ctx, flowkit.LatestBlockQuery)
		if err != nil {
			return err
		}
		heightBeforeTx = block.Height
		log.Warn().Uint64("heightBeforeTx", heightBeforeTx).Msg("Start TX")
	}

	res, err := txBuilder.RunE(ctx)
	if err != nil {
		return err
	}

	if res.Status == flow.TransactionStatusPending && l.conn.IsInMemoryEmulator() {
		// this only happens in unit tests
		log.Debug().Uint64("height", heightBeforeTx).Msg("Waiting for transaction to finish")
		for i := 0; i < emulatorTimeoutSeconds; i++ {
			block, err := l.conn.Services.GetBlock(ctx, flowkit.LatestBlockQuery)
			if err != nil {
				return err
			}

			if block.Height > heightBeforeTx {
				log.Debug().Uint64("height_before_tx", heightBeforeTx).Uint64("current_height", block.Height).Msg("New block found")
				return nil
			} else {
				log.Debug().Msg("No new blocks. Sleeping for 1 second...")
				time.Sleep(time.Second)
			}
		}

		return errors.New("in-memory transaction timed out")
	}

	return nil
}

func (l *Ledger) SubmitRecord(ctx context.Context, r *model.Record) error {
	_, err := l.recordWorkerPool.SubmitErr(func() (model.RecordStatus, error) {
		l.dequeMutex.Lock()
		keyIndex := l.keyDeque.PopFront()
		l.dequeMutex.Unlock()
		defer l.keyDeque.PushBack(keyIndex)

		var acct *accounts.Account
		if keyIndex != l.nodeAccount.Key.Index() {
			acct = &accounts.Account{
				Name:    "pool",
				Address: l.nodeAccount.Address,
				Key:     accounts.NewHexKeyFromPrivateKey(keyIndex, crypto.SHA3_256, l.privateKey),
			}
		} else {
			acct = l.nodeAccount
		}

		cadenceRecord := RecordToCadence(r, l.scriptsEngine.ContractAddress("MetaLocker"))

		for {
			txBuilder := l.scriptsEngine.NewTransaction("metalocker_submit_record").Argument(cadenceRecord)
			txBuilder.Proposer = acct
			txBuilder.Payer = acct
			txBuilder.MainSigner = l.nodeAccount

			err := l.runTx(ctx, &txBuilder)
			if err != nil {
				if strings.HasPrefix(err.Error(), "transaction is expired") {
					log.Warn().Err(err).Str("rid", r.ID).Msg("EXPIRED")
					continue
				}
				return "", err
			} else {
				break
			}
		}

		return model.StatusPublished, nil
	}).Wait()

	return err
}

func (l *Ledger) ImportBlock(ctx context.Context, blockNumber uint64, records []*model.Record) error {
	// check the imported block has a correct number

	prevBlock, err := l.GetTopBlock(ctx)
	if err != nil && !errors.Is(err, model.ErrBlockNotFound) {
		return err
	}
	if prevBlock != nil && blockNumber != prevBlock.Number+1 {
		return fmt.Errorf("block number %d is out of range. Top block number = %d", blockNumber, prevBlock.Number)
	}

	l.dequeMutex.Lock()
	keyIndex := l.keyDeque.PopFront()
	l.dequeMutex.Unlock()
	defer l.keyDeque.PushBack(keyIndex)

	var acct *accounts.Account
	if keyIndex != l.nodeAccount.Key.Index() {
		acct = &accounts.Account{
			Name:    "pool",
			Address: l.nodeAccount.Address,
			Key:     accounts.NewHexKeyFromPrivateKey(keyIndex, crypto.SHA3_256, l.privateKey),
		}
	} else {
		acct = l.nodeAccount
	}

	mlAddress := l.scriptsEngine.ContractAddress("MetaLocker")
	cadenceRecordList := make([]cadence.Value, len(records))
	for i, r := range records {
		cadenceRecordList[i] = RecordToCadence(r, mlAddress)
	}
	cadenceRecords := cadence.NewArray(cadenceRecordList)

	for {
		txBuilder := l.scriptsEngine.NewTransaction("metalocker_import_block").
			UInt64Argument(blockNumber).Argument(cadenceRecords)
		txBuilder.Proposer = acct
		txBuilder.Payer = acct
		txBuilder.MainSigner = l.nodeAccount

		err = l.runTx(ctx, &txBuilder)
		if err != nil {
			if strings.HasPrefix(err.Error(), "transaction is expired") {
				log.Warn().Err(err).Uint64("importedBlockNumber", blockNumber).Msg("EXPIRED")
				continue
			}
			return err
		} else {
			break
		}
	}

	return nil
}

func (l *Ledger) GetRecord(ctx context.Context, rid string) (*model.Record, error) {
	defer measure.ExecTime("onflow.GetRecord")()

	rec, err := l.loadRecord(rid)
	if err != nil {
		return nil, err
	}

	if rec != nil {
		return rec, nil
	} else {
		val, err := l.scriptsEngine.NewScript("metalocker_get_record").
			Argument(cadence.String(rid)).
			RunReturns(ctx)
		if err != nil {
			return nil, err
		}

		rec, err := RecordFromCadence(val)
		if err != nil {
			return nil, err
		}

		if rec != nil {
			return rec, nil
		} else {
			return nil, model.ErrRecordNotFound
		}
	}
}

func (l *Ledger) GetRecordState(ctx context.Context, rid string) (*model.RecordState, error) {
	val, err := l.scriptsEngine.NewScript("metalocker_get_record").
		Argument(cadence.String(rid)).
		RunReturns(ctx)
	if err != nil {
		return nil, err
	}

	if opt, ok := val.(cadence.Optional); ok {
		if opt.Value == nil {
			return nil, nil
		}
		val = opt.Value
	}

	valStruct, ok := val.(cadence.Struct)
	if !ok || valStruct.StructType.QualifiedIdentifier != "MetaLocker.Record" {
		return nil, errors.New("bad Record value")
	}

	return &model.RecordState{
		Status:      model.RecordStatus(valStruct.SearchFieldByName("status").(cadence.String)),
		BlockNumber: uint64(valStruct.SearchFieldByName("blockNumber").(cadence.UInt64)),
	}, nil
}

func (l *Ledger) GetBlock(ctx context.Context, bn uint64) (*model.Block, error) {
	b, err := l.loadBlock(bn)
	if err != nil {
		return nil, err
	}

	if b == nil || b.Status == model.BlockStatusOpen {
		val, err := l.scriptsEngine.NewScript("metalocker_get_block").
			Argument(cadence.UInt64(bn)).
			RunReturns(ctx)
		if err != nil {
			return nil, err
		}

		b, err = BlockFromCadence(val)
		if err != nil {
			return nil, err
		}
		if b == nil {
			return nil, model.ErrBlockNotFound
		}

		if err = l.addBlock(b, true); err != nil {
			return nil, err
		}
	}

	return b, nil
}

func (l *Ledger) GetBlockRecords(ctx context.Context, bn uint64) ([][]string, error) {
	recList, err := l.loadBlockRecords(bn)
	if err != nil {
		if !errors.Is(err, model.ErrBlockNotFound) {
			return nil, err
		}
		// load from Flow
		val, err := l.scriptsEngine.NewScript("metalocker_get_block").
			Argument(cadence.UInt64(bn)).
			RunReturns(ctx)
		if err != nil {
			return nil, err
		}

		recList, err = BlockRecordsFromCadence(val)
		if err != nil {
			return nil, err
		}
	}

	return recList, nil
}

func (l *Ledger) GetGenesisBlock(ctx context.Context) (*model.Block, error) {
	return l.GetBlock(ctx, 0)
}

func (l *Ledger) GetTopBlock(ctx context.Context) (*model.Block, error) {
	//return l.loadBlock(l.latestBlockNumber)

	val, err := l.scriptsEngine.NewScript("metalocker_get_top_block_number").
		RunReturns(ctx)
	if err != nil {
		return nil, err
	}

	topBlockNumber := uint64(val.(cadence.UInt64))

	return &model.Block{
		Number:     topBlockNumber,
		Hash:       "", // FIXME
		ParentHash: "",
		Status:     0,
	}, nil
}

func (l *Ledger) GetChain(ctx context.Context, startNumber uint64, depth int) ([]*model.Block, error) {
	defer measure.ExecTime("onflow.GetChain")()

	b, err := l.GetBlock(ctx, startNumber)
	if err != nil {
		return nil, err
	}

	top, err := l.GetTopBlock(ctx)
	if err != nil {
		return nil, err
	}

	result := []*model.Block{b}

	last := b.Number + uint64(depth-1)
	if last > top.Number {
		last = top.Number
	}

	for seqNo := b.Number + 1; seqNo <= last; seqNo++ {
		b, err = l.GetBlock(ctx, seqNo)
		if err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, nil
}

func (l *Ledger) GetDataAssetState(ctx context.Context, id string) (model.DataAssetState, error) {
	optVal, err := l.scriptsEngine.NewScript("metalocker_get_data_asset_counter").
		Argument(cadence.String(id)).
		RunReturns(ctx)
	if err != nil {
		return model.DataAssetStateNotFound, err
	}

	val, _ := optVal.(cadence.Optional)
	if val.Value == nil {
		return model.DataAssetStateNotFound, nil
	} else {
		counter := uint64(val.Value.(cadence.UInt64))
		if counter > 0 {
			return model.DataAssetStateKeep, nil
		} else {
			return model.DataAssetStateRemove, nil
		}
	}
}

func (l *Ledger) GetAssetHead(ctx context.Context, headID string) (*model.Record, error) {
	val, err := l.scriptsEngine.NewScript("metalocker_get_asset_head").
		Argument(cadence.String(headID)).
		RunReturns(ctx)
	if err != nil {
		return nil, err
	}

	rec, err := RecordFromCadence(val)
	if err != nil {
		return nil, err
	}

	if rec == nil {
		return nil, model.ErrAssetHeadNotFound
	} else {
		return rec, nil
	}
}

func NewLedger(ctx context.Context, connector *splash.Connector, network string, nodeAcct *accounts.Account, keyIndexList []uint32, dbFilePath string, ns notification.Service) (*Ledger, error) {
	log.Info().Msg("Initialising Flow ledger")

	scriptsEngine, err := NewTemplateEngine(connector)
	if err != nil {
		return nil, err
	}

	wellKnownAddresses := scriptsEngine.WellKnownAddresses()
	digitalArtAddr, ok := wellKnownAddresses["MetaLocker"]
	if !ok {
		return nil, errors.New("contract not found: MetaLocker")
	}

	eventTypes := []string{
		fmt.Sprintf("A.%s.MetaLocker.BlockOpened", digitalArtAddr[2:]),
		fmt.Sprintf("A.%s.MetaLocker.BlockConfirmed", digitalArtAddr[2:]),
		fmt.Sprintf("A.%s.MetaLocker.RecordPublished", digitalArtAddr[2:]),
		fmt.Sprintf("A.%s.MetaLocker.RecordRevoked", digitalArtAddr[2:]),
	}

	cacheDBClient, err := utils.NewBoltClient(dbFilePath, InstallLedgerSchema)
	if err != nil {
		return nil, err
	}

	genesisBlockHeight, err := cacheDBClient.FetchUint64(ControlsKey, GenesisBlockHeightKey)
	if err != nil {
		return nil, err
	}

	latestBlockNumber, err := cacheDBClient.FetchUint64(ControlsKey, TopBlockNumberKey)
	if err != nil {
		return nil, err
	}

	val, err := scriptsEngine.NewScript("metalocker_get_genesis_block_height").
		RunReturns(ctx)
	if err != nil {
		return nil, err
	}

	actualGenesisBlockHeight := uint64(val.(cadence.UInt64))

	var latestBlockHeight uint64
	if genesisBlockHeight == 0 {
		log.Warn().Uint64("height", actualGenesisBlockHeight).Msg("Initialising Flow ledger with genesis block")
		genesisBlockHeight = actualGenesisBlockHeight

		if err = cacheDBClient.UpdateUint64(ControlsKey, GenesisBlockHeightKey, genesisBlockHeight); err != nil {
			return nil, err
		}

		if err = cacheDBClient.UpdateUint64(ControlsKey, TopBlockNumberKey, genesisBlockHeight); err != nil {
			return nil, err
		}
		latestBlockHeight = genesisBlockHeight - 1
	} else {
		if actualGenesisBlockHeight != genesisBlockHeight {
			return nil, errors.New("genesis block mismatch")
		}
		latestBlockHeight, err = cacheDBClient.FetchUint64(ControlsKey, TopFlowBlockHeightKey)
		if err != nil {
			return nil, err
		}
	}

	recordWorkerPool := pond.NewResultPool[model.RecordStatus](len(keyIndexList))

	privateKey, err := nodeAcct.Key.PrivateKey()
	if err != nil {
		return nil, err
	}

	var keyDeque deque.Deque[uint32]
	for _, keyIndex := range keyIndexList {
		keyDeque.PushBack(keyIndex)
	}

	return &Ledger{
		network:           network,
		nodeAccount:       nodeAcct,
		conn:              connector,
		scriptsEngine:     scriptsEngine,
		ns:                ns,
		cacheDBClient:     cacheDBClient,
		privateKey:        *privateKey,
		dequeMutex:        &sync.Mutex{},
		keyDeque:          keyDeque,
		recordWorkerPool:  recordWorkerPool,
		latestBlockHeight: latestBlockHeight,
		latestBlockNumber: latestBlockNumber,
		eventTypes:        eventTypes,
		syncMutex:         &sync.Mutex{},
	}, nil
}

func loadFlowKitAccount(addrStr string, keyIndex uint32, keyStr string) (*accounts.Account, error) {
	acct := &accounts.Account{}

	key, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, keyStr)
	if err != nil {
		return acct, err
	}

	acct.Name = "node"
	acct.Address = flow.HexToAddress(addrStr)
	acct.Key = accounts.NewHexKeyFromPrivateKey(keyIndex, crypto.SHA3_256, key)

	return acct, nil
}

func CreateLedgerConnector(ctx context.Context, params ledger.Parameters, ns notification.Service, resolver cmdbase.ParameterResolver) (model.Ledger, error) {
	network, ok := params["network"].(string)
	if !ok {
		return nil, errors.New("parameter not found: network. Can't start Flow ledger connector")
	}
	dbFilePath, ok := params["dbFile"].(string)
	if !ok {
		return nil, errors.New("parameter not found: dbFile. Can't start Flow ledger connector")
	}

	if network == "local" {
		network = "emulator"
	}

	flowConnector, err := NewNetworkConnectorEmbedded(network)
	if err != nil {
		log.Err(err).Msg("Failed to create Flow connector")
		return nil, err
	}

	nodeAddress, err := resolver.ResolveString(params["nodeAddress"])
	if err != nil {
		return nil, err
	}

	if nodeAddress == "" {
		return nil, errors.New("node address is required (use 'nodeAddress' connector property)")
	}

	nodeKey, err := resolver.ResolveString(params["nodeKey"])
	if err != nil {
		return nil, err
	}

	if nodeKey == "" {
		return nil, errors.New("node key is required (use 'nodeKey' connector property)")
	}

	keyIndexStr, err := resolver.ResolveString(params["nodeKeyIndex"])
	if err != nil {
		return nil, err
	}

	if keyIndexStr == "" {
		keyIndexStr = "0"
	}

	keyIndex := uint32(utils.StringToUint64(keyIndexStr))

	nodeAcct, err := loadFlowKitAccount(nodeAddress, keyIndex, nodeKey)
	if err != nil {
		log.Err(err).Msg("Invalid Flow Node account")
		return nil, err
	}

	bl, err := NewLedger(ctx, flowConnector, network, nodeAcct, []uint32{0}, utils.AbsPathify(dbFilePath), ns)
	if err != nil {
		return nil, err
	}

	err = bl.StartLedgerEvents(ctx, true, 0)
	if err != nil {
		return nil, err
	}

	return bl, nil
}
