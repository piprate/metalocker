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

package actions

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/piprate/json-gold/ld"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/urfave/cli/v2"
)

func SampleCrypto(c *cli.Context) error {

	name := c.String("name")
	if name == "" {
		name = "Sample Identity"
	}

	skipLocker := c.Bool("skip-locker")

	a, err := model.GenerateNewSemanticAsset(false, false, "", "")
	if err != nil {
		return cli.Exit(err, OperationFailed)
	}

	key := model.NewEncryptionKey()
	keyBase64 := base64.StdEncoding.EncodeToString(key[:])

	passphrase := base58.Encode(model.NewEncryptionKey()[:])
	passphraseDoubleSha256 := account.HashUserPassword(passphrase)

	// generate sample ed25519 key

	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return cli.Exit(err, OperationFailed)
	}

	// generate sample identity

	did, err := model.GenerateDID()
	if err != nil {
		return cli.Exit(err, OperationFailed)
	}

	now := time.Now()
	var lockers []*model.Locker

	if !skipLocker {
		expiryTime := time.Now().AddDate(0, 120, 0).UTC()

		// add identity locker

		locker, err := model.GenerateLocker(model.AccessLevelHosted, name, &expiryTime, 1, model.Us(did, nil))
		if err != nil {
			return cli.Exit(err, OperationFailed)
		}

		lockers = append(lockers, locker)
	}

	idy := &account.Identity{
		DID:         did,
		Created:     &now,
		Name:        name,
		Type:        account.IdentityTypePersona,
		AccessLevel: model.AccessLevelHosted,
		Lockers:     lockers,
	}

	sample := map[string]any{
		"sampleApiKeys": map[string]any{
			"key":    randomKey(16),
			"secret": randomKey(32),
		},
		"sampleAESKeyBase64": keyBase64,
		"sampleEd25519Key": map[string]any{
			"publicKeyBase64":  base64.StdEncoding.EncodeToString(publicKey),
			"privateKeyBase64": base64.StdEncoding.EncodeToString(privateKey),
		},
		"sampleAssetID":  a.ID,
		"sampleIdentity": idy,
		"samplePassphrase": map[string]any{
			"value":        passphrase,
			"doubleSha256": passphraseDoubleSha256,
		},
	}

	ld.PrintDocument("", sample)

	return nil
}

func randomKey(length int) string {
	randBuffer := make([]byte, length)
	_, err := rand.Read(randBuffer)
	if err != nil {
		panic(err)
	}
	return base58.Encode(randBuffer)
}
