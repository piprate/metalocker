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

package examples

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/piprate/metalocker/node"
	"github.com/piprate/metalocker/remote/caller"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	_ "github.com/piprate/metalocker/index/bolt"
	_ "github.com/piprate/metalocker/ledger/local"
	_ "github.com/piprate/metalocker/storage/memory"
	_ "github.com/piprate/metalocker/vaults/fs"
	_ "github.com/piprate/metalocker/vaults/memory"
)

const DemoPort = 32000

func StartDemoServer(serverDir string, debugMode bool) (*node.MetaLockerServer, string, error) {
	// start MetaLocker server

	viper.SetConfigName("config")
	viper.AddConfigPath(serverDir)

	if err := os.WriteFile(filepath.Join(serverDir, "token.rsa"), []byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDPggCi6LD06XmaSQs4oI5Hk9wyuDzr7eoqIZ5FfHsyeWd7DWJ2
a2LdJrFe37brikBNRslrBiN+bQwOp5zZhXu+QiTHKX6JZMAiw7IyJykyE1mMBKwL
UctuaLGHjGNVogLHVTyVY1HxVpGSKWl1+nhpo3HXfAX1YBBXUb3XGJNZlQIDAQAB
AoGAE0N6U6VOaC4Uf+IwDH27N6HeW0cHQM/BYU/lpYW82h6zIJVJgrzNXMJuzOPv
00XuWj4sDKdxPBdbezDMOtVNe11VXja3QEcrTsqS+sTwV7j4zymQr4bcyakwMlLE
b51BOXKXltv0JReOt7xngY58Du6JFaGSJW6mHH19o/vHWrECQQDnhKIb/hcH3cS7
q74o3vRUGqmfUsjuY4T2TdOwhjOEW0NINpGXnLpuZEGZoVnzOVCVe2V+VAeJzRcV
Q5AtdWEPAkEA5XNkq6bdI1bsgtE14+cUHnymCnetc+MX0T0DAh9Y6fKjBxbI6Aah
9v8PIDlMwSlThHPUVz7Yw8SMqnaDHSITGwJBAJkulpvy2IYp45tQnPcp3XswUP7L
pYqlajoVcHUhtkBiqffDsz0fQ/L6frUJnxxg1cKx7ItTSdGRUy6Mj36kZV0CQGCs
7wS36MLEFCDGP1uH+F0kDd2pMSb7zwQ1HbheNttThUcuXXYNnV5xdxEPs3xLiknr
d9NOwowxm0cTagjzW3MCQQDT6v1Kpw5hdgmOE8ag32NpGavIntGofPCQWx7064IS
iFVOt4zaM3/EzLo4qTFMiuX0flK3aONodAyJpM49GY1e
-----END RSA PRIVATE KEY-----`), 0o600); err != nil {
		return nil, "", err
	}

	if err := os.WriteFile(filepath.Join(serverDir, "token.rsa.pub"), []byte(`-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDPggCi6LD06XmaSQs4oI5Hk9wy
uDzr7eoqIZ5FfHsyeWd7DWJ2a2LdJrFe37brikBNRslrBiN+bQwOp5zZhXu+QiTH
KX6JZMAiw7IyJykyE1mMBKwLUctuaLGHjGNVogLHVTyVY1HxVpGSKWl1+nhpo3HX
fAX1YBBXUb3XGJNZlQIDAQAB
-----END PUBLIC KEY-----`), 0o600); err != nil {
		return nil, "", err
	}

	cfg, _, err := node.GenerateConfig(DemoPort, serverDir)
	if err != nil {
		return nil, "", err
	}

	viper.SetConfigType("yaml")

	if err = viper.ReadConfig(bytes.NewReader(cfg)); err != nil {
		return nil, "", err
	}

	srv := node.NewMetaLockerServer(serverDir)

	if err := srv.InitServices(viper.GetViper(), debugMode); err != nil {
		return nil, "", err
	}

	if err := srv.InitAuthentication(viper.GetViper()); err != nil {
		return nil, "", err
	}

	if err := srv.InitStandardRoutes(viper.GetViper()); err != nil {
		return nil, "", err
	}

	go func() {
		if err := srv.Run(viper.GetViper()); err != nil {
			log.Err(err).Msg("Error starting MetaLocker server")
		}
	}()

	url := strings.ReplaceAll(srv.BaseURI(), "https", "https+insecure")

	waitUntilDemoServerIsUp(url)

	return srv, url, nil
}

func waitUntilDemoServerIsUp(url string) {
	httpCaller, err := caller.NewMetaLockerHTTPCaller(url, "MetaLocker Examples")
	if err != nil {
		panic(err)
	}

	// initialise context forwarding
	httpCaller.InitContextForwarding()

	retries := 0
	for retries < 10 {
		_, err = httpCaller.GetServerControls()
		if err == nil {
			return
		}
		time.Sleep(time.Millisecond * 200)
		retries++
	}

	panic("retries exceeded connecting to the demo server")
}
