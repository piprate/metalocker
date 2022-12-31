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

package dataset_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/piprate/metalocker/model"
)

func BenchmarkJSON_Load_Sonic(b *testing.B) {
	b.ReportAllocs()

	lease, err := os.ReadFile("testdata/builder/_results/prov_file_lease.json")
	if err != nil {
		panic("failed to read test file")
	}

	for i := 0; i < b.N; i++ {
		_, _ = sonic.Get(lease)
	}
}

func BenchmarkJSON_AttributeRead_Sonic(b *testing.B) {
	b.ReportAllocs()

	lease, err := os.ReadFile("testdata/builder/_results/prov_file_lease.json")
	if err != nil {
		panic("failed to read test file")
	}

	root, _ := sonic.Get(lease)

	for i := 0; i < b.N; i++ {
		_, _ = root.Get("id").String()
	}
}

func BenchmarkJSON_Unmarshal_Sonic(b *testing.B) {
	b.ReportAllocs()

	contents, err := os.ReadFile("testdata/builder/_results/prov_file_lease.json")
	if err != nil {
		panic("failed to read test file")
	}

	for i := 0; i < b.N; i++ {
		var lease model.Lease
		err = sonic.Unmarshal(contents, &lease)
		if err != nil {
			panic("failed to unmarshal")
		}
	}
}

func BenchmarkJSON_Unmarshal_SonicFast(b *testing.B) {
	b.ReportAllocs()

	contents, err := os.ReadFile("testdata/builder/_results/prov_file_lease.json")
	if err != nil {
		panic("failed to read test file")
	}

	for i := 0; i < b.N; i++ {
		var lease model.Lease
		err = sonic.ConfigFastest.Unmarshal(contents, &lease)
		if err != nil {
			panic("failed to unmarshal")
		}
	}
}

func BenchmarkJSON_Unmarshal_JSON(b *testing.B) {
	b.ReportAllocs()

	contents, err := os.ReadFile("testdata/builder/_results/prov_file_lease.json")
	if err != nil {
		panic("failed to read test file")
	}

	for i := 0; i < b.N; i++ {
		var lease model.Lease
		err = json.Unmarshal(contents, &lease)
		if err != nil {
			panic("failed to unmarshal")
		}
	}
}

func BenchmarkJSON_Unmarshal_Sonic_PreTouch(b *testing.B) {
	b.ReportAllocs()

	contents, err := os.ReadFile("testdata/builder/_results/prov_file_lease.json")
	if err != nil {
		panic("failed to read test file")
	}

	var pt model.Lease

	err = sonic.Pretouch(reflect.TypeOf(pt))
	if err != nil {
		panic("failed to pre-touch")
	}

	for i := 0; i < b.N; i++ {
		var lease model.Lease
		err = sonic.Unmarshal(contents, &lease)
		if err != nil {
			panic("failed to unmarshal")
		}
	}
}

func BenchmarkJSON_Marshal_Sonic(b *testing.B) {
	b.ReportAllocs()

	contents, err := os.ReadFile("testdata/builder/_results/prov_file_lease.json")
	if err != nil {
		panic("failed to read test file")
	}

	var lease model.Lease
	err = sonic.Unmarshal(contents, &lease)
	if err != nil {
		panic("failed to unmarshal")
	}

	for i := 0; i < b.N; i++ {
		_, _ = sonic.Marshal(lease)
	}
}

func BenchmarkJSON_Marshal_SonicFast(b *testing.B) {
	b.ReportAllocs()

	contents, err := os.ReadFile("testdata/builder/_results/prov_file_lease.json")
	if err != nil {
		panic("failed to read test file")
	}

	var lease model.Lease
	err = sonic.ConfigFastest.Unmarshal(contents, &lease)
	if err != nil {
		panic("failed to unmarshal")
	}

	for i := 0; i < b.N; i++ {
		_, _ = sonic.Marshal(lease)
	}
}

func BenchmarkJSON_Marshal_JSON(b *testing.B) {
	b.ReportAllocs()

	contents, err := os.ReadFile("testdata/builder/_results/prov_file_lease.json")
	if err != nil {
		panic("failed to read test file")
	}

	var lease model.Lease
	err = json.Unmarshal(contents, &lease)
	if err != nil {
		panic("failed to unmarshal")
	}

	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(lease)
	}
}
