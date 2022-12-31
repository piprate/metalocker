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

package caller

import (
	"fmt"
	"io"
	"net/http"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/sdk/httpsecure"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/vaults"
)

func (c *MetaLockerHTTPCaller) GetBlob(res *model.StoredResource, accessToken string) (io.ReadCloser, error) {
	return vaults.ReceiveBlob(res, accessToken, func(res *model.StoredResource, accessToken string) (io.ReadCloser, error) {
		url := fmt.Sprintf("/v1/vault/%s/serve", res.Vault)

		truncatedRes := model.StoredResource{
			ID:     res.ID,
			Method: res.Method,
			Params: res.Params,
		}

		rsp, err := c.client.SendRequest(http.MethodPost, url,
			httpsecure.WithHeaders(map[string]string{
				"X-Vault-Access-Token": accessToken,
			}),
			httpsecure.WithJSONBody(truncatedRes))
		if err != nil {
			return nil, err
		}

		switch rsp.StatusCode {
		case http.StatusOK:
			return rsp.Body, nil
		case http.StatusNotFound:
			return nil, model.ErrBlobNotFound
		case http.StatusUnauthorized:
			defer rsp.Body.Close()
			return nil, fmt.Errorf("unauthorised blob retrieval: %s", apibase.ParseResponseMessage(rsp))
		default:
			return nil, fmt.Errorf("bad response status code: %d", rsp.StatusCode)
		}
	})
}

func (c *MetaLockerHTTPCaller) PurgeBlob(res *model.StoredResource) error {
	url := fmt.Sprintf("/v1/vault/%s/purge", res.Vault)

	truncatedRes := model.StoredResource{
		ID:     res.ID,
		Method: res.Method,
		Params: res.Params,
	}

	rsp, err := c.client.SendRequest(http.MethodPost, url, httpsecure.WithJSONBody(truncatedRes))
	if err != nil {
		return err
	}

	switch rsp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return model.ErrBlobNotFound
	case http.StatusUnauthorized:
		defer rsp.Body.Close()
		return fmt.Errorf("unauthorised blob purge: %s", apibase.ParseResponseMessage(rsp))
	default:
		return fmt.Errorf("bad response status code: %d", rsp.StatusCode)
	}
}

func (c *MetaLockerHTTPCaller) SendBlob(data io.Reader, cleartext bool, vaultName string) (*model.StoredResource, error) {
	vaultMap, err := c.GetVaultMap()
	if err != nil {
		return nil, err
	}

	vault, found := vaultMap[vaultName]
	if !found {
		return nil, fmt.Errorf("vault not found: %s", vaultName)
	}

	return vaults.SendBlob(data, vault.ID, vault.SSE || cleartext, func(data io.Reader, vaultID string) (*model.StoredResource, error) {
		rsp, err := c.client.SendRequest(http.MethodPost, fmt.Sprintf("/v1/vault/%s/raw", vaultID), httpsecure.WithUnsignedBody(data))
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()

		switch rsp.StatusCode {
		case http.StatusOK:
			// do nothing
		case http.StatusNotFound:
			return nil, fmt.Errorf("endpoint not found for vault '%s'", vaultID)
		case http.StatusUnauthorized:
			return nil, ErrNotAuthorised
		default:
			errorMsg, err := io.ReadAll(rsp.Body)
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("store file operation failed with status code %d: %s", rsp.StatusCode, string(errorMsg))
		}

		var res model.StoredResource
		if err = jsonw.Decode(rsp.Body, &res); err != nil {
			return nil, err
		}

		return &res, nil
	})
}

func (c *MetaLockerHTTPCaller) GetDataAssetState(id string) (model.DataAssetState, error) {
	var rsp map[string]model.DataAssetState
	err := c.client.LoadContents(http.MethodGet, fmt.Sprintf("/v1/ledger/data-asset/%s/state", id), nil, &rsp)
	if err != nil {
		return model.DataAssetStateKeep, err
	}
	return rsp["state"], nil
}

func (c *MetaLockerHTTPCaller) GetVaultMap() (map[string]*model.VaultProperties, error) {
	if c.cachedVaultMap == nil {
		var vaultList []*model.VaultProperties
		err := c.client.LoadContents(http.MethodGet, "/v1/vault/list", nil, &vaultList)
		if err != nil {
			return nil, err
		}

		m := make(map[string]*model.VaultProperties, len(vaultList))
		for _, v := range vaultList {
			m[v.Name] = v
		}

		c.cachedVaultMap = m
	}

	return c.cachedVaultMap, nil
}
