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

package streams_test

import (
	"bytes"
	"encoding/base64"
	"io"
	"sync"
	"testing"

	. "github.com/piprate/metalocker/utils/streams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleStreamStatsWriter(t *testing.T) {
	message := []byte("message")
	msgReader := bytes.NewReader(message)

	ssw := NewStreamStatsWriter()

	writer := bytes.NewBuffer([]byte{})

	mw := io.MultiWriter(ssw, writer)

	_, err := io.Copy(mw, msgReader)
	require.NoError(t, err)

	stats := ssw.Stats()

	assert.Equal(t, int64(7), stats.Size)
	assert.Equal(t, "q1MKE+RZFJgrefm34/uplM/R8/si9xzqGvvwK0YMbR0=", base64.StdEncoding.EncodeToString(stats.SHA256Hash))
	assert.Equal(t, "text/plain; charset=utf-8", stats.ContentType)
}

func TestPipedStreamStatsWriter(t *testing.T) {
	message := []byte("message")
	msgReader := bytes.NewReader(message)

	pr, pw := io.Pipe()

	ssw := NewStreamStatsWriter()

	writer := bytes.NewBuffer([]byte{})

	mw := io.MultiWriter(ssw, pw)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	var copyErr error
	go func() {

		defer pw.Close()

		if _, copyErr = io.Copy(mw, msgReader); copyErr != nil {
			return
		}

		wg.Done()
	}()

	// use io.Reader interface (typically, in some external function that expects an io.Reader).

	_, err := io.Copy(writer, pr)
	require.NoError(t, err)

	// wait until the whole stream is read

	wg.Wait()

	require.Nil(t, copyErr)

	stats := ssw.Stats()

	assert.Equal(t, int64(7), stats.Size)
	assert.Equal(t, "q1MKE+RZFJgrefm34/uplM/R8/si9xzqGvvwK0YMbR0=", base64.StdEncoding.EncodeToString(stats.SHA256Hash))
	assert.Equal(t, "text/plain; charset=utf-8", stats.ContentType)
}
