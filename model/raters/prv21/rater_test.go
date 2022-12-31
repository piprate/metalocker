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

package prv21_test

import (
	"context"
	"testing"
	"time"

	"github.com/piprate/metalocker/model"
	. "github.com/piprate/metalocker/model/raters/prv21"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	MockRevision struct {
		recordID       string
		status         model.RecordStatus
		block          int64
		effectiveBlock int64
		locker         string
		participant    string
		assetID        string
		variantID      string
		impressionID   string
		revisionNumber int64
		createdAt      time.Time
		headFrom       int64
		headTo         int64
	}

	MockRevisionStore struct {
		revisions map[string]*MockRevision
	}
)

var _ Revision = (*MockRevision)(nil)
var _ RevisionStore = (*MockRevisionStore)(nil)

func (r *MockRevision) RecordID() string {
	return r.recordID
}

func (r *MockRevision) Status() model.RecordStatus {
	return r.status
}

func (r *MockRevision) Block() int64 {
	return r.block
}

func (r *MockRevision) EffectiveBlock() int64 {
	return r.effectiveBlock
}

func (r *MockRevision) Locker() string {
	return r.locker
}

func (r *MockRevision) Participant() string {
	return r.participant
}

func (r *MockRevision) AssetID() string {
	return r.assetID
}

func (r *MockRevision) VariantID() string {
	return r.variantID
}

func (r *MockRevision) ImpressionID() string {
	return r.impressionID
}

func (r *MockRevision) RevisionNumber() int64 {
	return r.revisionNumber
}

func (r *MockRevision) CreatedAt() time.Time {
	return r.createdAt
}

func (r *MockRevision) HeadFrom() int64 {
	return r.headFrom
}

func (r *MockRevision) HeadTo() int64 {
	return r.headTo
}

func (r *MockRevision) Update(status model.RecordStatus, headFrom, headTo int64) {
	r.status = status
	r.headFrom = headFrom
	r.headTo = headTo
}

func NewMockRevisionStore() *MockRevisionStore {
	return &MockRevisionStore{
		revisions: map[string]*MockRevision{},
	}
}

func (m *MockRevisionStore) Head(ctx context.Context, variantID string) (Revision, error) {
	for _, rev := range m.revisions {
		if rev.VariantID() == variantID &&
			rev.HeadFrom() != NoBlockNumber &&
			rev.HeadTo() == NoBlockNumber {
			return rev, nil
		}
	}
	return nil, model.ErrRecordNotFound
}

func (m *MockRevisionStore) HeadAt(ctx context.Context, variantID string, blockNumber int64) (Revision, error) {
	for _, rev := range m.revisions {
		if rev.VariantID() == variantID &&
			rev.HeadFrom() <= blockNumber &&
			rev.HeadTo() >= blockNumber {
			return rev, nil
		}
	}
	return nil, model.ErrRecordNotFound
}

func (m *MockRevisionStore) Revision(ctx context.Context, rid string) (Revision, error) {
	rev, found := m.revisions[rid]
	if !found {
		return nil, model.ErrRecordNotFound
	}
	return rev, nil
}

func (m *MockRevisionStore) CreateRevision(ctx context.Context, ds model.DataSet, effectiveBlockNumber int64, headFrom, headTo int64) error {
	imp := ds.Impression()
	varID := imp.GetVariantID()
	revNum := imp.Revision()

	rev := &MockRevision{
		recordID:       ds.ID(),
		status:         ds.Record().Status,
		block:          ds.BlockNumber(),
		effectiveBlock: effectiveBlockNumber,
		locker:         ds.LockerID(),
		participant:    ds.ParticipantID(),
		assetID:        imp.Asset,
		variantID:      varID,
		impressionID:   imp.ID,
		revisionNumber: revNum,
		createdAt:      *imp.GeneratedAtTime,
		headFrom:       headFrom,
		headTo:         headTo,
	}

	m.revisions[rev.recordID] = rev
	return nil
}

func (m *MockRevisionStore) UpdateRevision(ctx context.Context, rid string, status model.RecordStatus, headFrom, headTo int64) error {
	rev, found := m.revisions[rid]
	if !found {
		return model.ErrRecordNotFound
	}

	rev.Update(status, headFrom, headTo)

	return nil
}

func (m *MockRevisionStore) RevokeRevision(ctx context.Context, rid string) error {
	rev, found := m.revisions[rid]
	if !found {
		return model.ErrRecordNotFound
	}

	rev.status = model.StatusRevoked

	return nil
}

func (m *MockRevisionStore) SaveRevokedRevision(ctx context.Context, ds model.DataSet, effectiveBlockNumber int64) error {
	panic("not implemented")
}

func (m *MockRevisionStore) DataSet(ctx context.Context, rid string) (model.DataSet, error) {
	panic("not implemented")
}

func TestRaterImpl_AddRevision(t *testing.T) {
	locker := testbase.TestUniLocker(t)

	blockNumber := locker.FirstBlock + 1

	participantID := locker.Participants[0].ID

	store := NewMockRevisionStore()

	record1 := &model.Record{
		ID:        "record_1",
		Operation: model.OpTypeLease,
		Status:    model.StatusPublished,
	}

	ts := time.Unix(1000, 0).UTC()

	lease1 := &model.Lease{
		Impression: &model.Impression{
			ID:              "impression_1",
			Asset:           "asset_id",
			GeneratedAtTime: &ts,
		},
	}

	ctx := context.Background()

	r := NewRater(lease1.Impression.GetVariantID(), store)

	isNew, err := r.AddRevision(
		ctx,
		testbase.NewMockDataSet(record1, lease1, blockNumber, locker.ID, participantID, nil),
		blockNumber)
	require.NoError(t, err)
	assert.Equal(t, true, isNew)

	newRev, err := store.Revision(ctx, "record_1")
	require.NoError(t, err)
	assert.Equal(t, record1.ID, newRev.RecordID())

	// add the second revision (new head)

	record2 := &model.Record{
		ID:        "record_2",
		Operation: model.OpTypeLease,
		Status:    model.StatusPublished,
	}

	ts = time.Unix(2000, 0).UTC()

	lease2 := &model.Lease{
		Impression: &model.Impression{
			ID:               "impression_2",
			Asset:            "asset_id",
			SpecializationOf: lease1.Impression.ID,
			RevisionNumber:   2,
			GeneratedAtTime:  &ts,
		},
	}

	isNew, err = r.AddRevision(
		ctx,
		testbase.NewMockDataSet(record2, lease2, blockNumber, locker.ID, participantID, nil),
		blockNumber)
	require.NoError(t, err)
	assert.Equal(t, false, isNew)

	assert.Equal(t, "record_2", r.Head(ctx))

	newRev, err = store.Revision(ctx, "record_2")
	require.NoError(t, err)
	assert.Equal(t, record2.ID, newRev.RecordID())

	// add the fourth revision (skip one, to be added next)

	record3 := &model.Record{
		ID:        "record_3",
		Operation: model.OpTypeLease,
		Status:    model.StatusPublished,
	}

	ts = time.Unix(4000, 0).UTC()

	lease3 := &model.Lease{
		Impression: &model.Impression{
			ID:               "impression_3",
			Asset:            "asset_id",
			SpecializationOf: lease1.Impression.ID,
			RevisionNumber:   4,
			GeneratedAtTime:  &ts,
		},
	}

	isNew, err = r.AddRevision(
		ctx,
		testbase.NewMockDataSet(record3, lease3, blockNumber, locker.ID, participantID, nil),
		blockNumber)
	require.NoError(t, err)
	assert.Equal(t, false, isNew)

	assert.Equal(t, "record_3", r.Head(ctx))

	newRev, err = store.Revision(ctx, "record_3")
	require.NoError(t, err)
	assert.Equal(t, record3.ID, newRev.RecordID())

	// add the third revision (should be processed as an orphan)

	record4 := &model.Record{
		ID:        "record_4",
		Operation: model.OpTypeLease,
		Status:    model.StatusPublished,
	}

	ts = time.Unix(3000, 0).UTC()

	lease4 := &model.Lease{
		Impression: &model.Impression{
			ID:               "impression_4",
			Asset:            "asset_id",
			SpecializationOf: lease1.Impression.ID,
			RevisionNumber:   3,
			GeneratedAtTime:  &ts,
		},
	}

	isNew, err = r.AddRevision(
		ctx,
		testbase.NewMockDataSet(record4, lease4, blockNumber, locker.ID, participantID, nil),
		blockNumber)
	require.NoError(t, err)
	assert.Equal(t, false, isNew)

	assert.Equal(t, "record_3", r.Head(ctx))

	newRev, err = store.Revision(ctx, "record_4")
	require.NoError(t, err)
	assert.Equal(t, record4.ID, newRev.RecordID())
}

func TestRaterImpl_AddRevocation(t *testing.T) {
	locker := testbase.TestUniLocker(t)

	blockNumber := locker.FirstBlock + 1

	participantID := locker.Participants[0].ID

	store := NewMockRevisionStore()

	record1 := &model.Record{
		ID:        "record_1",
		Operation: model.OpTypeLease,
		Status:    model.StatusPublished,
	}

	ts := time.Unix(1000, 0).UTC()

	lease1 := &model.Lease{
		Impression: &model.Impression{
			ID:              "impression_1",
			Asset:           "asset_id",
			GeneratedAtTime: &ts,
		},
	}

	ctx := context.Background()

	r := NewRater(lease1.Impression.GetVariantID(), store)

	isNew, err := r.AddRevision(
		ctx,
		testbase.NewMockDataSet(record1, lease1, blockNumber, locker.ID, participantID, nil),
		blockNumber)
	require.NoError(t, err)
	assert.Equal(t, true, isNew)

	newRev, err := store.Revision(ctx, "record_1")
	require.NoError(t, err)
	assert.Equal(t, record1.ID, newRev.RecordID())

	err = r.AddRevocation(ctx, "record_1")
	require.NoError(t, err)

	rev, err := store.Revision(ctx, "record_1")
	require.NoError(t, err)
	assert.Equal(t, model.StatusRevoked, rev.Status())
}
