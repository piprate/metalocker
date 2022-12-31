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

package storage

import (
	"fmt"

	"github.com/piprate/metalocker/sdk/cmdbase"
	"github.com/rs/zerolog/log"
)

type Parameters map[string]any

type IdentityBackendConfig struct {
	Type   string     `json:"type"`
	Params Parameters `json:"params"`
}

type IdentityBackendConstructor func(def Parameters, resolver cmdbase.ParameterResolver) (IdentityBackend, error)

var identityBackendConstructors = make(map[string]IdentityBackendConstructor)

func Register(storageType string, ctor IdentityBackendConstructor) {
	if _, ok := identityBackendConstructors[storageType]; ok {
		panic("Identity backend constructor already registered for type: " + storageType)
	}

	identityBackendConstructors[storageType] = ctor
}

func CreateIdentityBackend(cfg *IdentityBackendConfig, resolver cmdbase.ParameterResolver) (IdentityBackend, error) {

	log.Info().Str("type", cfg.Type).Msg("Creating identity backend")

	ctor, ok := identityBackendConstructors[cfg.Type]
	if !ok {
		return nil, fmt.Errorf("identity backend %s not known or loaded", cfg.Type)
	}

	return ctor(cfg.Params, resolver)
}
