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
	"context"
	"errors"
	"io"
)

var (
	// ErrRecordNotFound indicates the ledger record was not found
	ErrRecordNotFound = errors.New("record not found")
	// ErrBlockNotFound indicates the ledger block was not found
	ErrBlockNotFound = errors.New("block not found")
	// ErrDataAssetNotFound indicates the data asset was not found
	ErrDataAssetNotFound = errors.New("data asset not found")
	// ErrAssetHeadNotFound indicates the record for an asset head is not found
	ErrAssetHeadNotFound = errors.New("asset head not found")
)

const (
	DataAssetStateKeep     DataAssetState = 1
	DataAssetStateRemove   DataAssetState = 2
	DataAssetStateNotFound DataAssetState = 3
)

type (
	DataAssetState int

	// AccessVerifier is used by MetaLocker vaults to retrieve record and data asset
	// information that is necessary to identify whether to serve the requested
	// data asset.
	AccessVerifier interface {
		// GetRecord returns a ledger record by its ID. Returns ErrRecordNotFound error
		// if record was not found.
		GetRecord(ctx context.Context, rid string) (*Record, error)
		// GetDataAssetState returns the state of the given data asset. Returns
		// ErrDataAssetNotFound error if data asset not found.
		GetDataAssetState(ctx context.Context, id string) (DataAssetState, error)
		// GetRecordState returns ledger record state for the given
		// record ID. It's useful to identify if the record
		// was published on the ledger (and its block ID) or if the lease
		// behind the record was revoked.
		GetRecordState(ctx context.Context, rid string) (*RecordState, error)
	}

	// Ledger is an interface to a MetaLocker ledger.
	Ledger interface {
		io.Closer

		// SubmitRecord adds a ledger records into the queue to be
		// included into the next block.
		SubmitRecord(ctx context.Context, r *Record) error
		// GetRecord returns a ledger record by its ID. Returns ErrRecordNotFound error
		// if record was not found.
		GetRecord(ctx context.Context, rid string) (*Record, error)
		// GetRecordState returns ledger record state for the given
		// record ID. It's useful to identify if the record
		// was published on the ledger (and its block ID) or if the lease
		// behind the record was revoked.
		GetRecordState(ctx context.Context, rid string) (*RecordState, error)
		// GetBlock returns a block definition for the given block number.
		GetBlock(ctx context.Context, bn int64) (*Block, error)
		// GetBlockRecords returns a list of all ledger records included
		// in the block as an array of arrays of strings:
		//     [record_id, routing_key, key_index]*
		// Returns ErrBlockNotFound error if block was not found.
		GetBlockRecords(ctx context.Context, bn int64) ([][]string, error)
		// GetGenesisBlock returns the definition of the genesis block.
		// If there is no genesis block yet, it will return nil as a block.
		GetGenesisBlock(ctx context.Context) (*Block, error)
		// GetTopBlock returns the definition of the top (latest) block.
		// If there are no blocks yet, it will return nil as a block.
		GetTopBlock(ctx context.Context) (*Block, error)
		// GetChain returns a sequence of block definitions of
		// the given length (depth), starting from the given block id
		GetChain(ctx context.Context, startNumber int64, depth int) ([]*Block, error)
		// GetDataAssetState returns the state of the given data asset. Returns
		// ErrDataAssetNotFound error if data asset not found.
		GetDataAssetState(ctx context.Context, id string) (DataAssetState, error)
		// GetAssetHead returns the record of type = head that defines the current asset head for the given ID.
		GetAssetHead(ctx context.Context, headID string) (*Record, error)
	}
)

const (
	NTopicNewBlock = "ledger.newBlock"

	MessageTypeNewBlockNotification = "NewBlockNotification"
)

type NewBlockMessage struct {
	Type   string `json:"type"`
	Number int64  `json:"number"`
}
