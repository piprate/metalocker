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
	"io"
	"time"

	"github.com/piprate/metalocker/model"
)

type (

	// RecordFuture represents the result of an asynchronous ledger record submission.
	RecordFuture interface {
		// Wait is a blocking operation that waits for the record to appear on the ledger
		// for up to the specified timeout value.
		Wait(timeout time.Duration) error

		// ID returns the ID of the ledger record.
		ID() string

		// Lease returns the lease of the submitted record (if applicable).
		Lease() *model.Lease

		// DataSet returns a fully featured model.DataSet for the submitted record.
		// This function will return nil if either the future state is !ready
		// or the submission returned an error.
		DataSet() model.DataSet

		// Heads returns a map Head ID ==> Head Record ID for operations that
		// create asset heads.
		Heads() map[string]string

		// IsReady returns true when the Wait function is guaranteed to not block
		IsReady() bool

		// WaitList returns a list of IDs of all ledger records produced
		// by the given submission.
		WaitList() []string

		// Error returns a non-nil value if the submission failed. It may not return
		// the actual error until the future is in 'ready' state.
		Error() error
	}

	// Builder enables interactive construction of a dataset, to be submitted
	// to MetaLocker ledger. Once all the required operations are completed,
	// call Submit to send the dataset definition to the ledger.
	Builder interface {
		// CreatorID returns the creator's identity ID (DID).
		CreatorID() string
		// AddResource attaches a new resource/blob to the dataset and returns
		// its content-addressable asset ID.
		AddResource(r io.Reader, opts ...BuilderOption) (string, error)
		// ImportResource imports an existing resource.
		ImportResource(res *model.StoredResource) error
		// AddMetaResource adds a meta resource to the dataset and returns
		// its content-addressable asset ID.
		AddMetaResource(meta any, opts ...BuilderOption) (string, error)
		// AddProvenance adds a provenance definition to the resource with the given ID.
		AddProvenance(id string, provenance any, override bool) error

		// SetHeads instructs the builder to set the dataset as a head with the given
		// names when it's submitted to the ledger.
		SetHeads(headName ...string) error

		// Submit submits the dataset to the MetaLocker ledger.
		Submit(expiryTime time.Time) RecordFuture

		// Cancel cancels the dataset creation. NOT IMPLEMENTED YET.
		Cancel() error
	}
)
