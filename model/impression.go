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
	"crypto/ed25519"
	"time"

	"github.com/piprate/metalocker/utils/jsonw"
)

const (
	impressionIDPrefix = ""
)

// MetaResource provides Impression with a link to its meta resource, the root document of
// the underlying dataset.
type MetaResource struct {
	// Asset is the meta resource's asset ID.
	Asset string `json:"id,omitempty"`
	// ContentType is the semantic type of the meta resource (and of the dataset).
	ContentType string `json:"contentType,omitempty"`
	// Fingerprint is the digital fingerprint of the meta resource. Because MetaResource
	// is signed as a part of impression, this fingerprint can verify if the meta resource
	// has been altered in any way.
	Fingerprint string `json:"fingerprint,omitempty"`
	// FingerprintAlgorithm is the Fingerprint's algorithm.
	FingerprintAlgorithm string `json:"fingerprintAlgorithm,omitempty"`
}

// Impression is a semantic definition of a dataset that contains verifiable information
// about its provenance, authorship, relation to other datasets, including revision data.
// Impression is signed by its creator using JSON-LD Signature scheme.
type Impression struct {
	Context         any           `json:"@context"`
	ID              string        `json:"id"`
	Type            []string      `json:"type"`
	Asset           string        `json:"asset,omitempty"`
	ProvGraph       any           `json:"graph,omitempty"`
	WasAttributedTo string        `json:"wasAttributedTo,omitempty"`
	GeneratedAtTime *time.Time    `json:"generatedAtTime,omitempty"`
	MetaResource    *MetaResource `json:"resource,omitempty"`

	RevisionNumber   int64  `json:"revisionNumber,omitempty"`
	RevisionMessage  string `json:"revisionMessage,omitempty"`
	WasRevisionOf    string `json:"wasRevisionOf,omitempty"`
	SpecializationOf string `json:"specializationOf,omitempty"`

	Proof *Proof `json:"proof,omitempty"`
}

func NewImpression(body []byte) (*Impression, error) {
	var imp Impression

	if err := jsonw.Unmarshal(body, &imp); err != nil {
		return nil, err
	}

	return &imp, nil
}

func NewBlankImpression() *Impression {
	return &Impression{
		Context: PiprateContextURL,
		Type:    []string{"Impression", "Entity", "Bundle"},
	}
}

func (ii *Impression) Copy() *Impression {
	cpy, _ := NewImpression(ii.Bytes())
	return cpy
}

func (ii *Impression) Revision() int64 {
	if ii.RevisionNumber == 0 {
		return 1
	} else {
		return ii.RevisionNumber
	}
}

func (ii *Impression) GetVariantID() string {
	if ii.SpecializationOf != "" {
		return ii.SpecializationOf
	} else {
		return ii.ID
	}
}

func (ii *Impression) IsRoot() bool {
	return ii.SpecializationOf == ""
}

func (ii *Impression) RevisionOf() string {
	return ii.WasRevisionOf
}

func (ii *Impression) GetProvenance(resourceID string) any {
	if ii.ProvGraph != nil {
		provList, ok := ii.ProvGraph.([]map[string]any)
		if ok {
			for _, p := range provList {
				if id, found := p["id"]; found && id == resourceID {
					return p
				}
			}
		}
	}
	return nil
}

func (ii *Impression) Bytes() []byte {
	b, _ := jsonw.Marshal(ii)
	return b
}

func (ii *Impression) Compact() ([]byte, error) {
	return CompactDocument(ii.Bytes(), PiprateContextURL)
}

func (ii *Impression) IsSigned() bool {
	return ii.Proof != nil
}

func (ii *Impression) MerkleSign(identity string, key ed25519.PrivateKey) error {
	signable, _ := NewSignableDocument(ii.Bytes())

	// sign and update document ID to make it content addressable
	id, proof, err := signable.MerkleSign(impressionIDPrefix, identity, key)
	if err != nil {
		return err
	}

	ii.ID = id
	ii.Proof = proof

	return nil
}

func (ii *Impression) MerkleVerify(key ed25519.PublicKey) (bool, error) {
	doc, _ := NewSignableDocument(ii.Bytes())
	return doc.MerkleVerify(impressionIDPrefix, key)
}
