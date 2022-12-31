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
	"io"
)

var (
	// ErrDataSetNotFound indicates the dataset was not found. It may mean the dataset is available
	// in MetaLocker, but not accessible by the given data wallet.
	ErrDataSetNotFound  = errors.New("dataset not found")
	ErrResourceNotFound = errors.New("resource not found")
)

// DataSet defines an interface to a MetaLocker dataset stored in the given record.
type DataSet interface {
	// ID returns the dataset's record ID.
	ID() string
	// MetaResource returns a reader for the dataset's meta resource.
	MetaResource() (io.ReadCloser, error)
	// DecodeMetaResource is a convenience function that unmarshals the dataset's metadata into the given structure.
	DecodeMetaResource(obj any) error
	// Resources returns a list of resource IDs that belong to the dataset.
	Resources() []string
	// Resource returns a reader for the given resource within the dataset.
	Resource(id string) (io.ReadCloser, error)
	// DecodeResource is a convenience function that unmarshals the requested resource into the given structure.
	DecodeResource(id string, obj any) error
	// Lease returns the dataset's lease document
	Lease() *Lease
	// Impression returns the dataset's impression document (also available through Lease() )
	Impression() *Impression

	// Record returns the dataset's record structure
	Record() *Record
	// BlockNumber returns the number (ID) of the block where the dataset's record appeared.
	BlockNumber() int64
	// LockerID returns the ID of the locker that contains the dataset.
	LockerID() string
	// ParticipantID returns the ID (the corresponding identity's DID) of the locker participant
	// that submitted the dataset.
	ParticipantID() string
}
