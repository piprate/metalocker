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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/apibase"
)

func (h *LedgerHandler) GetLedgerRecordHandler(c *gin.Context) {
	rid := c.Params.ByName("id")

	rec, err := h.ledger.GetRecord(rid)
	if err != nil {
		if errors.Is(err, model.ErrRecordNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		} else {
			log := apibase.CtxLogger(c)
			log.Err(err).Msg("Error when retrieving ledger record")
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	apibase.JSON(c, http.StatusOK, rec)
}

func (h *LedgerHandler) PostLedgerRecordHandler(c *gin.Context) {
	log := apibase.CtxLogger(c)

	var r model.Record
	err := apibase.BindJSON(c, &r)
	if err != nil {
		apibase.AbortWithError(c, http.StatusBadRequest, "Bad request body")
		return
	}

	err = h.ledger.SubmitRecord(&r)
	if err != nil {
		log.Err(err).Msg("Error when submitting ledger record")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	log.Debug().Str("id", r.ID).Interface("rec", r).Msg("Ledger record submitted")

	var result struct {
		ID string `json:"id"`
	}

	result.ID = r.ID

	apibase.JSON(c, http.StatusOK, &result)
}

func (h *LedgerHandler) GetLedgerRecordStateHandler(c *gin.Context) {
	rid := c.Params.ByName("id")

	rs, err := h.ledger.GetRecordState(rid)
	if err != nil {
		log := apibase.CtxLogger(c)
		log.Err(err).Msg("Error when getting ledger record state")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if rs == nil {
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		apibase.JSON(c, http.StatusOK, rs)
	}
}

func (h *LedgerHandler) GetAssetHeadHandler(c *gin.Context) {
	rid := c.Params.ByName("id")

	head, err := h.ledger.GetAssetHead(rid)
	if err != nil {
		if errors.Is(err, model.ErrAssetHeadNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		} else {
			log := apibase.CtxLogger(c)
			log.Err(err).Msg("Error when retrieving asset head")
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	apibase.JSON(c, http.StatusOK, head)
}
