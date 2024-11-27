package rdb

import (
	"errors"

	"github.com/piprate/metalocker/sdk/cmdbase"
	"github.com/piprate/metalocker/storage"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	storage.Register("rdb", CreateIdentityBackend)
}

const (
	ParameterURL            = "url"
	ParameterSyncSchema     = "sync_schema"
	ParameterMigrationsPath = "migrations_path"
	ParameterLogLevel       = "log_level"
)

func CreateIdentityBackend(params storage.Parameters, resolver cmdbase.ParameterResolver) (storage.IdentityBackend, error) {

	databaseURL, err := readDatabaseURL(params, resolver)
	if err != nil {
		return nil, err
	}

	var syncSchema bool
	syncSchemaVal := params[ParameterSyncSchema]
	if syncSchemaVal != nil {
		syncSchema, _ = syncSchemaVal.(bool)
	}

	var migrationsPath string
	migrationsPathVal := params[ParameterMigrationsPath]
	if migrationsPathVal != nil {
		migrationsPath, _ = migrationsPathVal.(string)
	}

	var logLevel int
	logLevelVal := params[ParameterLogLevel]
	if logLevelVal != nil {
		logLevel, _ = logLevelVal.(int)
	}

	entClient, err := NewEntClient(databaseURL, zerolog.Level(logLevel))
	if err != nil {
		return nil, err
	}

	if syncSchema {
		// run migrations
		_, _, err = MigrateSchemaWithScripts(databaseURL, migrationsPath)
		if err != nil {
			return nil, err
		}

		if err = CheckEntSchemaCompatibility(entClient); err != nil {
			return nil, err
		}
	}

	return NewRelationalBackend(entClient), nil
}

func readDatabaseURL(params map[string]any, resolver cmdbase.ParameterResolver) (string, error) {

	var dbURL string
	var err error
	url := params[ParameterURL]
	var isString bool
	if dbURL, isString = url.(string); !isString {
		if resolver != nil {
			var b []byte
			b, err = jsonw.Marshal(url)
			if err == nil {
				var urlParam map[string]any
				if err = jsonw.Unmarshal(b, &urlParam); err == nil {
					dbURL, err = resolver.ResolveString(urlParam)
				}
			}
		} else {
			err = errors.New("wrong type of url parameter")
		}
	}

	if err != nil {
		log.Err(err).Msg("Error reading RDB backend parameter:" + ParameterURL)
		return "", errors.New("can't read RDB backend parameter: " + ParameterURL)
	}

	if dbURL == "" {
		return "", errors.New("parameter not defined: " + ParameterURL +
			". Can't start RDB backend")
	}

	return dbURL, nil
}
