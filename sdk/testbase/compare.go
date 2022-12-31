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

package testbase

import (
	"testing"

	"github.com/piprate/json-gold/ld"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/stretchr/testify/assert"
)

func AssertEqualJSON(t *testing.T, expectedDoc, actualDoc any) (bool, []byte) {
	t.Helper()
	return assertEqualJSON(t, expectedDoc, actualDoc, false)
}

func AssertEqualUnorderedJSON(t *testing.T, expectedDoc, actualDoc any) (bool, []byte) {
	t.Helper()
	return assertEqualJSON(t, expectedDoc, actualDoc, true)
}

func assertEqualJSON(t *testing.T, expectedDoc, actualDoc any, ignoreListOrder bool) (bool, []byte) { //nolint:thelper
	expected := toStandardForm(expectedDoc)
	actual := toStandardForm(actualDoc)
	res := assert.True(t, ld.DeepCompare(expected, actual, !ignoreListOrder))
	if !res {
		ld.PrintDocument("Expected document", expected)
		ld.PrintDocument("Actual document", actual)

		// return pretty-printed JSON
		prettyForm, err := jsonw.MarshalIndent(actual, "", "  ")
		if err != nil {
			panic(err)
		}

		return res, prettyForm
	} else {
		return res, nil
	}
}

func toStandardForm(doc any) any {
	switch v := doc.(type) {
	case []byte:
		var res any
		err := jsonw.Unmarshal(v, &res)
		if err != nil {
			panic(err)
		}
		return res
	case string:
		return toStandardForm([]byte(v))
	default:
		b, err := jsonw.Marshal(doc)
		if err != nil {
			panic(err)
		}
		return toStandardForm(b)
	}
}
