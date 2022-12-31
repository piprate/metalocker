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
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/httpsecure"
	"github.com/piprate/metalocker/utils/jsonw"
)

func (c *MetaLockerHTTPCaller) GetOperation(opAddr string) ([]byte, error) {
	var op []byte
	err := c.client.LoadContents(http.MethodGet, fmt.Sprintf("/v1/lop/%s", opAddr), nil, &op)
	if err != nil {
		if errors.Is(err, httpsecure.ErrEntityNotFound) {
			return nil, model.ErrOperationNotFound
		} else {
			return nil, err
		}
	}
	return op, nil
}

func (c *MetaLockerHTTPCaller) SendOperation(opData []byte) (string, error) {
	res, err := c.client.SendRequest(http.MethodPost, "/v1/lop", httpsecure.WithBody(bytes.NewBuffer(opData)))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		// do nothing
	case http.StatusCreated:
		// do nothing
	case http.StatusUnauthorized:
		return "", ErrNotAuthorised
	case http.StatusConflict:
		return "", errors.New("operation already exists")
	default:
		return "", fmt.Errorf("operation upload failed with status code %d", res.StatusCode)
	}

	var rsp map[string]string
	if err = jsonw.Decode(res.Body, &rsp); err != nil {
		return "", err
	}
	leaseAddress := rsp["id"]

	return leaseAddress, nil
}

func (c *MetaLockerHTTPCaller) PurgeOperation(opAddr string) error {
	res, err := c.client.SendRequest(http.MethodPost, fmt.Sprintf("/v1/lop/%s/purge", opAddr))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		// do nothing
	case http.StatusNotFound:
		return model.ErrOperationNotFound
	case http.StatusUnauthorized:
		return ErrNotAuthorised
	default:
		return fmt.Errorf("operation purge failed with status code %d", res.StatusCode)
	}

	return nil
}
