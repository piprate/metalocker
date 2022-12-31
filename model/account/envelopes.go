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

package account

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog/log"
)

type DataEnvelope struct {
	Hash          string            `json:"hash"`
	AccessLevel   model.AccessLevel `json:"lvl"`
	EncryptedID   string            `json:"id,omitempty"`
	EncryptedBody string            `json:"data"`
}

func (ie DataEnvelope) Bytes() []byte {
	b, _ := jsonw.Marshal(ie)
	return b
}

func (ie DataEnvelope) Validate() error {
	if ie.Hash == "" {
		return errors.New("empty hash in data envelope")
	}
	if ie.AccessLevel != model.AccessLevelManaged && ie.AccessLevel != model.AccessLevelHosted {
		return fmt.Errorf("unsupported access level: %d", ie.AccessLevel)
	}
	if ie.EncryptedBody == "" {
		return errors.New("empty body in data envelope")
	}
	return nil
}

func HashID(id string, secret []byte) string {
	h := hmac.New(sha512.New512_256, secret)
	_, _ = h.Write([]byte(id))
	return hex.EncodeToString(h.Sum(nil))
}

func EncryptIdentity(idy *Identity, idSecret []byte, key *model.AESKey) (*DataEnvelope, error) {
	encBytes, err := model.EncryptAESCGM(idy.Bytes(), key)
	if err != nil {
		log.Err(err).Msg("Error encrypting an identity")
		return nil, errors.New("error encrypting an identity")
	}

	e := &DataEnvelope{
		Hash:          HashID(idy.DID.ID, idSecret),
		AccessLevel:   idy.AccessLevel,
		EncryptedBody: base64.StdEncoding.EncodeToString(encBytes),
	}

	return e, nil
}

func DecryptIdentity(envelope *DataEnvelope, key *model.AESKey) (*Identity, error) {
	encBytes, err := base64.StdEncoding.DecodeString(envelope.EncryptedBody)
	if err != nil {
		return nil, err
	}
	idyBytes, err := model.DecryptAESCGM(encBytes, key)
	if err != nil {
		log.Err(err).Msg("Error decrypting an identity")
		return nil, errors.New("error decrypting an identity")
	}
	var idy *Identity
	if err = jsonw.Unmarshal(idyBytes, &idy); err != nil {
		return nil, err
	}
	return idy, nil
}

func EncryptLocker(locker *model.Locker, idSecret []byte, key *model.AESKey) (*DataEnvelope, error) {
	encBytes, err := model.EncryptAESCGM(locker.Bytes(), key)
	if err != nil {
		log.Err(err).Msg("Error encrypting a locker")
		return nil, errors.New("error encrypting a locker")
	}

	e := &DataEnvelope{
		Hash:          HashID(locker.ID, idSecret),
		AccessLevel:   locker.AccessLevel,
		EncryptedBody: base64.StdEncoding.EncodeToString(encBytes),
	}

	return e, nil
}

func DecryptLocker(envelope *DataEnvelope, key *model.AESKey) (*model.Locker, error) {
	encBytes, err := base64.StdEncoding.DecodeString(envelope.EncryptedBody)
	if err != nil {
		return nil, err
	}
	lockerBytes, err := model.DecryptAESCGM(encBytes, key)
	if err != nil {
		log.Err(err).Msg("Error decrypting a locker")
		return nil, errors.New("error decrypting a locker")
	}
	var locker *model.Locker
	if err = jsonw.Unmarshal(lockerBytes, &locker); err != nil {
		return nil, err
	}
	return locker, nil
}

func EncryptValue(key string, val string, lvl model.AccessLevel, idSecret []byte, aesKey *model.AESKey) (*DataEnvelope, error) {
	encKeyBytes, err := model.EncryptAESCGM([]byte(key), aesKey)
	if err != nil {
		log.Err(err).Msg("Error encrypting property key")
		return nil, errors.New("error encrypting property key")
	}

	encBytes, err := model.EncryptAESCGM([]byte(val), aesKey)
	if err != nil {
		log.Err(err).Msg("Error encrypting property value")
		return nil, errors.New("error encrypting property value")
	}

	e := &DataEnvelope{
		Hash:          HashID(key, idSecret),
		AccessLevel:   lvl,
		EncryptedID:   base64.StdEncoding.EncodeToString(encKeyBytes),
		EncryptedBody: base64.StdEncoding.EncodeToString(encBytes),
	}

	return e, nil
}

func DecryptValue(envelope *DataEnvelope, key *model.AESKey, id *string) (string, error) {
	if id != nil {
		encIDBytes, err := base64.StdEncoding.DecodeString(envelope.EncryptedID)
		if err != nil {
			return "", err
		}
		clearIDBytes, err := model.DecryptAESCGM(encIDBytes, key)
		if err != nil {
			log.Err(err).Msg("Error decrypting property id")
			return "", errors.New("error decrypting property id")
		}
		idStr := string(clearIDBytes)
		*id = idStr
	}

	encBytes, err := base64.StdEncoding.DecodeString(envelope.EncryptedBody)
	if err != nil {
		return "", err
	}
	clearBytes, err := model.DecryptAESCGM(encBytes, key)
	if err != nil {
		log.Err(err).Msg("Error decrypting property value")
		return "", errors.New("error decrypting property value")
	}

	return string(clearBytes), nil
}
