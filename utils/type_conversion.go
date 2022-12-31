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
	"encoding/binary"
	"fmt"
	"strconv"
	"time"

	"github.com/piprate/metalocker/utils/jsonw"
)

func StringToTime(v string) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, v)
	return t
}

func TimeToString(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func StringToInt(v string) int {
	i, _ := strconv.Atoi(v)
	return i
}

func IntToString(v int) string {
	return strconv.Itoa(v)
}

func Int64ToString(v int64) string {
	return strconv.FormatInt(v, 10)
}

func StringToInt64(v string) int64 {
	i, _ := strconv.ParseInt(v, 10, 0)
	return i
}

func Float64ToString(v float64) string {
	return fmt.Sprintf("%1.15E", v)
}

func StringToFloat64(v string) float64 {
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

func BoolToInt(v bool) int {
	if v {
		return 1
	} else {
		return 0
	}
}

func IntToBool(v int) bool {
	return v == 1
}

func Uint32ToBytes(v uint32) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, v)
	return bs
}

func BytesToUint32(v []byte) uint32 {
	return binary.LittleEndian.Uint32(v)
}

func Int64ToBytes(v int64) []byte {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, uint64(v))
	return bs
}

func BytesToInt64(v []byte) int64 {
	return int64(binary.LittleEndian.Uint64(v))
}

func MarshalToType(val any, dest any, allowUnknownFields bool) error {
	if allowUnknownFields {
		b, err := jsonw.Marshal(val)
		if err != nil {
			return err
		}
		err = jsonw.Unmarshal(b, dest)
		if err != nil {
			return err
		}
		return nil
	} else {
		return jsonw.MarshalToTypeWithFieldValidation(val, dest)
	}
}
