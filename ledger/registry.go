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

package ledger

import (
	"context"
	"fmt"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/cmdbase"
	"github.com/piprate/metalocker/services/notification"
	"github.com/rs/zerolog/log"
)

type constructor func(context.Context, Parameters, notification.Service, cmdbase.ParameterResolver) (model.Ledger, error)

var ledgerConstructors = make(map[string]constructor)

func Register(ledgerType string, ctor constructor) {
	if _, ok := ledgerConstructors[ledgerType]; ok {
		panic("ledger constructor already registered for type: " + ledgerType)
	}

	ledgerConstructors[ledgerType] = ctor
}

func CreateLedgerConnector(ctx context.Context, cfg *Config, ns notification.Service, resolver cmdbase.ParameterResolver) (model.Ledger, error) {

	log.Info().Str("type", cfg.Type).Interface("params", cfg.Params).Msg("Creating ledger connector")

	ctor, ok := ledgerConstructors[cfg.Type]
	if !ok {
		return nil, fmt.Errorf("ledger connector %q not known or loaded", cfg.Type)
	}

	return ctor(ctx, cfg.Params, ns, resolver)
}
