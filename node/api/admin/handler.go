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

package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/storage"
)

type (
	Handler struct {
		identityBackend storage.IdentityBackend
	}
)

func InitRoutes(r *gin.Engine, path string, adminAuthFunc gin.HandlerFunc, identityBackend storage.IdentityBackend) {
	h := &Handler{
		identityBackend: identityBackend,
	}
	adm := r.Group(path)
	adm.Use(adminAuthFunc)
	adm.Use(apibase.ContextLoggerHandler)
	{
		adm.GET("/account", h.GetAccountListHandler)
		adm.GET("/account/:id", h.GetAccountHandler)
		adm.POST("/account", h.PostAccountHandler)
		adm.PATCH("/account/:id", h.PatchAccountHandler)
		adm.GET("/did", h.GetIdentityListHandler)
		adm.POST("/did", h.PostIdentityHandler)
	}
}
