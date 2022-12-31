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
	offerIDPrefix      = ""
	prospectusIDPrefix = ""
)

type Criteria struct {
	ID      string           `json:"id"`
	Type    string           `json:"type"`
	Name    string           `json:"name,omitempty"`
	Version string           `json:"version,omitempty"`
	Params  []map[string]any `json:"parameters,omitempty"`
}

type KnowledgeQuery struct {
	Context  any        `json:"@context"`
	ID       string     `json:"id"`
	Type     string     `json:"type"`
	Creator  string     `json:"creator,omitempty"`
	Created  *time.Time `json:"created,omitempty"`
	Criteria *Criteria  `json:"criteria,omitempty"`
	Proof    *Proof     `json:"proof,omitempty"`
}

type OfferTerms struct {
	Duration int64 `json:"duration,omitempty"`
}

type KnowledgeOffer struct {
	ID               string        `json:"id"`
	Type             string        `json:"type"`
	Sender           string        `json:"sender"`
	Recipient        string        `json:"recipient"`
	Expires          *time.Time    `json:"expires,omitempty"`
	Asset            any           `json:"asset"`
	DatasetType      string        `json:"datasetType"`
	DatasetPreview   any           `json:"datasetPreview,omitempty"`
	RevisionNumber   int64         `json:"revisionNumber,omitempty"`
	WasRevisionOf    string        `json:"wasRevisionOf,omitempty"`
	SpecializationOf string        `json:"specializationOf,omitempty"`
	Terms            []*OfferTerms `json:"terms"`
	Proof            *Proof        `json:"proof,omitempty"`
}

func NewKnowledgeOffer(body []byte) (*KnowledgeOffer, error) {
	var ko KnowledgeOffer

	if err := jsonw.Unmarshal(body, &ko); err != nil {
		return nil, err
	}

	return &ko, nil
}

func (ko *KnowledgeOffer) Bytes() []byte {
	b, _ := jsonw.Marshal(ko)
	return b
}

func (ko *KnowledgeOffer) MerkleSign(identity string, key ed25519.PrivateKey) error {
	signable, _ := NewSignableDocument(ko.Bytes())

	// sign and update document ID to make it content addressable
	id, proof, err := signable.MerkleSign(offerIDPrefix, identity, key)
	if err != nil {
		return err
	}

	ko.ID = id
	ko.Proof = proof

	return nil
}

func (ko *KnowledgeOffer) MerkleVerify(key ed25519.PublicKey) (bool, error) {
	doc, _ := NewSignableDocument(ko.Bytes())
	return doc.MerkleVerify(offerIDPrefix, key)
}

type KnowledgeProspectus struct {
	Context any               `json:"@context"`
	ID      string            `json:"id"`
	Type    string            `json:"type"`
	Creator string            `json:"creator,omitempty"`
	Created *time.Time        `json:"created,omitempty"`
	Query   *KnowledgeQuery   `json:"query,omitempty"`
	Offers  []*KnowledgeOffer `json:"offers,omitempty"`
	Proof   *Proof            `json:"proof,omitempty"`
}

func NewKnowledgeProspectus(body []byte) (*KnowledgeProspectus, error) {
	var kp KnowledgeProspectus

	if err := jsonw.Unmarshal(body, &kp); err != nil {
		return nil, err
	}

	return &kp, nil
}

func (kp *KnowledgeProspectus) Bytes() []byte {
	b, _ := jsonw.Marshal(kp)
	return b
}

func (kp *KnowledgeProspectus) MerkleSign(identity string, key ed25519.PrivateKey) error {
	signable, _ := NewSignableDocument(kp.Bytes())

	// sign and update document ID to make it content addressable
	id, proof, err := signable.MerkleSign(prospectusIDPrefix, identity, key)
	if err != nil {
		return err
	}

	kp.ID = id
	kp.Proof = proof

	return nil
}

func (kp *KnowledgeProspectus) MerkleVerify(key ed25519.PublicKey) (bool, error) {
	doc, _ := NewSignableDocument(kp.Bytes())
	return doc.MerkleVerify(prospectusIDPrefix, key)
}

type QuotedFromEntity struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	WasQuotedFrom string `json:"wasQuotedFrom"`
}

type Confirmation struct {
	Type       string              `json:"type"`
	Offer      *KnowledgeOffer     `json:"offer"`
	Terms      []*OfferTerms       `json:"terms,omitempty"`
	Provenance []*QuotedFromEntity `json:"provenance,omitempty"`
}

type KnowledgeSharingRequest struct {
	Context       any                  `json:"@context"`
	ID            string               `json:"id"`
	Type          string               `json:"type"`
	Creator       string               `json:"creator,omitempty"`
	Created       *time.Time           `json:"created,omitempty"`
	Prospectus    *KnowledgeProspectus `json:"prospectus,omitempty"`
	Locker        string               `json:"locker,omitempty"`
	Vault         string               `json:"vault,omitempty"`
	Confirmations []*Confirmation      `json:"confirmations"`
	Proof         *Proof               `json:"proof,omitempty"`
}

type KnowledgeSharingResult struct {
	RecordID     string `json:"recordId"`
	ImpressionID string `json:"impressionId"`
	Payload      any    `json:"payload,omitempty"`
}
