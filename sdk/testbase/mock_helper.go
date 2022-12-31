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

import "github.com/golang/mock/gomock"

type CustomFn func(x any) bool

// DynamicMatcher can be used with gomock to check for conditions dynamically
// Standard gomock matchers offer a limited set of comparison options
type DynamicMatcher struct {
	// description describes what the matcher matches.
	description string

	fn CustomFn
}

var _ gomock.Matcher = (*DynamicMatcher)(nil)

func NewDynamicMatcher(fn CustomFn, desc string) *DynamicMatcher {
	return &DynamicMatcher{
		description: desc,
		fn:          fn,
	}
}

func (dm DynamicMatcher) Matches(x any) bool {
	return dm.fn(x)
}

func (dm DynamicMatcher) String() string {
	return dm.description
}
