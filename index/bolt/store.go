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
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/piprate/metalocker/index"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/cmdbase"
	"github.com/piprate/metalocker/utils"
	"github.com/piprate/metalocker/utils/measure"
	"github.com/rs/zerolog/log"
	"go.etcd.io/bbolt"
)

const (
	Type = "bolt"

	ParameterFilePath = "file_path"

	Algorithm = "metalocker:root:1"
)

func init() {
	index.RegisterStoreType(Type, NewIndexStore)
}

type IndexStore struct {
	cfg              *index.StoreConfig
	props            *index.StoreProperties
	storeFilePath    string
	genesisBlockHash string
	client           *utils.BoltClient
}

var _ index.Store = (*IndexStore)(nil)

func NewIndexStore(cfg *index.StoreConfig, resolver cmdbase.ParameterResolver) (index.Store, error) {
	storeFilePath, ok := cfg.Params[ParameterFilePath].(string)
	if !ok {
		return nil, errors.New("parameter not found: " + ParameterFilePath +
			". Can't start index store")
	}

	storeFilePath = utils.AbsPathify(storeFilePath)

	// validate path and create missing directories, if required
	walletDir := filepath.Dir(storeFilePath)
	if _, err := os.Stat(walletDir); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(walletDir, 0o700); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	bc, err := utils.NewBoltClient(storeFilePath, InstallIndexStoreSchema)
	if err != nil {
		return nil, err
	}

	cfg.EncryptionMode = index.ModeNoEncryption

	indexStore := &IndexStore{
		cfg:           cfg,
		props:         index.NewStorePropertiesFromConfig(cfg),
		storeFilePath: storeFilePath,
		client:        bc,
	}

	return indexStore, nil
}

func (s *IndexStore) ID() string {
	return s.cfg.ID
}

func (s *IndexStore) Name() string {
	return s.cfg.Name
}

func (s *IndexStore) Properties() *index.StoreProperties {
	return s.props
}

func (s *IndexStore) CreateIndex(ctx context.Context, userID string, indexType string, accessLevel model.AccessLevel, opts ...index.Option) (index.Index, error) {

	if indexType != index.TypeRoot {
		return nil, fmt.Errorf("index type not supported in index store %s (%s): %s", s.ID(), Type, indexType)
	}

	idxOptions, err := index.NewOptions(opts...)
	if err != nil {
		return nil, err
	}

	// check the invoker didn't provide an encryption key
	if idxOptions.ClientKey != nil {
		return nil, errors.New("index encryption not supported")
	}

	// the logic below is for root index creation only

	indexID := index.RootIndexID(userID, accessLevel)

	installed := false
	err = s.client.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(indexID))
		if b != nil {
			installed = true
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if installed {
		return nil, index.ErrIndexExists
	}

	props := &index.Properties{
		IndexType:   indexType,
		Asset:       indexID,
		AccessLevel: accessLevel,
		Algorithm:   Algorithm,
	}

	if err := InstallIndexSchema(s.client, indexID); err != nil {
		return nil, err
	}

	if err := saveProperties(userID, indexID, props, s.client); err != nil {
		return nil, err
	}

	return &Index{
		id:     indexID,
		userID: userID,
		props:  props,
		client: s.client,
	}, nil
}

func (s *IndexStore) Index(ctx context.Context, userID string, id string) (index.Index, error) {
	props, err := loadProperties(id, s.client)
	if err != nil {
		return nil, err
	}

	return &Index{
		id:     id,
		userID: userID,
		props:  props,
		client: s.client,
	}, nil
}

func (s *IndexStore) ListIndexes(ctx context.Context, userID string) ([]*index.Properties, error) {
	panic("operation boltIndex::ListIndexes not implemented")
}

func (s *IndexStore) DeleteIndex(ctx context.Context, userID, id string) error {
	panic("operation boltIndex::DeleteIndex not implemented")
}

func (s *IndexStore) RootIndex(ctx context.Context, userID string, lvl model.AccessLevel) (index.RootIndex, error) {
	indexID := index.RootIndexID(userID, lvl)
	ix, err := s.Index(ctx, userID, indexID)
	if err != nil {
		return nil, err
	}
	return ix.(index.RootIndex), nil
}

func (s *IndexStore) Bind(ctx context.Context, gbHash string) error {
	walletBlockHash, err := s.fetchGenesisBlockHash()
	if err != nil {
		return err
	}

	if walletBlockHash == "" {
		// new wallet index store
		if err = s.storeGenesisBlock(gbHash); err != nil {
			return err
		}
	} else if walletBlockHash != gbHash {
		return fmt.Errorf(
			"genesis block hash mismatch between MetaLocker and local data wallet index: %s != %s",
			gbHash, walletBlockHash)
	}

	s.genesisBlockHash = gbHash

	return nil
}

func (s *IndexStore) GenesisBlockHash(ctx context.Context) string {
	return s.genesisBlockHash
}

func (s *IndexStore) fetchGenesisBlockHash() (string, error) {
	defer measure.ExecTime("store.fetchGenesisBlockHash")()

	b, err := s.client.FetchBytes(ControlsKey, GenesisBlockHashKey)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *IndexStore) storeGenesisBlock(genesisBlockID string) error {
	return s.client.Update(ControlsKey, GenesisBlockHashKey, []byte(genesisBlockID))
}

func (s *IndexStore) Close() error {
	if s.client != nil {
		log.Debug().Msg("Closing BoltDB data wallet index store file")
		err := s.client.Close()
		s.client = nil
		return err
	}
	return nil
}
