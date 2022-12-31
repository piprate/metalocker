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

package model

import (
	"errors"
	"time"

	"github.com/piprate/metalocker/utils/jsonw"
)

var (
	// ErrOperationNotFound indicates that operation was not found
	ErrOperationNotFound = errors.New("operation not found")
)

// OffChainStorage is an interface to a storage layer that is used to store ledger operation definitions.
// In contrast with ledger records which are permanent, offchain data can be deleted, if the underlying
// dataset lease expired, or it was revoked, or for any other reason that prohibits access to the given
// operation.
type OffChainStorage interface {
	GetOperation(opAddr string) ([]byte, error)
	SendOperation(opData []byte) (string, error)
	PurgeOperation(opAddr string) error
}

// Lease is a dataset lease as a MetaLocker operation. This lease is stored in OffChainStorage.
type Lease struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	ExpiresAt   *time.Time        `json:"expire,omitempty"`
	Resources   []*StoredResource `json:"storage"`
	DataSetType string            `json:"datasetType"`
	Impression  *Impression       `json:"impression"`
	Provenance  *ProvEntity       `json:"provenance,omitempty"`
	Proof       *Proof            `json:"proof,omitempty"`
}

func NewLease(body []byte) (*Lease, error) {
	var op Lease

	if err := jsonw.Unmarshal(body, &op); err != nil {
		return nil, err
	}

	return &op, nil
}

func (l *Lease) MetaResource() *StoredResource {
	for _, res := range l.Resources {
		if res.Asset == l.Impression.MetaResource.Asset {
			return res
		}
	}
	return nil
}

func (l *Lease) Resource(assetID string) *StoredResource {
	for _, res := range l.Resources {
		if res.Asset == assetID {
			return res
		}
	}
	return nil
}

func (l *Lease) DataAssetList(includeMetaAsset bool) []string {
	var assets []string
	for _, res := range l.Resources {
		if includeMetaAsset || res.Asset != l.Impression.MetaResource.Asset {
			assets = append(assets, res.Asset)
		}
	}
	return assets
}

func (l *Lease) GetResourceIDs() []string {
	var ids []string
	for _, res := range l.Resources {
		if id := res.StorageID(); id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func (l *Lease) GenerateAccessToken(rid string) string {
	var leaseExpiryTime int64 = 0
	if l.ExpiresAt != nil {
		leaseExpiryTime = l.ExpiresAt.Unix()
	}
	return GenerateAccessToken(rid, l.ID, time.Now().Unix(), leaseExpiryTime)
}
