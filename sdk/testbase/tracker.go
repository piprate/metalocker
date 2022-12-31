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
	"time"

	"github.com/rs/zerolog/log"
)

type (
	TestTracker struct {
		t         *testing.T
		startTime time.Time
	}
)

func StartTestTracker(t *testing.T) *TestTracker {
	t.Helper()

	now := time.Now()

	log.Warn().Msg("~~~~~ BEGIN TEST ~~~~~")

	return &TestTracker{
		t:         t,
		startTime: now,
	}
}

func (tt *TestTracker) Finish() {
	log.Warn().Msg("~~~~~ END TEST ~~~~~")
	log.Warn().Str("test", tt.t.Name()).Dur("duration", time.Since(tt.startTime)).Msg("Execution time")
}
