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

package contexts

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/piprate/metalocker/model"
)

var contextsLoaded = false

//go:embed files
var contextFS embed.FS

func PreloadContextsIntoMemory() error {
	if !contextsLoaded {

		for url, boxPath := range map[string]string{
			"https://piprate.org/context/piprate.jsonld":    "files/piprate.jsonld",
			"https://piprate.org/context/prov.jsonld":       "files/prov.jsonld",
			"https://piprate.org/context/metalocker.jsonld": "files/metalocker.jsonld",
			"http://schema.org/":                            "files/schema.jsonld",
			"https://w3id.org/security/v1":                  "files/security-v1.jsonld",
			"https://w3id.org/security/v2":                  "files/security-v2.jsonld",
			"https://w3id.org/did/v1":                       "files/did-v1.jsonld",
		} {
			b, err := contextFS.ReadFile(boxPath)
			if err != nil {
				return fmt.Errorf("context file not found: %s", boxPath)
			}

			err = model.PutBinaryContextIntoDefaultDocumentLoader(url, b)
			if err != nil {
				return err
			}
		}

		contextsLoaded = true
	}

	return nil
}

func HTTPFileSystem() http.FileSystem {
	subFS, err := fs.Sub(contextFS, "files")
	if err != nil {
		panic(err)
	}

	return http.FS(subFS)
}
