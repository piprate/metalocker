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

package graph

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/piprate/metalocker/cmd/metalo/datatypes"
	"github.com/piprate/metalocker/cmd/metalo/datatypes/directory"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/dataset"
	"github.com/piprate/metalocker/utils"
	"github.com/piprate/metalocker/utils/jsonw"
)

const (
	Type = datatypes.GraphType
)

type Uploader struct {
	filePath  string
	vaultName string
	mapping   map[string]string
}

func NewUploader(filePath, vaultName string, mapping map[string]string) (datatypes.Uploader, error) {
	return &Uploader{
		filePath:  filePath,
		vaultName: vaultName,
		mapping:   mapping,
	}, nil
}

func (gu *Uploader) Write(b dataset.Builder) error {
	fi, err := os.Stat(gu.filePath)
	if err != nil {
		return err
	}

	metaDataPath := gu.filePath
	provenancePath := ""
	if fi.Mode().IsDir() {
		metaDataPath = filepath.Join(gu.filePath, datatypes.DefaultMetadataFile)
		if _, err := os.Stat(metaDataPath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("%s file not found. The data set is not a graph", metaDataPath)
			} else {
				return err
			}
		}

		provPath := filepath.Join(gu.filePath, datatypes.DefaultProvenanceFile)
		if _, err := os.Stat(provPath); err == nil {
			// provenance file exists
			provenancePath = provPath
		}

		_, err = directory.ProcessDir(gu.filePath, gu.vaultName, b, []string{datatypes.DefaultMetadataFile, datatypes.DefaultImpressionFile,
			datatypes.DefaultProvenanceFile, datatypes.DefaultOperationFile})
		if err != nil {
			return err
		}
	}

	metaData, err := os.ReadFile(metaDataPath)
	if err != nil {
		return err
	}

	if provenancePath != "" {
		provTemplate, err := os.ReadFile(provenancePath)
		if err != nil {
			return err
		}

		provBytes, err := utils.SubstituteEntities(provTemplate, gu.mapping, nil)
		if err != nil {
			return err
		}

		var prov any
		if err = jsonw.Unmarshal(provBytes, &prov); err != nil {
			return fmt.Errorf("error reading provenance definition: %w", err)
		}

		if err = b.AddProvenance("", prov, true); err != nil {
			return err
		}

	}

	mrID, err := b.AddMetaResource(metaData, dataset.WithVault(gu.vaultName))
	if err != nil {
		return err
	}

	if provenancePath == "" {
		genTime := time.Now().UTC()

		err = b.AddProvenance(mrID, &model.ProvEntity{
			ID:              mrID,
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
	}

	return err
}

type Renderer struct {
	metaBytes []byte
	ds        model.DataSet
}

func NewRenderer(ds model.DataSet, resourceBytes []byte) (datatypes.Renderer, error) {
	res := &Renderer{
		metaBytes: resourceBytes,
		ds:        ds,
	}

	return res, nil
}

func (gr *Renderer) Print() error {
	var val any
	err := jsonw.Unmarshal(gr.metaBytes, &val)
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

func (gr *Renderer) ExportToDisk(ctx context.Context, destFolder string, includeMetaData bool) error {
	err := os.MkdirAll(destFolder, 0o700)
	if err != nil {
		return err
	}

	if includeMetaData {
		metaFilePath := filepath.Join(destFolder, datatypes.DefaultMetadataFile)

		if err := os.WriteFile(metaFilePath, gr.metaBytes, 0o600); err != nil {
			return err
		}
	}

	for _, assetID := range gr.ds.Resources() {
		fileName := filepath.Join(destFolder, strings.ReplaceAll(assetID, ":", "_"))

		r, err := gr.ds.Resource(ctx, assetID)
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
