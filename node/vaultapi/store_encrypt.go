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
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/utils/measure"
	"github.com/piprate/metalocker/utils/streams"
	"github.com/piprate/metalocker/vaults"
)

func PostStoreEncrypt(vaultAPI vaults.Vault) func(c *gin.Context) {
	return func(c *gin.Context) {
		defer measure.ExecTime("api.PostStoreEncrypt")()

		log := apibase.CtxLogger(c)

		defer c.Request.Body.Close()

		log.Debug().Str("vault", vaultAPI.Name()).Msg("Encrypting blob")

		var r io.Reader = c.Request.Body

		pr, pw := io.Pipe()

		ssw := streams.NewStreamStatsWriter()

		mw := io.MultiWriter(ssw, pw)

		wg := &sync.WaitGroup{}
		wg.Add(1)

		var copyErr error
		go func() {

			defer pw.Close()

			if _, copyErr = io.Copy(mw, r); copyErr != nil {
				return
			}

			wg.Done()
		}()

		var res *model.StoredResource
		if !vaultAPI.SSE() {

			// encrypt the blob on the client side

			encKey := model.NewEncryptionKey()

			// read blob into memory to calculate asset ID
			data, err := io.ReadAll(pr)
			if err != nil {
				apibase.AbortWithInternalServerError(c, err)
				return
			}

			encryptedData, err := model.EncryptAESCGM(data, encKey)
			if err != nil {
				apibase.AbortWithInternalServerError(c, err)
				return
			}

			res, err = vaultAPI.CreateBlob(c, bytes.NewReader(encryptedData))
			if err != nil {
				apibase.AbortWithInternalServerError(c, err)
				return
			}

			res.EncryptionKey = base64.StdEncoding.EncodeToString(encKey[:])
		} else {
			var err error
			res, err = vaultAPI.CreateBlob(c, pr)
			if err != nil {
				apibase.AbortWithInternalServerError(c, err)
				return
			}
		}

		wg.Wait()

		if copyErr != nil {
			apibase.AbortWithInternalServerError(c, copyErr)
			return
		}

		stats := ssw.Stats()

		res.Asset = model.BuildDigitalAssetIDWithFingerprint(stats.SHA256Hash, "")
		res.MIMEType = stats.ContentType
		res.Size = stats.Size

		c.JSON(http.StatusOK, res)
	}
}
