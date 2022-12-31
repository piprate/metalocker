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

package fingerprint

import (
	"crypto/sha256"
	"io"
)

const (
	AlgoSha256 = "fingerprints:sha256"
)

func GetSha256Fingerprint(r io.Reader) ([]byte, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, r); err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}
