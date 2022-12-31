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
	"fmt"
	"io"
)

func GetFingerprint(data io.Reader, algo string) ([]byte, error) {
	algoFunc, supportedAlgo := METHODS[algo]
	if !supportedAlgo {
		return nil, fmt.Errorf("unsupported data fingerprint algorithm: %s", algo)
	}
	return algoFunc(data)
}
