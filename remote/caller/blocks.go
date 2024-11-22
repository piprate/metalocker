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

package caller

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"

	"github.com/piprate/metalocker/model"
)

func (c *MetaLockerHTTPCaller) GetGenesisBlock(ctx context.Context) (*model.Block, error) {
	var b model.Block
	err := c.client.LoadContents(ctx, http.MethodGet, "/v1/ledger/genesis", nil, &b)
	if err != nil {
		return nil, err
	} else {
		return &b, nil
	}
}

func (c *MetaLockerHTTPCaller) GetTopBlock(ctx context.Context) (*model.Block, error) {
	var b model.Block
	err := c.client.LoadContents(ctx, http.MethodGet, "/v1/ledger/top", nil, &b)
	if err != nil {
		return nil, err
	} else {
		return &b, nil
	}
}

func (c *MetaLockerHTTPCaller) GetBlock(ctx context.Context, bn int64) (*model.Block, error) {
	var b model.Block
	err := c.client.LoadContents(ctx, http.MethodGet, fmt.Sprintf("/v1/ledger/block/%d", bn), nil, &b)
	if err != nil {
		return nil, err
	} else {
		return &b, nil
	}
}

func (c *MetaLockerHTTPCaller) GetChain(ctx context.Context, startNumber int64, depth int) ([]*model.Block, error) {
	var blocks []*model.Block
	err := c.client.LoadContents(ctx, http.MethodGet, fmt.Sprintf("/v1/ledger/chain/%d/%d", startNumber, depth), nil, &blocks)
	if err != nil {
		return nil, err
	} else {
		return blocks, nil
	}
}

func (c *MetaLockerHTTPCaller) GetBlockRecords(ctx context.Context, bn int64) ([][]string, error) {
	var recBytes []byte
	err := c.client.LoadContents(ctx, http.MethodGet, fmt.Sprintf("/v1/ledger/block/%d/records", bn), nil, &recBytes)
	if err != nil {
		return nil, err
	} else {
		return csv.NewReader(bytes.NewReader(recBytes)).ReadAll()
	}
}
