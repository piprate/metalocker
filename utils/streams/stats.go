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

package streams

import (
	"crypto/sha256"
	"hash"

	"github.com/gabriel-vasile/mimetype"
)

// ReadLimit is the maximum number of bytes mimetype library reads
// from the input stream when detecting the type.
const ReadLimit = 3072

type StreamStats struct {
	Size        int64
	SHA256Hash  []byte
	ContentType string
}

type StreamStatsWriter struct {
	hasher hash.Hash
	head   []byte
	size   int64
}

func NewStreamStatsWriter() *StreamStatsWriter {
	return &StreamStatsWriter{
		hasher: sha256.New(),
	}
}

func (ssw *StreamStatsWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	ssw.size += int64(n)

	if len(ssw.head) < ReadLimit {
		readSize := ReadLimit - len(ssw.head)
		if len(p) < readSize {
			readSize = len(p)
		}
		ssw.head = append(ssw.head, p[:readSize]...)
	}

	_, _ = ssw.hasher.Write(p)

	return
}

func (ssw *StreamStatsWriter) Stats() *StreamStats {

	mime := mimetype.Detect(ssw.head)

	return &StreamStats{
		Size:        ssw.size,
		SHA256Hash:  ssw.hasher.Sum(nil),
		ContentType: mime.String(),
	}
}
