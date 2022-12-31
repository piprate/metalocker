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

package hv

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/bytedance/sonic"
	"github.com/hashicorp/vault/api"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog/log"
)

type HCVaultClient struct {
	c     *api.Client
	ttl   int
	renew int
	done  chan struct{}
	mode  string
}

func NewHCVaultClient(mode string) (*HCVaultClient, error) {

	if mode != "prod" && mode != "dev" {
		return nil, fmt.Errorf("unsupported Vault mode: %s. The client can accept either prod of dev", mode)
	}

	// MetaLocker expects to take the Vault's address and token from the environment
	for _, envVar := range []string{api.EnvVaultAddress, api.EnvVaultToken} {
		if val := os.Getenv(envVar); val == "" {
			return nil, fmt.Errorf("environment variable not found: %s", envVar)
		}
	}

	c, err := api.NewClient(nil)
	if err != nil {
		return nil, err
	}

	hcv := &HCVaultClient{
		c:    c,
		mode: mode,
	}

	// enable token renewal, if it's renewable

	secret, err := c.Auth().Token().LookupSelf()
	if err != nil {
		return nil, err
	}

	isRenewable, err := secret.TokenIsRenewable()
	if err != nil {
		return nil, err
	}

	if isRenewable {
		creationTTL, err := secret.Data["creation_ttl"].(json.Number).Int64()
		if err != nil {
			return nil, err
		}

		ttl, err := secret.Data["ttl"].(json.Number).Int64()
		if err != nil {
			return nil, err
		}

		hcv.ttl = int(creationTTL)
		hcv.renew = int(ttl) / 2
		hcv.start()
	}

	return hcv, nil
}

// starts the renewal loop.
func (hvc *HCVaultClient) start() {
	if hvc.renew == 0 || hvc.ttl == 0 {
		log.Info().Msg("Vault: token renewal disabled")
		return
	}
	if hvc.done != nil {
		close(hvc.done)
	}
	log.Info().Int("ttl", hvc.ttl).Int("next_renewal", hvc.renew).Msg("Vault: token renewal enabled")
	hvc.done = make(chan struct{})
	if hvc.renew != 0 {
		go hvc.renewLoop()
	}
}

// stops the renewal loop.
func (hvc *HCVaultClient) stop() {
	close(hvc.done)
}

func (hvc *HCVaultClient) renewLoop() {
	for {
		select {
		case <-time.After(time.Duration(hvc.renew) * time.Second):
			incr := hvc.ttl
			hvc.renew = incr / 2

			log.Debug().Int("increment", hvc.ttl).Msg("Vault: refreshing token")
			_, err := hvc.c.Auth().Token().RenewSelf(incr)
			if err != nil {
				log.Err(err).Msg("Vault: refreshing token failed")
			} else {
				log.Debug().Int("increment", incr).Msg("Vault: refreshing token succeeded")
			}
		case <-hvc.done:
			return
		}
	}
}

func (hvc *HCVaultClient) Close() error {
	hvc.stop()
	return nil
}

func (hvc *HCVaultClient) ReadKey(keyPath string, key string) (string, error) {
	s, err := hvc.c.Logical().Read(keyPath)
	if err != nil {
		return "", err
	}

	if s == nil {
		return "", fmt.Errorf("path not found in Vault: %s. Can't start proxy ledger connector", keyPath)
	}

	if hvc.mode == "prod" {
		val, _ := s.Data[key].(string)
		return val, nil
	} else {
		// for some reason Vault in dev mode returns different key paths
		// and requires to use /secret/data instead of /secret

		// load AST
		b, _ := jsonw.Marshal(s.Data)
		vals, err := sonic.Get(b)
		if err != nil {
			return "", err
		}

		return vals.GetByPath("data", key).String()
	}
}
