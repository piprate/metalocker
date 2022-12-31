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
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/utils"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type PasswordHashFunction func(string) (string, error)

// Note: this call is expensive when invoked with the default hashing function (recommended).
func ReHashPassphrase(acct *Account, hashFunction PasswordHashFunction) error {
	if hashFunction == nil {
		hashFunction = func(password string) (string, error) {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
			if err != nil {
				return "", err
			} else {
				return string(hashedPassword), nil
			}
		}
	}

	hashedPassword, err := hashFunction(acct.EncryptedPassword)
	if err != nil {
		return err
	}
	acct.EncryptedPassword = hashedPassword

	return nil
}

func HashUserPassword(passphrase string) string {
	return base64.StdEncoding.EncodeToString(utils.DoubleSha256([]byte(passphrase)))
}

func ChangePassphrase(acct *Account, currentPassphrase, newPassphrase string, isHash bool) (*Account, error) {
	if acct.AccessLevel == model.AccessLevelHosted && isHash {
		return nil, errors.New("can't change passphrase using its hash for hosted accounts")
	}

	acct = acct.Copy()

	var currentPassphraseHash string
	var newPassphraseHash string
	if isHash {
		currentPassphraseHash = currentPassphrase
		newPassphraseHash = newPassphrase
	} else {
		currentPassphraseHash = HashUserPassword(currentPassphrase)
		newPassphraseHash = HashUserPassword(newPassphrase)
	}
	acct.EncryptedPassword = newPassphraseHash

	managedKey, err := acct.ManagedSecretStore.ExtractPayloadKey(currentPassphraseHash)
	if err != nil {
		return nil, err
	}

	switch acct.AccessLevel {
	case model.AccessLevelHosted:

		hostedMasterPassphrase := []byte(newPassphrase)
		newHostedMasterKey, err := newSecretKey(&hostedMasterPassphrase, hostedAccountConfig)
		if err != nil {
			log.Err(err).Msg("Failed to create master key")
			return nil, err
		}

		hostedKey, err := acct.HostedSecretStore.ExtractPayloadKey(currentPassphrase)
		if err != nil {
			return nil, err
		}

		// Encrypt the crypto keys with the associated master keys.
		cryptoKeyPrivEncrypted, err := newHostedMasterKey.Encrypt(hostedKey.Bytes())
		if err != nil {
			return nil, err
		}

		acct.HostedSecretStore.MasterKeyParams = base64.StdEncoding.EncodeToString(newHostedMasterKey.Marshal())
		acct.HostedSecretStore.EncryptedPayloadKey = base64.StdEncoding.EncodeToString(cryptoKeyPrivEncrypted)

		managedMasterPassphrase := []byte(newPassphraseHash)
		newManagedMasterKey, err := newSecretKey(&managedMasterPassphrase, managedAccountConfig)
		if err != nil {
			log.Err(err).Msg("Failed to create master key")
			return nil, err
		}

		// Encrypt the crypto keys with the associated master keys.
		managedCryptoKeyEncrypted, err := newManagedMasterKey.Encrypt(managedKey.Bytes())
		if err != nil {
			return nil, err
		}

		acct.ManagedSecretStore.MasterKeyParams = base64.StdEncoding.EncodeToString(newManagedMasterKey.Marshal())
		acct.ManagedSecretStore.EncryptedPayloadKey = base64.StdEncoding.EncodeToString(managedCryptoKeyEncrypted)
	case model.AccessLevelManaged:
		// managed account

		managedMasterPassphrase := []byte(newPassphraseHash)
		newManagedMasterKey, err := newSecretKey(&managedMasterPassphrase, managedAccountConfig)
		if err != nil {
			log.Err(err).Msg("Failed to create master key")
			return nil, err
		}

		// Encrypt the crypto keys with the associated master keys.
		managedCryptoKeyEncrypted, err := newManagedMasterKey.Encrypt(managedKey.Bytes())
		if err != nil {
			return nil, err
		}

		acct.ManagedSecretStore.MasterKeyParams = base64.StdEncoding.EncodeToString(newManagedMasterKey.Marshal())
		acct.ManagedSecretStore.EncryptedPayloadKey = base64.StdEncoding.EncodeToString(managedCryptoKeyEncrypted)
	default:
		return nil, fmt.Errorf("passphrase change not allowed for accounts with access level %d", acct.AccessLevel)
	}

	return acct, nil
}
