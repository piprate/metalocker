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

package jsonw_test

import (
	"bytes"
	"testing"

	. "github.com/piprate/metalocker/utils/jsonw"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestUnmarshal(t *testing.T) {
	data := []byte(`
{
	"id": 123,
    "name": "struct name"
}`)

	var res testStruct
	err := Unmarshal(data, &res)
	require.NoError(t, err)
	assert.Equal(t, 123, res.ID)
	assert.Equal(t, "struct name", res.Name)
}

func TestMarshal(t *testing.T) {
	res := testStruct{
		ID:   123,
		Name: "struct name",
	}

	b, err := Marshal(res)
	require.NoError(t, err)
	assert.Equal(t, `{"id":123,"name":"struct name"}`, string(b))
}

func TestMarshalIndent(t *testing.T) {
	res := testStruct{
		ID:   123,
		Name: "struct name",
	}

	b, err := MarshalIndent(res, "", "  ")
	require.NoError(t, err)
	assert.Equal(t, `{
  "id": 123,
  "name": "struct name"
}
`, string(b))
}

func TestDecode(t *testing.T) {
	r := bytes.NewReader([]byte(`
{
	"id": 123,
    "name": "struct name"
}`))

	var res testStruct
	err := Decode(r, &res)
	require.NoError(t, err)
	assert.Equal(t, 123, res.ID)
	assert.Equal(t, "struct name", res.Name)
}

func TestEncode(t *testing.T) {
	res := testStruct{
		ID:   123,
		Name: "struct name",
	}

	w := bytes.NewBuffer(nil)
	err := Encode(res, w)
	require.NoError(t, err)
	assert.Equal(t, `{"id":123,"name":"struct name"}
`, w.String())
}

func TestMarshalToTypeWithFieldValidation(t *testing.T) {
	r := map[string]any{
		"id":   123,
		"name": "struct name",
	}

	var res testStruct

	err := MarshalToTypeWithFieldValidation(r, &res)
	require.NoError(t, err)
	assert.Equal(t, 123, res.ID)
	assert.Equal(t, "struct name", res.Name)

	r = map[string]any{
		"id":          123,
		"name":        "struct name",
		"wrong_field": "value",
	}

	err = MarshalToTypeWithFieldValidation(r, &res)
	require.Error(t, err)
}
