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

package account_test

import (
	"testing"

	. "github.com/piprate/metalocker/model/account"
	"github.com/stretchr/testify/assert"
)

func TestHashUserPassword_TestPassword(t *testing.T) {
	hash := HashUserPassword("pass123")
	assert.Equal(t, "gm7FhMFxFD01wGd8dE1RqUAxx7noD8LvPQyBzK+27LA=", hash)

	hash = HashUserPassword("pass")
	assert.Equal(t, "rhTLDOANF7p9Yq8eB9Dt5EQpnMjS4K3usOzZD/veWAM=", hash)
}

func TestHashUserPassword(t *testing.T) {
	hash := HashUserPassword("testpassword")
	assert.Equal(t, "Y1rwA3Hl4PGoIepYTPpPF5TKRtJv8IxBILKQtk1buzQ=", hash)
}
