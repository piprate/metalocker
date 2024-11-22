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

package api

import (
	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/apibase"

	"net/http"
)

type ServerControls struct {
	Status           string `json:"status"`
	MaintenanceMode  bool   `json:"maintenanceMode"`
	JWTPublicKey     string `json:"jwtPublicKey"`
	GenesisBlockHash string `json:"genesis"`
	TopBlock         int64  `json:"top"`
}

func GetStatusHandler(controls *ServerControls, ledger model.Ledger) func(c *gin.Context) {
	return func(c *gin.Context) {
		gb, err := ledger.GetGenesisBlock(c)
		if err != nil {
			log := apibase.CtxLogger(c)
			log.Err(err).Msg("Error when getting genesis block")
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		b, err := ledger.GetTopBlock(c)
		if err != nil {
			log := apibase.CtxLogger(c)
			log.Err(err).Msg("Error when getting top block")
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		var res = *controls
		res.GenesisBlockHash = gb.Hash
		res.TopBlock = b.Number

		apibase.JSON(c, http.StatusOK, res)
	}
}
