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

package wallet

import (
	"github.com/piprate/metalocker/model"
)

const (
	AccountUpdateType = "AccountUpdate"

	MetaLeaseDurationYears = 100
)

type (
	AccountUpdate struct {
		Type        string            `json:"type"`
		AccountID   string            `json:"a"`
		AccessLevel model.AccessLevel `json:"lvl"`

		IdentitiesAdded   []string `json:"ida,omitempty"`
		IdentitiesRemoved []string `json:"idr,omitempty"`

		LockersOpened []string `json:"lop,omitempty"`
		LockersClosed []string `json:"lcl,omitempty"`

		SubAccountsAdded   []string `json:"saa,omitempty"`
		SubAccountsRemoved []string `json:"sar,omitempty"`

		IndexesAdded   []string `json:"ixa,omitempty"`
		IndexesRemoved []string `json:"ixr,omitempty"`
	}
)
