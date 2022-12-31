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

	mh "github.com/multiformats/go-multihash"
)

const (
	AlgoMultihash256 = "fingerprints:multihash256"
)

func GetMultihashFingerprint(r io.Reader) ([]byte, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, r); err != nil {
		return nil, err
	}
	h, err := mh.EncodeName(hasher.Sum(nil), "sha2-256")
	if err != nil {
		// this error can be safely ignored (panic) because multihash only fails
		// from the selection of hash function. If the fn + length are valid, it
		// won't error.
		panic("multihash failed to hash using SHA2_256.")
	}
	return h, nil
}
