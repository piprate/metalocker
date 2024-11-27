package rdb

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
	"github.com/piprate/metalocker/storage/rdb/ent"
	"github.com/piprate/metalocker/utils"
	"github.com/rs/zerolog"
)

func NewEntClient(databaseURL string, logLevel zerolog.Level) (*ent.Client, error) {

	if logLevel < 1 {
		// set log level to None
		logLevel = 1
	}

	schema, err := SchemeFromURL(databaseURL)
	if err != nil {
		return nil, err
	}

	var drv *entsql.Driver
	switch schema {
	case "postgres":
		connConfig, err := pgx.ParseConfig(databaseURL)
		if err != nil {
			return nil, err
		}

		connConfig.Tracer = utils.NewZerologQueryTracer(logLevel)
		connStr := stdlib.RegisterConnConfig(connConfig)

		sqlDB, err := sql.Open("pgx", connStr)
		if err != nil {
			return nil, err
		}

		drv = entsql.OpenDB(dialect.Postgres, sqlDB)
	case "sqlite3":
		str := strings.ReplaceAll(databaseURL, "sqlite3://", "file:")
		sqlDB, err := sql.Open("sqlite3", str)
		if err != nil {
			return nil, err
		}
		drv = entsql.OpenDB(dialect.SQLite, sqlDB)
	default:
		return nil, fmt.Errorf("unsupported database schema: %s", schema)
	}

	entClient := ent.NewClient(ent.Driver(drv))

	return entClient, nil
}

// SchemeFromURL returns the scheme from a URL string
func SchemeFromURL(url string) (string, error) {
	if url == "" {
		return "", errors.New("URL cannot be empty")
	}

	i := strings.Index(url, ":")

	// No : or : is the first character.
	if i < 1 {
		return "", errors.New("no scheme")
	}

	return url[0:i], nil
}
