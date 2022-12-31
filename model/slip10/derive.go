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

package slip10

import (
	"bytes"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/piprate/metalocker/utils/zero"
)

/*
	This SLIP-0010 (https://github.com/satoshilabs/slips/blob/master/slip-0010.md)
	implementation was adapted from https://github.com/anytypeio/go-slip10.
*/

const (
	// FirstHardenedIndex is the index of the first hardened key (2^31).
	// https://youtu.be/2HrMlVr1QX8?t=390
	FirstHardenedIndex = uint32(0x80000000)
	// As in https://github.com/satoshilabs/slips/blob/master/slip-0010.md
	seedModifier = "ed25519 seed"

	// RecommendedSeedLen is the recommended length in bytes for a seed
	// to a master node.
	RecommendedSeedLen = 32 // 256 bits

	// MinSeedBytes is the minimum number of bytes allowed for a seed to
	// a master node.
	MinSeedBytes = 16 // 128 bits

	// MaxSeedBytes is the maximum number of bytes allowed for a seed to
	// a master node.
	MaxSeedBytes = 64 // 512 bits
)

var (
	ErrInvalidPath        = fmt.Errorf("invalid derivation path")
	ErrNoPublicDerivation = fmt.Errorf("no public derivation for ed25519")
	// ErrInvalidSeedLen describes an error in which the provided seed or
	// seed length is not in the allowed range.
	ErrInvalidSeedLen = fmt.Errorf("seed length must be between %d and %d "+
		"bits", MinSeedBytes*8, MaxSeedBytes*8)

	pathRegex = regexp.MustCompile(`^m(/\d+')*$`)
)

// GenerateSeed returns a cryptographically secure random seed that can be used
// as the input for the NewMaster function to generate a new master node.
//
// The length is in bytes, and it must be between 16 and 64 (128 to 512 bits).
// The recommended length is 32 (256 bits) as defined by the RecommendedSeedLen
// constant.
func GenerateSeed(length uint8) ([]byte, error) {
	// Per [BIP32], the seed must be in range [MinSeedBytes, MaxSeedBytes].
	if length < MinSeedBytes || length > MaxSeedBytes {
		return nil, ErrInvalidSeedLen
	}

	buf := make([]byte, length)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

type Node interface {
	Derive(i uint32) (Node, error)

	KeyPair() (ed25519.PublicKey, ed25519.PrivateKey)
	PrivateKey() []byte
	PublicKeyWithPrefix() []byte
	RawSeed() []byte
	Serialize() string
	Bytes() []byte
	Zero()
}

type node []byte

// DeriveForPath derives key for a path in BIP-44 format and a seed.
// Ed25119 derivation operated on hardened keys only.
func DeriveForPath(path string, seed []byte) (Node, error) {
	if !IsValidPath(path) {
		return nil, ErrInvalidPath
	}

	key, err := NewMasterNode(seed)
	if err != nil {
		return nil, err
	}

	segments := strings.Split(path, "/")
	for _, segment := range segments[1:] {
		i64, err := strconv.ParseUint(strings.TrimRight(segment, "'"), 10, 32)
		if err != nil {
			return nil, err
		}

		// we operate on hardened keys
		i := uint32(i64) + FirstHardenedIndex
		key, err = key.Derive(i)
		if err != nil {
			return nil, err
		}
	}

	return key, nil
}

// NewMasterNode generates a new master key from seed.
func NewMasterNode(seed []byte) (Node, error) {
	hash := hmac.New(sha512.New, []byte(seedModifier))
	_, err := hash.Write(seed)
	if err != nil {
		return nil, err
	}
	return node(hash.Sum(nil)), nil
}

func NewNodeFromString(val string) (Node, error) {
	b, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return nil, err
	}
	return node(b), nil
}

func (k node) key() []byte {
	return k[:32]
}

func (k node) chainCode() []byte {
	return k[32:]
}

func (k node) Derive(i uint32) (Node, error) {
	// no public derivation for ed25519
	if i < FirstHardenedIndex {
		return nil, ErrNoPublicDerivation
	}

	iBytes := [4]byte{}
	binary.BigEndian.PutUint32(iBytes[:], i)
	key := append([]byte{0x0}, k.key()...)
	data := append(key, iBytes[:]...)

	hash := hmac.New(sha512.New, k.chainCode())
	_, err := hash.Write(data)
	if err != nil {
		return nil, err
	}
	return node(hash.Sum(nil)), nil
}

func (k node) KeyPair() (ed25519.PublicKey, ed25519.PrivateKey) {
	reader := bytes.NewReader(k.key())
	pub, priv, err := ed25519.GenerateKey(reader)
	if err != nil {
		// can't happens because we check the seed on NewMasterNode/DeriveForPath
		return nil, nil
	}

	return pub[:], priv[:]
}

// RawSeed returns raw seed bytes
func (k node) RawSeed() []byte {
	return k.key()
}

// PrivateKey returns private key seed bytes
func (k node) PrivateKey() []byte {
	_, priv := k.KeyPair()
	return priv.Seed()
}

// PublicKeyWithPrefix returns public key with 0x00 prefix, as specified in the slip-10
// https://github.com/satoshilabs/slips/blob/master/slip-0010/testvectors.py#L64
func (k node) PublicKeyWithPrefix() []byte {
	pub, _ := k.KeyPair()
	return append([]byte{0x00}, pub...)
}

func (k node) Serialize() string {
	return base64.StdEncoding.EncodeToString(k)
}

func (k node) Bytes() []byte {
	return k
}

func (k node) Zero() {
	zero.Bytes(k)
}

// IsValidPath check whether or not the path has valid segments.
func IsValidPath(path string) bool {
	if !pathRegex.MatchString(path) {
		return false
	}

	// check for overflows
	segments := strings.Split(path, "/")
	for _, segment := range segments[1:] {
		_, err := strconv.ParseUint(strings.TrimRight(segment, "'"), 10, 32)
		if err != nil {
			return false
		}
	}

	return true
}
