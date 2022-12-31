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

package directory_test

import (
	"os"
	"testing"

	. "github.com/piprate/metalocker/cmd/metalo/datatypes/directory"
	"github.com/piprate/metalocker/cmd/metalo/datatypes/test"
	"github.com/piprate/metalocker/contexts"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = contexts.PreloadContextsIntoMemory()
	testbase.SetupLogFormat()
}

func TestProcessDir(t *testing.T) {
	mlb := &test.MockLeaseBuilder{
		Resources: make(map[string][]byte),
	}

	ds, err := ProcessDir("testdata/folder", "local", mlb, nil)
	require.NoError(t, err)

	expectedBytes, err := os.ReadFile("testdata/_results/directory.json")
	require.NoError(t, err)

	testbase.AssertEqualJSON(t, expectedBytes, ds)

	require.Equal(t, 3, len(mlb.Resources))

	assert.Equal(t, 0, len(mlb.Resources["did:piprate:674wjTvPisYXNHNj1MbmqTKcekqttPwVuGoDDgpAbQUn"]))
	assert.Equal(t, 13, len(mlb.Resources["did:piprate:GGKW4B2zLbpxPJXE8TcD1okHFg4ymHvYbDfbHY2BthYc"]))
	assert.Equal(t, 6, len(mlb.Resources["did:piprate:85LM8giLGhiT31ZK4ugWkdLWzVTAFpxibWXN1v1jzZ7D"]))
}

func TestFolderDataSetWriter_Write(t *testing.T) {
	fw, err := NewUploader("testdata/folder", "local", nil)
	require.NoError(t, err)

	mlb := &test.MockLeaseBuilder{
		Resources: make(map[string][]byte),
	}

	err = fw.Write(mlb)
	require.NoError(t, err)

	require.Equal(t, 4, len(mlb.Resources))

	assert.Equal(t, 0, len(mlb.Resources["did:piprate:674wjTvPisYXNHNj1MbmqTKcekqttPwVuGoDDgpAbQUn"]))
	assert.Equal(t, 13, len(mlb.Resources["did:piprate:GGKW4B2zLbpxPJXE8TcD1okHFg4ymHvYbDfbHY2BthYc"]))
	assert.Equal(t, 6, len(mlb.Resources["did:piprate:85LM8giLGhiT31ZK4ugWkdLWzVTAFpxibWXN1v1jzZ7D"]))

	expectedFilePath := "testdata/_results/directory.json"
	expectedBytes, err := os.ReadFile(expectedFilePath)
	require.NoError(t, err)

	testbase.AssertEqualJSON(t, expectedBytes, mlb.Resources["did:piprate:EipbLjrDwnqDS3E5pWH5qP2VcsH4jTWnMfNWDeu54oE1"])
}
