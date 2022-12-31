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
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/utils/zero"
)

// Access Key algorithm is adapted from https://crypto.stackexchange.com/questions/72504/using-public-key-signature-instead-of-having-api-key#72521.
// Also, some ideas were taken from Authenticating Requests (AWS Signature Version 4) (https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-authenticating-requests.html).

const (
	AccessKeyType = "AccessKey"

	AccessKeyHeaderDate      = "X-Meta-Date"
	AccessKeyHeaderClientKey = "X-Meta-Client-Key"
	AccessKeyHeaderBodyHash  = "X-Meta-Body-Hash"
)

var (
	ErrMissingDateInHeader      = errors.New("missing " + AccessKeyHeaderDate + " in request header")
	ErrMissingClientKeyInHeader = errors.New("missing " + AccessKeyHeaderClientKey + " in request header")
)

// AccessKey defines a key that can be used to access MetaLocker.
// Access keys are useful for programmatic or temporary access to MetaLocker data
// without revealing its main encryption keys.
type AccessKey struct {
	ID          string      `json:"id"`
	AccountID   string      `json:"account"`
	AccessLevel AccessLevel `json:"level"`
	Secret      string      `json:"secret,omitempty"`
	Type        string      `json:"type"`

	ManagementKey       string `json:"mgmtKey"`
	EncryptedManagedKey string `json:"emk,omitempty"`
	EncryptedHostedKey  string `json:"ehk,omitempty"`

	// See https://www.algolia.com/doc/api-reference/api-methods/generate-secured-api-key/
	// for an example of key options
	//ExpiresAt int    `json:"expires,omitempty"`

	ManagementKeyPub ed25519.PublicKey  `json:"-"`
	ManagementKeyPrv ed25519.PrivateKey `json:"-"`
	ClientSecret     *AESKey            `json:"-"`
	ClientHMACKey    []byte             `json:"-"`
}

func (ak *AccessKey) AddManagedKey(key *AESKey) {
	cypherText := AnonEncrypt(key.Bytes(), ak.ManagementKeyPub)
	ak.EncryptedManagedKey = base64.StdEncoding.EncodeToString(cypherText)
}

func (ak *AccessKey) AddHostedKey(key *AESKey) {
	cypherText := AnonEncrypt(key.Bytes(), ak.ManagementKeyPub)
	ak.EncryptedHostedKey = base64.StdEncoding.EncodeToString(cypherText)
}

func (ak *AccessKey) Hydrate(secret string) error {
	mgmtKey, _, _, err := SplitClientSecret(secret)
	if err != nil {
		return err
	}
	ak.ManagementKeyPrv = mgmtKey
	ak.ManagementKeyPub = mgmtKey.Public().(ed25519.PublicKey)

	return nil
}

func (ak *AccessKey) Neuter() {
	zero.Bytes(ak.ManagementKeyPub)
	ak.ManagementKeyPub = nil
	zero.Bytes(ak.ManagementKeyPrv)
	ak.ManagementKeyPrv = nil
	zero.Bytea32((*[32]byte)(ak.ClientSecret))
	ak.ClientSecret = nil
	zero.Bytes(ak.ClientHMACKey)
	ak.ClientHMACKey = nil
}

// ClientKeys returns a pair <key-id> and <secret-string> that can be used by a remote client
// to gain access to a specific account with specific restrictions.
func (ak *AccessKey) ClientKeys() (string, string) {
	secret := base64.StdEncoding.EncodeToString(ak.ManagementKeyPrv)
	key := base64.StdEncoding.EncodeToString(ak.ClientHMACKey)
	return ak.ID, secret + "." + key
}

func (ak *AccessKey) Bytes() []byte {
	b, _ := jsonw.Marshal(ak)
	return b
}

func GenerateAccessKeyID() string {
	return base58.Encode(NewEncryptionKey()[:16])
}

func DeriveClientAESKey(pk ed25519.PrivateKey) *AESKey {
	return NewAESKey(Hash("access key secret", pk))
}

// GenerateAccessKey creates a new access key that can be used to connect to MetaLocker.
//
// Client will use: keyID, management key (64-byte private Ed-25519 key), HMAC key (64 bytes)
// Server will use: keyID, encrypted HMAC key
func GenerateAccessKey(accountID string, accessLevel AccessLevel) (*AccessKey, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, err
	}

	aesKey := DeriveClientAESKey(privateKey)

	hmacKey := make([]byte, 64)
	_, err = rand.Read(hmacKey)
	if err != nil {
		return nil, err
	}
	encHmacKey, err := EncryptAESCGM(hmacKey, aesKey)
	if err != nil {
		return nil, err
	}

	id := GenerateAccessKeyID()

	return &AccessKey{
		ID:          id,
		AccountID:   accountID,
		Secret:      base64.StdEncoding.EncodeToString(encHmacKey),
		Type:        AccessKeyType,
		AccessLevel: accessLevel,

		ManagementKey: base64.StdEncoding.EncodeToString(publicKey),

		ManagementKeyPub: publicKey,
		ManagementKeyPrv: privateKey,
		ClientSecret:     aesKey,
		ClientHMACKey:    hmacKey,
	}, nil
}

func SplitClientSecret(secret string) (ed25519.PrivateKey, *AESKey, []byte, error) {
	parts := strings.Split(secret, ".")
	if len(parts) != 2 {
		return nil, nil, nil, errors.New("bad client secret")
	}
	privateKey, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, nil, nil, err
	}

	aesKey := DeriveClientAESKey(privateKey)

	hmacKey, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, nil, nil, err
	}

	return privateKey, aesKey, hmacKey, nil
}

func DecodeAESKey(val string, privKey ed25519.PrivateKey) (*AESKey, error) {
	cypherText, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return nil, err
	}
	keyBytes, err := AnonDecrypt(cypherText, privKey)
	if err != nil {
		return nil, err
	}
	encKey := NewAESKey(keyBytes)
	zero.Bytes(keyBytes)

	return encKey, nil
}

func SignRequest(hdr http.Header, keyID string, clientSecret *AESKey, clientHMACKey []byte, now time.Time, url string, body []byte) (string, error) {
	dateStr := now.Format("20060102")
	aesKeyStr := base64.StdEncoding.EncodeToString(clientSecret[:])
	hdr.Set(AccessKeyHeaderDate, dateStr)
	hdr.Set(AccessKeyHeaderClientKey, aesKeyStr)

	h := hmac.New(sha512.New512_256, clientHMACKey)
	_, _ = h.Write([]byte(dateStr))
	dateKey := h.Sum(nil)
	h = hmac.New(sha512.New512_256, dateKey)
	_, _ = h.Write([]byte(url))

	if body != nil {
		bodyKey := h.Sum(nil)

		bodyH := hmac.New(sha512.New512_256, []byte("body hash"))
		_, _ = bodyH.Write(body)
		bodyHash := bodyH.Sum(nil)
		h = hmac.New(sha512.New512_256, bodyKey)
		_, _ = h.Write(bodyHash)

		hdr.Set(AccessKeyHeaderBodyHash, base64.StdEncoding.EncodeToString(bodyHash))
	}
	sig := h.Sum(nil)

	sigStr := base64.StdEncoding.EncodeToString(sig)

	authValue := "Meta " + keyID + ":" + sigStr

	hdr.Set("Authorization", authValue)

	return sigStr, nil
}

var ErrAuthorizationNotFound = errors.New("missing or invalid Authorization in request header")

func ExtractSignature(hdr http.Header) (string, string, error) {
	authValue := hdr.Get("Authorization")
	if authValue == "" {
		return "", "", ErrAuthorizationNotFound
	}

	// extract Key ID

	if !secureCompare(authValue[:5], "Meta ") {
		return "", "", ErrAuthorizationNotFound
	}

	parts := strings.Split(authValue[5:], ":")
	if subtle.ConstantTimeEq(int32(len(parts)), 2) == 0 {
		return "", "", errors.New("bad Authorization value")
	}

	return parts[0], parts[1], nil
}

func ValidateRequest(hdr http.Header, reqSig string, encryptedHMACKey []byte, reqTime time.Time, url string, bodyHash []byte) (bool, error) {

	// extract AES key and decrypt HMAC key

	clientSecretStr := hdr.Get(AccessKeyHeaderClientKey)
	if clientSecretStr == "" {
		return false, ErrMissingClientKeyInHeader
	}

	clientSecretBytes, err := base64.StdEncoding.DecodeString(clientSecretStr)
	if err != nil {
		return false, err
	}
	clientSecret := AESKey{}
	copy(clientSecret[:], clientSecretBytes)

	clientHMACKey, err := DecryptAESCGM(encryptedHMACKey, &clientSecret)
	if err != nil {
		return false, err
	}

	// extract request date and compare with server time

	dateStr := hdr.Get(AccessKeyHeaderDate)
	if clientSecretStr == "" {
		return false, ErrMissingDateInHeader
	}

	date, err := time.Parse("20060102", dateStr)
	if err != nil {
		return false, err
	}
	cutOffDate := date.Add(time.Hour*24 + time.Minute*15)
	if reqTime.After(cutOffDate) {
		// signature too old
		return false, nil
	}

	// re-run signature calculations

	h := hmac.New(sha512.New512_256, clientHMACKey)
	_, _ = h.Write([]byte(dateStr))
	dateKey := h.Sum(nil)
	h = hmac.New(sha512.New512_256, dateKey)
	_, _ = h.Write([]byte(url))

	if bodyHash != nil {
		bodyKey := h.Sum(nil)
		h = hmac.New(sha512.New512_256, bodyKey)
		_, _ = h.Write(bodyHash)
	}
	sigBytes := h.Sum(nil)

	sig := base64.StdEncoding.EncodeToString(sigBytes)

	return secureCompare(sig, reqSig), nil
}

func HashRequestBody(body []byte) []byte {
	bodyH := hmac.New(sha512.New512_256, []byte("body hash"))
	_, _ = bodyH.Write(body)
	return bodyH.Sum(nil)
}

// secureCompare performs a constant time compare of two strings to limit timing attacks.
func secureCompare(given, actual string) bool {
	if subtle.ConstantTimeEq(int32(len(given)), int32(len(actual))) == 1 {
		return subtle.ConstantTimeCompare([]byte(given), []byte(actual)) == 1
	} else {
		/* Securely compare actual to itself to keep constant time, but always return false */
		return subtle.ConstantTimeCompare([]byte(actual), []byte(actual)) == 1 && false
	}
}

func EncryptCredentials(recipient *DID, keyID, secret, subject string) string {
	msg := subject + "|" + keyID + "|" + secret
	encCredentials := AnonEncrypt([]byte(msg), recipient.VerKeyValue())
	return base64.StdEncoding.EncodeToString(encCredentials)
}

func DecryptCredentials(recipient *DID, credentials, subject string) (string, string, error) {
	if recipient.SignKey == "" {
		return "", "", errors.New("can't decrypt credentials with neutered key")
	}
	cypherBytes, err := base64.StdEncoding.DecodeString(credentials)
	if err != nil {
		return "", "", err
	}
	msg, err := AnonDecrypt(cypherBytes, recipient.SignKeyValue())
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(string(msg), "|")
	if len(parts) != 3 {
		return "", "", errors.New("wrong number of parts in encrypted credentials")
	}

	if parts[0] != "" && subject != "" && parts[0] != subject {
		return "", "", errors.New("failed to decrypt zone credentials: subject mismatch")
	}
	return parts[1], parts[2], nil
}
