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
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/piprate/metalocker/index"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/dataset"
	"github.com/piprate/metalocker/services/notification"
	"github.com/piprate/metalocker/storage"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/utils/measure"
	"github.com/rs/zerolog/log"
)

type (
	localStoreImpl struct {
		dataWallet      DataWallet
		ledger          model.Ledger
		offChainStorage model.OffChainStorage
		blobManager     model.BlobManager
		ns              notification.Service
	}
)

var _ DataStore = (*localStoreImpl)(nil)
var _ dataset.LeaseBuilderBackend = (*localStoreImpl)(nil)

var (
	ErrRecordNotFoundInRootIndex = errors.New("record not found in root index")
	ErrSenderNotFound            = errors.New("sender not found")
	ErrRecipientNotFound         = errors.New("recipient not found")
	ErrLeaseRevokedAndPurged     = errors.New("lease revoked and purged")
)

func newLocalDataStore(dataWallet DataWallet, services Services) (DataStore, error) {
	ns, err := services.NotificationService()
	if err != nil {
		return nil, err
	}
	return &localStoreImpl{
		ledger:          services.Ledger(),
		offChainStorage: services.OffChainStorage(),
		blobManager:     services.BlobManager(),
		dataWallet:      dataWallet,
		ns:              ns,
	}, nil
}

func (c *localStoreImpl) getRootIndexRecord(ctx context.Context, id string) (*index.RecordState, error) {
	rootIndex, err := c.dataWallet.RootIndex(ctx)
	if err != nil {
		if errors.Is(err, index.ErrIndexNotFound) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	rec, err := rootIndex.GetRecord(id)
	if err != nil {
		return nil, err
	}
	return rec, nil
}

type recordProcessor func(blockNumber int64, lockerID, participantID string, symKey *model.AESKey) error

func (c *localStoreImpl) processRecord(ctx context.Context, lr *model.Record, suggestedLockerID string, fn recordProcessor) error {
	rec, err := c.getRootIndexRecord(ctx, lr.ID)
	if err != nil {
		return err
	}

	var (
		blockNumber   int64
		lockerID      string
		participantID string
		publicKey     *btcec.PublicKey
		symKey        *model.AESKey
	)

	if rec == nil {
		var lockers []*model.Locker
		if suggestedLockerID != "" {
			l, err := c.dataWallet.GetLocker(ctx, suggestedLockerID)
			if err != nil {
				if !errors.Is(err, storage.ErrLockerNotFound) {
					return err
				} else {
					log.Warn().Str("rid", lr.ID).Str("lid", suggestedLockerID).
						Msg("Attempted to load a dataset from inaccessible or non-existent locker")
				}
			} else {
				lockers = []*model.Locker{l.Raw()}
			}
		}
		if len(lockers) == 0 {
			// record is not yet in the index. Retrieve crypto material from the lockers, if available
			lockers, err = c.dataWallet.GetLockers(ctx)
			if err != nil {
				return err
			}
		}
	OuterLoop:
		for _, l := range lockers {
			for _, p := range l.Participants {
				publicKey, symKey, err = p.IsRecordOwner(lr.RoutingKey, lr.KeyIndex)
				if err != nil {
					return err
				}
				if publicKey != nil {
					lockerID = l.ID
					participantID = p.ID
					break OuterLoop
				}
			}
		}
		if publicKey == nil && lr.Flags&model.RecordFlagPublic == 0 {
			// the given record doesn't belong to any of the lockers
			return model.ErrDataSetNotFound
		}

		// identify block number

		state, err := c.ledger.GetRecordState(lr.ID)
		if err != nil {
			return err
		}
		blockNumber = state.BlockNumber
	} else {
		if lr.Flags&model.RecordFlagPublic == 0 {
			// retrieve crypto material from the locker this record belongs to
			l, err := c.dataWallet.GetLocker(ctx, rec.LockerID)
			if err != nil {
				return err
			}

			p := l.Raw().GetParticipant(rec.ParticipantID)
			publicKey, err = p.GetRecordPublicKey(rec.Index)
			if err != nil {
				return err
			}

			lockerID = rec.LockerID
			participantID = rec.ParticipantID

			symKey = p.GetOperationSymKey(rec.Index)
		}

		blockNumber = rec.BlockNumber
	}

	if lr.Flags&model.RecordFlagPublic == 0 {
		valid, err := lr.Verify(publicKey)
		if err != nil {
			return err
		} else if !valid {
			log.Warn().Str("rec", string(lr.Bytes())).Msg("Invalid record found")
			return errors.New("ledger record validation failed")
		}
	}

	return fn(blockNumber, lockerID, participantID, symKey)
}

func (c *localStoreImpl) Load(ctx context.Context, recordID string, opts ...dataset.LoadOption) (model.DataSet, error) {
	var options dataset.LoadOptions
	for _, fn := range opts {
		if err := fn(&options); err != nil {
			return nil, err
		}
	}

	rec, err := c.ledger.GetRecord(recordID)
	if err != nil {
		if errors.Is(err, model.ErrRecordNotFound) {
			return nil, model.ErrDataSetNotFound
		} else {
			return nil, err
		}
	}

	if rec.Operation != model.OpTypeLease {
		return nil, errors.New("invalid operation type")
	}

	var ds model.DataSet
	err = c.processRecord(ctx, rec, options.LockerID, func(blockNumber int64, lockerID, participantID string, symKey *model.AESKey) error {
		opRecBytes, err := c.offChainStorage.GetOperation(rec.OperationAddress)
		if err != nil {
			if rec.Status == model.StatusRevoked {
				return ErrLeaseRevokedAndPurged
			} else {
				return err
			}
		}

		if rec.Flags&model.RecordFlagPublic == 0 {
			opRecBytes, err = model.DecryptAESCGM(opRecBytes, symKey)
			if err != nil {
				return err
			}
		}

		lease, err := model.NewLease(opRecBytes)
		if err != nil {
			return err
		}

		ds = dataset.NewDataSetImpl(rec, lease, blockNumber, lockerID, participantID, c.blobManager)

		return nil
	})

	return ds, err
}

func (c *localStoreImpl) Submit(ctx context.Context, lease *model.Lease, cleartext bool, lockerID string, sender *model.LockerParticipant, headName ...string) dataset.RecordFuture {
	defer measure.ExecTime("store.Submit")()

	log.Debug().Str("lid", lockerID).Msg("Processing a new lease submission")

	rec, err := c.submitLease(lease, cleartext, sender)
	if err != nil {
		log.Error().Str("lid", lockerID).Err(err).Msg("Failed to submit a lease")
		return dataset.RecordFutureWithError(err)
	}

	log.Debug().Str("rid", rec.ID).Msg("Record submitted")

	waitList := []string{rec.ID}
	heads := make(map[string]string, len(headName))

	assetID := lease.Impression.Asset
	for _, name := range headName {
		headID := model.HeadID(assetID, lockerID, sender, name)
		headRecID, err := c.submitHead(assetID, lockerID, sender, name, rec.ID)
		if err != nil {
			return dataset.RecordFutureWithError(err)
		}
		heads[headID] = headRecID
		waitList = append(waitList, headRecID)
	}

	ds := dataset.NewDataSetImpl(rec, lease, 0, lockerID, sender.ID, c.blobManager)

	return dataset.RecordFutureWithResult(c.ledger, c.ns, rec.ID, ds, heads, waitList)
}

func (c *localStoreImpl) submitLease(lease *model.Lease, cleartext bool, p *model.LockerParticipant) (*model.Record, error) {
	keyIndex := model.RandomKeyIndex()

	recordPrivKey, err := p.GetRecordPrivateKey(keyIndex)
	if err != nil {
		return nil, err
	}
	recordPubKey, err := recordPrivKey.ECPubKey()
	if err != nil {
		return nil, err
	}

	opRecBytes, err := jsonw.Marshal(lease)
	if err != nil {
		return nil, err
	}

	if !cleartext {
		// derive symmetrical key

		sharedSecretBytes, err := base64.StdEncoding.DecodeString(p.SharedSecret)
		if err != nil {
			return nil, err
		}
		symKey := model.DeriveSymmetricalKey(sharedSecretBytes, recordPubKey)

		// encrypt operation record

		opRecBytes, err = model.EncryptAESCGM(opRecBytes, symKey)
		if err != nil {
			return nil, err
		}
	}

	leaseAddress, err := c.offChainStorage.SendOperation(opRecBytes)
	if err != nil {
		return nil, err
	}

	// generate authorising commitment

	ac := sha256.Sum256(
		model.BuildAuthorisingCommitmentInput(recordPrivKey, leaseAddress),
	)

	// generate requesting commitment

	rc := sha256.Sum256(
		model.BuildRequestingCommitmentInput(
			lease.ID,
			lease.ExpiresAt,
		),
	)

	// generate new record routing key

	routingKey, _ := model.BuildRoutingKey(recordPubKey)

	rec := &model.Record{
		RoutingKey:                routingKey,
		KeyIndex:                  keyIndex,
		Operation:                 model.OpTypeLease,
		OperationAddress:          leaseAddress,
		AuthorisingCommitment:     base64.StdEncoding.EncodeToString(ac[:]),
		AuthorisingCommitmentType: 0,
		RequestingCommitment:      base64.StdEncoding.EncodeToString(rc[:]),
		RequestingCommitmentType:  model.RcTypeAlgo1,
		DataAssets:                lease.GetResourceIDs(),
	}

	if cleartext {
		rec.Flags |= model.RecordFlagPublic
	}

	// seal the record

	pk, err := recordPrivKey.ECPrivKey()
	if err != nil {
		return nil, err
	}
	err = rec.Seal(pk)
	if err != nil {
		return nil, err
	}

	if err = c.ledger.SubmitRecord(rec); err != nil {
		return nil, err
	}

	return rec, nil
}

func (c *localStoreImpl) submitLeaseRevocation(recordID string, p *model.LockerParticipant) (string, error) {
	keyIndex := model.RandomKeyIndex()

	recordPrivKey, err := p.GetRecordPrivateKey(keyIndex)
	if err != nil {
		return "", err
	}
	recordPubKey, err := recordPrivKey.ECPubKey()
	if err != nil {
		return "", err
	}

	subj, err := c.ledger.GetRecord(recordID)
	if err != nil {
		return "", err
	}

	subjPrivKey, err := p.GetRecordPrivateKey(subj.KeyIndex)
	if err != nil {
		return "", err
	}

	acInput := model.BuildAuthorisingCommitmentInput(subjPrivKey, subj.OperationAddress)
	subjAC := sha256.Sum256(acInput)

	if subj.AuthorisingCommitment != base64.StdEncoding.EncodeToString(subjAC[:]) {
		return "", errors.New(
			"authorising commitment check failed. You are not authorised to revoke this lease")
	}

	// generate new record routing key

	routingKey, _ := model.BuildRoutingKey(recordPubKey)

	rec := &model.Record{
		RoutingKey: routingKey,
		KeyIndex:   keyIndex,
		Operation:  model.OpTypeLeaseRevocation,

		SubjectRecord: recordID,
		RevocationProof: []string{
			base64.StdEncoding.EncodeToString(acInput),
		},
	}

	// seal the record

	pk, err := recordPrivKey.ECPrivKey()
	if err != nil {
		return "", err
	}
	err = rec.Seal(pk)
	if err != nil {
		return "", err
	}

	if err = c.ledger.SubmitRecord(rec); err != nil {
		return "", err
	}

	return rec.ID, nil
}

func (c *localStoreImpl) Share(ctx context.Context, ds model.DataSet, locker Locker, vaultName string, expiryTime time.Time) dataset.RecordFuture {
	sender := locker.Us()
	if sender == nil {
		return dataset.RecordFutureWithError(fmt.Errorf("read-only locker"))
	}
	us := sender.ID
	idy, err := c.dataWallet.GetIdentity(ctx, us)
	if err != nil {
		return dataset.RecordFutureWithError(err)
	}
	creator := idy.DID()

	var recipient []*model.LockerParticipant
	if locker.IsUniLocker() {
		recipient = append(recipient, sender)
	} else {
		recipient = locker.Them()
	}

	builder, err := dataset.NewLeaseBuilderForSharing(ctx, ds, c, c.blobManager, dataset.CopyModeDeep, creator,
		nil, recipient[0].ID, vaultName, nil)
	if err != nil {
		return dataset.RecordFutureWithError(err)
	}

	lease, err := builder.Build(expiryTime)
	if err != nil {
		return dataset.RecordFutureWithError(err)
	}

	return c.Submit(ctx, lease, false, locker.ID(), sender)
}

func (c *localStoreImpl) Revoke(ctx context.Context, id string) dataset.RecordFuture {
	defer measure.ExecTime("store.RevokeLease")()

	rs, err := c.getRootIndexRecord(ctx, id)
	if err != nil {
		return dataset.RecordFutureWithError(err)
	}

	if rs == nil {
		return dataset.RecordFutureWithError(ErrRecordNotFoundInRootIndex)
	}

	if rs.Operation != model.OpTypeLease {
		return dataset.RecordFutureWithError(fmt.Errorf("only leases can be revoked. Found optype: %d", rs.Operation))
	}

	if rs.Status == model.StatusRevoked {
		return dataset.RecordFutureWithError(fmt.Errorf("lease already revoked: %s", id))
	}

	locker, err := c.dataWallet.GetLocker(ctx, rs.LockerID)
	if err != nil {
		return dataset.RecordFutureWithError(err)
	}

	usParty := locker.Us()
	if usParty == nil {
		return dataset.RecordFutureWithError(ErrSenderNotFound)
	}

	if usParty.ID != rs.ParticipantID {
		log.Error().Str("rid", id).Msg("Third party requested to revoke a lease")
		return dataset.RecordFutureWithError(errors.New("only lease owner can revoke leases"))
	}

	log.Debug().Str("rid", id).Msg("Record owner found. Revoking the lease...")

	recID, err := c.submitLeaseRevocation(id, usParty)
	if err != nil {
		log.Error().Str("rec", id).Err(err).Msg("Failed to submit lease revocation")
		return dataset.RecordFutureWithError(err)
	}

	return dataset.RecordFutureWithResult(c.ledger, c.ns, recID, nil, nil, []string{recID})
}

func (c *localStoreImpl) PurgeDataAssets(ctx context.Context, recordID string) error {
	log.Debug().Str("rid", recordID).Msg("Purging data assets from record")

	// purge blobs

	ds, err := c.Load(ctx, recordID)
	if err != nil {
		if errors.Is(err, model.ErrOperationNotFound) {
			log.Warn().Str("rid", recordID).Msg("Operation already purged")
			return nil
		}
		return err
	}

	lease := ds.Lease()

	for _, res := range lease.Resources {
		state, err := c.ledger.GetDataAssetState(res.StorageID())
		if err != nil {
			return err
		}
		if state == model.DataAssetStateRemove {
			if err = c.blobManager.PurgeBlob(res); err != nil {
				if errors.Is(err, model.ErrBlobNotFound) {
					log.Warn().Str("id", res.StorageID()).Msg("Blob already purged")
				} else {
					return err
				}
			}
		}
	}

	// purge operation

	lr, err := c.ledger.GetRecord(recordID)
	if err != nil {
		return err
	}

	if err = c.offChainStorage.PurgeOperation(lr.OperationAddress); err != nil {
		if errors.Is(err, model.ErrOperationNotFound) {
			log.Warn().Str("op_addr", lr.OperationAddress).Msg("Operation already purged")
		} else {
			return err
		}
	}

	return nil
}

func (c *localStoreImpl) NewDataSetBuilder(ctx context.Context, lockerID string, opts ...dataset.BuilderOption) (dataset.Builder, error) {
	locker, err := c.dataWallet.GetLocker(ctx, lockerID)
	if err != nil {
		return nil, err
	}

	usParty := locker.Us()
	if usParty == nil {
		return nil, fmt.Errorf("read-only locker")
	}

	us := usParty.ID
	creator, err := c.dataWallet.GetIdentity(ctx, us)
	if err != nil {
		return nil, err
	}

	lb, err := dataset.NewLeaseBuilder(ctx, c, c.blobManager, locker.Raw(), creator.DID(), opts...)
	if err != nil {
		return nil, err
	}

	return lb, nil
}

func (c *localStoreImpl) AssetHead(ctx context.Context, headID string, opts ...dataset.LoadOption) (model.DataSet, error) {
	defer measure.ExecTime("store.AssetHead")()

	var options dataset.LoadOptions
	for _, fn := range opts {
		if err := fn(&options); err != nil {
			return nil, err
		}
	}

	headRec, err := c.ledger.GetAssetHead(headID)
	if err != nil {
		return nil, err
	}

	var recordID string
	err = c.processRecord(ctx, headRec, options.LockerID, func(blockNumber int64, lockerID, participantID string, symKey *model.AESKey) error {
		headBodyBytes, err := base64.StdEncoding.DecodeString(headRec.HeadBody)
		if err != nil {
			return err
		}

		decryptedHeadBodyBytes, err := model.DecryptAESCGM(headBodyBytes, symKey)
		if err != nil {
			return err
		}

		_, _, _, _, recordID = model.UnpackHeadBody(decryptedHeadBodyBytes)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return c.Load(ctx, recordID)
}

func (c *localStoreImpl) SetAssetHead(ctx context.Context, assetID string, locker *model.Locker, headName, recordID string) dataset.RecordFuture {
	defer measure.ExecTime("store.SetAssetHead")()

	headDataSetRecord, err := c.ledger.GetRecord(recordID)
	if err != nil {
		return dataset.RecordFutureWithError(err)
	}

	if headDataSetRecord.Status == model.StatusRevoked {
		return dataset.RecordFutureWithError(fmt.Errorf("can't set asset head to a revoked record: %s", recordID))
	}

	sender := locker.Us()
	if sender == nil {
		return dataset.RecordFutureWithError(errors.New("can't find an identity to send the data from"))
	}

	recID, err := c.submitHead(assetID, locker.ID, sender, headName, recordID)
	if err != nil {
		return dataset.RecordFutureWithError(err)
	}

	return dataset.RecordFutureWithResult(c.ledger, c.ns, recID, nil, nil, []string{recID})
}

func (c *localStoreImpl) submitHead(assetID, lockerID string, sender *model.LockerParticipant, headName, recordID string) (string, error) {

	headID := model.HeadID(assetID, lockerID, sender, headName)

	prevHead, err := c.ledger.GetAssetHead(headID)
	if err != nil && !errors.Is(err, model.ErrAssetHeadNotFound) {
		return "", err
	}

	var prevHeadRecordID string
	var prevHeadRevocationProof []string
	if prevHead != nil {
		if prevHead.Status == model.StatusRevoked {
			// Timing issues...
			return "", fmt.Errorf("asset head record already revoked: %s", prevHead.ID)
		}

		prevHeadPrivKey, err := sender.GetRecordPrivateKey(prevHead.KeyIndex)
		if err != nil {
			return "", err
		}

		acInput := model.BuildAuthorisingCommitmentInput(prevHeadPrivKey, prevHead.OperationAddress)
		subjAC := sha256.Sum256(acInput)

		if prevHead.AuthorisingCommitment != base64.StdEncoding.EncodeToString(subjAC[:]) {
			return "", errors.New(
				"authorising commitment check failed. You are not authorised to update the asset head")
		}

		prevHeadRecordID = prevHead.ID
		prevHeadRevocationProof = []string{
			base64.StdEncoding.EncodeToString(acInput),
		}
	}

	keyIndex := model.RandomKeyIndex()

	recordPrivKey, err := sender.GetRecordPrivateKey(keyIndex)
	if err != nil {
		return "", err
	}
	recordPubKey, err := recordPrivKey.ECPubKey()
	if err != nil {
		return "", err
	}

	// assetID, lockerID, participantID, name, recordID string
	headBodyBytes := model.PackHeadBody(assetID, lockerID, sender.ID, headName, recordID)

	// derive symmetrical key

	sharedSecretBytes, err := base64.StdEncoding.DecodeString(sender.SharedSecret)
	if err != nil {
		return "", err
	}
	symKey := model.DeriveSymmetricalKey(sharedSecretBytes, recordPubKey)

	// encrypt head body

	encryptedHeadBody, err := model.EncryptAESCGM(headBodyBytes, symKey)
	if err != nil {
		return "", err
	}

	// generate authorising commitment

	ac := sha256.Sum256(
		model.BuildAuthorisingCommitmentInput(recordPrivKey, ""),
	)

	// generate new record routing key

	routingKey, _ := model.BuildRoutingKey(recordPubKey)

	rec := &model.Record{
		RoutingKey: routingKey,
		KeyIndex:   keyIndex,
		Operation:  model.OpTypeAssetHead,

		AuthorisingCommitment:     base64.StdEncoding.EncodeToString(ac[:]),
		AuthorisingCommitmentType: 0,

		SubjectRecord:   prevHeadRecordID,
		RevocationProof: prevHeadRevocationProof,

		HeadID:   headID,
		HeadBody: base64.StdEncoding.EncodeToString(encryptedHeadBody),
	}

	// seal the record

	pk, err := recordPrivKey.ECPrivKey()
	if err != nil {
		return "", err
	}
	err = rec.Seal(pk)
	if err != nil {
		return "", err
	}

	if err = c.ledger.SubmitRecord(rec); err != nil {
		return "", err
	}

	return rec.ID, nil
}
