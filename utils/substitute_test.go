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

package utils_test

import (
	"testing"
	"time"

	"github.com/piprate/metalocker/sdk/testbase"
	. "github.com/piprate/metalocker/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildMapFromStringSuccess(t *testing.T) {
	res, err := BuildMapFromString("v1/abc,v2/xyz")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"v1": "abc", "v2": "xyz"}, res)
}

func TestBuildMapFromStringFailure(t *testing.T) {
	_, err := BuildMapFromString("v1,v2/xyz")
	assert.NotNil(t, err)
}

func TestSubstituteEntities(t *testing.T) {
	templ := []byte(`
[
  {
    "id": "did:piprate:2HzAMUFWpvqCEeEQhtc5Fjw7PsfkNm4SJzXwZoJ2zEL8",
    "type": "Entity",
    "generatedAtTime": "2019-04-09T14:02:00.464345Z",
    "wasGeneratedBy": {
      "type": "Activity",
      "wasAssociatedWith": "did:piprate:GMavEk4K2CDHSmixB3UTuCaL1cqfqEjvSU6vygdrz7ih",
      "algorithm": "Geocoding",
      "qualifiedUsage": [
        {
          "entity": "%v1%",
          "hadRole": {
            "label": "source",
            "type": "Role"
          },
          "type": "Usage"
        }
      ],
      "used": [
        "%v1%"
      ]
    }
  },
  {
    "id": "did:piprate:GMavEk4K2CDHSmixB3UTuCaL1cqfqEjvSU6vygdrz7ih",
    "type": "Agent"
  }
]
`)
	res, err := SubstituteEntities(templ, map[string]string{
		"v1": "xxx",
	}, nil)
	require.NoError(t, err)

	expected := `
[
  {
    "id": "did:piprate:2HzAMUFWpvqCEeEQhtc5Fjw7PsfkNm4SJzXwZoJ2zEL8",
    "type": "Entity",
    "generatedAtTime": "2019-04-09T14:02:00.464345Z",
    "wasGeneratedBy": {
      "type": "Activity",
      "wasAssociatedWith": "did:piprate:GMavEk4K2CDHSmixB3UTuCaL1cqfqEjvSU6vygdrz7ih",
      "algorithm": "Geocoding",
      "qualifiedUsage": [
        {
          "entity": "xxx",
          "hadRole": {
            "label": "source",
            "type": "Role"
          },
          "type": "Usage"
        }
      ],
      "used": [
        "xxx"
      ]
    }
  },
  {
    "id": "did:piprate:GMavEk4K2CDHSmixB3UTuCaL1cqfqEjvSU6vygdrz7ih",
    "type": "Agent"
  }
]
`

	testbase.AssertEqualJSON(t, expected, res)
}

func TestSubstituteEntitiesWithNowVar(t *testing.T) {
	templ := []byte(`
[
  {
    "id": "did:piprate:2HzAMUFWpvqCEeEQhtc5Fjw7PsfkNm4SJzXwZoJ2zEL8",
    "type": "Entity",
    "generatedAtTime": "%%now%%",
    "wasGeneratedBy": {
      "type": "Activity",
      "wasAssociatedWith": "did:piprate:GMavEk4K2CDHSmixB3UTuCaL1cqfqEjvSU6vygdrz7ih",
      "algorithm": "Geocoding",
      "qualifiedUsage": [
        {
          "entity": "%v1%",
          "hadRole": {
            "label": "source",
            "type": "Role"
          },
          "type": "Usage"
        }
      ],
      "used": [
        "%v1%"
      ]
    }
  },
  {
    "id": "did:piprate:GMavEk4K2CDHSmixB3UTuCaL1cqfqEjvSU6vygdrz7ih",
    "type": "Agent"
  }
]
`)

	ts := time.Unix(1000, 200).UTC()

	res, err := SubstituteEntities(templ, map[string]string{
		"v1": "xxx",
	}, &ts)
	require.NoError(t, err)

	expected := `
[
  {
    "id": "did:piprate:2HzAMUFWpvqCEeEQhtc5Fjw7PsfkNm4SJzXwZoJ2zEL8",
    "type": "Entity",
    "generatedAtTime": "1970-01-01T00:16:40.0000002Z",
    "wasGeneratedBy": {
      "type": "Activity",
      "wasAssociatedWith": "did:piprate:GMavEk4K2CDHSmixB3UTuCaL1cqfqEjvSU6vygdrz7ih",
      "algorithm": "Geocoding",
      "qualifiedUsage": [
        {
          "entity": "xxx",
          "hadRole": {
            "label": "source",
            "type": "Role"
          },
          "type": "Usage"
        }
      ],
      "used": [
        "xxx"
      ]
    }
  },
  {
    "id": "did:piprate:GMavEk4K2CDHSmixB3UTuCaL1cqfqEjvSU6vygdrz7ih",
    "type": "Agent"
  }
]
`

	testbase.AssertEqualJSON(t, expected, res)
}
