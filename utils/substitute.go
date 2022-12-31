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

package utils

import (
	"fmt"
	"strings"
	"time"
)

func BuildMapFromString(mappingStr string) (map[string]string, error) {
	res := make(map[string]string)
	if mappingStr != "" {
		for _, m := range strings.Split(mappingStr, ",") {
			kv := strings.Split(m, "/")
			if len(kv) != 2 {
				return nil, fmt.Errorf("invalid mapping: %s", m)
			}
			res[kv[0]] = kv[1]
		}
	}
	return res, nil
}

func SubstituteEntities(template []byte, entityMap map[string]string, nowTimestamp *time.Time) ([]byte, error) {
	res := string(template)

	// replace %%now%% by either provided timestamp or current time
	if strings.Contains(res, "%%now%%") {
		if nowTimestamp == nil {
			n := time.Now().UTC()
			nowTimestamp = &n
		}
		res = strings.ReplaceAll(res, "%%now%%", nowTimestamp.Format(time.RFC3339Nano))
	}

	for id, val := range entityMap {
		res = strings.ReplaceAll(res, fmt.Sprintf("%%%s%%", id), val)
	}
	return []byte(res), nil
}
