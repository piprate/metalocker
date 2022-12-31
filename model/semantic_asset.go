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
	"io"
	"os"
	"strings"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/piprate/metalocker/utils/fingerprint"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog/log"
)

/*
	NOTE: Semantic assets are a part of the original Piprate platform prototype. Currently, we don't use them.
	We left the code here in case if it's useful in the future, particularly the idea of Value Assets
	as a way to identify results of a parameterised function.
*/

const (
	NonceLength = 32
)

type SemanticAsset struct {
	Context any    `json:"@context,omitempty"`
	ID      string `json:"id"`
	Type    any    `json:"type"`
	Nonce   string `json:"nonce,omitempty"`

	Serial     bool `json:"isSerial,omitempty"`
	IsIdentity bool `json:"isIdentity,omitempty"`

	IsDigital            bool   `json:"isDigital,omitempty"`
	Fingerprint          string `json:"fingerprint,omitempty"`
	FingerprintAlgorithm string `json:"fingerprintAlgorithm,omitempty"`

	WasGeneratedBy map[string]any `json:"wasGeneratedBy,omitempty"`

	Proof *Proof `json:"proof,omitempty"`
}

func GenerateNewSemanticAsset(serial, isIdentity bool, didMethod, nonce string) (*SemanticAsset, error) {
	if nonce == "" {
		nonceBuffer := make([]byte, NonceLength)
		_, err := rand.Read(nonceBuffer)
		if err != nil {
			return nil, err
		}
		nonce = base58.Encode(nonceBuffer)
	}

	a := &SemanticAsset{
		Type:       "Asset",
		Nonce:      nonce,
		Serial:     serial,
		IsIdentity: isIdentity,
	}
	if err := a.MerkleSetID(didMethod); err != nil {
		return nil, err
	}

	return a, nil
}

func GenerateValueAsset(functionID string, entityArgs, valueArgs map[string]any, didMethod string) (*SemanticAsset, error) {
	qualifiedUsageList := make([]map[string]any, 0)
	for argID, val := range entityArgs {
		qualifiedUsageList = append(qualifiedUsageList, map[string]any{
			"type": "Usage",
			"entity": map[string]any{
				"id":   val,
				"type": "Entity",
			},
			"hadRole": map[string]any{
				"label": argID,
				"type":  "Role",
			},
		})
	}
	for argID, val := range valueArgs {
		qualifiedUsageList = append(qualifiedUsageList, map[string]any{
			"type": "Usage",
			"entity": map[string]any{
				"type":  "Entity",
				"value": val,
			},
			"hadRole": map[string]any{
				"label": argID,
				"type":  "Role",
			},
		})
	}
	a := &SemanticAsset{
		Type: []string{"ValueAsset", "Entity"},
		WasGeneratedBy: map[string]any{
			"algorithm":      functionID,
			"type":           "Activity",
			"qualifiedUsage": qualifiedUsageList,
		},
	}
	if err := a.MerkleSetID(didMethod); err != nil {
		return nil, err
	}

	return a, nil
}

// GenerateNewSemanticDigitalAsset creates a new instance for Digital Asset
func GenerateNewSemanticDigitalAsset(data []byte, fingerprintAlgorithm, didMethod string) (*SemanticAsset, error) {
	return GenerateNewSemanticDigitalAssetFromReader(bytes.NewReader(data), fingerprintAlgorithm, didMethod)
}

// GenerateNewSemanticDigitalAssetFromReader creates a new instance for Digital Asset from io.Reader
func GenerateNewSemanticDigitalAssetFromReader(r io.Reader, fingerprintAlgorithm, didMethod string) (*SemanticAsset, error) {
	fp, err := fingerprint.GetFingerprint(r, fingerprintAlgorithm)
	if err != nil {
		return nil, err
	}

	return GenerateNewSemanticDigitalAssetWithHash(fp, fingerprintAlgorithm, didMethod)
}

// GenerateNewSemanticDigitalAssetWithHash creates a new instance for Digital Asset from pre-calculated hash.
func GenerateNewSemanticDigitalAssetWithHash(fp []byte, fingerprintAlgorithm, didMethod string) (*SemanticAsset, error) {
	a := &SemanticAsset{
		Type:                 "Asset",
		IsDigital:            true,
		Fingerprint:          base58.Encode(fp),
		FingerprintAlgorithm: fingerprintAlgorithm,
		Serial:               false,
		IsIdentity:           false,
	}
	if err := a.MerkleSetID(didMethod); err != nil {
		return nil, err
	}

	return a, nil
}

// GenerateNewSemanticDigitalAssetFromFile creates a new instance for Digital Asset
func GenerateNewSemanticDigitalAssetFromFile(filename, fingerprintAlgorithm, didMethod string) (*SemanticAsset, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return GenerateNewSemanticDigitalAssetFromReader(file, fingerprintAlgorithm, didMethod)
}

func VerifySemanticDigitalAssetID(id, fingerprintAlgorithm string, data []byte) (bool, error) {
	method, err := ExtractDIDMethod(id)
	if err != nil {
		return false, err
	}

	a, err := GenerateNewSemanticDigitalAsset(data, fingerprintAlgorithm, method)
	if err != nil {
		return false, err
	}

	if a.ID != id {
		log.Error().Str("expected", a.ID).Str("actual", id).Msg("Invalid digital asset ID")
		return false, nil
	}

	return true, nil
}

func (sa *SemanticAsset) MerkleSetID(didMethod string) error {
	b, err := jsonw.Marshal(sa)
	if err != nil {
		panic("Failed to marshal asset JSON: " + err.Error())
	}
	doc, _ := NewSignableDocument(b)
	doc.SetContext(PiprateContextURL)

	// update document ID to make it content addressable
	prefix := BuildDIDPrefix(didMethod)

	id, err := doc.MerkleSetID(prefix)
	if err != nil {
		return err
	}

	sa.ID = id

	return nil
}

func (sa *SemanticAsset) MerkleVerify() (bool, error) {
	b, _ := jsonw.Marshal(sa)

	doc, _ := NewSignableDocument(b)
	doc.SetContext(PiprateContextURL)

	prefix := sa.ID[:strings.LastIndex(sa.ID, ":")+1]

	// update document ID to make it content addressable
	id, err := doc.MerkleSetID(prefix)
	if err != nil {
		return false, err
	}

	result := sa.ID == id

	if !result {
		log.Error().Str("expected", sa.ID).Str("actual", id).
			Msg("Asset merkle ID verification failed")
	}

	return result, nil
}
