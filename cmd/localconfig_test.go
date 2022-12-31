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

package cmd_test

import (
	"strings"
	"testing"

	. "github.com/piprate/metalocker/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetWalletPath(t *testing.T) {
	configDir := GetMetaLockerConfigDir()

	p, err := GetWalletPath("https://metalocker.example.com:4000", "1234",
		"test@example.com")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(p, configDir))
	assert.Equal(t,
		"/wallet/https___metalocker.example.com_4000/1234/test_at_example.com/data_wallet.bolt",
		p[len(configDir):],
	)
}
