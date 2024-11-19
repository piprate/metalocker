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
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/vaults"
)

type (
	LedgerHandler struct {
		ledger        model.Ledger
		offChainVault vaults.Vault
	}
)

func InitLedgerRoutes(rg *gin.RouterGroup, ledger model.Ledger, offChainVault vaults.Vault) {

	h := &LedgerHandler{
		ledger:        ledger,
		offChainVault: offChainVault,
	}

	rg.POST("/lop", h.PostLedgerOperationHandler)
	rg.GET("/lop/:id", h.GetLedgerOperationHandler)
	rg.POST("/lop/:id/purge", h.PostPurgeLedgerOperationHandler)

	rg.POST("/lrec", h.PostLedgerRecordHandler)
	rg.GET("/lrec/:id", h.GetLedgerRecordHandler)
	rg.GET("/lrec/:id/state", h.GetLedgerRecordStateHandler)

	rg.GET("/head/:id", h.GetAssetHeadHandler)

	rg.GET("/ledger/genesis", h.GetLedgerGenesisHandler)
	rg.GET("/ledger/top", h.GetLedgerTopHandler)
	rg.GET("/ledger/block/:number", h.GetLedgerBlockHandler)
	rg.GET("/ledger/block/:number/records", h.GetLedgerBlockRecordsHandler)
	rg.GET("/ledger/chain/:start/:depth", h.GetLedgerChainHandler)
	rg.GET("/ledger/data-asset/:id/state", h.GetDataAssetStateHandler)
}

func (h *LedgerHandler) GetLedgerGenesisHandler(c *gin.Context) {
	log := apibase.CtxLogger(c)

	b, err := h.ledger.GetGenesisBlock(c)
	if err != nil {
		log.Err(err).Msg("Error when getting genesis block state")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	apibase.JSON(c, http.StatusOK, b)
}

func (h *LedgerHandler) GetLedgerTopHandler(c *gin.Context) {
	log := apibase.CtxLogger(c)

	b, err := h.ledger.GetTopBlock(c)
	if err != nil {
		log.Err(err).Msg("Error when getting top block state")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	apibase.JSON(c, http.StatusOK, b)
}

func (h *LedgerHandler) GetLedgerChainHandler(c *gin.Context) {
	log := apibase.CtxLogger(c)

	startBlockNumberStr := c.Params.ByName("start")
	depthStr := c.Params.ByName("depth")

	if startBlockNumberStr == "" {
		log.Error().Msg("Start block number is empty")
		_ = c.AbortWithError(http.StatusBadRequest, errors.New("start block number is empty"))
		return
	}

	if depthStr == "" {
		log.Error().Msg("Depth is not specified")
		_ = c.AbortWithError(http.StatusBadRequest, errors.New("depth is not specified"))
		return
	}

	startBlockNumber, err := strconv.ParseInt(startBlockNumberStr, 10, 0)
	if err != nil {
		log.Err(err).Str("number", startBlockNumberStr).Msg("Error when parsing start block number")
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	depth, err := strconv.Atoi(depthStr)
	if err != nil {
		log.Err(err).Msg("Error when getting block chain")
		_ = c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid chain depth: %s", depthStr))
		return
	}

	chain, err := h.ledger.GetChain(c, startBlockNumber, depth)
	if err != nil {
		log.Err(err).Msg("Error when getting block chain")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if chain == nil {
		c.String(http.StatusOK, "[]")
	} else {
		apibase.JSON(c, http.StatusOK, chain)
	}
}

func (h *LedgerHandler) GetLedgerBlockHandler(c *gin.Context) {
	log := apibase.CtxLogger(c)

	blockNumberStr := c.Params.ByName("number")

	if blockNumberStr == "" {
		log.Error().Msg("Block number is empty")
		_ = c.AbortWithError(http.StatusBadRequest, errors.New("block number is empty"))
		return
	}

	blockNumber, err := strconv.ParseInt(blockNumberStr, 10, 0)
	if err != nil {
		log.Err(err).Str("number", blockNumberStr).Msg("Error when parsing block number")
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	b, err := h.ledger.GetBlock(c, blockNumber)
	if err != nil {
		log.Err(err).Msg("Error when reading a block")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	apibase.JSON(c, http.StatusOK, b)
}

func (h *LedgerHandler) GetLedgerBlockRecordsHandler(c *gin.Context) {
	log := apibase.CtxLogger(c)

	blockNumberStr := c.Params.ByName("number")
	if blockNumberStr == "" {
		log.Error().Msg("Block number is empty")
		_ = c.AbortWithError(http.StatusBadRequest, errors.New("block number is empty"))
		return
	}

	blockNumber, err := strconv.ParseInt(blockNumberStr, 10, 0)
	if err != nil {
		log.Err(err).Str("number", blockNumberStr).Msg("Error when parsing block number")
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	recs, err := h.ledger.GetBlockRecords(c, blockNumber)
	if err != nil {
		log.Err(err).Msg("Error when reading block records")
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if recs == nil {
		c.String(http.StatusOK, "")
	} else {
		buf := bytes.NewBuffer(nil)
		w := csv.NewWriter(buf)
		for _, rec := range recs {
			_ = w.Write(rec)
		}
		w.Flush()

		c.Data(http.StatusOK, "text/csv", buf.Bytes())
	}
}
