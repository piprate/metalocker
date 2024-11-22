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
	"context"
	"errors"

	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/storage"
)

func (c *MetaLockerHTTPCaller) StoreLocker(ctx context.Context, locker *account.DataEnvelope) error {
	return c.storeDataEnvelope(ctx, "locker", locker)
}

func (c *MetaLockerHTTPCaller) GetLocker(ctx context.Context, hash string) (*account.DataEnvelope, error) {
	return c.getDataEnvelope(ctx, "locker", hash, storage.ErrLockerNotFound)
}

func (c *MetaLockerHTTPCaller) ListLockers(ctx context.Context) ([]*account.DataEnvelope, error) {
	return c.listDataEnvelopes(ctx, "locker")
}

func (c *MetaLockerHTTPCaller) ListLockerHashes(ctx context.Context) ([]string, error) {
	return nil, errors.New("operation ListLockerHashes not implemented")
}
