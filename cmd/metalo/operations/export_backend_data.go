package operations

import (
	"context"
	"os"
	"path/filepath"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/sdk/cmdbase"
	"github.com/piprate/metalocker/utils/jsonw"
)

var (
	allLevels = []model.AccessLevel{
		model.AccessLevelRestricted,
		model.AccessLevelManaged,
		model.AccessLevelHosted,
	}
)

//func InitAerospikeBackend(cfg *koanf.Koanf, resolver cmdbase.ParameterResolver) (*aerospike.Persister, error) {
//	if cfg.Exists("accountStore") {
//		var backendCfg storage.IdentityBackendConfig
//		err := cfg.Unmarshal("accountStore", &backendCfg)
//		if err != nil {
//			log.Err(err).Msg("Failed to read account storage configuration")
//			return nil, cli.Exit(err, 1)
//		}
//
//		identityBackend, err := aerospike.CreateIdentityBackend(backendCfg.Params, resolver)
//		if err != nil {
//			log.Err(err).Msg("Failed to create storage backend")
//			return nil, cli.Exit(err, 1)
//		}
//
//		return identityBackend.(*aerospike.Persister), nil
//	} else {
//		return nil, cli.Exit("account store not defined", 1)
//	}
//}

func ExportBackendData(ctx context.Context, exportFolderPath, configFilePath string) error {

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

	// export accounts

	accountList, err := backend.ListAccounts(ctx, "", "")
	if err != nil {
		return err
	}

	b, _ := jsonw.MarshalIndent(accountList, "", "  ")

	if err = os.WriteFile(filepath.Join(exportFolderPath, "accounts.json"), b, 0o600); err != nil {
		return err
	}

	iidList := map[string][]*account.DataEnvelope{}
	lockerList := map[string][]*account.DataEnvelope{}
	propertyList := map[string][]*account.DataEnvelope{}
	var accessKeyList []*model.AccessKey
	for _, acct := range accountList {
		for _, lvl := range allLevels {
			accountIdentityList, err := backend.ListIdentities(ctx, acct.ID, lvl)
			if err != nil {
				return err
			}
			if len(accountIdentityList) > 0 {
				iidList[acct.ID] = append(iidList[acct.ID], accountIdentityList...)
			}

			accountLockerList, err := backend.ListLockers(ctx, acct.ID, lvl)
			if err != nil {
				return err
			}
			if len(accountLockerList) > 0 {
				lockerList[acct.ID] = append(lockerList[acct.ID], accountLockerList...)
			}

			accountPropertyList, err := backend.ListProperties(ctx, acct.ID, lvl)
			if err != nil {
				return err
			}
			if len(accountPropertyList) > 0 {
				propertyList[acct.ID] = append(propertyList[acct.ID], accountPropertyList...)
			}
		}

		accountKeyList, err := backend.ListAccessKeys(ctx, acct.ID)
		if err != nil {
			return err
		}
		accessKeyList = append(accessKeyList, accountKeyList...)
	}

	b, _ = jsonw.MarshalIndent(iidList, "", "  ")
	if err = os.WriteFile(filepath.Join(exportFolderPath, "identities.json"), b, 0o600); err != nil {
		return err
	}

	b, _ = jsonw.MarshalIndent(lockerList, "", "  ")
	if err = os.WriteFile(filepath.Join(exportFolderPath, "lockers.json"), b, 0o600); err != nil {
		return err
	}

	b, _ = jsonw.MarshalIndent(accessKeyList, "", "  ")
	if err = os.WriteFile(filepath.Join(exportFolderPath, "access_keys.json"), b, 0o600); err != nil {
		return err
	}

	b, _ = jsonw.MarshalIndent(propertyList, "", "  ")
	if err = os.WriteFile(filepath.Join(exportFolderPath, "properties.json"), b, 0o600); err != nil {
		return err
	}

	// export DIDs

	didList, err := backend.ListDIDDocuments(ctx)
	if err != nil {
		return err
	}

	b, _ = jsonw.MarshalIndent(didList, "", "  ")

	if err = os.WriteFile(filepath.Join(exportFolderPath, "dids.json"), b, 0o600); err != nil {
		return err
	}

	return nil
}
