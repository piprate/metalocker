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
	"errors"
	"fmt"
	"time"

	"github.com/piprate/metalocker/index"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/utils"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/utils/measure"
	"github.com/rs/zerolog/log"
	"go.etcd.io/bbolt"
)

type Index struct {
	id     string
	userID string
	props  *index.Properties
	client *utils.BoltClient
}

var _ index.RootIndex = (*Index)(nil)
var _ index.Writer = (*Index)(nil)

func (dwi *Index) ID() string {
	return dwi.id
}

func (dwi *Index) Properties() *index.Properties {
	return dwi.props
}

func (dwi *Index) Close() error {
	log.Debug().Msg("Closing BoltDB data wallet index")
	return nil
}

func (dwi *Index) IsLocked() bool {
	return false
}

func (dwi *Index) Unlock(key []byte) error {
	return nil
}

func (dwi *Index) Lock() {
	// do nothing
}

func (dwi *Index) indexBucket(tx *bbolt.Tx, bucketID string) *bbolt.Bucket {
	b := tx.Bucket([]byte(dwi.id))
	if b != nil {
		b = b.Bucket([]byte(bucketID))
	} else {
		log.Warn().Str("id", dwi.id).Msg("Index bucket not found in Bolt data wallet")
	}
	return b
}

func (dwi *Index) AddLease(ds model.DataSet, effectiveBlockNumber int64) error {
	defer measure.ExecTime("index.AddLease")()

	r := ds.Record()
	lockerID := ds.LockerID()
	participantID := ds.ParticipantID()
	blockNumber := ds.BlockNumber()

	if r.Status == model.StatusRevoked {
		rs := &index.RecordState{
			ID:            r.ID,
			Operation:     r.Operation,
			Status:        model.StatusRevoked,
			LockerID:      lockerID,
			ParticipantID: participantID,
			BlockNumber:   blockNumber,
			Index:         r.KeyIndex,
		}

		if err := dwi.client.DB.Update(func(tx *bbolt.Tx) error {

			// update record lookup

			rlb := dwi.indexBucket(tx, RecordLookupKey)
			if rlb == nil {
				return fmt.Errorf("bucket %s not found", RecordLookupKey)
			}
			err := rlb.Put([]byte(rs.ID), rs.Bytes())
			if err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}
	} else {
		lease := ds.Lease()

		impID := lease.Impression.ID
		assetID := lease.Impression.Asset
		revNum := lease.Impression.Revision()
		parentID := lease.Impression.WasRevisionOf
		variantID := lease.Impression.SpecializationOf
		contentType := lease.Impression.MetaResource.ContentType

		rs := &index.RecordState{
			ID:            r.ID,
			Operation:     r.Operation,
			Status:        model.StatusPublished,
			LockerID:      lockerID,
			ParticipantID: participantID,
			BlockNumber:   blockNumber,
			Index:         r.KeyIndex,
			ImpressionID:  impID,
			ContentType:   contentType,
		}

		as := &index.AssetState{
			ImpressionID:     impID,
			AssetID:          assetID,
			ContentType:      contentType,
			RevisionNumber:   revNum,
			WasRevisionOf:    parentID,
			SpecializationOf: variantID,
		}

		vrs := &index.VariantRecordState{
			ID:             r.ID,
			Operation:      r.Operation,
			Status:         model.StatusPublished,
			LockerID:       lockerID,
			ParticipantID:  participantID,
			BlockNumber:    blockNumber,
			Index:          r.KeyIndex,
			AssetID:        lease.Impression.Asset,
			ImpressionID:   lease.Impression.ID,
			RevisionNumber: lease.Impression.Revision(),
			ContentType:    lease.Impression.MetaResource.ContentType,
			CreatedAt:      lease.Impression.GeneratedAtTime,
		}

		if err := dwi.client.DB.Update(func(tx *bbolt.Tx) error {

			// add record state

			rb := dwi.indexBucket(tx, RecordsKey)
			if rb == nil {
				return fmt.Errorf("bucket %s not found", RecordsKey)
			}
			rlb, err := rb.CreateBucketIfNotExists([]byte(lockerID))
			if err != nil {
				return err
			}
			rlpb, err := rlb.CreateBucketIfNotExists([]byte(participantID))
			if err != nil {
				return err
			}
			err = rlpb.Put([]byte(r.ID), rs.Bytes())
			if err != nil {
				return err
			}

			// update record lookup

			rlb = dwi.indexBucket(tx, RecordLookupKey)
			if rlb == nil {
				return fmt.Errorf("bucket %s not found", RecordLookupKey)
			}
			err = rlb.Put([]byte(rs.ID), rs.Bytes())
			if err != nil {
				return err
			}

			// update impression lookup

			ilb := dwi.indexBucket(tx, ImpressionLookupKey)
			if ilb == nil {
				return fmt.Errorf("bucket %s not found", ImpressionLookupKey)
			}

			impb, err := ilb.CreateBucketIfNotExists([]byte(impID))
			if err != nil {
				return err
			}

			err = impb.Put([]byte(r.ID), []byte(lockerID)) // store this pair for now
			if err != nil {
				return err
			}

			// update resource lookup

			rslb := dwi.indexBucket(tx, ResourceLookupKey)
			if rslb == nil {
				return fmt.Errorf("bucket %s not found", ResourceLookupKey)
			}

			for _, res := range lease.Resources {
				resb, err := rslb.CreateBucketIfNotExists([]byte(res.Asset))
				if err != nil {
					return err
				}

				err = resb.Put([]byte(r.ID), []byte(lockerID))
				if err != nil {
					return err
				}
			}

			// update asset record lookup

			alb := dwi.indexBucket(tx, AssetLookupKey)
			if alb == nil {
				return fmt.Errorf("bucket %s not found", AssetLookupKey)
			}

			ab, err := alb.CreateBucketIfNotExists([]byte(as.AssetID))
			if err != nil {
				return err
			}

			err = ab.Put([]byte(r.ID), as.Bytes())
			if err != nil {
				return err
			}

			// update variant lookup

			vb := dwi.indexBucket(tx, VariantsKey)
			if vb == nil {
				return fmt.Errorf("bucket %s not found", VariantsKey)
			}
			vvb, err := vb.CreateBucketIfNotExists([]byte(lease.Impression.GetVariantID()))
			if err != nil {
				return err
			}

			err = vvb.Put([]byte(vrs.ID), vrs.Bytes())
			if err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func (dwi *Index) AddLeaseRevocation(ds model.DataSet) error {
	defer measure.ExecTime("index.AddLeaseRevocation")()

	r := ds.Record()
	lockerID := ds.LockerID()
	participantID := ds.ParticipantID()
	blockNumber := ds.BlockNumber()

	rs := &index.RecordState{
		ID:            r.ID,
		Operation:     r.Operation,
		Status:        model.StatusPublished,
		LockerID:      lockerID,
		ParticipantID: participantID,
		BlockNumber:   blockNumber,
		Index:         r.KeyIndex,
	}

	if err := dwi.client.DB.Update(func(tx *bbolt.Tx) error {

		// add the revocation record to record lookup

		rlb := dwi.indexBucket(tx, RecordLookupKey)
		if rlb == nil {
			return fmt.Errorf("bucket %s not found", RecordLookupKey)
		}
		err := rlb.Put([]byte(rs.ID), rs.Bytes())
		if err != nil {
			return err
		}

		// update subject record state (locker participant)

		rb := dwi.indexBucket(tx, RecordsKey)
		if rb == nil {
			return fmt.Errorf("bucket %s not found", RecordsKey)
		}
		rlb = rb.Bucket([]byte(lockerID))
		if rlb == nil {
			return fmt.Errorf("lease revocation failed for %s: source locker not found: %s", r.SubjectRecord, lockerID)
		}
		rlpb := rlb.Bucket([]byte(participantID))
		if rlpb == nil {
			return fmt.Errorf("lease revocation failed for %s: source participant not found: %s", r.SubjectRecord, participantID)
		}

		subj := rlpb.Get([]byte(r.SubjectRecord))

		if subj != nil {

			// subject record was previously saved in the wallet

			var subjRS index.RecordState
			err = jsonw.Unmarshal(subj, &subjRS)
			if err != nil {
				return err
			}

			subjRS.Status = model.StatusRevoked

			err = rlpb.Put([]byte(r.SubjectRecord), subjRS.Bytes())
			if err != nil {
				return err
			}

			// update subject record lookup

			rlb = dwi.indexBucket(tx, RecordLookupKey)
			if rlb == nil {
				return fmt.Errorf("bucket %s not found", RecordLookupKey)
			}
			err = rlb.Put([]byte(r.SubjectRecord), subjRS.Bytes())
			if err != nil {
				return err
			}

			// drop record from impression lookup

			ilb := dwi.indexBucket(tx, ImpressionLookupKey)
			if ilb == nil {
				return fmt.Errorf("bucket %s not found", ImpressionLookupKey)
			}

			impb := ilb.Bucket([]byte(subjRS.ImpressionID))

			err = impb.Delete([]byte(r.SubjectRecord))
			if err != nil {
				return err
			}

			// drop record from resource lookup

			rslb := dwi.indexBucket(tx, ResourceLookupKey)
			if rslb == nil {
				return fmt.Errorf("bucket %s not found", ResourceLookupKey)
			}

			err = rslb.ForEach(func(resourceID, empty []byte) error {
				resourceBucket := rslb.Bucket(resourceID)
				err = resourceBucket.Delete([]byte(r.SubjectRecord))
				return err
			})
			if err != nil {
				return err
			}

			// drop from asset record lookup

			alb := dwi.indexBucket(tx, AssetLookupKey)
			if alb == nil {
				return fmt.Errorf("bucket %s not found", AssetLookupKey)
			}

			err = alb.ForEach(func(assetID, empty []byte) error {
				assetBucket := alb.Bucket(assetID)
				err = assetBucket.Delete([]byte(r.SubjectRecord))
				return err
			})
			if err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (dwi *Index) UpdateTopBlock(blockNumber int64) error {
	if blockNumber <= 0 {
		return errors.New("no block ID provided when updating locker stats " +
			"(maybe there were no new blocks processed?)")
	}

	return dwi.client.DB.Update(func(tx *bbolt.Tx) error {
		b := dwi.indexBucket(tx, LockersKey)
		if b == nil {
			return fmt.Errorf("bucket %s not found", LockersKey)
		}

		return b.ForEach(func(k, v []byte) error {
			var ls index.LockerState
			err := jsonw.Unmarshal(v, &ls)
			if err != nil {
				return err
			}

			ls.TopBlock = blockNumber

			if err = b.Put(k, ls.Bytes()); err != nil {
				return err
			}

			return nil
		})
	})
}

func (dwi *Index) GetRecord(recordID string) (*index.RecordState, error) {
	var val []byte
	var rs index.RecordState
	found := false
	err := dwi.client.DB.View(func(tx *bbolt.Tx) error {
		b := dwi.indexBucket(tx, RecordLookupKey)
		if b == nil {
			return fmt.Errorf("bucket %s not found", RecordLookupKey)
		}

		val = b.Get([]byte(recordID))

		if val != nil {
			found = true
			return jsonw.Unmarshal(val, &rs)
		} else {
			return nil
		}
	})
	if err != nil {
		return nil, err
	}

	if found {
		return &rs, nil
	} else {
		return nil, nil
	}
}

func (dwi *Index) TraverseRecords(lockerFilter, participantFilter string, vFunc index.RecordVisitor, maxRecords uint64) error {
	err := dwi.client.DB.View(func(tx *bbolt.Tx) error {
		lb := dwi.indexBucket(tx, RecordsKey)
		if lb == nil {
			return fmt.Errorf("bucket %s not found", RecordsKey)
		}

		err := lb.ForEach(func(lockerID, empty []byte) error {
			if lockerFilter != "" && string(lockerID) != lockerFilter {
				return nil
			}

			lockerBucket := lb.Bucket(lockerID)
			if err := lockerBucket.ForEach(func(pid, v []byte) error {
				if participantFilter != "" && string(pid) != participantFilter {
					return nil
				}

				partyBucket := lockerBucket.Bucket(pid)
				if err := partyBucket.ForEach(func(recID, v []byte) error {
					var rs index.RecordState
					err := jsonw.Unmarshal(v, &rs)
					if err != nil {
						return err
					}

					return vFunc(&rs)
				}); err != nil {
					return err
				} else {
					return nil
				}
			}); err != nil {
				return err
			} else {
				return nil
			}
		})

		return err
	})

	return err
}

func (dwi *Index) TraverseAssetRecords(assetID string, vFunc index.AssetRecordVisitor, maxRecords uint64) error {
	err := dwi.client.DB.View(func(tx *bbolt.Tx) error {
		rb := dwi.indexBucket(tx, AssetLookupKey)
		if rb == nil {
			return fmt.Errorf("bucket %s not found", AssetLookupKey)
		}

		b := rb.Bucket([]byte(assetID))
		if b == nil {
			// no records were inserted yet
			return nil
		}

		err := b.ForEach(func(k, v []byte) error {
			var as index.AssetState
			err := jsonw.Unmarshal(v, &as)
			if err != nil {
				return err
			}

			err = vFunc(string(k), &as)
			if err != nil {
				return err
			}

			return nil
		})

		return err
	})

	return err
}

func (dwi *Index) TraverseVariants(lockerFilter, participantFilter string, vFunc index.VariantVisitor, includeHistory bool, maxVariants uint64) error {
	err := dwi.client.DB.View(func(tx *bbolt.Tx) error {
		vb := dwi.indexBucket(tx, VariantsKey)
		if vb == nil {
			return fmt.Errorf("bucket %s not found", VariantsKey)
		}

		err := vb.ForEach(func(variantID, empty []byte) error {
			var hist []*index.VariantRecordState
			var masterRec *index.VariantRecordState
			var maxRevision int64 = -1
			var maxCreatedAt *time.Time = nil

			if err := vb.Bucket(variantID).ForEach(func(recID, v []byte) error {
				var rs index.VariantRecordState
				err := jsonw.Unmarshal(v, &rs)
				if err != nil {
					return err
				}

				if lockerFilter != "" && rs.LockerID != lockerFilter {
					return nil
				}
				if participantFilter != "" && rs.ParticipantID != participantFilter {
					return nil
				}

				if rs.RevisionNumber > maxRevision || (maxCreatedAt != nil && rs.RevisionNumber == maxRevision && rs.CreatedAt.After(*maxCreatedAt)) {
					masterRec = &rs
					maxRevision = rs.RevisionNumber
				}

				if includeHistory {
					hist = append(hist, &rs)
				}

				return nil
			}); err != nil {
				return err
			}

			if masterRec != nil {
				return vFunc(string(variantID), masterRec, hist)
			} else {
				// this condition may happen if locker or participant filters are present
				return nil
			}
		})

		return err
	})

	return err
}

func (dwi *Index) GetVariant(variantID string, includeHistory bool) (*index.VariantRecordState,
	[]*index.VariantRecordState, error) {

	// WARNING: if there are multiple instances of the same impression in the index, this function
	// return a random master and ALL copies of all revisions.

	var hist []*index.VariantRecordState
	var masterRec *index.VariantRecordState

	err := dwi.client.DB.View(func(tx *bbolt.Tx) error {
		vb := dwi.indexBucket(tx, VariantsKey)
		if vb == nil {
			return fmt.Errorf("bucket %s not found", VariantsKey)
		}

		vvb := vb.Bucket([]byte(variantID))
		if vvb == nil {
			// no variants were inserted yet
			return fmt.Errorf("variant %s not found in index", variantID)
		}

		var maxRevision int64 = -1
		var maxCreatedAt *time.Time = nil

		if err := vvb.ForEach(func(recID, v []byte) error {
			var rs index.VariantRecordState
			err := jsonw.Unmarshal(v, &rs)
			if err != nil {
				return err
			}

			if rs.RevisionNumber > maxRevision || (maxCreatedAt != nil && rs.RevisionNumber == maxRevision && rs.CreatedAt.After(*maxCreatedAt)) {
				masterRec = &rs
				maxRevision = rs.RevisionNumber
			}

			if includeHistory {
				hist = append(hist, &rs)
			}

			return nil
		}); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return masterRec, hist, nil
}

func (dwi *Index) GetRecordsByResourceID(resourceID string, lockerFilter map[string]bool) ([]string, error) {
	recordIDs := make([]string, 0)
	err := dwi.client.DB.View(func(tx *bbolt.Tx) error {
		rb := dwi.indexBucket(tx, ResourceLookupKey)
		if rb == nil {
			return fmt.Errorf("bucket %s not found", ResourceLookupKey)
		}

		b := rb.Bucket([]byte(resourceID))
		if b == nil {
			// no records were inserted yet
			return nil
		}

		err := b.ForEach(func(k, v []byte) error {
			if lockerFilter != nil {
				if _, found := lockerFilter[string(v)]; !found {
					return nil
				}
			}
			recordIDs = append(recordIDs, string(k))
			return nil
		})

		return err
	})

	if err != nil {
		return nil, err
	}

	return recordIDs, nil
}

func (dwi *Index) GetRecordsByImpressionID(impID string, lockerFilter map[string]bool) ([]string, error) {
	recordIDs := make([]string, 0)
	err := dwi.client.DB.View(func(tx *bbolt.Tx) error {
		rb := dwi.indexBucket(tx, ImpressionLookupKey)
		if rb == nil {
			return fmt.Errorf("bucket %s not found", ImpressionLookupKey)
		}

		b := rb.Bucket([]byte(impID))
		if b == nil {
			// no records were inserted yet
			return nil
		}

		err := b.ForEach(func(k, v []byte) error {
			if lockerFilter != nil {
				if _, found := lockerFilter[string(v)]; !found {
					return nil
				}
			}
			recordIDs = append(recordIDs, string(k))
			return nil
		})

		return err
	})

	if err != nil {
		return nil, err
	}

	return recordIDs, nil
}

func (dwi *Index) IsWritable() bool {
	return true
}

func (dwi *Index) Writer() (index.Writer, error) {
	return dwi, nil
}

func (dwi *Index) LockerStates() ([]index.LockerState, error) {
	defer measure.ExecTime("index.LockerStates")()

	states := make([]index.LockerState, 0)
	err := dwi.client.DB.View(func(tx *bbolt.Tx) error {
		b := dwi.indexBucket(tx, LockersKey)
		if b == nil {
			return fmt.Errorf("bucket %s not found", LockersKey)
		}

		err := b.ForEach(func(k, v []byte) error {
			var ls index.LockerState
			err := jsonw.Unmarshal(v, &ls)
			if err != nil {
				return err
			}
			states = append(states, ls)
			return nil
		})

		return err
	})
	if err != nil {
		return nil, err
	}

	return states, nil
}

func (dwi *Index) AddLockerState(accountID, lockerID string, firstBlock int64) error {
	defer measure.ExecTime("index.AddLockerState")()

	found := false
	err := dwi.client.DB.View(func(tx *bbolt.Tx) error {
		b := dwi.indexBucket(tx, LockersKey)
		if b == nil {
			return fmt.Errorf("bucket %s not found", LockersKey)
		}

		if b.Get([]byte(lockerID)) != nil {
			found = true
		}

		return nil
	})
	if err != nil {
		return err
	}

	if found {
		return index.ErrLockerStateExists
	}

	ls := &index.LockerState{
		ID:         lockerID,
		IndexID:    dwi.id,
		AccountID:  accountID,
		FirstBlock: firstBlock,
		TopBlock:   firstBlock,
	}

	if err := dwi.client.DB.Update(func(tx *bbolt.Tx) error {
		b := dwi.indexBucket(tx, LockersKey)
		if b == nil {
			return fmt.Errorf("bucket %s not found", LockersKey)
		}

		err = b.Put([]byte(lockerID), ls.Bytes())
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return err
}

func saveProperties(userID, indexID string, props *index.Properties, client *utils.BoltClient) error {
	return client.DB.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(indexID))
		if b != nil {
			b = b.Bucket([]byte(PropertiesKey))
		} else {
			return errors.New("properties bucket not found")
		}

		err := b.Put([]byte(AccessLevelKey), utils.Uint32ToBytes(uint32(props.AccessLevel)))
		if err != nil {
			return err
		}

		err = b.Put([]byte(AccountKey), []byte(userID))
		if err != nil {
			return err
		}

		return nil
	})
}

func loadProperties(indexID string, client *utils.BoltClient) (*index.Properties, error) {
	var val []byte
	props := &index.Properties{
		IndexType:   index.TypeRoot,
		Asset:       indexID,
		AccessLevel: 0,
		Algorithm:   Algorithm,
	}
	err := client.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(indexID))
		if b != nil {
			b = b.Bucket([]byte(PropertiesKey))
		} else {
			return index.ErrIndexNotFound
		}

		val = b.Get([]byte(AccessLevelKey))

		if val != nil {
			props.AccessLevel = model.AccessLevel(utils.BytesToUint32(val))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return props, nil
}
