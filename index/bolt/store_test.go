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

package bolt_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/piprate/metalocker/contexts"
	"github.com/piprate/metalocker/index"
	. "github.com/piprate/metalocker/index/bolt"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = contexts.PreloadContextsIntoMemory()
	testbase.SetupLogFormat()
}

func newTestIndexStore(t *testing.T) (index.Store, string) {
	t.Helper()

	dir, err := os.MkdirTemp(".", "tempdir_")
	require.NoError(t, err)

	indexStore, err := NewIndexStore(
		&index.StoreConfig{
			ID:   testbase.IndexStoreID,
			Name: testbase.IndexStoreName,
			Type: Type,
			Params: map[string]any{
				ParameterFilePath: filepath.Join(dir, "wallet.bolt"),
			},
		}, nil)

	require.NoError(t, err)
	require.NoError(t, indexStore.Bind("abc"))
	assert.NotNil(t, indexStore)

	return indexStore, dir
}

func TestNewIndexStore(t *testing.T) {
	store, dir := newTestIndexStore(t)
	defer func() {
		_ = store.Close()
		_ = os.RemoveAll(dir)
	}()

	hash := store.GenesisBlockHash()
	assert.Equal(t, "abc", hash)
}

func TestIndexStore_RootIndex(t *testing.T) {
	store, dir := newTestIndexStore(t)
	defer func() {
		_ = store.Close()
		_ = os.RemoveAll(dir)
	}()

	userID := "did:piprate:QgH6CZvhjTUFvCbRUw4N6Z"
	_, err := store.RootIndex(userID, model.AccessLevelHosted)
	require.Error(t, index.ErrIndexNotFound, err)

	newIdx, err := store.CreateIndex(userID, index.TypeRoot, model.AccessLevelHosted)
	require.NoError(t, err)

	// check index properties were created correctly

	props := newIdx.Properties()
	assert.Equal(t, model.AccessLevelHosted, props.AccessLevel)
	assert.Equal(t, Algorithm, props.Algorithm)
	assert.Equal(t, index.TypeRoot, props.IndexType)
	assert.Equal(t, newIdx.ID(), props.Asset)

	idx, err := store.RootIndex(userID, model.AccessLevelHosted)
	require.NoError(t, err)
	assert.NotEmpty(t, idx)

	// check index properties were loaded correctly

	props = idx.Properties()
	assert.Equal(t, model.AccessLevelHosted, props.AccessLevel)
	assert.Equal(t, Algorithm, props.Algorithm)
	assert.Equal(t, index.TypeRoot, props.IndexType)
	assert.Equal(t, newIdx.ID(), props.Asset)
}
