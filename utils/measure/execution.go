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

package measure

import (
	"time"

	"github.com/rs/zerolog/log"
)

func ExecTime(name string) func() {
	start := time.Now()
	return func() {
		log.Debug().Str("func", name).Dur("dur", time.Since(start)).Msg("Execution time")
	}
}

func CustomExecTime(name string, nowFn func() time.Time) func() {
	start := nowFn()
	return func() {
		finish := nowFn()
		log.Debug().Str("func", name).Dur("dur", finish.Sub(start)).Msg("Execution time")
	}
}
