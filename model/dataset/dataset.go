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

package dataset

import (
	"errors"
	"io"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog/log"
)

type (
	DataSetImpl struct {
		blobManager   model.BlobManager
		record        *model.Record
		lease         *model.Lease
		blockNumber   int64
		lockerID      string
		participantID string
		accessToken   string
	}

	LoadOptions struct {
		LockerID string
	}

	// LoadOption is for defining optional parameters for loading datasets
	LoadOption func(opts *LoadOptions) error
)

var _ model.DataSet = (*DataSetImpl)(nil)

func FromLocker(id string) LoadOption {
	return func(opts *LoadOptions) error {
		opts.LockerID = id
		return nil
	}
}

func WithLoadOptions(options LoadOptions) LoadOption {
	return func(opts *LoadOptions) error {
		opts.LockerID = options.LockerID
		return nil
	}
}

func NewDataSetImpl(r *model.Record, lease *model.Lease, blockNumber int64, lockerID, participantID string, blobManager model.BlobManager) *DataSetImpl {
	return &DataSetImpl{
		record:        r,
		lease:         lease,
		blockNumber:   blockNumber,
		lockerID:      lockerID,
		participantID: participantID,
		blobManager:   blobManager,
	}
}

func NewRevokedDataSetImpl(r *model.Record, blockNumber int64, lockerID, participantID string) *DataSetImpl {
	return &DataSetImpl{
		record:        r,
		blockNumber:   blockNumber,
		lockerID:      lockerID,
		participantID: participantID,
	}
}

func (d *DataSetImpl) ID() string {
	return d.record.ID
}

func (d *DataSetImpl) getAccessToken() string {
	if d.lease == nil {
		log.Warn().Msg("Attempted to get an access token for revoked record")
		return ""
	}
	if d.accessToken == "" {
		d.accessToken = d.lease.GenerateAccessToken(d.record.ID)
	}
	return d.accessToken
}

func (d *DataSetImpl) MetaResource() (io.ReadCloser, error) {
	if d.lease == nil {
		return nil, errors.New("data access forbidden for revoked records")
	}
	metaRes := d.lease.MetaResource()
	if metaRes == nil {
		log.Error().Str("id", d.lease.Impression.MetaResource.Asset).Msg("Meta resource not found")
		return nil, model.ErrResourceNotFound
	}

	r, err := d.blobManager.GetBlob(metaRes, d.getAccessToken())
	if err != nil {
		log.Err(err).Str("id", metaRes.ID).Interface("params", metaRes.Params).Msg("Error serving blob")
	}

	return r, err
}

func (d *DataSetImpl) DecodeMetaResource(obj any) error {
	r, err := d.MetaResource()
	if err != nil {
		return err
	}
	return jsonw.Decode(r, obj)
}

func (d *DataSetImpl) Resources() []string {
	if d.lease == nil {
		log.Warn().Msg("Attempted to get resource list for revoked record")
		return nil
	}
	resList := make([]string, len(d.lease.Resources)-1)
	i := 0
	for _, res := range d.lease.Resources {
		if res.Asset != d.lease.Impression.MetaResource.Asset {
			resList[i] = res.Asset
			i++
		}
	}
	return resList
}

func (d *DataSetImpl) Resource(id string) (io.ReadCloser, error) {
	if d.lease == nil {
		return nil, errors.New("data access forbidden for revoked records")
	}

	var requestedResource *model.StoredResource
	for _, res := range d.lease.Resources {
		if res.Asset == id {
			requestedResource = res
			break
		}
	}

	if requestedResource == nil {
		log.Error().Str("id", id).Msg("Resource not found")
		return nil, model.ErrResourceNotFound
	}

	r, err := d.blobManager.GetBlob(requestedResource, d.getAccessToken())
	if err != nil {
		log.Err(err).Str("id", requestedResource.ID).Interface("params", requestedResource.Params).Msg("Error serving blob")
	}

	return r, err
}

func (d *DataSetImpl) DecodeResource(id string, obj any) error {
	r, err := d.Resource(id)
	if err != nil {
		return err
	}
	return jsonw.Decode(r, obj)
}

func (d *DataSetImpl) Lease() *model.Lease {
	return d.lease
}

func (d *DataSetImpl) Impression() *model.Impression {
	if d.lease != nil {
		return d.lease.Impression
	} else {
		return nil
	}
}

func (d *DataSetImpl) Record() *model.Record {
	return d.record
}

func (d *DataSetImpl) BlockNumber() int64 {
	return d.blockNumber
}

func (d *DataSetImpl) LockerID() string {
	return d.lockerID
}

func (d *DataSetImpl) ParticipantID() string {
	return d.participantID
}
