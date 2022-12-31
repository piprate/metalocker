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

package index

import (
	"fmt"
	"time"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/utils/jsonw"
)

const (
	TypeRoot = "idx:root"
)

type (
	RecordState struct {
		ID            string             `json:"id"`
		Operation     model.OpType       `json:"op"`
		Status        model.RecordStatus `json:"status"`
		LockerID      string             `json:"locker"`
		ParticipantID string             `json:"participant"`
		BlockNumber   int64              `json:"blockNumber"`
		Index         uint32             `json:"index"`
		ImpressionID  string             `json:"impression,omitempty"`
		ContentType   string             `json:"contentType,omitempty"`
	}

	AssetState struct {
		ImpressionID     string `json:"impression"`
		AssetID          string `json:"asset"`
		ContentType      string `json:"contentType"`
		RevisionNumber   int64  `json:"revisionNumber,omitempty"`
		WasRevisionOf    string `json:"wasRevisionOf,omitempty"`
		SpecializationOf string `json:"specializationOf,omitempty"`
	}

	VariantRecordState struct {
		ID             string             `json:"id"`
		Operation      model.OpType       `json:"op"`
		Status         model.RecordStatus `json:"status"`
		LockerID       string             `json:"locker"`
		ParticipantID  string             `json:"participant"`
		BlockNumber    int64              `json:"block"`
		Index          uint32             `json:"index"`
		AssetID        string             `json:"asset"`
		ImpressionID   string             `json:"imp"`
		RevisionNumber int64              `json:"rev"`
		CreatedAt      *time.Time         `json:"at"`
		ContentType    string             `json:"ctype"`
	}

	AssetRecordVisitor func(recordID string, r *AssetState) error
	VariantVisitor     func(variantID string, master *VariantRecordState, history []*VariantRecordState) error
	RecordVisitor      func(r *RecordState) error

	// RootIndex is a special type of index used by data wallets to provide fast access to ledger records
	// and facilitate generic operations over available lockers and records.
	RootIndex interface {
		Index

		GetRecord(recordID string) (*RecordState, error)
		TraverseRecords(lockerFilter, participantFilter string, vFunc RecordVisitor, maxRecords uint64) error
		TraverseVariants(lockerFilter, participantFilter string, vFunc VariantVisitor, includeHistory bool, maxVariants uint64) error
		TraverseAssetRecords(assetID string, vFunc AssetRecordVisitor, maxRecords uint64) error
		GetRecordsByImpressionID(impID string, lockerFilter map[string]bool) ([]string, error)
		GetVariant(variantID string, includeHistory bool) (*VariantRecordState, []*VariantRecordState, error)
	}

	RootIndexParameters struct {
		ClientKey []byte `json:"key,omitempty"`
	}
)

func (ls *LockerState) Bytes() []byte {
	b, _ := jsonw.Marshal(ls)
	return b
}

func (rs *RecordState) Bytes() []byte {
	b, _ := jsonw.Marshal(rs)
	return b
}

func (as *AssetState) Bytes() []byte {
	b, _ := jsonw.Marshal(as)
	return b
}

func (rs *VariantRecordState) Bytes() []byte {
	b, _ := jsonw.Marshal(rs)
	return b
}

func RootIndexID(userID string, accessLevel model.AccessLevel) string {
	return fmt.Sprintf("%s#%d", userID, accessLevel)
}
