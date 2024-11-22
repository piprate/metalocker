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
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/sdk/httpsecure"
	"github.com/piprate/metalocker/storage"
)

func (c *MetaLockerHTTPCaller) StoreIdentity(ctx context.Context, idy *account.DataEnvelope) error {
	return c.storeDataEnvelope(ctx, "identity", idy)
}

func (c *MetaLockerHTTPCaller) GetIdentity(ctx context.Context, hash string) (*account.DataEnvelope, error) {
	return c.getDataEnvelope(ctx, "identity", hash, storage.ErrIdentityNotFound)
}

func (c *MetaLockerHTTPCaller) ListIdentities(ctx context.Context) ([]*account.DataEnvelope, error) {
	return c.listDataEnvelopes(ctx, "identity")
}

func (c *MetaLockerHTTPCaller) storeDataEnvelope(ctx context.Context, entityType string, env *account.DataEnvelope) error {
	if !c.client.IsAuthenticated() {
		return errors.New("you need to log in before performing any operations")
	}

	url := "/v1/account/" + c.currentAccountID + "/" + entityType
	res, err := c.client.SendRequest(ctx, http.MethodPost, url, httpsecure.WithJSONBody(env))
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

func (c *MetaLockerHTTPCaller) getDataEnvelope(ctx context.Context, entityType, hash string, entityNotFoundErr error) (*account.DataEnvelope, error) {
	if !c.client.IsAuthenticated() {
		return nil, errors.New("you need to log in before performing any operations")
	}

	var env account.DataEnvelope
	url := "/v1/account/" + c.currentAccountID + "/" + entityType + "/" + hash
	err := c.client.LoadContents(ctx, http.MethodGet, url, nil, &env)
	if err != nil {
		if errors.Is(err, httpsecure.ErrEntityNotFound) {
			return nil, entityNotFoundErr
		}
		return nil, err
	} else {
		return &env, nil
	}
}

func (c *MetaLockerHTTPCaller) listDataEnvelopes(ctx context.Context, entityType string) ([]*account.DataEnvelope, error) {
	if !c.client.IsAuthenticated() {
		return nil, errors.New("you need to log in before performing any operations")
	}

	var envList []*account.DataEnvelope
	url := "/v1/account/" + c.currentAccountID + "/" + entityType
	err := c.client.LoadContents(ctx, http.MethodGet, url, nil, &envList)
	if err != nil {
		return nil, err
	} else {
		return envList, nil
	}
}

func (c *MetaLockerHTTPCaller) deleteDataEnvelope(ctx context.Context, entityType, hash string, entityNotFoundErr error) error {
	if !c.client.IsAuthenticated() {
		return errors.New("you need to log in before performing any operations")
	}

	url := "/v1/account/" + c.currentAccountID + "/" + entityType + "/" + hash
	res, err := c.client.SendRequest(ctx, http.MethodDelete, url)
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
