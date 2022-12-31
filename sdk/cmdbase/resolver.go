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

	"github.com/piprate/metalocker/utils/hv"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	EPTypeViper = "ViperParam"
	EPTypeVault = "VaultParam"
)

type ParameterResolver interface {
	ResolveString(param any) (string, error)
}

type ExternalParameter struct {
	Type  string `json:"type"`
	Viper string `json:"viper"`
	Path  string `json:"path"`
	Key   string `json:"key"`
}

type SecureParameterResolver struct {
	sideVipers map[string]*viper.Viper
	hcv        *hv.HCVaultClient
}

func NewSecureParameterResolver(hcv *hv.HCVaultClient, sideVipers map[string]*viper.Viper) *SecureParameterResolver {
	if sideVipers == nil {
		sideVipers = make(map[string]*viper.Viper)
	}

	return &SecureParameterResolver{
		sideVipers: sideVipers,
		hcv:        hcv,
	}
}

func (spr *SecureParameterResolver) Close() error {
	if spr.hcv != nil {
		return spr.hcv.Close()
	}
	return nil
}

func (spr *SecureParameterResolver) AddSideViper(name string, v *viper.Viper) {
	spr.sideVipers[name] = v
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
		case EPTypeViper:
			viperName, found := val["viper"]
			if !found {
				return "", fmt.Errorf("viper name not found")
			}
			v, found := spr.sideVipers[viperName.(string)]
			if !found {
				return "", fmt.Errorf("unknown side viper file: '%s'", viperName)
			}
			key, found := val["key"]
			if !found {
				return "", fmt.Errorf("key name not found")
			}
			return v.GetString(key.(string)), nil
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

func ConfigureParameterResolver(viperCfg *viper.Viper, secretsPath string) (*SecureParameterResolver, error) {

	// initialise HashiCorp Vault client

	var hcv *hv.HCVaultClient
	var err error

	vaultMode := viperCfg.GetString("vaultMode")

	if vaultMode != "" {
		hcv, err = hv.NewHCVaultClient(vaultMode)
		if err != nil {
			log.Err(err).Msg("Failed to create HashiCorp Vault client")
			return nil, err
		}
	}

	// load secrets from local file, if provided

	sideVipers := make(map[string]*viper.Viper)

	if secretsConfigName := viperCfg.GetString("secretsConfig"); secretsConfigName != "" {
		secretsViper := viper.New()
		secretsViper.SetConfigName(secretsConfigName)
		secretsViper.AddConfigPath(secretsPath)
		err = secretsViper.ReadInConfig()
		if err != nil {
			panic(fmt.Errorf("fatal error when reading secrets config file: %w", err))
		}

		sideVipers["secrets"] = secretsViper
	}

	return NewSecureParameterResolver(hcv, sideVipers), nil
}
