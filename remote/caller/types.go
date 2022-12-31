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

import "github.com/piprate/metalocker/model/account"

type (
	LoginForm struct {
		Username          string `json:"username"`
		Password          string `json:"password"`
		Audience          string `json:"audience"`
		AudiencePublicKey string `json:"audienceKey"`
	}

	AccountPatch struct {
		Email                string `json:"email,omitempty"`
		OldEncryptedPassword string `json:"oldEncryptedPassword,omitempty"`
		NewEncryptedPassword string `json:"newEncryptedPassword,omitempty"`
		Name                 string `json:"name,omitempty"`
		GivenName            string `json:"givenName,omitempty"`
		FamilyName           string `json:"familyName,omitempty"`
	}

	GetRecoveryCodeResponse struct {
		Code string `json:"code"`
	}

	AccountRecoveryResponse struct {
		Account *account.Account `json:"account"`
	}

	ServerControls struct {
		Status           string `json:"status"`
		MaintenanceMode  bool   `json:"maintenanceMode"`
		JWTPublicKey     string `json:"jwtPublicKey"`
		GenesisBlockHash string `json:"genesis"`
		TopBlock         int64  `json:"top"`
	}
)
