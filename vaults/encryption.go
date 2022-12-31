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

package vaults

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"sync"

	"github.com/piprate/metalocker/model"
	fp "github.com/piprate/metalocker/utils/fingerprint"
	"github.com/piprate/metalocker/utils/streams"
	"github.com/rs/zerolog/log"
)

type blobSenderFn func(data io.Reader, vaultID string) (*model.StoredResource, error)

// SendBlob sends the given blob to a vault and takes care of building StoredResource and applying
// encryption where necessary.
func SendBlob(r io.Reader, vaultID string, cleartext bool, senderFn blobSenderFn) (*model.StoredResource, error) {

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
	if cleartext {
		var err error
		res, err = senderFn(pr, vaultID)
		if err != nil {
			return nil, err
		}
	} else {

		// encrypt the blob on the client side

		encKey := model.NewEncryptionKey()

		b, err := io.ReadAll(pr)
		if err != nil {
			return nil, err
		}

		encryptedData, err := model.EncryptAESCGM(b, encKey)
		if err != nil {
			return nil, err
		}

		res, err = senderFn(bytes.NewReader(encryptedData), vaultID)
		if err != nil {
			return nil, err
		}

		res.EncryptionKey = base64.StdEncoding.EncodeToString(encKey[:])
	}

	wg.Wait()

	if copyErr != nil {
		return nil, copyErr
	}

	stats := ssw.Stats()

	res.Asset = model.BuildDigitalAssetIDWithFingerprint(stats.SHA256Hash, "")
	res.MIMEType = stats.ContentType
	res.Size = stats.Size

	return res, nil
}

type blobReceiverFn func(res *model.StoredResource, accessToken string) (io.ReadCloser, error)

// ReceiveBlob returns a decrypted blob stream from the vault (either local or remote)
func ReceiveBlob(res *model.StoredResource, accessToken string, receiverFn blobReceiverFn) (io.ReadCloser, error) {
	if res.EncryptionKey == "" {
		// server side encryption
		return receiverFn(res, accessToken)
	} else {
		// client side encryption
		r, err := receiverFn(res, accessToken)
		if err != nil {
			return nil, err
		}
		defer r.Close()

		encryptedFileBytes, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		fileBytes, err := model.DecryptAESCGM(encryptedFileBytes, res.GetEncryptionKey())
		if err != nil {
			return nil, err
		}

		valid, err := model.VerifyDigitalAssetID(res.Asset, fp.AlgoSha256, fileBytes)
		if err != nil {
			return nil, err
		}

		if !valid {
			log.Warn().Str("id", res.Asset).Msg("Mismatch found between a blob and its asset ID")
			return nil, fmt.Errorf("mismatch found between a blob and its asset ID: %s", res.Asset)
		}

		return io.NopCloser(bytes.NewReader(fileBytes)), nil
	}
}
