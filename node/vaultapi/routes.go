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

package vaultapi

import (
	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/vaults"
)

func InitRoutes(vaultGrp *gin.RouterGroup, lbm *vaults.LocalBlobManager) {
	vaultMap, _ := lbm.GetVaultMap()
	for _, prop := range vaultMap {
		vault, err := lbm.GetVault(prop.ID)
		if err != nil {
			panic(err)
		}

		v := vaultGrp.Group(vault.ID())

		v.POST("/raw", PostStoreRaw(vault))
		v.POST("/encrypt", PostStoreEncrypt(vault))
		v.POST("/serve", PostServeBlobHandler(vault))
		v.POST("/purge", PostPurge(vault))
	}

	vaultGrp.GET("/list", GetVaultListHandler(vaultMap))
}
