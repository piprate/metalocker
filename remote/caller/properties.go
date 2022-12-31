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
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/storage"
)

func (c *MetaLockerHTTPCaller) StoreProperty(prop *account.DataEnvelope) error {
	return c.storeDataEnvelope("property", prop)
}

func (c *MetaLockerHTTPCaller) GetProperty(hash string) (*account.DataEnvelope, error) {
	return c.getDataEnvelope("property", hash, storage.ErrPropertyNotFound)
}

func (c *MetaLockerHTTPCaller) ListProperties() ([]*account.DataEnvelope, error) {
	return c.listDataEnvelopes("property")
}

func (c *MetaLockerHTTPCaller) DeleteProperty(hash string) error {
	return c.deleteDataEnvelope("property", hash, storage.ErrPropertyNotFound)
}
