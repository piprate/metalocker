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
	"errors"
	"fmt"
	"net/http"

	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/sdk/httpsecure"
	"github.com/piprate/metalocker/storage"
)

func (c *MetaLockerHTTPCaller) StoreIdentity(idy *account.DataEnvelope) error {
	return c.storeDataEnvelope("identity", idy)
}

func (c *MetaLockerHTTPCaller) GetIdentity(hash string) (*account.DataEnvelope, error) {
	return c.getDataEnvelope("identity", hash, storage.ErrIdentityNotFound)
}

func (c *MetaLockerHTTPCaller) ListIdentities() ([]*account.DataEnvelope, error) {
	return c.listDataEnvelopes("identity")
}

func (c *MetaLockerHTTPCaller) storeDataEnvelope(entityType string, env *account.DataEnvelope) error {
	if !c.client.IsAuthenticated() {
		return errors.New("you need to log in before performing any operations")
	}

	url := "/v1/account/" + c.currentAccountID + "/" + entityType
	res, err := c.client.SendRequest(http.MethodPost, url, httpsecure.WithJSONBody(env))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusCreated:
	// do nothing
	case http.StatusUnauthorized:
		return ErrNotAuthorised
	default:
		return fmt.Errorf("%s submission failed with status code %d", entityType, res.StatusCode)
	}

	return nil
}

func (c *MetaLockerHTTPCaller) getDataEnvelope(entityType, hash string, entityNotFoundErr error) (*account.DataEnvelope, error) {
	if !c.client.IsAuthenticated() {
		return nil, errors.New("you need to log in before performing any operations")
	}

	var env account.DataEnvelope
	url := "/v1/account/" + c.currentAccountID + "/" + entityType + "/" + hash
	err := c.client.LoadContents(http.MethodGet, url, nil, &env)
	if err != nil {
		if errors.Is(err, httpsecure.ErrEntityNotFound) {
			return nil, entityNotFoundErr
		}
		return nil, err
	} else {
		return &env, nil
	}
}

func (c *MetaLockerHTTPCaller) listDataEnvelopes(entityType string) ([]*account.DataEnvelope, error) {
	if !c.client.IsAuthenticated() {
		return nil, errors.New("you need to log in before performing any operations")
	}

	var envList []*account.DataEnvelope
	url := "/v1/account/" + c.currentAccountID + "/" + entityType
	err := c.client.LoadContents(http.MethodGet, url, nil, &envList)
	if err != nil {
		return nil, err
	} else {
		return envList, nil
	}
}

func (c *MetaLockerHTTPCaller) deleteDataEnvelope(entityType, hash string, entityNotFoundErr error) error {
	if !c.client.IsAuthenticated() {
		return errors.New("you need to log in before performing any operations")
	}

	url := "/v1/account/" + c.currentAccountID + "/" + entityType + "/" + hash
	res, err := c.client.SendRequest(http.MethodDelete, url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusNoContent:
	// do nothing
	case http.StatusNotFound:
		return entityNotFoundErr
	case http.StatusUnauthorized:
		return ErrNotAuthorised
	default:
		return fmt.Errorf("%s deletion failed with status code %d", entityType, res.StatusCode)
	}

	return nil
}
