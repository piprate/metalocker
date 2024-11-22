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

package memory

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/cmdbase"
	"github.com/piprate/metalocker/utils/fingerprint"
	"github.com/piprate/metalocker/vaults"
	"github.com/rs/zerolog/log"
)

const VaultType = "memory"

func init() {
	vaults.Register(VaultType, CreateVault)
}

// InMemoryVault keeps all submitted data in memory. It doesn't survive restarts.
// This vault type is useful for testing to avoid disk or network operations
// SSE mode is not supported. The value of SSE parameter will be ignored.
type InMemoryVault struct {
	id       string
	name     string
	sse      bool
	cas      bool
	verifier model.AccessVerifier

	blobMtx sync.RWMutex
	blobs   map[string][]byte
}

func (v *InMemoryVault) CAS() bool {
	return v.cas
}

func (v *InMemoryVault) SSE() bool {
	return v.sse
}

func (v *InMemoryVault) CreateBlob(ctx context.Context, r io.Reader) (*model.StoredResource, error) {

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var id string
	if v.cas {
		id, err = model.BuildDigitalAssetID(data, fingerprint.AlgoSha256, "")
		if err != nil {
			return nil, err
		}
	} else {
		id = model.NewAssetID("")
	}

	res := &model.StoredResource{
		ID:     id,
		Type:   model.TypeResource,
		Vault:  v.id,
		Method: VaultType,
	}

	v.blobMtx.Lock()
	v.blobs[id] = data
	v.blobMtx.Unlock()

	return res, nil
}

func (v *InMemoryVault) PurgeBlob(ctx context.Context, id string, params map[string]any) error {
	if v.verifier != nil {
		state, err := v.verifier.GetDataAssetState(ctx, id)
		if err != nil {
			return err
		}

		switch state {
		case model.DataAssetStateKeep:
			return errors.New("data asset in use and can't be purged")
		case model.DataAssetStateNotFound:
			return model.ErrBlobNotFound
		case model.DataAssetStateRemove:
			// all fine
		}
	}

	if _, found := v.blobs[id]; !found {
		return model.ErrBlobNotFound
	}

	v.blobMtx.Lock()
	delete(v.blobs, id)
	v.blobMtx.Unlock()

	return nil
}

func (v *InMemoryVault) ServeBlob(ctx context.Context, id string, params map[string]any, accessToken string) (io.ReadCloser, error) {
	if v.verifier != nil {
		if !model.VerifyAccessToken(ctx, accessToken, id, time.Now().Unix(), model.DefaultMaxDistanceSeconds, v.verifier) {
			return nil, model.ErrDataAssetAccessDenied
		}
	}

	v.blobMtx.Lock()
	val, found := v.blobs[id]
	v.blobMtx.Unlock()
	if !found {
		return nil, model.ErrBlobNotFound
	}

	r := io.NopCloser(bytes.NewReader(val))

	return r, nil
}

func (v *InMemoryVault) ID() string {
	return v.id
}

func (v *InMemoryVault) Name() string {
	return v.name
}

func (v *InMemoryVault) Close() error {
	log.Info().Msg("Closing in-memory vault")
	return nil
}

func CreateVault(cfg *vaults.Config, resolver cmdbase.ParameterResolver, verifier model.AccessVerifier) (vaults.Vault, error) {
	log.Info().Msg("Initialising in-memory vault")

	return &InMemoryVault{
		id:       cfg.ID,
		name:     cfg.Name,
		blobs:    make(map[string][]byte),
		sse:      cfg.SSE,
		cas:      cfg.CAS,
		verifier: verifier,
	}, nil
}
