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

package main

import (
	"github.com/piprate/metalocker/cmd/metalo/datatypes"
	"github.com/piprate/metalocker/cmd/metalo/datatypes/directory"
	"github.com/piprate/metalocker/cmd/metalo/datatypes/file"
	"github.com/piprate/metalocker/cmd/metalo/datatypes/graph"
)

func init() {

	// register a basic set of supported data types to read and write to disk.

	datatypes.RegisterUploader(graph.Type, graph.NewUploader)
	datatypes.RegisterRenderer(graph.Type, graph.NewRenderer)

	datatypes.RegisterUploader(file.Type, file.NewUploader)
	datatypes.RegisterRenderer(file.Type, file.NewRenderer)

	datatypes.RegisterUploader(directory.Type, directory.NewUploader)
	datatypes.RegisterRenderer(directory.Type, directory.NewRenderer)
}
