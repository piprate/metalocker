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

package pgembed

import (
	"fmt"
	"os"
	"testing"
	"time"

	embeddedPostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

const (
	Database = "metalocker"
	Username = "metalocker"
	Password = "password"
)

type TestTearDownFunc func(t *testing.T)

// Sometimes, if a test fails, the embedded database remains in memory.
// On OSX, run:
// 		lsof -nP -iTCP:9876 | grep LISTEN
// to find the offending process and kill it.

func NewEmbeddedDatabase(t *testing.T, connStringTemplate string) (string, *embeddedPostgres.EmbeddedPostgres, string) {
	t.Helper()

	dir, err := os.MkdirTemp(".", "tempdir_edb_")
	require.NoError(t, err)

	var port uint32 = 9876
	for ; port < 9876+5; port++ {
		database := embeddedPostgres.NewDatabase(embeddedPostgres.DefaultConfig().
			Username(Username).
			Password(Password).
			Database(Database).
			Version(embeddedPostgres.V13).
			RuntimePath(dir).
			Port(port).
			StartTimeout(15 * time.Second))
		if err = database.Start(); err == nil {
			connString := fmt.Sprintf(connStringTemplate, port)
			return connString, database, dir
		}
		log.Warn().Uint32("port", port).Msg("Postgres port is taken, trying the next one...")
	}

	t.Fatal(err)

	return "", nil, dir
}

func StopEmbeddedDatabase(t *testing.T, database *embeddedPostgres.EmbeddedPostgres, tempDir string) {
	t.Helper()

	err := database.Stop()
	_ = os.RemoveAll(tempDir)

	if err != nil {
		t.Fatal(err)
	}
}
