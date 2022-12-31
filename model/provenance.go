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
	ProvTypeAgent       = "Agent"
	ProvTypeRole        = "Role"
	ProvTypeUsage       = "Usage"
	ProvTypeActivity    = "Activity"
	ProvTypeEntity      = "Entity"
	ProvTypeAssociation = "Association"
)

type ProvAgent struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ActedOnBehalfOf string `json:"actedOnBehalfOf,omitempty"`
}

type ProvRole struct {
	Type  string `json:"type"`
	Label string `json:"label"`
}

type ProvUsage struct {
	Type    string    `json:"type"`
	Entity  any       `json:"entity"`
	HadRole *ProvRole `json:"hadRole,omitempty"`
}

type ProvAssociation struct {
	Type    string    `json:"type"`
	Agent   any       `json:"agent"`
	HadRole *ProvRole `json:"hadRole,omitempty"`
}

type ProvActivity struct {
	ID                   string             `json:"id,omitempty"`
	Type                 string             `json:"type"`
	Algorithm            string             `json:"algorithm"`
	Used                 any                `json:"used,omitempty"`
	QualifiedUsage       []*ProvUsage       `json:"qualifiedUsage,omitempty"`
	WasAssociatedWith    string             `json:"wasAssociatedWith,omitempty"`
	QualifiedAssociation []*ProvAssociation `json:"qualifiedAssociation,omitempty"`
}

type ProvPrimarySource struct {
	Type      string `json:"type"`
	Entity    any    `json:"entity"`
	Algorithm string `json:"algorithm"`
}

type ProvEntity struct {
	Context         any           `json:"@context,omitempty"`
	ID              string        `json:"id,omitempty"`
	Type            string        `json:"type"`
	WasAttributedTo string        `json:"wasAttributedTo,omitempty"`
	GeneratedAtTime *time.Time    `json:"generatedAtTime,omitempty"`
	WasGeneratedBy  *ProvActivity `json:"wasGeneratedBy,omitempty"`
	WasQuotedFrom   any           `json:"wasQuotedFrom,omitempty"`
	WasAccessibleTo any           `json:"wasAccessibleTo,omitempty"`
	ContentType     string        `json:"contentType,omitempty"`
	MentionOf       string        `json:"mentionOf,omitempty"`
	AsInBundle      string        `json:"asInBundle,omitempty"`

	Proof *Proof `json:"proof,omitempty"`
}

type ProvBundle struct {
	Context                any                  `json:"@context,omitempty"`
	ID                     string               `json:"id,omitempty"`
	Type                   string               `json:"type"`
	GeneratedAtTime        *time.Time           `json:"generatedAtTime,omitempty"`
	WasAttributedTo        string               `json:"wasAttributedTo,omitempty"`
	HadPrimarySource       string               `json:"hadPrimarySource,omitempty"`
	QualifiedPrimarySource []*ProvPrimarySource `json:"qualifiedPrimarySource,omitempty"`
	Graph                  any                  `json:"graph,omitempty"`

	Proof *Proof `json:"proof,omitempty"`
}

func (pe *ProvEntity) Bytes() []byte {
	b, _ := jsonw.Marshal(pe)
	return b
}

func (pe *ProvEntity) Copy() *ProvEntity {
	var cpy *ProvEntity
	_ = jsonw.Unmarshal(pe.Bytes(), &cpy)
	return cpy
}

func (pe *ProvEntity) MerkleSign(identity string, key ed25519.PrivateKey) error {
	signable, _ := NewSignableDocument(pe.Bytes())

	// sign and update document ID to make it content addressable
	id, proof, err := signable.MerkleSign("", identity, key)
	if err != nil {
		return err
	}

	pe.ID = id
	pe.Proof = proof

	return nil
}

func (pe *ProvEntity) MerkleVerify(key ed25519.PublicKey) (bool, error) {
	doc, _ := NewSignableDocument(pe.Bytes())
	return doc.MerkleVerify("", key)
}
