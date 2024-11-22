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
	"context"
	"errors"
	"fmt"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/cmdbase"
	"github.com/rs/zerolog/log"
)

type StoreConstructor func(cfg *StoreConfig, resolver cmdbase.ParameterResolver) (Store, error)

var storeConstructors = make(map[string]StoreConstructor)

func RegisterStoreType(storeType string, ctor StoreConstructor) {
	if _, ok := storeConstructors[storeType]; ok {
		panic("index store constructor already registered for type: " + storeType)
	}

	storeConstructors[storeType] = ctor
}

func CreateStore(cfg *StoreConfig, resolver cmdbase.ParameterResolver) (Store, error) {

	log.Debug().Str("id", cfg.ID).Str("name", cfg.Name).Str("type", cfg.Type).Msg("Creating index store")

	ctor, ok := storeConstructors[cfg.Type]
	if !ok {
		return nil, fmt.Errorf("index store type %s not known or loaded", cfg.Type)
	}

	return ctor(cfg, resolver)
}

type clientImpl struct {
	stores           map[string]Store
	priorityList     []Store
	genesisBlockHash string
}

var _ Client = (*clientImpl)(nil)

func NewLocalIndexClient(ctx context.Context, storeConfigs []*StoreConfig, resolver cmdbase.ParameterResolver, genesisBlockHash string) (Client, error) {
	indexClient := &clientImpl{
		stores:           make(map[string]Store),
		genesisBlockHash: genesisBlockHash,
	}

	for _, cfg := range storeConfigs {
		if err := indexClient.AddIndexStore(ctx, cfg, resolver); err != nil {
			_ = indexClient.Close()
			return nil, err
		}
	}

	return indexClient, nil
}

func (ic *clientImpl) Bind(ctx context.Context, gbHash string) error {
	if ic.genesisBlockHash != "" {
		if ic.genesisBlockHash != gbHash {
			return errors.New("genesis block mismatch")
		}
	} else {
		for _, s := range ic.priorityList {
			if err := s.Bind(ctx, gbHash); err != nil {
				return err
			}
		}
		ic.genesisBlockHash = gbHash
	}

	return nil
}

func (ic *clientImpl) AddIndexStore(ctx context.Context, cfg *StoreConfig, resolver cmdbase.ParameterResolver) error {
	store, err := CreateStore(cfg, resolver)
	if err != nil {
		return err
	}
	ic.stores[cfg.Name] = store
	ic.priorityList = append(ic.priorityList, store)

	if ic.genesisBlockHash != "" {
		if err = store.Bind(ctx, ic.genesisBlockHash); err != nil {
			return err
		}
	}

	return nil
}

func (ic *clientImpl) RootIndex(ctx context.Context, userID string, lvl model.AccessLevel) (RootIndex, error) {
	if len(ic.priorityList) == 0 {
		return nil, ErrIndexNotFound
	}

	for _, store := range ic.priorityList {
		ix, err := store.RootIndex(ctx, userID, lvl)
		if err != nil {
			if errors.Is(err, ErrIndexNotFound) {
				continue
			}
			return nil, err
		}
		return ix, nil
	}

	return nil, ErrIndexNotFound
}

func (ic *clientImpl) Index(ctx context.Context, userID string, id string) (Index, error) {
	if len(ic.priorityList) == 0 {
		return nil, ErrIndexNotFound
	}

	for _, store := range ic.priorityList {
		ix, err := store.Index(ctx, userID, id)
		if err != nil {
			if errors.Is(err, ErrIndexNotFound) {
				continue
			}
			return nil, err
		}
		return ix, nil
	}

	return nil, ErrIndexNotFound
}

func (ic *clientImpl) ListIndexes(ctx context.Context, userID string) ([]*Properties, error) {
	var res []*Properties
	for _, store := range ic.stores {
		propsList, err := store.ListIndexes(ctx, userID)
		if err != nil {
			return nil, err
		}
		res = append(res, propsList...)
	}
	return res, nil
}

func (ic *clientImpl) DeleteIndex(ctx context.Context, userID, id string) error {
	for _, store := range ic.stores {
		_, err := store.Index(ctx, userID, id)
		if err != nil {
			if errors.Is(err, ErrIndexNotFound) {
				continue
			} else {
				return err
			}
		}
		return store.DeleteIndex(ctx, userID, id)
	}
	return ErrIndexNotFound
}

func (ic *clientImpl) IndexStore(ctx context.Context, storeName string) (Store, error) {
	store, found := ic.stores[storeName]
	if !found {
		return nil, ErrIndexStoreNotFound
	}
	return store, nil
}

func (ic *clientImpl) IndexStores(ctx context.Context) []*StoreProperties {
	res := make([]*StoreProperties, len(ic.priorityList))
	for i, store := range ic.priorityList {
		res[i] = store.Properties()
	}

	return res
}

func (ic *clientImpl) Close() error {
	var returnError error
	for _, store := range ic.stores {
		if err := store.Close(); err != nil {
			returnError = errors.New("one of the index stores failed to close")
		}
	}
	return returnError
}
