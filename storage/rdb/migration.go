package rdb

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"strings"

	"github.com/golang-migrate/migrate/v4/source"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/piprate/metalocker/storage/rdb/ent"
	"github.com/rs/zerolog/log"
)

//go:embed migrations/*.sql
var fs embed.FS

func MigrateSchemaWithScripts(databaseURL, migrationsPath string) (uint, uint, error) {
	var m *migrate.Migrate
	var err error

	if strings.Contains(databaseURL, "?") {
		databaseURL += "&x-migrations-table=schema_migrations"
	} else {
		databaseURL += "?x-migrations-table=schema_migrations"
	}

	schema, err := SchemeFromURL(databaseURL)
	if err != nil {
		return 0, 0, err
	}
	if schema == "sqlite3" {
		// see https://github.com/golang-migrate/migrate/tree/master/database/sqlite3
		databaseURL += "&x-no-tx-wrap=true"
	}

	if migrationsPath == "" {
		var d source.Driver
		d, err = iofs.New(fs, "migrations")
		if err != nil {
			return 0, 0, err
		}

		m, err = migrate.NewWithSourceInstance("iofs", d, databaseURL)
	} else {
		m, err = migrate.New("file://"+migrationsPath, databaseURL)
	}

	if err != nil {
		return 0, 0, err
	}

	previousVersion, _, err := m.Version()
	newVersion := previousVersion
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return 0, 0, err
	}

	err = m.Up()
	closeFn := func() {
		_, _ = m.Close()
	}
	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return 0, 0, err
		}
		defer closeFn()
	} else {
		defer closeFn()

		newVersion, _, err = m.Version()
		if err != nil {
			return 0, 0, err
		}

		log.Info().Uint("before", previousVersion).Uint("after", newVersion).Msg("Database schema updated")
	}

	return previousVersion, newVersion, nil
}

var linesToSkip = []string{
	"BEGIN;", "COMMIT;",
	"PRAGMA foreign_keys = off;", "PRAGMA foreign_keys = on;", // SQLite
}

func okToSkip(line string) bool {
	for _, s := range linesToSkip {
		if s == line {
			return true
		}
	}
	return false
}

func CheckEntSchemaCompatibility(client *ent.Client) error {
	ctx := context.Background()
	buf := &bytes.Buffer{}
	if err := client.Schema.WriteTo(ctx, buf); err != nil {
		log.Err(err).Msg("Failed printing schema changes")
		return err
	}

	statementsStr := buf.String()
	if len(statementsStr) == 0 {
		return nil
	}

	lines := strings.Split(statementsStr, "\n")
	hasStatementsToExecute := false
	for _, line := range lines {
		if len(line) > 0 && !okToSkip(line) {
			hasStatementsToExecute = true
			break
		}
	}

	if hasStatementsToExecute {
		println("---------------------")
		println(statementsStr)
		println("---------------------")
		return errors.New("schema migration required")
	}

	return nil
}
