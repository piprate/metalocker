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
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/utils/zero"
)

const (
	piprateDIDPrefix               = "did:piprate:"
	Ed25519VerificationKey2018Type = "Ed25519VerificationKey2018"
)

var ErrInvalidDID = errors.New("invalid DID identifier")

type (
	// DID is a Hyperledger Indy-style definition of a decentralised identifier (DID).
	// DID is a globally unique persistent identifier that does not require a centralized
	// registration authority because it is generated and/or registered cryptographically.
	DID struct {
		// ID is a decentralised identifier.
		ID string `json:"id" example:"did:piprate:9JA3ukzEXQeoTLyH9N2Jdp"`
		// VerKey public Ed25519 key in base58 encoding
		VerKey string `json:"verKey" example:"<public key in base58 encoding>"`
		// SignKey (optional) private Ed25519 key in base58 encoding
		SignKey string `json:"signKey,omitempty" example:"<private key in base58 encoding>"`
		verKey  ed25519.PublicKey
		signKey ed25519.PrivateKey
	}

	// Ed25519VerificationKey2018 is based on https://w3c-dvcg.github.io/lds-ed25519-2018/
	Ed25519VerificationKey2018 struct {
		Context         any        `json:"@context,omitempty"`
		ID              string     `json:"id"`
		Type            string     `json:"type"`
		Controller      string     `json:"controller"`
		Expires         *time.Time `json:"expires,omitempty"`
		PublicKeyBase58 string     `json:"publicKeyBase58"`
	}

	DIDDocument struct {
		Context        any        `json:"@context,omitempty"`
		ID             string     `json:"id"`
		PublicKey      []any      `json:"publicKey,omitempty"`
		Authentication []any      `json:"authentication,omitempty"`
		Service        []any      `json:"service,omitempty"`
		Created        *time.Time `json:"created,omitempty"`
		Updated        *time.Time `json:"updated,omitempty"`
		Proof          *Proof     `json:"proof,omitempty"`
	}
)

var _ Signer = (*DID)(nil)
var _ Verifier = (*DID)(nil)

func (did *DID) SignKeyValue() ed25519.PrivateKey {
	if did.signKey == nil {
		did.signKey = base58.Decode(did.SignKey)
	}
	return did.signKey
}

func (did *DID) VerKeyValue() ed25519.PublicKey {
	if did.verKey == nil {
		did.verKey = base58.Decode(did.VerKey)
	}
	return did.verKey
}

func (did *DID) Sign(message []byte) []byte {
	return ed25519.Sign(did.SignKeyValue(), message)
}

func (did *DID) Verify(message, signature []byte) bool {
	return ed25519.Verify(did.VerKeyValue(), message, signature)
}

func (did *DID) Bytes() []byte {
	b, _ := jsonw.Marshal(did)
	return b
}

func (did *DID) Zero() {
	did.SignKey = ""
	zero.Bytes(did.signKey)
	did.signKey = nil
}

func (did *DID) NeuteredCopy() *DID {
	return &DID{
		ID:     did.ID,
		VerKey: did.VerKey,
	}
}

func (did *DID) Copy() *DID {
	cp := *did
	return &cp
}

type (
	didOptions struct {
		seed   string
		method string
	}

	DIDOption func(opts *didOptions)
)

func WithSeed(seed string) DIDOption {
	return func(opts *didOptions) {
		opts.seed = seed
	}
}

func WithMethod(method string) DIDOption {
	return func(opts *didOptions) {
		opts.method = method
	}
}

func GenerateDID(options ...DIDOption) (*DID, error) {
	var opts didOptions
	for _, o := range options {
		o(&opts)
	}

	var randSeed io.Reader
	if opts.seed != "" {
		// pad the seed, if needed
		if len(opts.seed) < ed25519.SeedSize {
			opts.seed = strings.Repeat("0", ed25519.SeedSize-len(opts.seed)) + opts.seed
		}
		randSeed = strings.NewReader(opts.seed)
	}
	publicKey, privateKey, err := ed25519.GenerateKey(randSeed)
	if err != nil {
		return nil, err
	} else {
		return &DID{
			ID:      BuildDIDPrefix(opts.method) + base58.Encode(publicKey[0:16]),
			VerKey:  base58.Encode(publicKey),
			SignKey: base58.Encode(privateKey),
			signKey: privateKey,
		}, nil
	}
}

var didPrefixCache = map[string]string{}

func BuildDIDPrefix(method string) string {
	if method == "" {
		return piprateDIDPrefix
	}

	prefix, found := didPrefixCache[method]
	if !found {
		prefix = "did:" + method + ":"
		didPrefixCache[method] = prefix
	}

	return prefix
}

func ValidateDIDMethodPrefix(methodPrefix string) error {
	if methodPrefix == "" {
		return errors.New("empty DID method prefix")
	}
	if !strings.HasPrefix(methodPrefix, "did:") {
		return errors.New("did method prefix should start with 'did:'")
	}
	if methodPrefix[len(methodPrefix)-1] != ':' {
		return errors.New("did method prefix should end with ':'")
	}
	return nil
}

func NewDID(did, verKey, signKey string) *DID {
	return &DID{
		ID:      did,
		VerKey:  verKey,
		SignKey: signKey,
	}
}

func ExtractDIDMethod(didID string) (string, error) {
	parts := strings.SplitN(didID, ":", 3)
	if len(parts) < 3 || parts[0] != "did" {
		return "", ErrInvalidDID
	}
	return parts[1], nil
}

type (
	DIDProvider interface {
		CreateDIDDocument(ctx context.Context, ddoc *DIDDocument) error
		GetDIDDocument(ctx context.Context, iid string) (*DIDDocument, error)
	}
)

func (d *DIDDocument) Bytes() []byte {
	b, _ := jsonw.Marshal(d)
	return b
}

func (d *DIDDocument) Equals(anotherD *DIDDocument) bool {
	return anotherD != nil && anotherD.ID == d.ID && anotherD.Proof != nil && d.Proof != nil && anotherD.Proof.Value == d.Proof.Value
}

// Sign signs the document and assigns Merkle ID to the root JSON-LD element
func (d *DIDDocument) Sign(identity string, key ed25519.PrivateKey) error {
	signable, _ := NewSignableDocument(d.Bytes())

	proof, err := signable.Sign(identity, key)
	if err != nil {
		return err
	}

	d.Proof = proof

	return nil
}

func (d *DIDDocument) Verify(key ed25519.PublicKey) (bool, error) {
	doc, _ := NewSignableDocument(d.Bytes())
	return doc.Verify(key)
}

func (d *DIDDocument) ExtractIndyStyleDID() (*DID, error) {
	if len(d.PublicKey) > 0 {
		for _, k := range d.PublicKey {
			switch val := k.(type) {
			case *Ed25519VerificationKey2018:
				if val.Type == Ed25519VerificationKey2018Type {
					if val.PublicKeyBase58 != "" {
						return &DID{
							ID:     d.ID,
							VerKey: val.PublicKeyBase58,
						}, nil
					}
				}
			case map[string]any:
				if val["type"] == Ed25519VerificationKey2018Type {
					if verKey := val["publicKeyBase58"].(string); verKey != "" {
						return &DID{
							ID:     d.ID,
							VerKey: verKey,
						}, nil
					}
				}
			}
		}
	}
	return nil, errors.New("no instances of Ed25519VerificationKey2018 found")
}

func SimpleDIDDocument(did *DID, created *time.Time) (*DIDDocument, error) {
	if created == nil {
		now := time.Now().UTC()
		created = &now
	}

	dd := &DIDDocument{
		Context: []string{"https://w3id.org/did/v1", "https://w3id.org/security/v1"},
		ID:      did.ID,
		PublicKey: []any{
			&Ed25519VerificationKey2018{
				ID:              fmt.Sprintf("%s#key-1", did.ID),
				Type:            Ed25519VerificationKey2018Type,
				Controller:      did.ID,
				PublicKeyBase58: did.VerKey,
			},
		},
		Created: created,
	}

	key := base58.Decode(did.SignKey)
	err := dd.Sign(did.ID, key)
	if err != nil {
		return nil, err
	}

	return dd, nil
}
