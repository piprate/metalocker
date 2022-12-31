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
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/piprate/metalocker/utils"
	"github.com/rs/zerolog/log"
)

func GenerateAccessToken(recordID, leaseID string, now, leaseExpiryTime int64) string {
	pub, priv := DeriveStorageAccessKey(leaseID)
	msg := fmt.Sprintf(
		"%s.%d.%s.%d",
		recordID,
		leaseExpiryTime,
		base64.StdEncoding.EncodeToString(pub),
		now,
	)
	sig := ed25519.Sign(priv, []byte(msg))
	return fmt.Sprintf("%s.%s", msg, base64.StdEncoding.EncodeToString(sig))
}

var (
	// DefaultMaxDistanceSeconds is 5 minutes
	DefaultMaxDistanceSeconds int64 = 5 * 60
)

func VerifyAccessToken(at, dataAssetID string, now, maxDistanceSeconds int64, ledger AccessVerifier) bool {
	if at == "" {

		// check if this data asset isn't attached to any leases yet

		state, err := ledger.GetDataAssetState(dataAssetID)
		if err != nil {
			log.Err(err).Str("daid", dataAssetID).Msg("Error when reading data asset state from ledger")
			return false
		}

		if state == DataAssetStateNotFound {
			log.Info().Str("daid", dataAssetID).
				Msg("Data asset isn't attached to any leases yet. Allow access.")
			return true
		}
	}
	s := strings.Split(at, ".")
	if len(s) != 5 {
		log.Error().Str("at", at).Msg("Bad access token format")
		return false
	}

	recordID := s[0]

	// verify that distance between token creation and now is less than allowed threshold
	// (to prevent replay attacks)

	tokenTime := utils.StringToInt64(s[3])
	if now-tokenTime > maxDistanceSeconds {
		log.Error().Msg("Access token expired")
		return false
	}

	// verify that the lease hasn't expired yet

	leaseExpiryTime := utils.StringToInt64(s[1])
	if leaseExpiryTime != 0 && leaseExpiryTime < now {
		log.Error().Str("rid", recordID).Msg("Lease expired")
		return false
	}

	// check the token signature

	var pub ed25519.PublicKey
	var err error
	pub, err = base64.StdEncoding.DecodeString(s[2])
	if err != nil {
		log.Err(err).Msg("Error decoding public key in access token")
		return false
	}

	sig, err := base64.StdEncoding.DecodeString(s[4])
	if err != nil {
		log.Err(err).Msg("Error decoding signature in access token")
		return false
	}

	if !ed25519.Verify(pub, []byte(at[:strings.LastIndex(at, ".")]), sig) {
		log.Error().Msg("Signature verification failed")
		return false
	}

	// verify that the requester has access to the relevant record

	rec, err := ledger.GetRecord(recordID)
	if err != nil {
		if !errors.Is(err, ErrRecordNotFound) {
			log.Err(err).Str("rid", recordID).Msg("Error when reading record from ledger")
		} else {
			log.Err(err).Msg("Error reading record from ledger")
		}
		return false
	}

	if rec.Status == StatusRevoked {
		return false
	}

	expectedRC, _ := base64.StdEncoding.DecodeString(rec.RequestingCommitment)

	rcInput := pub
	if leaseExpiryTime != 0 {
		rcInput = append(rcInput, utils.Int64ToBytes(leaseExpiryTime)...)
	}

	actualRC := sha256.Sum256(Hash(RequestingCommitmentTag, rcInput))

	if !bytes.Equal(expectedRC, actualRC[:]) {
		log.Error().Msg("Requesting commitment check failed")
		return false
	}

	for _, id := range rec.DataAssets {
		if id == dataAssetID {
			return true
		}
	}

	// data asset not found

	log.Error().Str("rid", recordID).Str("daid", dataAssetID).Msg("Data asset not found in record")
	return false
}
