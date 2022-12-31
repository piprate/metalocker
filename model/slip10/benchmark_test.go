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

package slip10_test

import (
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	. "github.com/piprate/metalocker/model/slip10"
)

func BenchmarkNode_Derive(b *testing.B) {
	seed, _ := hex.DecodeString("fffcf9f6f3f0edeae7e4e1dedbd8d5d2cfccc9c6c3c0bdbab7b4b1aeaba8a5a29f9c999693908d8a8784817e7b7875726f6c696663605d5a5754514e4b484542")
	node, _ := NewMasterNode(seed)

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		childNode, err := node.Derive(uint32(i) + FirstHardenedIndex)
		if err != nil {
			b.Fail()
			break
		}
		childNode.KeyPair()
	}
}

func BenchmarkHDKeyChain_Child(b *testing.B) {
	privKey, _ := hdkeychain.NewKeyFromString("xprv9s21ZrQH143K3QTDL4LXw2F7HEK3wJUD2nW2nRk4stbPy6cq3jPPqjiChkVvvNKmPGJxWUtg6LnF5kejMRNNU3TGtRBeJgk33yuGBxrMPHi")

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		extKey, err := privKey.Derive(uint32(i))
		if err != nil {
			b.Fail()
			break
		}
		_, err = extKey.ECPubKey()
		if err != nil {
			b.Fail()
			break
		}
	}
}
