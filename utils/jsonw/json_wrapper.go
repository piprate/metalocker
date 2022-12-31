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

package jsonw

import (
	"bytes"
	"io"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/decoder"
	"github.com/bytedance/sonic/encoder"
)

var (
	Marshal   = sonic.Marshal
	Unmarshal = sonic.Unmarshal
)

func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return encoder.EncodeIndented(v, prefix, indent, 0)
}

func Decode(reader io.Reader, obj any) error {
	return decoder.NewStreamDecoder(reader).Decode(obj)
}

func Encode(val any, writer io.Writer) error {
	return encoder.NewStreamEncoder(writer).Encode(val)
}

func MarshalToTypeWithFieldValidation(val any, dest any) error {
	b, err := sonic.Marshal(val)
	if err != nil {
		return err
	}
	d := decoder.NewStreamDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()

	err = d.Decode(dest)
	if err != nil {
		return err
	}
	return nil
}
