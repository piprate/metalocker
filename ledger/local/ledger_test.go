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

package local_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/piprate/metalocker/contexts"
	. "github.com/piprate/metalocker/ledger/local"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/piprate/metalocker/services/notification"
	"github.com/piprate/metalocker/vaults"
	_ "github.com/piprate/metalocker/vaults/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = contexts.PreloadContextsIntoMemory()
	testbase.SetupLogFormat()
}

func NewTestBoltLedger(t *testing.T, blockCheckInterval uint64) (*BoltLedger, vaults.Vault, string) {
	t.Helper()

	dir, err := os.MkdirTemp(".", "tempdir_")
	require.NoError(t, err)

	dbFilepath := filepath.Join(dir, "ledger.bolt")

	offchainDir := filepath.Join(dir, "offchain")
	err = os.Mkdir(offchainDir, 0o755)
	require.NoError(t, err)

	offchainAPI, err := vaults.CreateVault(&vaults.Config{
		ID:   "HjxvefRJ9Vj5tn9Eq5yc8dYHdwG2xM3MUDUuRRdnPV1F",
		Name: "offchain",
		Type: "fs",
		Params: map[string]any{
			"root_dir": offchainDir,
		},
	}, nil, nil)
	require.NoError(t, err)

	ns := notification.NewLocalNotificationService(100)

	bl, err := NewBoltLedger(dbFilepath, ns, 1000, blockCheckInterval)
	require.NoError(t, err)

	return bl, offchainAPI, dir
}

func TestBoltLedger_GetGenesisBlock(t *testing.T) {
	bl, _, dir := NewTestBoltLedger(t, 0)
	defer os.RemoveAll(dir) // clean up
	defer bl.Close()

	gb, err := bl.GetGenesisBlock()
	require.NoError(t, err)
	assert.NotNil(t, gb)
}

func TestBoltLedger_SaveRecord(t *testing.T) {
	bl, _, dir := NewTestBoltLedger(t, 0)
	defer os.RemoveAll(dir) // clean up
	defer bl.Close()

	rec := &model.Record{
		ID: "xx",
	}

	_, err := bl.OpenNewBlockSession()
	require.NoError(t, err)

	err = bl.SaveRecord(rec)
	require.NoError(t, err)

	r, err := bl.GetRecord("xx")
	require.NoError(t, err)
	assert.Equal(t, "xx", r.ID)

	rs, err := bl.GetRecordState("xx")
	require.NoError(t, err)
	assert.Equal(t, model.StatusPending, rs.Status)
	assert.Equal(t, int64(0), rs.BlockNumber)
}
