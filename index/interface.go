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

package index

import (
	"errors"
	"fmt"
	"io"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/cmdbase"
)

var (
	ErrIndexNotFound       = errors.New("index not found")
	ErrIndexExists         = errors.New("index already exists")
	ErrIndexStoreNotFound  = errors.New("index store not found")
	ErrLockerStateNotFound = errors.New("locker state not found")
	ErrLockerStateExists   = errors.New("locker state already exists")
)

const (
	ModeNoEncryption      EncryptionMode = "none"
	ModeManagedEncryption EncryptionMode = "managed"
	ModeClientEncryption  EncryptionMode = "client"
)

type (
	EncryptionMode string

	LockerState struct {
		ID         string `json:"id"`
		IndexID    string `json:"indexID"`
		AccountID  string `json:"accountID"`
		FirstBlock int64  `json:"firstBlock,omitempty"`
		TopBlock   int64  `json:"topBlock,omitempty"`
	}

	Properties struct {
		IndexType   string            `json:"type"`
		Asset       string            `json:"asset,omitempty"`
		AccessLevel model.AccessLevel `json:"accessLevel"`
		Algorithm   string            `json:"algo"`
		Params      any               `json:"params"`
	}

	// Index is a database or any other storage system that provided an index of all
	// or selected ledger records for easy and efficient access. Indexes can be generic
	// or purpose-built.
	Index interface {
		io.Closer

		ID() string
		Properties() *Properties

		IsLocked() bool
		Unlock(key []byte) error
		Lock()

		IsWritable() bool
		Writer() (Writer, error)
	}

	// Writer is an interface for adding new records into an index.
	Writer interface {
		io.Closer

		ID() string

		LockerStates() ([]LockerState, error)
		AddLockerState(accountID, lockerID string, firstBlock int64) error
		AddLease(ds model.DataSet, effectiveBlockNumber int64) error
		AddLeaseRevocation(ds model.DataSet) error
		UpdateTopBlock(blockNumber int64) error
	}

	StoreProperties struct {
		ID             string         `json:"id"`
		Name           string         `json:"name"`
		Type           string         `json:"type"`
		EncryptionMode EncryptionMode `json:"encryptionMode"`
	}

	StoreConfig struct {
		ID             string         `mapstructure:"id" json:"id"`
		Name           string         `mapstructure:"name" json:"name"`
		Type           string         `mapstructure:"type" json:"type"`
		EncryptionMode EncryptionMode `mapstructure:"encryptionMode" json:"encryptionMode"`
		Params         map[string]any `mapstructure:"params" json:"params,omitempty"`
	}

	Options struct {
		ClientKey  []byte `json:"key,omitempty"`
		Parameters any    `json:"parameters,omitempty"`
	}

	Option func(opts *Options) error

	// Store is a facility that can store indexes. Typically, a store is based on a specific
	// database or storage technology that provides the desired runtime properties.
	Store interface {
		io.Closer

		ID() string
		Name() string
		Properties() *StoreProperties

		CreateIndex(userID string, indexType string, accessLevel model.AccessLevel, opts ...Option) (Index, error)
		RootIndex(userID string, lvl model.AccessLevel) (RootIndex, error)
		Index(userID string, id string) (Index, error)
		ListIndexes(userID string) ([]*Properties, error)
		DeleteIndex(userID, id string) error

		// Bind links the index store to a particular ledger instance by storing genesis block ID in the store.
		// If the store is already bound to a ledger, if will check the provided hash and return an error
		// if there is a mismatch. This is useful to catch conditions when the index store was used
		// in the context of a different ledger (i.e. in the development environment.
		Bind(gbHash string) error

		GenesisBlockHash() string
	}

	// Client provides an interface to a group of index stores, accessed using a priority list.
	// It hides the implementation details of the store types available, and actual location
	// of the index stores.
	Client interface {
		io.Closer

		// Bind links all underlying index states to a specific genesis block hash. If any of the stores
		// were already linked to a different hash, the call would fail.
		Bind(gbHash string) error

		// RootIndex returns a root index for the given account and requested access level.
		// If the index is not found, it will return ErrIndexNotFound.
		RootIndex(userID string, lvl model.AccessLevel) (RootIndex, error)

		// Index returns an index with the given id for the given account.
		// If the index is not found, it will return ErrIndexNotFound.
		Index(userID string, id string) (Index, error)
		// ListIndexes return a list of index definitions for all the indexes that are available
		// through this index client. We return a list of Properties to avoid construction of
		// all index instances that may be an expensive operation.
		ListIndexes(userID string) ([]*Properties, error)
		// DeleteIndex deletes the index from all the underlying index stores.
		DeleteIndex(userID, id string) error

		// IndexStore returns the index store with the given name.
		// If the store is not found, it will return ErrIndexStoreNotFound
		IndexStore(storeName string) (Store, error)
		// IndexStores return a list of index store definitions for all the stores that are available
		// through this index client. We return a list of StoreProperties to avoid construction of
		// all store instances that may be an expensive operation.
		IndexStores() []*StoreProperties
		// AddIndexStore instantiates and adds a new index store based on the provided configuration.
		AddIndexStore(cfg *StoreConfig, resolver cmdbase.ParameterResolver) error
	}
)

func WithEncryption(key []byte) Option {
	return func(o *Options) error {
		o.ClientKey = key
		return nil
	}
}

func WithOptions(opts Options) Option {
	return func(o *Options) error {
		*o = opts
		return nil
	}
}

// NewOptions reads Options from Option array
func NewOptions(opts ...Option) (Options, error) {
	optsStruct := Options{}
	for _, option := range opts {
		err := option(&optsStruct)
		if err != nil {
			return optsStruct, fmt.Errorf("failed to read options: %w", err)
		}
	}
	return optsStruct, nil
}

func NewStorePropertiesFromConfig(cfg *StoreConfig) *StoreProperties {
	return &StoreProperties{
		ID:             cfg.ID,
		Name:           cfg.Name,
		Type:           cfg.Type,
		EncryptionMode: cfg.EncryptionMode,
	}
}
