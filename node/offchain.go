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

package node

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/vaults"
)

type OffChainStorageProxy struct {
	offChainVault vaults.Vault
}

var _ model.OffChainStorage = (*OffChainStorageProxy)(nil)

func NewOffChainStorageProxy(v vaults.Vault) *OffChainStorageProxy {
	return &OffChainStorageProxy{
		offChainVault: v,
	}
}

func (p *OffChainStorageProxy) GetOperation(ctx context.Context, opAddr string) ([]byte, error) {
	rdr, err := p.offChainVault.ServeBlob(ctx, opAddr, nil, "")
	if err != nil {
		if errors.Is(err, model.ErrBlobNotFound) {
			return nil, model.ErrOperationNotFound
		}
		return nil, err
	}
	defer rdr.Close()

	return io.ReadAll(rdr)
}

func (p *OffChainStorageProxy) SendOperation(ctx context.Context, opData []byte) (string, error) {
	res, err := p.offChainVault.CreateBlob(ctx, bytes.NewReader(opData))
	if err != nil {
		return "", err
	}
	return res.ID, nil
}

func (p *OffChainStorageProxy) PurgeOperation(ctx context.Context, opAddr string) error {
	return p.offChainVault.PurgeBlob(ctx, opAddr, nil)
}
