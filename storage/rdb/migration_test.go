package rdb_test

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/piprate/metalocker/sdk/testbase/pgembed"
	. "github.com/piprate/metalocker/storage/rdb"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestMigrateSchemaWithScripts_EmbeddedScripts(t *testing.T) {
	databaseURL, edb, dir := pgembed.NewEmbeddedDatabase(t)
	defer func() {
		pgembed.StopEmbeddedDatabase(t, edb, dir)
	}()

	_, _, err := MigrateSchemaWithScripts(databaseURL, "")
	require.NoError(t, err)
}

func TestMigrateSchemaWithScripts_ScriptFolder(t *testing.T) {
	databaseURL, edb, dir := pgembed.NewEmbeddedDatabase(t)
	defer func() {
		pgembed.StopEmbeddedDatabase(t, edb, dir)
	}()

	_, _, err := MigrateSchemaWithScripts(databaseURL, "migrations")
	require.NoError(t, err)
}

func TestCheckEntSchemaCompatibility(t *testing.T) {
	databaseURL, edb, dir := pgembed.NewEmbeddedDatabase(t)
	defer func() {
		pgembed.StopEmbeddedDatabase(t, edb, dir)
	}()

	_, _, err := MigrateSchemaWithScripts(databaseURL, "migrations")
	require.NoError(t, err)

	client, err := NewEntClient(databaseURL, zerolog.DebugLevel)
	require.NoError(t, err)

	err = CheckEntSchemaCompatibility(client)
	require.NoError(t, err)
}
