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
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"io"
	"os"
	"strings"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/piprate/metalocker/utils/fingerprint"
	"github.com/rs/zerolog/log"
)

// NewAssetID generates a new random asset ID.
func NewAssetID(method string) string {
	idBuf := make([]byte, 32)
	_, err := rand.Read(idBuf)
	if err != nil {
		panic(err)
	}
	return BuildDIDPrefix(method) + base58.Encode(idBuf)
}

// BuildDigitalAssetID creates a new instance for Digital Asset from a pre-calculated fingerprint.
func BuildDigitalAssetID(data []byte, fingerprintAlgorithm, didMethod string) (string, error) {
	fp, err := fingerprint.GetFingerprint(bytes.NewReader(data), fingerprintAlgorithm)
	if err != nil {
		return "", err
	}
	return BuildDigitalAssetIDWithFingerprint(fp, didMethod), nil
}

// BuildDigitalAssetIDFromReader creates a new instance for Digital Asset from a pre-calculated fingerprint.
func BuildDigitalAssetIDFromReader(r io.Reader, fingerprintAlgorithm, didMethod string) (string, error) {
	fp, err := fingerprint.GetFingerprint(r, fingerprintAlgorithm)
	if err != nil {
		return "", err
	}
	return BuildDigitalAssetIDWithFingerprint(fp, didMethod), nil
}

// BuildDigitalAssetIDFromFile creates a new instance for Digital Asset
func BuildDigitalAssetIDFromFile(filename, fingerprintAlgorithm, didMethod string) (string, string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	fp, err := fingerprint.GetFingerprint(file, fingerprintAlgorithm)
	if err != nil {
		return "", "", err
	}

	assetID := BuildDigitalAssetIDWithFingerprint(fp, didMethod)

	return assetID, base64.StdEncoding.EncodeToString(fp), nil
}

// BuildDigitalAssetIDWithFingerprint creates a new instance for Digital Asset from a pre-calculated fingerprint.
func BuildDigitalAssetIDWithFingerprint(fp []byte, didMethod string) string {
	prefix := BuildDIDPrefix(didMethod)
	hash := Hash(prefix, fp)
	return prefix + base58.Encode(hash)
}

func VerifyDigitalAssetID(id, fingerprintAlgorithm string, data []byte) (bool, error) {
	method, err := ExtractDIDMethod(id)
	if err != nil {
		return false, err
	}

	controlID, err := BuildDigitalAssetID(data, fingerprintAlgorithm, method)
	if err != nil {
		return false, err
	}

	if controlID != id {
		log.Error().Str("expected", controlID).Str("actual", id).Msg("Invalid digital asset ID")
		return false, nil
	}

	return true, nil
}

// UnwrapDigitalAssetID removes 'did:method:' component from the given DID
func UnwrapDigitalAssetID(id string) string {
	return id[strings.LastIndex(id, ":")+1:]
}
