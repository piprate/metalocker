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

package testbase

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/stretchr/testify/require"
)

type MockDataSet struct {
	record        *model.Record
	lease         *model.Lease
	blockNumber   int64
	lockerID      string
	participantID string
	meta          []byte
}

var _ model.DataSet = (*MockDataSet)(nil)

func NewMockDataSet(r *model.Record, lease *model.Lease, blockNumber int64, lockerID, participantID string, meta any) *MockDataSet {
	ds := &MockDataSet{
		record:        r,
		lease:         lease,
		blockNumber:   blockNumber,
		lockerID:      lockerID,
		participantID: participantID,
	}

	if meta != nil {
		metaBytes, err := jsonw.Marshal(meta)
		if err != nil {
			panic(err)
		}
		ds.meta = metaBytes
	}

	return ds
}

func (d *MockDataSet) ID() string {
	return d.record.ID
}

func (d *MockDataSet) MetaResource(ctx context.Context) (io.ReadCloser, error) {
	if d.meta == nil {
		return nil, errors.New("no meta resource in mock dataset")
	}

	return io.NopCloser(bytes.NewReader(d.meta)), nil
}

func (d *MockDataSet) DecodeMetaResource(ctx context.Context, obj any) error {
	r, err := d.MetaResource(ctx)
	if err != nil {
		return err
	}
	return jsonw.Decode(r, obj)
}

func (d *MockDataSet) Resources() []string {
	return []string{}
}

func (d *MockDataSet) Resource(ctx context.Context, id string) (io.ReadCloser, error) {
	return nil, errors.New("resource retrieval not supported in mock dataset")
}

func (d *MockDataSet) DecodeResource(ctx context.Context, id string, obj any) error {
	r, err := d.Resource(ctx, id)
	if err != nil {
		return err
	}
	return jsonw.Decode(r, obj)
}

func (d *MockDataSet) Lease() *model.Lease {
	return d.lease
}

func (d *MockDataSet) Impression() *model.Impression {
	return d.lease.Impression
}

func (d *MockDataSet) Record() *model.Record {
	return d.record
}

func (d *MockDataSet) BlockNumber() int64 {
	return d.blockNumber
}

func (d *MockDataSet) LockerID() string {
	return d.lockerID
}

func (d *MockDataSet) ParticipantID() string {
	return d.participantID
}

func (d *MockDataSet) AddMockLeaseFromFile(t *testing.T, id string, filename string) {
	t.Helper()

	d.record = &model.Record{ID: id}
	leaseDoc, err := os.ReadFile(filename)
	require.NoError(t, err)

	var lease model.Lease
	err = jsonw.Unmarshal(leaseDoc, &lease)
	require.NoError(t, err)

	d.lease = &lease
}
