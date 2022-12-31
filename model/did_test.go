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

package model_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/piprate/metalocker/contexts"
	. "github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDID     = "did:piprate:Rxbe5BURPeWwYdJUvRRHMm"
	testVerKey  = "EbzQKk7VvynkUbcGBFRNmXfwGHrByucPawRZjvsR6UYv"
	testSignKey = "2BakCg631vruxqmVxRorQ3LiN3S7jN9GovP33zLtZtriRHQWzW7pBEXVoc7ME75QFiYuQvwdSFZdgCnTWVej9Ghc"
)

func TestDID_Sign_Verify(t *testing.T) {
	did := NewDID(testDID, testVerKey, testSignKey)
	msg := []byte("test")
	signature := did.Sign(msg)
	require.NotEmpty(t, signature)
	assert.True(t, did.Verify(msg, signature))
}

func TestDID_Bytes(t *testing.T) {
	did := NewDID(testDID, testVerKey, testSignKey)
	assert.NotEmpty(t, did.Bytes())
}

func TestDID_Copy(t *testing.T) {
	did := NewDID(testDID, testVerKey, testSignKey)
	cpy := did.Copy()
	did.VerKey = ""
	did.SignKey = ""
	assert.Equal(t, testVerKey, cpy.VerKey)
	assert.Equal(t, testSignKey, cpy.SignKey)
}

func TestGenerateDID(t *testing.T) {
	_ = contexts.PreloadContextsIntoMemory()

	// happy path
	did, err := GenerateDID(WithSeed("Steward1"))
	require.NoError(t, err)
	assert.Equal(t, "did:piprate:Th7MpTaRZVRYnPiabds81Y", did.ID)
	assert.Equal(t, "FYmoFw55GeQH7SRFa37dkx1d2dZ3zUF8ckg7wmL7ofN4", did.VerKey)
	assert.Equal(t, "xt19s1sp2UZCGhy9rNyb1FtxdKiDGZZPQ1RLsDSvcomTyZh1EFYHaUoo19qKunQEhkTSzGztovCC3QXma1foGRr", did.SignKey)
}

func TestGenerateDID_CustomMethod(t *testing.T) {
	did, err := GenerateDID(WithMethod("example"))
	require.Nil(t, err)
	assert.True(t, strings.HasPrefix(did.ID, "did:example:"))

	// default method

	did, err = GenerateDID()
	require.Nil(t, err)
	assert.True(t, strings.HasPrefix(did.ID, "did:piprate:"))
}

func TestGenerateDID_SeedPadding(t *testing.T) {
	_, err := GenerateDID(WithSeed("xxx"))
	require.Nil(t, err)
}

func TestGenerateDID_NoSeed(t *testing.T) {
	_ = contexts.PreloadContextsIntoMemory()

	fixedDid, err := GenerateDID(WithSeed("Steward1"))
	require.NoError(t, err)

	// ensure that empty seed will be producing different answers every time
	did1, err := GenerateDID()
	require.NoError(t, err)
	did2, err := GenerateDID()
	require.NoError(t, err)
	assert.NotEqual(t, did1.VerKey, did2.VerKey)
	assert.NotEqual(t, fixedDid.VerKey, did1.VerKey)
	assert.NotEqual(t, fixedDid.VerKey, did2.VerKey)
}

func TestValidateDIDMethodPrefix(t *testing.T) {
	assert.NoError(t, ValidateDIDMethodPrefix("did:test:"))
	assert.Error(t, ValidateDIDMethodPrefix("xxx:test:"))
	assert.Error(t, ValidateDIDMethodPrefix("did:test"))
	assert.Error(t, ValidateDIDMethodPrefix("test"))
	assert.Error(t, ValidateDIDMethodPrefix(""))
}

func TestNewDID(t *testing.T) {
	did := NewDID(testDID, testVerKey, testSignKey)
	assert.Equal(t, testDID, did.ID)
	assert.Equal(t, testVerKey, did.VerKey)
	assert.Equal(t, testSignKey, did.SignKey)
}

func TestExtractDIDMethod(t *testing.T) {
	method, err := ExtractDIDMethod("did:example:xxx")
	require.NoError(t, err)
	assert.Equal(t, "example", method)

	method, err = ExtractDIDMethod("did:example:xxx:tail")
	require.NoError(t, err)
	assert.Equal(t, "example", method)

	_, err = ExtractDIDMethod("did:xxx")
	require.Error(t, err)

	_, err = ExtractDIDMethod("xxx:yyy")
	require.Error(t, err)

	_, err = ExtractDIDMethod("xxx")
	require.Error(t, err)
}

func TestDIDDocument_Sign(t *testing.T) {
	_ = contexts.PreloadContextsIntoMemory()

	did, err := GenerateDID(WithSeed("Test0001"))
	require.NoError(t, err)

	created, _ := time.Parse(time.RFC3339Nano, "2016-09-25T21:58:23Z")

	dd := &DIDDocument{
		Context: []string{"https://w3id.org/did/v1", "https://w3id.org/security/v1"},
		ID:      did.ID,
		PublicKey: []any{
			&Ed25519VerificationKey2018{
				ID:              fmt.Sprintf("%s#key-1", did.ID),
				Type:            "Ed25519VerificationKey2018",
				Controller:      did.ID,
				PublicKeyBase58: did.VerKey,
			},
		},
		Created: &created,
	}

	signKey := did.SignKeyValue()
	err = dd.Sign(did.ID, signKey)
	require.NoError(t, err)

	assert.Equal(t, dd.Proof.Type, "Ed25519Signature2018")
	assert.Equal(t, dd.Proof.Creator, "did:piprate:GmNuLkqi3NY9h5PbeLqHBq")
	assert.Equal(t, dd.Proof.Value, "2PxcGvEXyap9jx99iqcPvAoojK6xRa3CdE9kYpnL6U7ufJ78fNG1d7Tc6vy97vZR1h4suzkJGYAAnMTp5eyZxpaa")

	verKey := base58.Decode(did.VerKey)
	verified, err := dd.Verify(verKey)
	require.NoError(t, err)
	assert.True(t, verified)

	dd.Created = nil

	verified, err = dd.Verify(verKey)
	require.NoError(t, err)
	assert.False(t, verified)
}

func TestSimpleDIDDocument(t *testing.T) {
	did, err := GenerateDID(WithSeed("Steward1"))
	require.NoError(t, err)

	didDoc, err := SimpleDIDDocument(did, nil)
	require.NoError(t, err)
	assert.Equal(t, did.ID, didDoc.ID)

	extractedDID, err := didDoc.ExtractIndyStyleDID()
	require.NoError(t, err)
	assert.Equal(t, did.ID, extractedDID.ID)
}

func TestDIDDocument_Equals(t *testing.T) {
	did1, err := GenerateDID(WithSeed("Steward1"))
	require.NoError(t, err)

	did1Doc, err := SimpleDIDDocument(did1, nil)
	require.NoError(t, err)

	did2, err := GenerateDID(WithSeed("Steward1"))
	require.NoError(t, err)

	did2Doc, err := SimpleDIDDocument(did2, nil)
	require.NoError(t, err)

	did3, err := GenerateDID(WithSeed("Steward2"))
	require.NoError(t, err)

	did3Doc, err := SimpleDIDDocument(did3, nil)
	require.NoError(t, err)

	assert.True(t, did1Doc.Equals(did1Doc)) //nolint:gocritic
	assert.False(t, did2Doc.Equals(did1Doc))
	assert.False(t, did1Doc.Equals(did3Doc))
}

func TestDIDDocument_ExtractIndyStyleDID(t *testing.T) {
	did, err := GenerateDID(WithSeed("Steward1"))
	require.NoError(t, err)

	didDoc, err := SimpleDIDDocument(did, nil)
	require.NoError(t, err)

	extractedDID, err := didDoc.ExtractIndyStyleDID()
	require.NoError(t, err)
	assert.Equal(t, did.ID, extractedDID.ID)
	assert.Equal(t, did.VerKey, extractedDID.VerKey)
	assert.Equal(t, "", extractedDID.SignKey)

	var docCpy *DIDDocument
	require.NoError(t, jsonw.Unmarshal(didDoc.Bytes(), &docCpy))

	extractedDID, err = docCpy.ExtractIndyStyleDID()
	require.NoError(t, err)
	assert.Equal(t, did.ID, extractedDID.ID)
	assert.Equal(t, did.VerKey, extractedDID.VerKey)
	assert.Equal(t, "", extractedDID.SignKey)

	didDoc.PublicKey = nil
	_, err = didDoc.ExtractIndyStyleDID()
	assert.Error(t, err)
}
