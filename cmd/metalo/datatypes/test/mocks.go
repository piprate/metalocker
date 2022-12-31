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

package test

import (
	"io"
	"time"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/dataset"
	"github.com/piprate/metalocker/utils/fingerprint"
	"github.com/piprate/metalocker/utils/jsonw"
)

type MockLeaseBuilder struct {
	ImpressionMetaResource *model.MetaResource
	Resources              map[string][]byte
	Heads                  []string
}

func (mlb *MockLeaseBuilder) ImportResource(res *model.StoredResource) error {
	return nil
}

func (mlb *MockLeaseBuilder) SetHeads(headName ...string) error {
	mlb.Heads = headName
	return nil
}

func (mlb *MockLeaseBuilder) Submit(expiryTime time.Time) dataset.RecordFuture {
	panic("not supported")
}

func (mlb *MockLeaseBuilder) Store() (string, error) {
	panic("not supported")
}

func (mlb *MockLeaseBuilder) OverrideImpression(imp *model.Impression) {
}

func (mlb *MockLeaseBuilder) SetImpressionMetaResource(res *model.MetaResource) error {
	mlb.ImpressionMetaResource = res
	return nil
}

func (mlb *MockLeaseBuilder) AddResource(r io.Reader, opts ...dataset.BuilderOption) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	assetID, err := model.BuildDigitalAssetID(data, fingerprint.AlgoSha256, "")
	if err != nil {
		return "", err
	}

	mlb.Resources[assetID] = data
	return assetID, nil
}

func (mlb *MockLeaseBuilder) AddMetaResource(meta any, opts ...dataset.BuilderOption) (string, error) {
	var data []byte
	var err error
	var ok bool

	// marshal, if not []byte
	if data, ok = meta.([]byte); !ok {
		data, err = jsonw.Marshal(meta)
		if err != nil {
			return "", err
		}
	}

	assetID, err := model.BuildDigitalAssetID(data, fingerprint.AlgoSha256, "")
	if err != nil {
		return "", err
	}

	mlb.Resources[assetID] = data
	return assetID, nil
}

func (mlb *MockLeaseBuilder) AddProvenance(id string, provenance any, override bool) error {
	return nil
}

func (mlb *MockLeaseBuilder) CreatorID() string {
	return "unknown creator"
}

func (mlb *MockLeaseBuilder) Cancel() error {
	panic("not supported")
}
