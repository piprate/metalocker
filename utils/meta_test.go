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

	"github.com/piprate/metalocker/sdk/testbase"
	. "github.com/piprate/metalocker/utils"
	"github.com/stretchr/testify/assert"
)

func init() {
	testbase.SetupLogFormat()
}

func TestDiscoverMetaGraph(t *testing.T) {
	id, ct := DiscoverMetaGraph([]byte(`
{
	"type": "File"
}
	`))
	assert.Equal(t, "", id)
	assert.Equal(t, "File", ct)

	id, ct = DiscoverMetaGraph([]byte(`
{
	"id": "did:piprate:xxx",
	"type": "File"
}
	`))
	assert.Equal(t, "did:piprate:xxx", id)
	assert.Equal(t, "File", ct)

	id, ct = DiscoverMetaGraph([]byte(`
{
	"id": "did:piprate:xxx"
}
	`))
	assert.Equal(t, "", id)
	assert.Equal(t, "", ct)

	id, ct = DiscoverMetaGraph([]byte(`
[
	{
		"type": "File"
	}
]
	`))
	assert.Equal(t, "", id)
	assert.Equal(t, "File", ct)

	id, ct = DiscoverMetaGraph([]byte(`
[
	{
		"id": "did:piprate:xxx",
		"type": "File"
	}
]
	`))
	assert.Equal(t, "did:piprate:xxx", id)
	assert.Equal(t, "File", ct)

	id, ct = DiscoverMetaGraph([]byte(`
[
	{
		"id": "did:piprate:xxx"
	}
]
	`))
	assert.Equal(t, "", id)
	assert.Equal(t, "", ct)

	id, ct = DiscoverMetaGraph([]byte(`
[]
	`))
	assert.Equal(t, "", id)
	assert.Equal(t, "", ct)

	id, ct = DiscoverMetaGraph([]byte(`
some text
	`))
	assert.Equal(t, "", id)
	assert.Equal(t, "", ct)
}
