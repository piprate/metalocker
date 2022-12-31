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

package bolt_test

import (
	"errors"
	"os"
	"testing"

	"github.com/piprate/metalocker/index"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/stretchr/testify/require"
)

func TestIndex_AddLockerState(t *testing.T) {
	store, dir := newTestIndexStore(t)
	defer func() {
		_ = store.Close()
		_ = os.RemoveAll(dir)
	}()

	userID := "did:piprate:QgH6CZvhjTUFvCbRUw4N6Z"

	ix, err := store.CreateIndex(userID, index.TypeRoot, model.AccessLevelHosted)
	require.NoError(t, err)

	iw, _ := ix.Writer()

	locker := testbase.TestUniLocker(t)

	err = iw.AddLockerState(userID, locker.ID, locker.FirstBlock)
	require.NoError(t, err)

	// try adding the same locker again

	err = iw.AddLockerState(userID, locker.ID, locker.FirstBlock)
	require.Error(t, err)
	require.True(t, errors.Is(err, index.ErrLockerStateExists))
}
