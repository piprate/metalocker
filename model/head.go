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

import (
	"strings"

	"github.com/btcsuite/btcd/btcutil/base58"
)

func HeadID(assetID string, lockerID string, sender *LockerParticipant, headName string) string {
	data := strings.Join([]string{assetID, lockerID, sender.SharedSecret}, "|")
	return base58.Encode(Hash(headName, []byte(data)))
}

func PackHeadBody(assetID, lockerID, participantID, name, recordID string) []byte {
	return []byte(strings.Join([]string{assetID, lockerID, participantID, name, recordID}, "|"))
}

func UnpackHeadBody(val []byte) (string, string, string, string, string) {
	parts := strings.Split(string(val), "|")
	return parts[0], parts[1], parts[2], parts[3], parts[4]
}
