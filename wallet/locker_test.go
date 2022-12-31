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

package wallet_test

import (
	"testing"
	"time"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/dataset"
	"github.com/piprate/metalocker/model/expiry"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLockerWrapper_Store(t *testing.T) {
	env := testbase.SetUpTestEnvironment(t)
	defer func() { _ = env.Close() }()

	dw := testHostedAccount(t, env, true)

	idy, err := dw.NewIdentity(model.AccessLevelHosted, "John XXX")
	require.NoError(t, err)

	locker, err := idy.NewLocker(idy.Name())
	require.NoError(t, err)

	f := locker.Store(map[string]string{
		"type": "Map",
		"name": "Test Dataset",
	}, expiry.FromNow("1h"), dataset.WithVault(testbase.TestVaultName))
	require.NoError(t, f.Wait(time.Second*10))

	var res map[string]string
	require.NoError(t, f.DataSet().DecodeMetaResource(&res))
	assert.Equal(t, "Test Dataset", res["name"])
}
