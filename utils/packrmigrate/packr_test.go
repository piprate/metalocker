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

package packrmigrate_test

import (
	"testing"

	"github.com/gobuffalo/packr/v2"
	st "github.com/golang-migrate/migrate/v4/source/testing"
	. "github.com/piprate/metalocker/utils/packrmigrate"
)

func Test(t *testing.T) {
	// wrap assets into Resource first
	Box := packr.New("testbox", "./testdata/")

	d, err := WithInstance(Box)
	if err != nil {
		t.Fatal(err)
	}
	st.Test(t, d)
}

func TestOpen(t *testing.T) {
	b := &Packr{}
	_, err := b.Open("")
	if err == nil {
		t.Fatal("expected err, because it's not implemented yet")
	}
}
