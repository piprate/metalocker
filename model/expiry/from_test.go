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

package expiry_test

import (
	"testing"
	"time"

	. "github.com/piprate/metalocker/model/expiry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromDateErr(t *testing.T) {
	base := time.Date(2022, time.December, 26, 11, 30, 15, 0, time.UTC)
	offset, err := FromDateErr(base, "20y1m1h")
	require.NoError(t, err)
	assert.Equal(t, "2043-01-26T12:30:15Z", offset.Format(time.RFC3339Nano))

	offset, err = FromDateErr(base, "1min5s")
	require.NoError(t, err)
	assert.Equal(t, "2022-12-26T11:31:20Z", offset.Format(time.RFC3339Nano))

	offset, err = FromDateErr(base, "5s500ms")
	require.NoError(t, err)
	assert.Equal(t, "2022-12-26T11:30:20.5Z", offset.Format(time.RFC3339Nano))

	offset, err = FromDateErr(base, "0")
	require.NoError(t, err)
	assert.True(t, offset.IsZero())

	offset, err = FromDateErr(base, "never")
	require.NoError(t, err)
	assert.True(t, offset.IsZero())

	_, err = FromDateErr(base, "10.5y")
	require.Error(t, err)

	_, err = FromDateErr(base, "10.5m")
	require.Error(t, err)

	_, err = FromDateErr(base, "10.5d")
	require.Error(t, err)

	_, err = FromDateErr(base, "10yr")
	require.Error(t, err)

	_, err = FromDateErr(base, "1min5sec")
	require.Error(t, err)
}

func TestFromDate(t *testing.T) {
	base := time.Date(2022, time.December, 26, 11, 30, 15, 0, time.UTC)
	offset := FromDate(base, "1min5s")
	assert.Equal(t, "2022-12-26T11:31:20Z", offset.Format(time.RFC3339Nano))

	offset = FromDate(base, "10yr")
	require.Equal(t, BadTime, offset)
}

func TestFromNow(t *testing.T) {
	offset := FromNow("1min5s")
	assert.InDelta(t, 65.0, time.Until(offset).Seconds(), 1.0)

	offset = FromNow("10yr")
	require.Equal(t, BadTime, offset)
}
