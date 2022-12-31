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

package fs

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/cmdbase"
	"github.com/piprate/metalocker/utils"
	"github.com/piprate/metalocker/utils/fingerprint"
	"github.com/piprate/metalocker/vaults"
	"github.com/rs/zerolog/log"
)

const VaultType = "fs"

func init() {
	vaults.Register(VaultType, CreateVault)
}

type FileSystemVault struct {
	id       string
	name     string
	root     string
	prefix   string
	sse      bool
	cas      bool
	verifier model.AccessVerifier
}

func (v *FileSystemVault) CAS() bool {
	return v.cas
}

func (v *FileSystemVault) SSE() bool {
	return v.sse
}

func (v *FileSystemVault) CreateBlob(r io.Reader) (*model.StoredResource, error) {

	var id string
	var fileName string
	var err error
	if v.cas {
		b, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		id, err = model.BuildDigitalAssetID(b, fingerprint.AlgoSha256, "")
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(b)
		fileName = model.UnwrapDigitalAssetID(id)
	} else {
		fileName, err = utils.RandomID(32)
		if err != nil {
			return nil, err
		}

		id = v.prefix + fileName
	}

	fileName = filepath.Join(v.root, fileName)
	res := &model.StoredResource{
		ID:     id,
		Type:   model.TypeResource,
		Vault:  v.id,
		Method: VaultType,
	}

	if v.sse {
		encKey := model.NewEncryptionKey()

		b, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		encryptedData, err := model.EncryptAESCGM(b, encKey)
		if err != nil {
			return nil, err
		}

		r = bytes.NewReader(encryptedData)

		res.Params = map[string]any{
			"sseKey": base64.StdEncoding.EncodeToString(encKey[:]),
		}
	}

	w, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	defer w.Close()

	n, err := io.Copy(w, r)
	if err != nil {
		return nil, err
	}

	log.Info().Str("fileName", fileName).Int64("size", n).Msg("Saved blob file")

	return res, nil
}

func (v *FileSystemVault) PurgeBlob(id string, params map[string]any) error {
	if v.verifier != nil {
		state, err := v.verifier.GetDataAssetState(id)
		if err != nil {
			return err
		}

		switch state {
		case model.DataAssetStateKeep:
			return errors.New("data asset in use and can't be purged")
		case model.DataAssetStateNotFound:
			return model.ErrBlobNotFound
		case model.DataAssetStateRemove:
			// all fine
		}
	}

	fileName := filepath.Join(v.root, model.UnwrapDigitalAssetID(id))

	// detect if file exists
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return model.ErrBlobNotFound
	}

	return os.Remove(fileName)
}

func (v *FileSystemVault) ServeBlob(id string, params map[string]any, accessToken string) (io.ReadCloser, error) {
	if v.verifier != nil {
		if !model.VerifyAccessToken(accessToken, id, time.Now().Unix(), model.DefaultMaxDistanceSeconds, v.verifier) {
			return nil, model.ErrDataAssetAccessDenied
		}
	}

	fileName := filepath.Join(v.root, model.UnwrapDigitalAssetID(id))

	f, err := os.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, model.ErrBlobNotFound
		} else {
			return nil, err
		}
	}
	r := io.ReadCloser(f)

	if v.sse {
		defer r.Close()

		sseKeyStr, hasSSEKey := params["sseKey"]
		if !hasSSEKey {
			return nil, errors.New("missing SSE encryption key in requested storage parameters")
		}

		encKey := &model.AESKey{}
		keyBytes, _ := base64.StdEncoding.DecodeString(sseKeyStr.(string))
		copy(encKey[:], keyBytes)

		encryptedFileBytes, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		fileBytes, err := model.DecryptAESCGM(encryptedFileBytes, encKey)
		if err != nil {
			return nil, err
		}

		r = io.NopCloser(bytes.NewReader(fileBytes))
	}

	return r, nil
}

func (v *FileSystemVault) ID() string {
	return v.id
}

func (v *FileSystemVault) Name() string {
	return v.name
}

func (v *FileSystemVault) Close() error {
	log.Info().Msg("Closing file system based vault")
	return nil
}

func CreateVault(cfg *vaults.Config, resolver cmdbase.ParameterResolver, verifier model.AccessVerifier) (vaults.Vault, error) {
	rootDir, parameterFound := cfg.Params["root_dir"]
	if !parameterFound {
		return nil, fmt.Errorf("parameter not found: root_dir. Can't start the vault")
	}
	rootDirStr := utils.AbsPathify(rootDir.(string))

	log.Info().Str("path", rootDirStr).Msg("Initialising file system based vault")

	if _, err := os.Stat(rootDirStr); os.IsNotExist(err) {
		// create root directory, if doesn't exist already
		if err = os.MkdirAll(rootDirStr, 0o755); err != nil {
			return nil, fmt.Errorf("error creating folder %s: %w", rootDirStr, err)
		}
	}

	return &FileSystemVault{
		id:       cfg.ID,
		name:     cfg.Name,
		root:     rootDirStr,
		prefix:   model.BuildDIDPrefix(""),
		sse:      cfg.SSE,
		cas:      cfg.CAS,
		verifier: verifier,
	}, nil
}
