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

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/httpsecure"
	"github.com/piprate/metalocker/storage"
)

func (c *MetaLockerHTTPCaller) GetDIDDocument(ctx context.Context, id string) (*model.DIDDocument, error) {
	var ddoc model.DIDDocument
	err := c.client.LoadContents(http.MethodGet, fmt.Sprintf("/v1/did/%s", id), nil, &ddoc)
	if err != nil {
		if errors.Is(err, httpsecure.ErrEntityNotFound) {
			return nil, storage.ErrDIDNotFound
		}
		return nil, err
	} else {
		return &ddoc, nil
	}
}

func (c *MetaLockerHTTPCaller) CreateDIDDocument(ctx context.Context, didDoc *model.DIDDocument) error {
	if !c.client.IsAuthenticated() {
		return errors.New("you need to log in before performing any operations")
	}

	res, err := c.client.SendRequest(http.MethodPost, "/v1/did", httpsecure.WithJSONBody(didDoc))
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
		return fmt.Errorf("identity submission failed with status code %d", res.StatusCode)
	}

	return nil
}

func (c *MetaLockerHTTPCaller) ListDIDDocuments(ctx context.Context) ([]*model.DIDDocument, error) {
	if !c.client.IsAuthenticated() {
		return nil, errors.New("you need to log in before performing any operations")
	}

	var iidList []*model.DIDDocument
	err := c.client.LoadContents(http.MethodGet, "/v1/admin/did", nil, &iidList)
	if err != nil {
		return nil, err
	} else {
		return iidList, nil
	}
}

func (c *MetaLockerHTTPCaller) AdminStoreIdentity(didDoc *model.DIDDocument) error {
	if !c.client.IsAuthenticated() {
		return errors.New("you need to log in before performing any operations")
	}

	res, err := c.client.SendRequest(http.MethodPost, "/v1/admin/did", httpsecure.WithJSONBody(didDoc))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
	// do nothing
	case http.StatusUnauthorized:
		return ErrNotAuthorised
	default:
		return fmt.Errorf("identity upload failed with status code %d", res.StatusCode)
	}

	return nil
}
