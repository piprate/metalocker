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

package expiry

import "time"

func Years(years int) time.Time {
	return time.Now().UTC().AddDate(years, 0, 0)
}

func Months(months int) time.Time {
	return time.Now().UTC().AddDate(0, months, 0)
}

func Days(days int) time.Time {
	return time.Now().UTC().AddDate(0, 0, days)
}

func Never() time.Time {
	return time.Time{}
}
