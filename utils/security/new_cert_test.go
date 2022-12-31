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

package security_test

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/piprate/metalocker/utils/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustRemoveAll(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		panic(err)
	}
}

func TestGenerateCertificate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatalf("TempDir: %v", err)
	}
	defer mustRemoveAll(tempDir)

	hosts := make([]string, 0)
	validFrom, _ := time.Parse("Jan 2 15:04:05 2006", "Aug 2 22:46:05 2014")
	validFor := 24 * time.Hour

	err = GenerateCertificate(2048, hosts, validFrom, validFor, tempDir)
	if err != nil {
		t.Fatalf("Failed GenerateCertificate call: %v", err)
	}

	// check files exist

	certPath := filepath.Join(tempDir, "cert.pem")
	keyPath := filepath.Join(tempDir, "key.pem")
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Fatalf("Certificate file not created: %v", certPath)
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Fatalf("Key file not created: %v", keyPath)
	}

	// check some certificate fields

	certFile, err := os.Open(certPath)
	require.NoError(t, err)
	defer certFile.Close()

	buf := make([]byte, 2048)
	_, _ = certFile.Read(buf)
	block, _ := pem.Decode(buf)

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("Error parsing certificate: %v", err)
	}

	assert.Equal(t, cert.NotBefore, validFrom)

	expectedNotAfter, _ := time.Parse("Jan 2 15:04:05 2006", "Aug 3 22:46:05 2014")
	assert.Equal(t, cert.NotAfter, expectedNotAfter)

	assert.True(t, cert.IsCA)
}
