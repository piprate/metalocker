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
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	. "github.com/piprate/metalocker/model"
	"github.com/stretchr/testify/assert"
)

type MockLedger struct {
	rec                      *Record
	dataAssetState           DataAssetState
	errorAtGetDataAssetState error
	status                   RecordStatus
	errorAtGetRecord         error
}

func (ml *MockLedger) GetRecord(rid string) (*Record, error) {
	if ml.errorAtGetRecord != nil {
		return nil, ml.errorAtGetRecord
	} else {
		return ml.rec, nil
	}
}

func (ml *MockLedger) GetDataAssetState(id string) (DataAssetState, error) {
	if ml.errorAtGetDataAssetState != nil {
		return 0, ml.errorAtGetDataAssetState
	} else {
		return ml.dataAssetState, nil
	}
}

func (ml *MockLedger) GetRecordState(rid string) (*RecordState, error) {
	return &RecordState{
		Status: ml.status,
	}, nil
}

func TestGenerateAccessToken(t *testing.T) {
	recordID := "5WPE1m5rKnhjZqgRPRRWrfddBwWP3w3Q1rXYQXBVZXd5"
	leaseID := "FB4PFMJuqCoW5WxwW2sXw6y48gt8WkiPDf4KKVcHYpGf"
	tokenTime := time.Unix(1000, 0).UTC().Unix()
	leaseExpiryTime := time.Unix(2000, 0).UTC().Unix()
	at := GenerateAccessToken(recordID, leaseID, tokenTime, leaseExpiryTime)
	assert.Equal(t, "5WPE1m5rKnhjZqgRPRRWrfddBwWP3w3Q1rXYQXBVZXd5.2000.pgvwS872MxHt+wR0+y0FDFWvlkpXNMy1As8loGoUuRE=.1000.Bu94Hv44pZaUNUBcPHh6zAqTs7ndS2OcxlGI1dba8SPHQg2oVWbDrRle+fRRwmUmwo+naRQu6bxQT5bzOuYfDg==", at)
}

func TestVerifyAccessToken(t *testing.T) {
	recordID := "5WPE1m5rKnhjZqgRPRRWrfddBwWP3w3Q1rXYQXBVZXd5"
	leaseID := "FB4PFMJuqCoW5WxwW2sXw6y48gt8WkiPDf4KKVcHYpGf"
	dataAssetID := "2uGCmHBRJjqUY8LK26DzZ9yCnexFtprGkM2CJfQbqajA"
	leaseExpiryTime := time.Unix(1025, 0).UTC()

	// test empty access token, error when reading data asset state

	verifier := &MockLedger{
		errorAtGetDataAssetState: errors.New("some error"),
	}

	assert.False(t, VerifyAccessToken("", "asset-id", time.Now().Unix(), 30, verifier))

	// test empty access token, data asset state not found

	verifier = &MockLedger{
		dataAssetState: DataAssetStateNotFound,
	}

	assert.True(t, VerifyAccessToken("", "asset-id", time.Now().Unix(), 30, verifier))

	// bad asset token

	assert.False(t, VerifyAccessToken("bad.format", "asset-id", time.Now().Unix(), 30, verifier))

	// build requesting commitment

	rc := sha256.Sum256(
		BuildRequestingCommitmentInput(
			leaseID,
			&leaseExpiryTime,
		),
	)

	verifier = &MockLedger{
		rec: &Record{
			ID:                   recordID,
			RequestingCommitment: base64.StdEncoding.EncodeToString(rc[:]),
			DataAssets: []string{
				dataAssetID,
			},
			Status: StatusPublished,
		},
		dataAssetState: DataAssetStateKeep,
		status:         StatusPublished,
	}

	// build access token

	tokenTime := time.Unix(1000, 0).UTC().Unix()
	at := GenerateAccessToken(recordID, leaseID, tokenTime, leaseExpiryTime.Unix())

	// verify token (status == published)

	now := time.Unix(1020, 0).UTC().Unix()

	res := VerifyAccessToken(at, dataAssetID, now, 30, verifier)
	assert.True(t, res)

	// verify token (token expired)

	res = VerifyAccessToken(at, dataAssetID, now, 10, verifier)
	assert.False(t, res)

	// verify token (bad public key)

	s := strings.Split(at, ".")
	s[2] = "bad_key"

	res = VerifyAccessToken(strings.Join(s, "."), dataAssetID, now, 30, verifier)
	assert.False(t, res)

	// verify token (bad signature format)

	s = strings.Split(at, ".")
	s[4] = "bad_sig"

	res = VerifyAccessToken(strings.Join(s, "."), dataAssetID, now, 30, verifier)
	assert.False(t, res)

	// verify token (invalid signature)

	s = strings.Split(at, ".")
	s[4] = base64.StdEncoding.EncodeToString([]byte("bad_sig"))

	res = VerifyAccessToken(strings.Join(s, "."), dataAssetID, now, 30, verifier)
	assert.False(t, res)

	// verify token (failed to read the record)

	verifier.errorAtGetRecord = errors.New("some error")

	res = VerifyAccessToken(at, dataAssetID, now, 30, verifier)
	assert.False(t, res)

	verifier.errorAtGetRecord = ErrRecordNotFound

	res = VerifyAccessToken(at, dataAssetID, now, 30, verifier)
	assert.False(t, res)

	verifier.errorAtGetRecord = nil

	// verify token (lease expired)

	now = time.Unix(1027, 0).UTC().Unix()

	res = VerifyAccessToken(at, dataAssetID, now, 30, verifier)
	assert.False(t, res)

	// verify token (status == revoked)

	now = time.Unix(1020, 0).UTC().Unix()
	verifier.rec.Status = StatusRevoked

	res = VerifyAccessToken(at, dataAssetID, now, 30, verifier)
	assert.False(t, res)

	verifier.rec.Status = StatusPublished

	// verify token (invalid RC)

	badToken := GenerateAccessToken(recordID, leaseID, tokenTime, leaseExpiryTime.Unix()+1)

	res = VerifyAccessToken(badToken, dataAssetID, now, 30, verifier)
	assert.False(t, res)

	// verify token (asset not found)

	verifier.rec.DataAssets[0] = "123"

	res = VerifyAccessToken(at, dataAssetID, now, 30, verifier)
	assert.False(t, res)
}
