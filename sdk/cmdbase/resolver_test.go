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

package cmdbase_test

import (
	. "github.com/piprate/metalocker/sdk/cmdbase"
	"github.com/piprate/metalocker/utils/hv"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"testing"
)

func TestNewSecureParameterResolver_ResolveString(t *testing.T) {
	rootViper := viper.New()

	var hcv *hv.HCVaultClient
	var err error

	res := NewSecureParameterResolver(hcv, nil)

	// test simple string

	rootViper.Set("string", "val")

	val, err := res.ResolveString(rootViper.Get("string"))
	require.NoError(t, err)
	assert.Equal(t, "val", val)

	// test Viper parameter (first level)

	rootViper.Set("viper_param.type", "ViperParam")
	rootViper.Set("viper_param.viper", "test")
	rootViper.Set("viper_param.key", "field1")

	sideViper1 := viper.New()
	sideViper1.Set("field1", "val1")

	res.AddSideViper("test", sideViper1)

	val, err = res.ResolveString(rootViper.Get("viper_param"))
	require.NoError(t, err)
	assert.Equal(t, "val1", val)

	// test Viper parameter (multi-level)

	rootViper.Set("viper_param.type", "ViperParam")
	rootViper.Set("viper_param.viper", "test")
	rootViper.Set("viper_param.key", "level1.level2.field2")

	sideViper1.Set("level1.level2.field2", "val2")

	val, err = res.ResolveString(rootViper.Get("viper_param"))
	require.NoError(t, err)
	assert.Equal(t, "val2", val)

	// test Vault parameter (Vault is missing)

	rootViper.Set("vault_param.type", "VaultParam")
	rootViper.Set("vault_param.path", "piprate/metalocker/test")
	rootViper.Set("vault_param.key", "field3")

	_, err = res.ResolveString(rootViper.Get("vault_param"))
	assert.Error(t, err)

	// test direct value provision (string)

	val, err = res.ResolveString("val1")
	require.NoError(t, err)
	assert.Equal(t, "val1", val)

	// test direct value provision (map)

	val, err = res.ResolveString(map[string]any{
		"type":  "ViperParam",
		"viper": "test",
		"key":   "field1",
	})
	require.NoError(t, err)
	assert.Equal(t, "val1", val)

	// test no value

	val, err = res.ResolveString(nil)
	require.NoError(t, err)
	assert.Equal(t, "", val)
}
