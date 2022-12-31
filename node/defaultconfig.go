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

package node

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/urfave/cli/v2"
)

const defaultConfigTemplate = `
version: 1.0
host: 127.0.0.1
port: %d
https: true
production: true
httpsCert: cert.pem
httpsKey: key.pem
tokenPublicKey: token.rsa.pub
tokenPrivateKey: token.rsa
issuer: metalocker
defaultAudience: test
defaultAudiencePublicKey: %s
defaultAudiencePrivateKey: %s
jwtTimeout: 43200
allowedHttpOrigins:
  - "*"
administration:
  apiKey: %s
  apiSecret: %s
accountStore:
  type: memory
offChainStore:
  id: %s
  name: offchain
  type: fs
  cas: true
  params:
    root_dir: %s/state/offchain
secondLevelRecoveryKey: %s
ledger:
  type: local
  params:
    dbFile: %s/state/ledger.bolt
vaults:
  - id: %s
    name: local
    type: fs
    params:
      root_dir: %s/state/fs_vault
defaultVaultName: local
`

func GenerateConfig(port int, baseDir string) ([]byte, []byte, error) {
	audiencePublicKey, audiencePrivateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, err
	}

	slrcPublicKey, slrcPrivateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, err
	}

	cfg := fmt.Sprintf(defaultConfigTemplate,
		port,
		base64.StdEncoding.EncodeToString(audiencePublicKey),  // defaultAudiencePublicKey
		base64.StdEncoding.EncodeToString(audiencePrivateKey), // defaultAudiencePrivateKey
		randomBytes(16), // admin key
		randomBytes(32), // admin secret
		randomBytes(32), // offChain store ID
		baseDir,
		base64.StdEncoding.EncodeToString(slrcPublicKey), // secondLevelRecoveryKey
		baseDir,
		randomBytes(32), // fs vault ID
		baseDir,
	)

	return []byte(cfg), slrcPrivateKey, nil
}

func SafeWriteConfigToFile(configDir, configName string) error {
	configFile := filepath.Join(configDir, fmt.Sprintf("%s.yaml", configName))

	_, err := os.Stat(configFile)
	if err == nil {
		return fmt.Errorf("config file already exists: %s", configFile)
	} else if !os.IsNotExist(err) {
		return err
	}

	if _, err = os.Stat(configDir); os.IsNotExist(err) {
		err = os.MkdirAll(configDir, 0o700)
		if err != nil {
			return cli.Exit(err, 1)
		}
	}

	cfg, slrcPrivateKey, err := GenerateConfig(4000, "$HOME/.metalocker")
	if err != nil {
		return cli.Exit(err, 1)
	}

	if err = os.WriteFile(configFile, cfg, 0o600); err != nil {
		return err
	}

	println(`{
  "secondLevelRecoveryPrivateKey": "` + base64.StdEncoding.EncodeToString(slrcPrivateKey) + `"
}`)

	return nil
}

func randomBytes(length int) string {
	randBuffer := make([]byte, length)
	_, err := rand.Read(randBuffer)
	if err != nil {
		panic(err)
	}
	return base58.Encode(randBuffer)
}
