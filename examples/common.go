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

package examples

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/piprate/metalocker/index"
	"github.com/piprate/metalocker/index/bolt"
	"github.com/piprate/metalocker/remote"
	"github.com/piprate/metalocker/remote/caller"
)

const DemoIndexStoreID = "did:piprate:B7hqrwxrCnsWqtBnbYnhABwzktnT5oo1ZWbtby3My886"
const DemoIndexStoreName = "examples"

func WithDemoIndexStore(baseDir string) remote.IndexClientSourceFn {
	return func(ctx context.Context, userID string, mlc *caller.MetaLockerHTTPCaller) (index.Client, error) {
		gb, err := mlc.GetGenesisBlock(ctx)
		if err != nil {
			return nil, err
		}

		return index.NewLocalIndexClient(ctx, []*index.StoreConfig{
			{
				ID:   DemoIndexStoreID,
				Name: DemoIndexStoreName,
				Type: bolt.Type,
				Params: map[string]any{
					bolt.ParameterFilePath: filepath.Join(baseDir, fmt.Sprintf("examples_%s.bolt", userID)),
				},
			},
		}, nil, gb.Hash)
	}
}
