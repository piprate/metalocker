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
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	. "github.com/piprate/metalocker/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAccessKey(t *testing.T) {
	accessKey, err := GenerateAccessKey("did:piprate:xxx", AccessLevelHosted)
	require.NoError(t, err)

	keyID, clientSecret := accessKey.ClientKeys()
	assert.Equal(t, keyID, accessKey.ID)
	assert.Equal(t, AccessKeyType, accessKey.Type)

	mgmtKey, aesKey, hmacKey, err := SplitClientSecret(clientSecret)
	require.NoError(t, err)
	assert.Equal(t, mgmtKey, accessKey.ManagementKeyPrv)
	assert.Equal(t, aesKey, accessKey.ClientSecret)
	assert.Equal(t, hmacKey, accessKey.ClientHMACKey)
	assert.Equal(t, AccessLevelHosted, accessKey.AccessLevel)
}

func TestSplitClientSecret(t *testing.T) {
	_, _, _, err := SplitClientSecret("bad secret")
	assert.Error(t, err)

	_, _, _, err = SplitClientSecret("bad.secret")
	assert.Error(t, err)

	_, _, _, err = SplitClientSecret("ynPPwBn+/NpiCViJ1ryb3cD07T7q0SyZVRy7e7IimHV+Llzy/sS8/PxxwHtlxuyG0ev5rJF1EteUegGmt6ET3Q==.bad-part")
	assert.Error(t, err)

	mgmtKey, aesKey, hmacKey, err := SplitClientSecret("ynPPwBn+/NpiCViJ1ryb3cD07T7q0SyZVRy7e7IimHV+Llzy/sS8/PxxwHtlxuyG0ev5rJF1EteUegGmt6ET3Q==.V07XomkwrvnBOY/cczzlEH21Qx3kqcxf5b7hyKeG0/1JKyaTfC3BqDIeaHkMLi04afZZXukxYU9MYAX8l17R+g==")
	assert.NoError(t, err)
	assert.NotNil(t, mgmtKey)
	assert.NotNil(t, aesKey)
	assert.NotNil(t, hmacKey)
}

func TestAccessKey_Hydrate(t *testing.T) {
	accessKey, err := GenerateAccessKey("did:piprate:xxx", AccessLevelHosted)
	require.NoError(t, err)

	err = accessKey.Hydrate("bad secret")
	require.Error(t, err)

	// happy path

	_, secret := accessKey.ClientKeys()
	err = accessKey.Hydrate(secret)
	require.NoError(t, err)
}

func TestAccessKey_Neuter(t *testing.T) {
	accessKey, err := GenerateAccessKey("did:piprate:xxx", AccessLevelHosted)
	require.NoError(t, err)

	accessKey.Neuter()

	assert.Empty(t, accessKey.ManagementKeyPub)
	assert.Empty(t, accessKey.ManagementKeyPrv)
	assert.Empty(t, accessKey.ClientSecret)
	assert.Empty(t, accessKey.ClientHMACKey)
}

func TestAccessKey_Bytes(t *testing.T) {
	accessKey, err := GenerateAccessKey("did:piprate:xxx", AccessLevelHosted)
	require.NoError(t, err)

	b := accessKey.Bytes()

	assert.NotEmpty(t, b)
}

func TestSignRequest(t *testing.T) {
	keyID := "HYJWc4FMvHt56o3K9RHCpt"
	//nolint:gosec
	clientSecret := "UsjKWSbhqwiasmSOY5Jk6djS3GCpSAxJZSHOwRMH2aKkxqAIB+snrIzGgXtLmGhzhRrbtLRMq3s92oJgQTmdUw==.Nr5P22pC3lW3sTdhgwJeYopbJEc+G5M3hlILH9Te5tp/0dxHAbI8q9EUO+lVJRti0Tm6zRODsX1ti7EPE68ktw=="
	// secret := "Kv2w9cS9rtMU88cGLwdqo3duNvvUhZC1Pk/KGodKVDS0O+5T5aD079m7FMNdra4BfX+UN0Ld2GfdrgVznu3KxjnSUWbm7pCLALYopnSZYfvEH0XFGwIIQ/qS0Ug="
	_, aesKey, hmacKey, err := SplitClientSecret(clientSecret)
	require.NoError(t, err)

	now := time.Unix(10000000, 0)

	hdr := http.Header{}
	_, err = SignRequest(hdr, keyID, aesKey, hmacKey, now, "/v1/account", []byte("test"))
	require.NoError(t, err)

	assert.Equal(t, "Meta HYJWc4FMvHt56o3K9RHCpt:Ea58gE9YL3Oo1hLyoi6BPHNWRPC6YP1nDo/3qRLjXas=", hdr.Get("Authorization"))
	assert.Equal(t, "IvZOgH9mrFaxDYqwGF5P4dpS+yZ5RuHeUfvTJHjXEAk=", hdr.Get("X-Meta-Body-Hash"))
	assert.Equal(t, "cCx+SYim/H/L26YXlWWz4InWWz7hbu0whvB/LSiBjiA=", hdr.Get("X-Meta-Client-Key"))
	assert.Equal(t, "19700426", hdr.Get("X-Meta-Date"))
}

func TestValidateRequest(t *testing.T) {
	hdr := http.Header{
		"Authorization": []string{
			"Meta HYJWc4FMvHt56o3K9RHCpt:Ea58gE9YL3Oo1hLyoi6BPHNWRPC6YP1nDo/3qRLjXas=",
		},
		"X-Meta-Body-Hash": []string{
			"IvZOgH9mrFaxDYqwGF5P4dpS+yZ5RuHeUfvTJHjXEAk=",
		},
		"X-Meta-Client-Key": []string{
			"cCx+SYim/H/L26YXlWWz4InWWz7hbu0whvB/LSiBjiA=",
		},
		"X-Meta-Date": []string{
			"19700426",
		},
	}

	//nolint:gosec
	secret := "Kv2w9cS9rtMU88cGLwdqo3duNvvUhZC1Pk/KGodKVDS0O+5T5aD079m7FMNdra4BfX+UN0Ld2GfdrgVznu3KxjnSUWbm7pCLALYopnSZYfvEH0XFGwIIQ/qS0Ug="

	encryptedHMACKey, err := base64.StdEncoding.DecodeString(secret)
	require.NoError(t, err)

	serverTime := time.Unix(10020000, 0)
	url := "/v1/account"
	body := []byte("test")
	bodyHash := HashRequestBody(body)

	keyID, sig, err := ExtractSignature(hdr)
	require.NoError(t, err)
	assert.Equal(t, "HYJWc4FMvHt56o3K9RHCpt", keyID)
	assert.Equal(t, "Ea58gE9YL3Oo1hLyoi6BPHNWRPC6YP1nDo/3qRLjXas=", sig)

	// happy path

	valid, err := ValidateRequest(hdr, sig, encryptedHMACKey, serverTime, url, bodyHash)
	require.NoError(t, err)
	assert.True(t, valid)

	// signature too old

	serverTime = time.Unix(20002000, 0)
	valid, err = ValidateRequest(hdr, sig, encryptedHMACKey, serverTime, url, bodyHash)
	require.NoError(t, err)
	assert.False(t, valid)

	serverTime = time.Unix(10020000, 0)

	// missing header fields

	delete(hdr, AccessKeyHeaderDate)
	_, err = ValidateRequest(hdr, sig, encryptedHMACKey, serverTime, url, bodyHash)
	require.Error(t, err)

	delete(hdr, AccessKeyHeaderClientKey)
	_, err = ValidateRequest(hdr, sig, encryptedHMACKey, serverTime, url, bodyHash)
	require.Error(t, err)
}

func TestEncryptCredentials(t *testing.T) {
	did, err := GenerateDID(WithSeed("Test0001"))
	require.NoError(t, err)

	cypherText := EncryptCredentials(did.NeuteredCopy(), "key1", "secret1", "")

	key, secret, err := DecryptCredentials(did, cypherText, "")
	require.NoError(t, err)
	assert.Equal(t, "key1", key)
	assert.Equal(t, "secret1", secret)

	// test subject verification

	cypherText = EncryptCredentials(did.NeuteredCopy(), "key1", "secret1", "test-dt")

	_, _, err = DecryptCredentials(did, cypherText, "")
	require.NoError(t, err)

	key, secret, err = DecryptCredentials(did, cypherText, "test-dt")
	require.NoError(t, err)
	assert.Equal(t, "key1", key)
	assert.Equal(t, "secret1", secret)

	_, _, err = DecryptCredentials(did, cypherText, "wrong-test-dt")
	require.Error(t, err)
}

func TestDecryptCredentials(t *testing.T) {
	did, err := GenerateDID(WithSeed("Test0001"))
	require.NoError(t, err)

	cypherText := EncryptCredentials(did, "key1", "secret1", "test-dt")

	// cannot decrypt bad cypher text

	_, _, err = DecryptCredentials(did, "someText", "test-dt")
	assert.Error(t, err)

	someText := AnonEncrypt([]byte("someText"), did.VerKeyValue())

	_, _, err = DecryptCredentials(did, string(someText), "test-dt")
	assert.Error(t, err)

	_, _, err = DecryptCredentials(did, base64.StdEncoding.EncodeToString(someText), "test-dt")
	assert.Error(t, err)

	// cannot decrypt with a different DID

	anotherDID, err := GenerateDID(WithSeed("Test0002"))
	require.NoError(t, err)

	_, _, err = DecryptCredentials(anotherDID, cypherText, "test-dt")
	assert.Error(t, err)

	// cannot decrypt with a neutered DID

	_, _, err = DecryptCredentials(did.NeuteredCopy(), cypherText, "test-dt")
	assert.Error(t, err)

	// cannot decrypt with a wrong subject

	_, _, err = DecryptCredentials(did, cypherText, "bad-dt")
	assert.Error(t, err)

	// happy path

	_, _, err = DecryptCredentials(did, cypherText, "test-dt")
	assert.NoError(t, err)
}
