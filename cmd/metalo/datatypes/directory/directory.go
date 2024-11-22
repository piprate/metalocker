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

package directory

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/piprate/metalocker/cmd/metalo/datatypes"
	"github.com/piprate/metalocker/cmd/metalo/datatypes/file"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/dataset"
	fp "github.com/piprate/metalocker/utils/fingerprint"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog/log"
)

const (
	Type = "Directory"
)

type MetaLockerDirectory struct {
	Context              any                    `json:"@context,omitempty"`
	ID                   string                 `json:"id,omitempty"`
	Type                 string                 `json:"type"`
	Fingerprint          string                 `json:"fingerprint,omitempty"`
	FingerprintAlgorithm string                 `json:"fingerprintAlgorithm,omitempty"`
	Files                []*file.MetaLockerFile `json:"files"`
}

type Uploader struct {
	path      string
	vaultName string
}

func NewUploader(path, vaultName string, mapping map[string]string) (datatypes.Uploader, error) {
	return &Uploader{
		path:      path,
		vaultName: vaultName,
	}, nil
}

func (du *Uploader) Write(b dataset.Builder) error {
	dir, err := ProcessDir(du.path, du.vaultName, b, nil)
	if err != nil {
		return err
	}

	mrID, err := b.AddMetaResource(dir,
		dataset.WithAsset(dir.ID),
		dataset.WithContentType(dir.Type),
		dataset.WithVault(du.vaultName),
	)
	if err != nil {
		return err
	}

	genTime := time.Now().UTC()

	err = b.AddProvenance(mrID, &model.ProvEntity{
		ID:              mrID,
		Type:            model.ProvTypeEntity,
		GeneratedAtTime: &genTime,
		WasGeneratedBy: &model.ProvActivity{
			Type:              model.ProvTypeActivity,
			WasAssociatedWith: b.CreatorID(),
			Algorithm:         "Capture",
		},
	}, false)
	if err != nil {
		return err
	}

	return err
}

func ProcessDir(dirPath, vaultName string, b dataset.Builder, exclusionList []string) (*MetaLockerDirectory, error) {
	fileList := make([]*file.MetaLockerFile, 0)
	assets := make(map[string]bool)
	fingerprintAlgo := fp.AlgoSha256

	exclusionMap := make(map[string]bool)
	for _, val := range exclusionList {
		exclusionMap[val] = true
	}

	if err := filepath.Walk(dirPath, func(filePath string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.Mode().IsRegular() {
			log.Debug().Str("path", filePath).Msg("Processing file")
			if _, exclude := exclusionMap[f.Name()]; exclude {
				log.Debug().Str("path", filePath).Msg("Exclude file")
				return nil
			}

			assetID, fingerprint, err := model.BuildDigitalAssetIDFromFile(filePath, fingerprintAlgo, "")
			if err != nil {
				return err
			}

			if _, found := assets[assetID]; !found {
				fileHandler, err := os.Open(filePath)
				if err != nil {
					return err
				}
				defer fileHandler.Close()

				_, err = b.AddResource(fileHandler, dataset.WithVault(vaultName))
				if err != nil {
					return err
				}

				genTime := f.ModTime().UTC()

				err = b.AddProvenance(assetID, &model.ProvEntity{
					ID:              assetID,
					Type:            model.ProvTypeEntity,
					GeneratedAtTime: &genTime,
					WasGeneratedBy: &model.ProvActivity{
						Type:              model.ProvTypeActivity,
						WasAssociatedWith: b.CreatorID(),
						Algorithm:         "Import",
					},
				}, false)
				if err != nil {
					return err
				}

				assets[assetID] = true
			}

			relPath, err := filepath.Rel(dirPath, filePath)
			if err != nil {
				return err
			}

			mime, err := mimetype.DetectFile(filePath)
			if err != nil {
				return err
			}

			fileList = append(fileList, &file.MetaLockerFile{
				ID:                   assetID,
				Type:                 file.Type,
				Path:                 filepath.Dir(relPath),
				Name:                 f.Name(),
				ContentSize:          f.Size(),
				FileFormat:           mime.String(),
				Fingerprint:          fingerprint,
				FingerprintAlgorithm: fingerprintAlgo,
			})
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// We assume directories don't have an implicit asset ID outside of specific context
	// and what this directory is for.

	return &MetaLockerDirectory{
		Type:  Type,
		Files: fileList,
	}, nil
}

type Renderer struct {
	dirBytes []byte
	dir      *MetaLockerDirectory
	ds       model.DataSet
}

func NewRenderer(ds model.DataSet, resourceBytes []byte) (datatypes.Renderer, error) {
	var dir MetaLockerDirectory
	err := jsonw.Unmarshal(resourceBytes, &dir)
	if err != nil {
		return nil, err
	}

	res := &Renderer{
		dir:      &dir,
		dirBytes: resourceBytes,
		ds:       ds,
	}

	return res, nil
}

func (dr *Renderer) Print() error {
	var val any
	err := jsonw.Unmarshal(dr.dirBytes, &val)
	if err != nil {
		return err
	}
	db, err := jsonw.MarshalIndent(val, "", "  ")
	if err != nil {
		return err
	}

	println(string(db))

	return nil
}

func (dr *Renderer) ExportToDisk(ctx context.Context, destFolder string, includeMetaData bool) error {
	if includeMetaData {
		metaFilePath := filepath.Join(destFolder, datatypes.DefaultMetadataFile)

		err := os.MkdirAll(destFolder, 0o700)
		if err != nil {
			return err
		}

		if err := os.WriteFile(metaFilePath, dr.dirBytes, 0o600); err != nil {
			return err
		}
	}

	for _, f := range dr.dir.Files {
		folderPath := filepath.Join(destFolder, f.Path)

		err := os.MkdirAll(folderPath, 0o700)
		if err != nil {
			return err
		}

		fileName := filepath.Join(folderPath, f.Name)

		r, err := dr.ds.Resource(ctx, f.ID)
		if err != nil {
			return err
		}
		defer r.Close()

		w, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer w.Close()

		_, err = io.Copy(w, r)
		if err != nil {
			return err
		}
	}

	return nil
}
