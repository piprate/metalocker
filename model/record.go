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
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/piprate/metalocker/utils"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog/log"
)

type OpType uint32

var (
	OpTypeLease           OpType = 1
	OpTypeLeaseRevocation OpType = 2
	OpTypeAssetHead       OpType = 3
)

const (
	RcTypeAlgo0 = 0
	RcTypeAlgo1 = 1

	// RecordFlagPublic bit is set to true, if the underlying operation
	// and data assets are available in a clear-text form. Use this
	// flag to publish data that needs to be accessed by third parties
	// that don't know the locker secrets.
	RecordFlagPublic uint32 = 0x00000001
)

// Record represents one data transaction (lease, revocation, etc).
// It contains no details which would allow a third party observer to identify
// the participants or the nature of this transaction.
type Record struct {
	// ID of the record. Currently, it's a hash of the record generated
	// by the Seal function (see below).
	ID string `json:"id"`
	// RoutingKey is a public key from the locker HD structure. It can be
	// used to filter specific messages from the ledger.
	RoutingKey string `json:"routingKey"`
	// KeyIndex is the index of the HD key used to produce the routing key.
	KeyIndex uint32 `json:"keyIndex"`
	// Operation is the type of the operation. May be removed from
	// the record in the future.
	Operation OpType `json:"operationType"`
	// OperationAddress is the address of the Operation (can be an asset ID, IPFS address, etc.)
	OperationAddress string `json:"address,omitempty"`
	// Flags contain a set of flags that modify the record's behaviour.
	// See RecordFlagXXX constants for examples.
	Flags uint32 `json:"flags,omitempty"`
	// AuthorisingCommitment is binary data which allows the originator
	// of the transaction to prove their role, without disclosing any
	// other information about this transaction
	AuthorisingCommitment string `json:"ac,omitempty"`
	// AuthorisingCommitmentType. for future use: there may be different
	// types of commitment structures.
	AuthorisingCommitmentType byte `json:"acType,omitempty"`
	// RequestingCommitment is binary data which allows the recipient of
	// the transaction to prove their right to access data without disclosing
	// any other information about this transaction
	RequestingCommitment string `json:"rc,omitempty"`
	// RequestingCommitmentType. for future use: there may be different types
	// of commitment structures.
	RequestingCommitmentType byte `json:"rcType,omitempty"`
	// ImpressionCommitment is binary data which allows a party to prove
	// that this record contains a specific impression, by combining
	// the impression ID with another artifact, a trapdoor.
	ImpressionCommitment string `json:"ic,omitempty"`
	// ImpressionCommitmentType. for future use: there may be different
	// types of commitment structures.
	ImpressionCommitmentType byte `json:"icType,omitempty"`

	// DataAssets is a list of data assets (blobs) attached to the record
	DataAssets []string `json:"dataAssets,omitempty"`

	// Lease Revocation / Head fields

	SubjectRecord   string   `json:"subjectRecord,omitempty"`
	RevocationProof []string `json:"revocationProof,omitempty"`

	// Head fields

	// HeadID is unique ID of the asset head.
	HeadID string `json:"headID,omitempty"`
	// HeadBody contains base64-encoded, encrypted head body (see PackHeadBody() ).
	HeadBody string `json:"headBody,omitempty"`

	// Signature contains a digital signature of the record, signed by
	// the record's private HD key
	Signature string `json:"signature"`

	Status RecordStatus `json:"status,omitempty"`
}

type RecordStatus string

const (
	StatusUnknown   RecordStatus = "unknown"
	StatusPending   RecordStatus = "pending"
	StatusPublished RecordStatus = "published"
	StatusRevoked   RecordStatus = "revoked"
	StatusFailed    RecordStatus = "failed"
)

type RecordState struct {
	Status      RecordStatus `json:"status"`
	BlockNumber int64        `json:"number"`
}

func (r *Record) Seal(pk *btcec.PrivateKey) error {

	buf := r.bodyBuffer()

	hash := Hash("ledger record construction", buf.Bytes())

	sig := ecdsa.Sign(pk, hash)

	sigValue := sig.Serialize()

	r.Signature = base58.Encode(sigValue)

	buf.Write(sigValue)

	hash = Hash("ledger record construction", buf.Bytes())

	r.ID = base58.Encode(hash)

	return nil
}

func (r *Record) Verify(publicKey *btcec.PublicKey) (bool, error) {
	sigValue := base58.Decode(r.Signature)
	sig, err := ecdsa.ParseSignature(sigValue)
	if err != nil {
		return false, err
	}

	buf := r.bodyBuffer()

	hash := Hash("ledger record construction", buf.Bytes())

	if signatureVerified := sig.Verify(hash, publicKey); !signatureVerified {
		log.Error().Msg("Signature verification failed for ledger record")
		return false, nil
	}

	buf.Write(sigValue)

	hash = Hash("ledger record construction", buf.Bytes())

	validID := r.ID == base58.Encode(hash)

	if !validID {
		log.Error().Str("expected", base58.Encode(hash)).Str("actual", r.ID).Msg("ID verification failed for ledger record")
	}
	return validID, nil
}

func (r *Record) bodyBuffer() *bytes.Buffer {
	buf := bytes.NewBuffer(nil)
	routingKeyVal := base58.Decode(r.RoutingKey)
	buf.Write(routingKeyVal)
	buf.Write(utils.Uint32ToBytes(r.KeyIndex))
	buf.WriteString(r.OperationAddress)
	if r.Flags > 0 {
		buf.Write(utils.Uint32ToBytes(r.Flags))
	}
	ac, _ := base64.StdEncoding.DecodeString(r.AuthorisingCommitment)
	buf.Write(ac)
	buf.WriteByte(r.AuthorisingCommitmentType)
	rc, _ := base64.StdEncoding.DecodeString(r.RequestingCommitment)
	buf.Write(rc)
	buf.WriteByte(r.RequestingCommitmentType)
	ic, _ := base64.StdEncoding.DecodeString(r.ImpressionCommitment)
	buf.Write(ic)
	buf.WriteByte(r.ImpressionCommitmentType)
	buf.WriteByte(0)

	for _, da := range r.DataAssets {
		val := base58.Decode(da)
		buf.Write(val)
	}

	subjectRecordVal := base58.Decode(r.SubjectRecord)
	buf.Write(subjectRecordVal)
	for _, p := range r.RevocationProof {
		pval, _ := base64.StdEncoding.DecodeString(p)
		buf.Write(pval)
	}

	headID, _ := base64.StdEncoding.DecodeString(r.HeadID)
	buf.Write(headID)
	headBody, _ := base64.StdEncoding.DecodeString(r.HeadBody)
	buf.Write(headBody)

	return buf
}

func (r *Record) Bytes() []byte {
	b, _ := jsonw.Marshal(r)
	return b
}

func (r *Record) ToSlice() []string {
	return []string{
		r.ID,
		r.RoutingKey,
		fmt.Sprintf("%d", r.KeyIndex),
		strconv.Itoa(int(r.Operation)),
		r.OperationAddress,

		r.AuthorisingCommitment,
		strconv.Itoa(int(r.AuthorisingCommitmentType)),
		r.RequestingCommitment,
		strconv.Itoa(int(r.RequestingCommitmentType)),
		r.ImpressionCommitment,
		strconv.Itoa(int(r.ImpressionCommitmentType)),
		r.Signature,
		strings.Join(r.DataAssets, "|"),
		r.SubjectRecord,
		strings.Join(r.RevocationProof, "|"),
	}
}

func (r *Record) Copy() *Record {
	cp := *r
	return &cp
}

func (r *Record) Validate() error {
	switch r.Operation {
	case OpTypeLease:
	case OpTypeLeaseRevocation:
	case OpTypeAssetHead:
		if r.SubjectRecord != "" {
			if len(r.RevocationProof) != 1 {
				return errors.New("invalid head revocation proof")
			}
		}
		if r.HeadID == "" {
			return errors.New("empty head ID")
		}
		if r.HeadBody == "" {
			return errors.New("empty head body")
		}
	default:
		return errors.New("invalid operation type")
	}

	return nil
}

func (r *RecordState) Bytes() []byte {
	b, _ := jsonw.Marshal(r)
	return b
}

func RecordsToCSV(recs []*Record) []byte {
	buf := bytes.NewBuffer(nil)
	w := csv.NewWriter(buf)
	for _, rec := range recs {
		_ = w.Write(rec.ToSlice())
	}
	w.Flush()
	return buf.Bytes()
}

func RandomKeyIndex() uint32 {
	randBuffer := make([]byte, 4)
	if _, err := rand.Read(randBuffer); err != nil {
		panic("failed to generate random uint32")
	}

	idx := utils.BytesToUint32(randBuffer)

	return idx & 0x7fffffff // should be less than 0x80000000 to generate a non-hardened key
}
