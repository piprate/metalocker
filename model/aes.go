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

/*
  This code is taken from https://github.com/gtank/cryptopasta
*/

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"io"

	"github.com/piprate/metalocker/utils/zero"
)

const KeySize = 32

type AESKey [32]byte

func (k AESKey) Bytes() []byte {
	return k[:]
}

func (k *AESKey) Zero() {
	zero.Bytes(k[:])
}

func (k *AESKey) Base64() string {
	return base64.StdEncoding.EncodeToString(k[:])
}

func NewAESKey(val []byte) *AESKey {
	key := AESKey{}
	copy(key[:], val)
	return &key
}

// NewEncryptionKey generates a random 256-bit key for Encrypt() and
// Decrypt(). It panics if the source of randomness fails.
func NewEncryptionKey() *AESKey {
	key := AESKey{}
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		panic(err)
	}
	return &key
}

func DeriveEncryptionKey(secret1, secret2 []byte) *AESKey {
	v := Hash("combine", append(secret1, secret2...))
	key := AESKey{}
	copy(key[:], v)
	return &key
}

// EncryptAESCGM encrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Output takes the
// form nonce|ciphertext|tag where '|' indicates concatenation.
func EncryptAESCGM(plaintext []byte, key *AESKey) (ciphertext []byte, err error) {
	if key == nil {
		return nil, errors.New("empty AES key")
	}
	block, err := aes.NewCipher(key.Bytes())
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// DecryptAESCGM decrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Expects input
// form nonce|ciphertext|tag where '|' indicates concatenation.
func DecryptAESCGM(ciphertext []byte, key *AESKey) (plaintext []byte, err error) {
	if key == nil {
		return nil, errors.New("empty AES key")
	}
	block, err := aes.NewCipher(key.Bytes())
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	return gcm.Open(nil,
		ciphertext[:gcm.NonceSize()],
		ciphertext[gcm.NonceSize():],
		nil,
	)
}

// Hash generates a hash of data using HMAC-SHA-512/256. The tag is intended to
// be a natural-language string describing the purpose of the hash, such as
// "hash file for lookup key" or "master secret to client secret".  It serves
// as an HMAC "key" and ensures that different purposes will have different
// hash output. This function is NOT suitable for hashing passwords.
func Hash(tag string, data []byte) []byte {
	h := hmac.New(sha512.New512_256, []byte(tag))
	_, _ = h.Write(data)
	return h.Sum(nil)
}

func Pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}
