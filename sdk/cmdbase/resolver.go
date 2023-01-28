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

package cmdbase

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/piprate/metalocker/utils/hv"
	"github.com/rs/zerolog/log"
)

const (
	EPTypeSideConfig = "SideConfigParam"
	EPTypeVault      = "VaultParam"
)

type ParameterResolver interface {
	ResolveString(param any) (string, error)
}

type SecureParameterResolver struct {
	sideConfigs map[string]*koanf.Koanf
	hcv         *hv.HCVaultClient
}

func NewSecureParameterResolver(hcv *hv.HCVaultClient, sideConfigs map[string]*koanf.Koanf) *SecureParameterResolver {
	if sideConfigs == nil {
		sideConfigs = make(map[string]*koanf.Koanf)
	}

	return &SecureParameterResolver{
		sideConfigs: sideConfigs,
		hcv:         hcv,
	}
}

func (spr *SecureParameterResolver) Close() error {
	if spr.hcv != nil {
		return spr.hcv.Close()
	}
	return nil
}

func (spr *SecureParameterResolver) AddSideConfig(name string, cfg *koanf.Koanf) {
	spr.sideConfigs[name] = cfg
}

func (spr *SecureParameterResolver) ResolveString(param any) (string, error) {
	if param == nil {
		return "", nil
	}
	switch val := param.(type) {
	case string:
		return val, nil
	case map[string]any:
		switch val["type"] {
		case EPTypeSideConfig:
			cfgName, found := val["cfg"]
			if !found {
				return "", fmt.Errorf("side config name not found")
			}
			cfg, found := spr.sideConfigs[cfgName.(string)]
			if !found {
				return "", fmt.Errorf("unknown side config file: '%s'", cfgName)
			}
			key, found := val["key"]
			if !found {
				return "", fmt.Errorf("key name not found")
			}
			return cfg.String(key.(string)), nil
		case EPTypeVault:
			if spr.hcv == nil {
				return "", errors.New("no HashiCorp Vault not provided")
			}

			vaultPath, found := val["path"]
			if !found {
				return "", fmt.Errorf("vault key path not found")
			}
			key, found := val["key"]
			if !found {
				return "", fmt.Errorf("vault key name not found")
			}
			v, err := spr.hcv.ReadKey(vaultPath.(string), key.(string))
			if err != nil {
				return "", err
			}
			return v, nil
		default:
			return "", fmt.Errorf("unknown parameter type: '%s'", val["type"])
		}
	default:
		return "", fmt.Errorf("unsupported parameter shape: '%s'", val)
	}
}

func ConfigureParameterResolver(cfg *koanf.Koanf, secretsPath string) (*SecureParameterResolver, error) {

	// initialise HashiCorp Vault client

	var hcv *hv.HCVaultClient
	var err error

	vaultMode := cfg.String("vaultMode")

	if vaultMode != "" {
		hcv, err = hv.NewHCVaultClient(vaultMode)
		if err != nil {
			log.Err(err).Msg("Failed to create HashiCorp Vault client")
			return nil, err
		}
	}

	// load secrets from local file, if provided

	sideConfigs := make(map[string]*koanf.Koanf)

	if secretsConfigName := cfg.String("secretsConfig"); secretsConfigName != "" {
		sideCfg := koanf.New(".")
		err = sideCfg.Load(
			file.Provider(
				filepath.Join(secretsPath, fmt.Sprintf("%s.yaml", secretsConfigName)),
			),
			yaml.Parser(),
		)
		if err != nil {
			panic(fmt.Errorf("fatal error when reading secrets config file: %w", err))
		}

		sideConfigs["secrets"] = sideCfg
	}

	return NewSecureParameterResolver(hcv, sideConfigs), nil
}
