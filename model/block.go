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

package model

// Block defines a block of the MetaLocker ledger.
// Blocks are identified by their sequential numbers, starting with 0.
// Hash and ParentHash fields allow connecting a specific block
// with the underlying block implementation.
type Block struct {
	Number     int64  `json:"number"`
	Hash       string `json:"hash"`
	ParentHash string `json:"parentHash,omitempty"`
}
