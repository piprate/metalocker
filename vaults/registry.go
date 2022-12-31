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

package vaults

import (
	"fmt"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/cmdbase"
	"github.com/rs/zerolog/log"
)

type VaultConstructor func(cfg *Config, resolver cmdbase.ParameterResolver, verifier model.AccessVerifier) (Vault, error)

var vaultConstructors = make(map[string]VaultConstructor)

func Register(vaultType string, ctor VaultConstructor) {
	if _, ok := vaultConstructors[vaultType]; ok {
		panic("vault constructor already registered for type: " + vaultType)
	}

	vaultConstructors[vaultType] = ctor
}

func CreateVault(cfg *Config, resolver cmdbase.ParameterResolver, verifier model.AccessVerifier) (Vault, error) {

	log.Info().Str("id", cfg.ID).Str("name", cfg.Name).Str("type", cfg.Type).Msg("Creating vault")

	ctor, ok := vaultConstructors[cfg.Type]
	if !ok {
		return nil, fmt.Errorf("vault %q not known or loaded", cfg.Type)
	}

	return ctor(cfg, resolver, verifier)
}
