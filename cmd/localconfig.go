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

package cmd

import (
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
)

var (
	configDirName = ".metalocker"
)

func GetMetaLockerConfigDir() string {
	currentUser, err := user.Current()
	if err != nil {
		panic(err)
	}
	return path.Join(currentUser.HomeDir, configDirName)
}

func SetConfigDirName(name string) {
	configDirName = name
}

func GetWalletPath(connectionURL, genesisBlockID, accountID string) (string, error) {

	homeDir := GetMetaLockerConfigDir()

	serverFolder := strings.ReplaceAll(
		strings.ReplaceAll(connectionURL, ":", "_"),
		"/",
		"_",
	)

	accountID = strings.ReplaceAll(accountID, "@", "_at_")

	walletFilePath := filepath.Join(homeDir, "wallet", serverFolder, genesisBlockID, accountID, "data_wallet.bolt")

	return walletFilePath, nil
}

func InstallLocalFile(relativePath []string, fileName string, data []byte) error {
	configDir := GetMetaLockerConfigDir()
	pathElem := []string{configDir}
	if relativePath != nil {
		pathElem = append(pathElem, relativePath...)
	}
	fullPath := path.Join(pathElem...)
	err := os.MkdirAll(fullPath, 0o700)
	if err != nil {
		return err
	}

	pathElem = append(pathElem, fileName)
	fullFileName := path.Join(pathElem...)

	return os.WriteFile(fullFileName, data, 0o600)
}

func ReadLocalFile(relativePath []string, fileName string) ([]byte, error) {
	configDir := GetMetaLockerConfigDir()
	pathElem := []string{configDir}
	if relativePath != nil {
		pathElem = append(pathElem, relativePath...)
	}
	pathElem = append(pathElem, fileName)
	fullFileName := path.Join(pathElem...)

	return os.ReadFile(fullFileName)
}
