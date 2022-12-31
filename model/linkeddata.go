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

package model

import (
	"errors"
	"sync"

	"github.com/piprate/json-gold/ld"
	"github.com/piprate/metalocker/utils/jsonw"
)

const (
	crvyBase = "http://crvy.org/"
)

var (
	documentLoaderLock    sync.Mutex
	defaultDocumentLoader = ld.DocumentLoader(ld.NewCachingDocumentLoader(ld.NewDefaultDocumentLoader(nil)))
)

func DefaultDocumentLoader() ld.DocumentLoader {
	return defaultDocumentLoader
}

func SetDefaultDocumentLoader(l ld.DocumentLoader) {
	defaultDocumentLoader = l
}

func PutBinaryContextIntoDefaultDocumentLoader(url string, ctx []byte) error {
	documentLoaderLock.Lock()
	defer documentLoaderLock.Unlock()

	cdl, correctType := defaultDocumentLoader.(*ld.CachingDocumentLoader)
	if !correctType {
		return errors.New("failed to put context into cache: wrong loader type")
	}

	var ctxDoc any
	if err := jsonw.Unmarshal(ctx, &ctxDoc); err != nil {
		return err
	}

	cdl.AddDocument(url, ctxDoc)

	return nil
}

func PutContextIntoDefaultDocumentLoader(url, filePath string) error {
	documentLoaderLock.Lock()
	defer documentLoaderLock.Unlock()

	cdl, correctType := defaultDocumentLoader.(*ld.CachingDocumentLoader)
	if !correctType {
		return errors.New("failed to put context into cache: wrong loader type")
	}
	return cdl.PreloadWithMapping(map[string]string{url: filePath})
}

func PutContextMapIntoDefaultDocumentLoader(contextMap map[string]string) error {
	for url, filePath := range contextMap {
		if err := PutContextIntoDefaultDocumentLoader(url, filePath); err != nil {
			return err
		}
	}
	return nil
}

func ExpandDocument(input []byte) ([]byte, error) {
	var val any
	err := jsonw.Unmarshal(input, &val)
	if err != nil {
		return nil, err
	}

	proc := ld.NewJsonLdProcessor()

	opts := ld.NewJsonLdOptions(crvyBase)
	opts.ProcessingMode = ld.JsonLd_1_1
	opts.DocumentLoader = DefaultDocumentLoader()

	newData, err := proc.Expand(val, opts)
	if err != nil {
		return nil, err
	}

	expandedDoc, err := jsonw.Marshal(newData)
	if err != nil {
		return nil, err
	}

	return expandedDoc, nil
}

func CompactDocument(input []byte, ctxURL string) ([]byte, error) {
	var val any
	err := jsonw.Unmarshal(input, &val)
	if err != nil {
		return nil, err
	}

	proc := ld.NewJsonLdProcessor()

	opts := ld.NewJsonLdOptions(crvyBase)
	opts.ProcessingMode = ld.JsonLd_1_1
	opts.DocumentLoader = DefaultDocumentLoader()

	ctxDoc, err := opts.DocumentLoader.LoadDocument(ctxURL)
	if err != nil {
		return nil, err
	}

	newData, err := proc.Compact(val, ctxDoc.Document, opts)
	if err != nil {
		return nil, err
	}

	newData["@context"] = ctxURL

	res, err := jsonw.Marshal(newData)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// GenerateDocumentNodeMap exposes GenerateNodeMap method from JSON-LD API.
// It shouldn't be really used directly (use Flatten instead),
// but it's sometimes useful for debugging JSON-LD schema related issues.
func GenerateDocumentNodeMap(input []byte) (map[string]any, error) {
	var val any
	err := jsonw.Unmarshal(input, &val)
	if err != nil {
		return nil, err
	}

	proc := ld.NewJsonLdProcessor()

	opts := ld.NewJsonLdOptions(crvyBase)
	opts.ProcessingMode = ld.JsonLd_1_1
	opts.DocumentLoader = DefaultDocumentLoader()

	idGen := ld.NewIdentifierIssuer("_:b")

	expandedInput, err := proc.Expand(val, opts)
	if err != nil {
		return nil, err
	}

	api := ld.NewJsonLdApi()

	nodeMap := make(map[string]any)
	nodeMap["@default"] = make(map[string]any)
	_, err = api.GenerateNodeMap(expandedInput, nodeMap, "@default", idGen, "", "", nil)
	if err != nil {
		return nil, err
	}

	return nodeMap, nil
}

func FlattenDocument(input []byte, ctx any) ([]byte, error) {
	var val any
	err := jsonw.Unmarshal(input, &val)
	if err != nil {
		return nil, err
	}

	proc := ld.NewJsonLdProcessor()

	opts := ld.NewJsonLdOptions(crvyBase)
	opts.ProcessingMode = ld.JsonLd_1_1
	opts.DocumentLoader = DefaultDocumentLoader()

	if ctxURL, isString := ctx.(string); isString {
		ctxDoc, err := opts.DocumentLoader.LoadDocument(ctxURL)
		if err != nil {
			return nil, err
		}
		ctx = ctxDoc.Document
	}

	newData, err := proc.Flatten(val, ctx, opts)
	if err != nil {
		return nil, err
	}

	res, err := jsonw.Marshal(newData)
	if err != nil {
		return nil, err
	}

	return res, nil
}
