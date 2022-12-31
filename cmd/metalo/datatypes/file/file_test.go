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

package file_test

import (
	"testing"

	. "github.com/piprate/metalocker/cmd/metalo/datatypes/file"
	"github.com/piprate/metalocker/cmd/metalo/datatypes/test"
	"github.com/piprate/metalocker/contexts"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/stretchr/testify/require"
)

func TestUploader_Write(t *testing.T) {
	_ = contexts.PreloadContextsIntoMemory()

	fw, err := NewUploader("testdata/file1.txt", "local", nil)
	require.NoError(t, err)

	mlb := &test.MockLeaseBuilder{
		Resources: make(map[string][]byte),
	}

	err = fw.Write(mlb)
	require.NoError(t, err)

	require.Equal(t, 2, len(mlb.Resources))

	expectedDoc := `
{
  "@context": "https://piprate.org/context/metalocker.jsonld",
  "contentSize": 13,
  "fileFormat": "text/plain; charset=utf-8",
  "fingerprint": "xP20vnDo1soN221BU5YDQRjgiQoVjOlIx6O4NVf+4Ns=",
  "fingerprintAlgorithm": "fingerprints:sha256",
  "id": "did:piprate:GGKW4B2zLbpxPJXE8TcD1okHFg4ymHvYbDfbHY2BthYc",
  "name": "file1.txt",
  "type": "File"
}`

	testbase.AssertEqualJSON(t, expectedDoc, mlb.Resources["did:piprate:HUd27KFwiYXrVgBhZW8Q1kLoKEBR88BX2EWgYh4atumr"])
}
