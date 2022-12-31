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
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/apibase"
)

func (h *LedgerHandler) GetLedgerOperationHandler(c *gin.Context) {
	id := c.Params.ByName("id")

	opReader, err := h.offChainVault.ServeBlob(id, nil, "")
	if err != nil {
		if errors.Is(err, model.ErrBlobNotFound) {
			apibase.AbortWithError(c, http.StatusNotFound, "operation not found")
		} else {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	defer opReader.Close()

	c.Writer.Header().Set("Content-Type", "application/octet-stream")

	_, err = io.Copy(c.Writer, opReader)
	if err != nil {
		log := apibase.CtxLogger(c)
		log.Err(err).Str("id", id).Msg("Error serving operation stream")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
}

func (h *LedgerHandler) PostPurgeLedgerOperationHandler(c *gin.Context) {
	id := c.Params.ByName("id")

	err := h.offChainVault.PurgeBlob(id, nil)
	if err != nil {
		if errors.Is(err, model.ErrBlobNotFound) {
			c.Status(http.StatusNotFound)
		} else {
			log := apibase.CtxLogger(c)
			log.Err(err).Msg("Error purging operation")
			apibase.AbortWithInternalServerError(c, err)
		}
		return
	}

	c.Status(http.StatusOK)
}

func (h *LedgerHandler) PostLedgerOperationHandler(c *gin.Context) {
	defer c.Request.Body.Close()

	// persist operation
	res, err := h.offChainVault.CreateBlob(c.Request.Body)
	if err != nil {
		apibase.AbortWithInternalServerError(c, err)
		return
	}

	var result struct {
		ID string `json:"id"`
	}

	result.ID = res.ID

	apibase.JSON(c, http.StatusOK, &result)
}
