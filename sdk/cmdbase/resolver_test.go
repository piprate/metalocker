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
	"github.com/knadh/koanf"
	. "github.com/piprate/metalocker/sdk/cmdbase"
	"github.com/piprate/metalocker/utils/hv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"testing"
)

func TestNewSecureParameterResolver_ResolveString(t *testing.T) {
	rootCfg := koanf.New(".")

	var hcv *hv.HCVaultClient
	var err error

	res := NewSecureParameterResolver(hcv, nil)

	// test simple string

	_ = rootCfg.Set("string", "val")

	val, err := res.ResolveString(rootCfg.Get("string"))
	require.NoError(t, err)
	assert.Equal(t, "val", val)

	// test Side Config parameter (first level)

	_ = rootCfg.Set("side_param.type", EPTypeSideConfig)
	_ = rootCfg.Set("side_param.cfg", "test")
	_ = rootCfg.Set("side_param.key", "field1")

	sideCfg1 := koanf.New(".")
	_ = sideCfg1.Set("field1", "val1")

	res.AddSideConfig("test", sideCfg1)

	val, err = res.ResolveString(rootCfg.Get("side_param"))
	require.NoError(t, err)
	assert.Equal(t, "val1", val)

	// test Side Config parameter (multi-level)

	_ = rootCfg.Set("side_param.type", EPTypeSideConfig)
	_ = rootCfg.Set("side_param.cfg", "test")
	_ = rootCfg.Set("side_param.key", "level1.level2.field2")

	_ = sideCfg1.Set("level1.level2.field2", "val2")

	val, err = res.ResolveString(rootCfg.Get("side_param"))
	require.NoError(t, err)
	assert.Equal(t, "val2", val)

	// test Vault parameter (Vault is missing)

	_ = rootCfg.Set("vault_param.type", "VaultParam")
	_ = rootCfg.Set("vault_param.path", "piprate/metalocker/test")
	_ = rootCfg.Set("vault_param.key", "field3")

	_, err = res.ResolveString(rootCfg.Get("vault_param"))
	assert.Error(t, err)

	// test direct value provision (string)

	val, err = res.ResolveString("val1")
	require.NoError(t, err)
	assert.Equal(t, "val1", val)

	// test direct value provision (map)

	val, err = res.ResolveString(map[string]any{
		"type": EPTypeSideConfig,
		"cfg":  "test",
		"key":  "field1",
	})
	require.NoError(t, err)
	assert.Equal(t, "val1", val)

	// test no value

	val, err = res.ResolveString(nil)
	require.NoError(t, err)
	assert.Equal(t, "", val)
}
