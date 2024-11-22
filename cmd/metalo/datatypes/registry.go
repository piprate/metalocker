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

package datatypes

import (
	"context"
	"errors"
	"io"
	"sort"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/utils"
	"github.com/rs/zerolog/log"
)

const (
	GraphType = "Graph"
)

var (
	ErrTypeNotSupported = errors.New("data type not supported")
)

var defaultType = GraphType

type rendererConstructor func(ds model.DataSet, resourceBytes []byte) (Renderer, error)
type uploaderConstructor func(sourcePath, vaultName string, mapping map[string]string) (Uploader, error)

var rendererConstructors = make(map[string]rendererConstructor)
var uploaderConstructors = make(map[string]uploaderConstructor)

func RegisterRenderer(metaType string, ctor rendererConstructor) {
	if _, ok := rendererConstructors[metaType]; ok {
		panic("dataset renderer already registered for meta type: " + metaType)
	}

	rendererConstructors[metaType] = ctor
}

func RegisterUploader(metaType string, ctor uploaderConstructor) {
	if _, ok := uploaderConstructors[metaType]; ok {
		panic("dataset uploader already registered for meta type: " + metaType)
	}

	uploaderConstructors[metaType] = ctor
}

func NewRenderer(ctx context.Context, ds model.DataSet) (Renderer, error) {

	r, err := ds.MetaResource(ctx)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	resourceBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	_, metaType := utils.DiscoverMetaGraph(resourceBytes)

	log.Debug().Str("metaType", metaType).Msg("Creating dataset renderer")

	ctor, ok := rendererConstructors[metaType]
	if !ok {
		ctor, ok = rendererConstructors[defaultType]
		if !ok {
			return nil, ErrTypeNotSupported
		}
	}

	return ctor(ds, resourceBytes)
}

func NewUploader(metaType, sourcePath, vaultName string, mapping map[string]string) (Uploader, error) {

	log.Debug().Str("metaType", metaType).Msg("Creating dataset uploader")

	ctor, ok := uploaderConstructors[metaType]
	if !ok {
		ctor, ok = uploaderConstructors[defaultType]
		if !ok {
			return nil, ErrTypeNotSupported
		}
	}

	return ctor(sourcePath, vaultName, mapping)
}

func SupportedDataTypes() []string {
	list := make([]string, 0, len(uploaderConstructors))
	for k := range uploaderConstructors {
		list = append(list, k)
	}
	sort.Strings(list)
	return list
}

func SetDefaultType(metaType string) {
	defaultType = metaType
}
