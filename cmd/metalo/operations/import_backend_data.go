package operations

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/sdk/cmdbase"
	"github.com/piprate/metalocker/storage"
	"github.com/piprate/metalocker/utils"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func ImportBackendData(ctx context.Context, importDirPath, configFilePath string) error { //nolint:gocyclo

	// read configuration
	cfg, err := readConfigFile(configFilePath)
	if err != nil {
		return err
	}

	resolver, err := cmdbase.ConfigureParameterResolver(cfg, filepath.Dir(configFilePath))
	if err != nil {
		return err
	}

	backend, err := initBackendInstance(cfg, resolver)
	if err != nil {
		return err
	}

	// import accounts

	accountsBody, err := os.ReadFile(filepath.Join(importDirPath, "accounts.json"))
	if err != nil {
		return err
	}

	var acctList []*account.Account
	err = jsonw.Unmarshal(accountsBody, &acctList)
	if err != nil {
		return err
	}

	for _, acct := range acctList {
		err = backend.CreateAccount(ctx, acct)
		if err != nil {
			return err
		}
	}

	// import identities

	identitiesBody, err := os.ReadFile(filepath.Join(importDirPath, "identities.json"))
	if err != nil {
		return err
	}

	iidMap := map[string][]*account.DataEnvelope{}
	err = jsonw.Unmarshal(identitiesBody, &iidMap)
	if err != nil {
		return err
	}

	identityCount := 0
	for acctID, iidList := range iidMap {
		for _, iid := range iidList {
			err = backend.StoreIdentity(ctx, acctID, iid)
			if err != nil {
				return err
			}
			identityCount++
		}
	}

	// import lockers

	lockersBody, err := os.ReadFile(filepath.Join(importDirPath, "lockers.json"))
	if err != nil {
		return err
	}

	lockerMap := map[string][]*account.DataEnvelope{}
	err = jsonw.Unmarshal(lockersBody, &lockerMap)
	if err != nil {
		return err
	}

	lockerCount := 0
	for acctID, lockerList := range lockerMap {
		for _, l := range lockerList {
			err = backend.StoreLocker(ctx, acctID, l)
			if err != nil {
				return err
			}
			lockerCount++
		}
	}

	// import properties

	propertiesBody, err := os.ReadFile(filepath.Join(importDirPath, "properties.json"))
	if err != nil {
		return err
	}

	propertyMap := map[string][]*account.DataEnvelope{}
	err = jsonw.Unmarshal(propertiesBody, &propertyMap)
	if err != nil {
		return err
	}

	propertyCount := 0
	for acctID, propertyList := range propertyMap {
		for _, prop := range propertyList {
			err = backend.StoreProperty(ctx, acctID, prop)
			if err != nil {
				return err
			}
			propertyCount++
		}
	}

	// import access keys

	accessKeyBody, err := os.ReadFile(filepath.Join(importDirPath, "access_keys.json"))
	if err != nil {
		return err
	}

	var accessKeyList []*model.AccessKey
	err = jsonw.Unmarshal(accessKeyBody, &accessKeyList)
	if err != nil {
		return err
	}

	for _, key := range accessKeyList {
		err = backend.StoreAccessKey(ctx, key)
		if err != nil {
			return err
		}
	}

	// import DID documents

	didBody, err := os.ReadFile(filepath.Join(importDirPath, "dids.json"))
	if err != nil {
		return err
	}

	var didList []*model.DIDDocument
	err = jsonw.Unmarshal(didBody, &didList)
	if err != nil {
		return err
	}

	for _, d := range didList {
		err = backend.CreateDIDDocument(ctx, d)
		if err != nil {
			return err
		}
	}

	log.Info().Int("accounts", len(acctList)).Int("identities", identityCount).
		Int("lockers", lockerCount).Int("properties", propertyCount).
		Int("access_keys", len(accessKeyList)).Int("dids", len(didList)).
		Msg("Imported backend data")

	return nil
}

func readConfigFile(configFilePath string) (*koanf.Koanf, error) {
	var cfg = koanf.New(".")

	configFilePath = utils.AbsPathify(configFilePath)

	// read configuration

	var err error
	if strings.HasSuffix(configFilePath, ".json") {
		err = cfg.Load(file.Provider(configFilePath), json.Parser())
	} else if strings.HasSuffix(configFilePath, ".yaml") {
		err = cfg.Load(file.Provider(configFilePath), yaml.Parser())
	} else {
		return nil, errors.New("only config files in JSON and YAML format are supported")
	}

	return cfg, err
}

func initBackendInstance(cfg *koanf.Koanf, resolver cmdbase.ParameterResolver) (storage.IdentityBackend, error) {
	if cfg.Exists("accountStore") {
		var backendCfg *storage.IdentityBackendConfig

		err := cfg.Unmarshal("accountStore", &backendCfg)
		if err != nil {
			log.Err(err).Msg("Failed to read account storage configuration")
			return nil, cli.Exit(err, 1)
		}

		identityBackend, err := storage.CreateIdentityBackend(backendCfg, resolver)
		if err != nil {
			log.Err(err).Msg("Failed to create storage backend")
			return nil, cli.Exit(err, 1)
		}

		return identityBackend, nil
	} else {
		return nil, cli.Exit("account store not defined", 1)
	}
}
