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

package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/piprate/metalocker/cmd/metalo/datatypes"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/dataset"
	fp "github.com/piprate/metalocker/utils/fingerprint"
	"github.com/piprate/metalocker/utils/jsonw"
)

const (
	Type = "File"
)

type Uploader struct {
	filePath  string
	vaultName string
}

type MetaLockerFile struct {
	Context              any    `json:"@context,omitempty"`
	ID                   string `json:"id,omitempty"`
	Type                 string `json:"type"`
	Path                 string `json:"path,omitempty"`
	Name                 string `json:"name,omitempty"`
	ContentSize          int64  `json:"contentSize,omitempty"`
	FileFormat           string `json:"fileFormat,omitempty"`
	Fingerprint          string `json:"fingerprint,omitempty"`
	FingerprintAlgorithm string `json:"fingerprintAlgorithm,omitempty"`
}

var (
	supportedExtensions = map[string]bool{ //nolint: unused
		"docx": true,
		"txt":  true,
		"pdf":  true,
		"csv":  true,
	}
	mimeToExtensionMapping = map[string]string{
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": "docx",
		"text/csv":        "csv",
		"text/plain":      "txt",
		"application/pdf": "pdf",
	}
)

func (bf *MetaLockerFile) BuildFileName() string {
	ext := filepath.Ext(bf.Name)
	if ext != "" {
		return bf.Name
	} else {
		ext, found := mimeToExtensionMapping[bf.FileFormat]
		if !found {
			ext = "out"
		}
		return fmt.Sprintf("%s.%s", bf.Name, ext)
	}
}

func NewUploader(filePath, vaultName string, mapping map[string]string) (datatypes.Uploader, error) {
	return &Uploader{
		filePath:  filePath,
		vaultName: vaultName,
	}, nil
}

func (u *Uploader) Write(b dataset.Builder) error {
	fingerprintAlgo := fp.AlgoSha256

	assetID, fingerprint, err := model.BuildDigitalAssetIDFromFile(u.filePath, fingerprintAlgo, "")
	if err != nil {
		return err
	}

	// save blob

	file, err := os.Open(u.filePath)
	if err != nil {
		return err
	}

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	_, err = b.AddResource(file, dataset.WithVault(u.vaultName))
	if err != nil {
		return err
	}

	genTime := stat.ModTime().UTC()

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

	// create MetaLockerFile definition

	mime, err := mimetype.DetectFile(u.filePath)
	if err != nil {
		return err
	}

	fileName := filepath.Base(u.filePath)

	d := &MetaLockerFile{
		Context:              "https://piprate.org/context/metalocker.jsonld",
		ID:                   assetID,
		Type:                 Type,
		Name:                 fileName,
		ContentSize:          stat.Size(),
		FileFormat:           mime.String(),
		Fingerprint:          fingerprint,
		FingerprintAlgorithm: fingerprintAlgo,
	}

	mrID, err := b.AddMetaResource(d,
		dataset.WithAsset(d.ID),
		dataset.WithContentType(d.Type),
		dataset.WithVault(u.vaultName),
	)
	if err != nil {
		return err
	}

	genTime = time.Now().UTC()

	err = b.AddProvenance(mrID, &model.ProvEntity{
		ID:              mrID,
		Type:            model.ProvTypeEntity,
		GeneratedAtTime: &genTime,
		WasGeneratedBy: &model.ProvActivity{
			Type:              model.ProvTypeActivity,
			WasAssociatedWith: b.CreatorID(),
			Algorithm:         "Capture",
		},
	}, true)
	if err != nil {
		return err
	}

	return nil
}

type Renderer struct {
	resourceBytes []byte
	ds            model.DataSet
	blobID        string
}

func NewRenderer(ds model.DataSet, resourceBytes []byte) (datatypes.Renderer, error) {
	res := &Renderer{
		resourceBytes: resourceBytes,
		ds:            ds,
	}

	resourceList := ds.Resources()
	if len(resourceList) == 1 {
		res.blobID = resourceList[0]
	} else {
		return nil, fmt.Errorf("unexpected number of storage configs: %d", len(resourceList)+1)
	}

	return res, nil
}

func (r *Renderer) Print() error {
	var val any
	err := jsonw.Unmarshal(r.resourceBytes, &val)
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

func (r *Renderer) ExportToDisk(ctx context.Context, destFolder string, includeMetaData bool) error {
	err := os.MkdirAll(destFolder, 0o700)
	if err != nil {
		return err
	}

	var meta MetaLockerFile
	err = jsonw.Unmarshal(r.resourceBytes, &meta)
	if err != nil {
		return err
	}

	reader, err := r.ds.Resource(ctx, r.blobID)
	if err != nil {
		return err
	}
	defer reader.Close()

	filePath := filepath.Join(destFolder, meta.BuildFileName())

	w, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, reader)
	if err != nil {
		return err
	}

	if includeMetaData {
		metaFilePath := filepath.Join(destFolder, datatypes.DefaultMetadataFile)

		err = os.WriteFile(metaFilePath, r.resourceBytes, 0o600)
		if err != nil {
			return err
		}
	}

	return nil
}
