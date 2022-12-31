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

package model_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/piprate/metalocker/contexts"
	. "github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	updateTestFiles = false
)

func TestNewImpression(t *testing.T) {
	testDoc, err := os.ReadFile("testdata/impression/folder.json")
	require.NoError(t, err)

	imp, err := NewImpression(testDoc)
	require.NoError(t, err)
	require.NotNil(t, imp)
	assert.Equal(t, []string{"Impression", "Entity", "Bundle"}, imp.Type)
	assert.Equal(t, "did:piprate:BKsCwWhKKZyjNk98ew6SQxvaviSryYVBNJnu7xwsEuLy", imp.Asset)
}

func TestNewImpressionBroken(t *testing.T) {
	testDoc, err := os.ReadFile("testdata/impression/broken.json")
	require.NoError(t, err)

	imp, err := NewImpression(testDoc)
	assert.NotNil(t, err)
	assert.Nil(t, imp)
}

func TestNewBlankImpression(t *testing.T) {
	imp := NewBlankImpression()
	assert.Equal(t, "", imp.ID)

	expectedJSON := `
	{
		"@context": "https://piprate.org/context/piprate.jsonld",
		"id": "",
		"type": [ "Impression", "Entity", "Bundle" ]
	}`

	testbase.AssertEqualJSON(t, expectedJSON, imp)
}

func TestImpressionCompact(t *testing.T) {
	_ = contexts.PreloadContextsIntoMemory()

	testDoc, err := os.ReadFile("testdata/impression/folder.json")
	require.NoError(t, err)

	imp, err := NewImpression(testDoc)
	require.NoError(t, err)
	assert.NotNil(t, imp)

	actual, err := imp.Compact()
	require.NoError(t, err)

	expectedFilePath := "testdata/_results/impression/folder_compacted.json"
	expectedBytes, err := os.ReadFile(expectedFilePath)
	require.NoError(t, err)

	if ok, pretty := testbase.AssertEqualJSON(t, expectedBytes, actual); !ok && updateTestFiles {
		// update expected file - useful for updating the schema
		err = os.WriteFile(expectedFilePath, pretty, 0o600)
		assert.NoError(t, err)
	}
}

//nolint:thelper
func testExpandedForm(t *testing.T, testName string, updateExpectedFile bool) {
	_ = contexts.PreloadContextsIntoMemory()

	SetDebugMode(false)

	testFilePath := fmt.Sprintf("testdata/impression/%s.json", testName)
	expectedFilePath := fmt.Sprintf("testdata/_results/impression/%s_expanded.json", testName)

	testDoc, err := os.ReadFile(testFilePath)
	require.NoError(t, err)

	imp, err := NewImpression(testDoc)
	require.NoError(t, err)
	assert.NotNil(t, imp)

	expanded, err := ExpandDocument(imp.Bytes())
	require.NoError(t, err)

	expectedBytes, err := os.ReadFile(expectedFilePath)
	require.NoError(t, err)

	if ok, pretty := testbase.AssertEqualJSON(t, expectedBytes, expanded); !ok && updateExpectedFile {
		// update expected file - useful for updating the schema
		err = os.WriteFile(expectedFilePath, pretty, 0o600)
		require.NoError(t, err)
	}
}

func TestImpressionExpandedForm(t *testing.T) {
	testExpandedForm(t, "folder", updateTestFiles)
}

func TestImpressionWithRevisionsExpandedForm(t *testing.T) {
	testExpandedForm(t, "folder_second_revision", updateTestFiles)
}

func TestImpressionMerkleSign(t *testing.T) {
	_ = contexts.PreloadContextsIntoMemory()

	testDoc, err := os.ReadFile("testdata/impression/folder.json")
	require.NoError(t, err)

	imp, err := NewImpression(testDoc)
	require.NoError(t, err)
	assert.NotNil(t, imp)

	did, err := GenerateDID(WithSeed("Test0001"))
	require.NoError(t, err)

	signKey := did.SignKeyValue()
	err = imp.MerkleSign(did.ID, signKey)
	require.NoError(t, err)

	expectedFilePath := "testdata/_results/impression/folder_signed.json"
	expectedBytes, err := os.ReadFile(expectedFilePath)
	require.NoError(t, err)

	if ok, pretty := testbase.AssertEqualJSON(t, expectedBytes, imp); !ok && updateTestFiles {
		// update expected file - useful for updating the schema
		err = os.WriteFile(expectedFilePath, pretty, 0o600)
		assert.NoError(t, err)
	}
}

func TestImpressionMerkleVerify(t *testing.T) {
	_ = contexts.PreloadContextsIntoMemory()

	testDoc, err := os.ReadFile("testdata/_results/impression/folder_signed.json")
	require.NoError(t, err)

	imp, err := NewImpression(testDoc)
	require.NoError(t, err)
	assert.NotNil(t, imp)

	did, err := GenerateDID(WithSeed("Test0001"))
	require.NoError(t, err)

	verKey := base58.Decode(did.VerKey)
	ver, err := imp.MerkleVerify(verKey)
	require.NoError(t, err)
	assert.True(t, ver)
}
