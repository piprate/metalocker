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
	"context"
	"fmt"
	"io"

	"github.com/piprate/metalocker/model"
)

type LocalBlobManager struct {
	vaultMap map[string]Vault
	propMap  map[string]*model.VaultProperties
}

var _ model.BlobManager = (*LocalBlobManager)(nil)

func NewLocalBlobManager() *LocalBlobManager {
	return &LocalBlobManager{
		vaultMap: make(map[string]Vault),
		propMap:  make(map[string]*model.VaultProperties),
	}
}

func (lbm *LocalBlobManager) AddVault(v Vault, cfg *Config) {
	lbm.vaultMap[v.ID()] = v
	lbm.propMap[cfg.Name] = &model.VaultProperties{
		ID:   cfg.ID,
		Name: cfg.Name,
		Type: cfg.Type,
		SSE:  cfg.SSE,
		CAS:  cfg.CAS,
	}
}

func (lbm *LocalBlobManager) GetBlob(ctx context.Context, res *model.StoredResource, accessToken string) (io.ReadCloser, error) {
	v, found := lbm.vaultMap[res.Vault]
	if !found {
		return nil, fmt.Errorf("vault not found: %s", res.Vault)
	}

	return ReceiveBlob(res, accessToken, func(res *model.StoredResource, accessToken string) (io.ReadCloser, error) {
		return v.ServeBlob(ctx, res.StorageID(), res.Params, accessToken)
	})
}

func (lbm *LocalBlobManager) PurgeBlob(ctx context.Context, res *model.StoredResource) error {
	v, found := lbm.vaultMap[res.Vault]
	if !found {
		return fmt.Errorf("vault not found: %s", res.Vault)
	}

	return v.PurgeBlob(ctx, res.ID, res.Params)
}

func (lbm *LocalBlobManager) SendBlob(ctx context.Context, data io.Reader, cleartext bool, vaultName string) (*model.StoredResource, error) {
	props, found := lbm.propMap[vaultName]
	if !found {
		return nil, fmt.Errorf("vault not found: %s", vaultName)
	}

	v, found := lbm.vaultMap[props.ID]
	if !found {
		return nil, fmt.Errorf("vault not found: %s", props.ID)
	}

	return SendBlob(data, v.ID(), v.SSE() || cleartext, func(data io.Reader, vaultID string) (*model.StoredResource, error) {
		return v.CreateBlob(ctx, data)
	})
}

func (lbm *LocalBlobManager) GetVaultMap(ctx context.Context) (map[string]*model.VaultProperties, error) {
	return lbm.propMap, nil
}

func (lbm *LocalBlobManager) GetVault(id string) (Vault, error) {
	v, found := lbm.vaultMap[id]
	if !found {
		return nil, fmt.Errorf("vault not found: %s", id)
	}
	return v, nil
}
