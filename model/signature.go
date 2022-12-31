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
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/piprate/json-gold/ld"
	"github.com/piprate/metalocker/utils/jsonw"
)

const (
	// MerkleID is an ID used for signing and verification of Merkle documents.
	// We set the document ID to MerkleID before signing (because final Merkle ID of the document
	// is a hash of the document including its signature
	MerkleID = "_:merkle_root_2016"
)

var (
	// Setting this flag to true will advise Hash() function to print out normalised documents - useful for debugging
	// This is a necessary hack because logging framework doesn't allow multi-line messages
	debugMode = false
)

type (
	Proof struct {
		Type    string `json:"type"`
		Creator string `json:"creator"`
		Value   string `json:"proofValue"`
	}

	Signer interface {
		Sign(message []byte) []byte
	}

	Verifier interface {
		Verify(message, signature []byte) bool
	}

	MerkleSigner interface {
		MerkleSign(identity string, key ed25519.PrivateKey) error
	}

	MerkleVerifier interface {
		GetProof() *Proof
		MerkleVerify(key ed25519.PublicKey) (bool, error)
	}

	SignableDocument struct {
		data map[string]any
	}
)

func SetDebugMode(v bool) {
	debugMode = v
}

func NewSignableDocument(b []byte) (*SignableDocument, error) {
	p := &SignableDocument{}
	err := jsonw.Unmarshal(b, &p.data)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (dp *SignableDocument) Context() any {
	return dp.data["@context"]
}

func (dp *SignableDocument) SetContext(ctx any) {
	dp.data["@context"] = ctx
}

func (dp *SignableDocument) ID() string {
	return dp.data["id"].(string)
}

// Copy return a deep copy of the document
func (dp *SignableDocument) Copy() (*SignableDocument, error) {
	copyBytes, err := jsonw.Marshal(dp.data)
	if err != nil {
		return nil, err
	}
	return NewSignableDocument(copyBytes)
}

func (dp *SignableDocument) Hash() ([]byte, error) {
	proc := ld.NewJsonLdProcessor()

	opts := ld.NewJsonLdOptions(crvyBase)
	opts.ProcessingMode = ld.JsonLd_1_1
	opts.DocumentLoader = DefaultDocumentLoader()
	opts.Format = "application/n-quads"

	normDoc, err := proc.Normalize(dp.data, opts)
	if err != nil {
		return nil, err
	}

	if debugMode {
		println("===== start normalised doc =====")
		print(normDoc.(string))
		println("===== finish normalised doc =====")
	}

	hash32 := sha256.Sum256([]byte(normDoc.(string)))
	return hash32[:], nil
}

// MerkleSetID assigns Merkle ID to the root JSON-LD element
func (dp *SignableDocument) MerkleSetID(idPrefix string) (string, error) {
	dp.data["id"] = MerkleID

	hash, err := dp.Hash()
	if err != nil {
		return "", err
	}

	newID := idPrefix + base58.Encode(hash)

	return newID, nil
}

// Sign signs the document as per JSON-LD signatures specification.
func (dp *SignableDocument) Sign(identity string, key ed25519.PrivateKey) (*Proof, error) {
	delete(dp.data, "proof")

	hash, err := dp.Hash()
	if err != nil {
		return nil, err
	}

	sig := ed25519.Sign(key, hash)

	return &Proof{
		Type:    "Ed25519Signature2018",
		Creator: identity,
		Value:   base58.Encode(sig),
	}, nil
}

// Verify verifies document signature as per JSON-LD signatures specification and returns false if verification fails.
func (dp *SignableDocument) Verify(publicKey ed25519.PublicKey) (bool, error) {
	sigVal, hasSignature := dp.data["proof"]
	if !hasSignature {
		return false, fmt.Errorf("signature not found")
	}
	sigMap, ok := sigVal.(map[string]any)
	if !ok || sigMap["proofValue"] == nil {
		return false, fmt.Errorf("bad signature shape")
	}
	proofVal, ok := sigMap["proofValue"].(string)
	if !ok {
		return false, fmt.Errorf("bad signature type")
	}
	sig := base58.Decode(proofVal)

	dCopy, err := dp.Copy()
	if err != nil {
		return false, err
	}

	delete(dCopy.data, "proof")

	hash, err := dCopy.Hash()
	if err != nil {
		return false, err
	}

	if signatureVerified := ed25519.Verify(publicKey, hash, sig); !signatureVerified {
		return false, nil
	}

	return true, nil
}

// MerkleSign signs the document and assigns Merkle ID to the root JSON-LD element
func (dp *SignableDocument) MerkleSign(idPrefix string, identity string, key ed25519.PrivateKey) (string, *Proof, error) {
	delete(dp.data, "proof")

	dp.data["id"] = MerkleID

	hash, err := dp.Hash()
	if err != nil {
		return "", nil, err
	}

	sig := ed25519.Sign(key, hash)

	dp.data["proof"] = map[string]any{
		"type":       "Ed25519Signature2018",
		"creator":    identity,
		"proofValue": base58.Encode(sig),
	}

	proof := &Proof{
		Type:    "Ed25519Signature2018",
		Creator: identity,
		Value:   base58.Encode(sig),
	}

	hash, err = dp.Hash()
	if err != nil {
		return "", nil, err
	}

	newID := idPrefix + base58.Encode(hash)

	return newID, proof, nil
}

// MerkleVerify verifies document signature and returns false if verification fails.
func (dp *SignableDocument) MerkleVerify(idPrefix string, publicKey ed25519.PublicKey) (bool, error) {
	sigVal, hasSignature := dp.data["proof"]
	if !hasSignature {
		return false, fmt.Errorf("signature not found")
	}
	sigMap, ok := sigVal.(map[string]any)
	if !ok {
		return false, errors.New("invalid signature shape")
	}
	sig := base58.Decode(sigMap["proofValue"].(string))

	dCopy, err := dp.Copy()
	if err != nil {
		return false, err
	}

	dCopy.data["id"] = MerkleID

	hash, err := dCopy.Hash()
	if err != nil {
		return false, err
	}

	merkleID := idPrefix + base58.Encode(hash)
	if idStr, ok := dp.data["id"].(string); !ok || merkleID != idStr {
		return false, fmt.Errorf("not Merkle ID. Document: %s, Merkle: %s", idStr, merkleID)
	}

	delete(dCopy.data, "proof")

	hash, err = dCopy.Hash()
	if err != nil {
		return false, err
	}

	if signatureVerified := ed25519.Verify(publicKey, hash, sig); !signatureVerified {
		return false, nil
	}

	return true, nil
}
