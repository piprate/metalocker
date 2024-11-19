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
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/storage"
)

type (
	DIDHandler struct {
		didStorage storage.DIDBackend
	}
)

func InitDIDRoutes(rg *gin.RouterGroup, didStorage storage.DIDBackend) {

	h := &DIDHandler{
		didStorage: didStorage,
	}

	rg.POST("/did", h.PostDIDDocumentHandler)
	rg.GET("/did/:id", h.GetDIDDocumentHandler)
}

func (h *DIDHandler) GetDIDDocumentHandler(c *gin.Context) {
	iid := c.Params.ByName("id")

	ddoc, err := h.didStorage.GetDIDDocument(c, iid)
	if err != nil {
		if errors.Is(err, storage.ErrDIDNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		} else {
			log := apibase.CtxLogger(c)
			log.Err(err).Msg("Error when retrieving identity")
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	apibase.JSON(c, http.StatusOK, ddoc)
}

func (h *DIDHandler) PostDIDDocumentHandler(c *gin.Context) {
	log := apibase.CtxLogger(c)

	var ddoc model.DIDDocument
	err := apibase.BindJSON(c, &ddoc)
	if err != nil {
		apibase.AbortWithError(c, http.StatusBadRequest, "Bad request body")
		return
	}

	err = h.didStorage.CreateDIDDocument(c, &ddoc)
	if err != nil {
		log.Err(err).Msg("Error when saving identity")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Writer.Header().Add("Location", fmt.Sprintf("%s/%s", c.Request.URL.RequestURI(), ddoc.ID))
	c.Status(http.StatusCreated)
}
