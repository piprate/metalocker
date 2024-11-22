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
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/sdk/httpsecure"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog/log"
)

func (c *MetaLockerHTTPCaller) GetRecord(ctx context.Context, rid string) (*model.Record, error) {
	var lr model.Record
	err := c.client.LoadContents(ctx, http.MethodGet, fmt.Sprintf("/v1/lrec/%s", rid), nil, &lr)
	if err != nil {
		return nil, err
	} else {
		return &lr, nil
	}
}

func (c *MetaLockerHTTPCaller) SubmitRecord(ctx context.Context, r *model.Record) error {
	res, err := c.client.SendRequest(ctx, http.MethodPost, "/v1/lrec", httpsecure.WithJSONBody(r))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		// do nothing
	case http.StatusCreated:
		// do nothing
	case http.StatusUnauthorized:
		return ErrNotAuthorised
	case http.StatusConflict:
		return errors.New("ledger record already exists")
	default:
		return fmt.Errorf("ledger record upload failed with status code %d", res.StatusCode)
	}

	return nil
}

func (c *MetaLockerHTTPCaller) GetRecordState(ctx context.Context, rid string) (*model.RecordState, error) {
	url := fmt.Sprintf("/v1/lrec/%s/state", rid)
	res, err := c.client.SendRequest(ctx, http.MethodGet, url)
	if err != nil {
		return &model.RecordState{
			Status: model.StatusUnknown,
		}, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		var rs model.RecordState
		err = jsonw.Decode(res.Body, &rs)
		if err != nil {
			rs.Status = model.StatusUnknown
			return &rs, err
		}

		return &rs, nil

	case http.StatusNotFound:
		// the record may not have reached the ledger
		// continue waiting
		return &model.RecordState{
			Status: model.StatusUnknown,
		}, nil

	case http.StatusUnauthorized:
		return &model.RecordState{
			Status: model.StatusUnknown,
		}, fmt.Errorf("unauthorised call: GET %s", url)

	default:
		msg := apibase.ParseResponseMessage(res)
		log.Error().Str("url", url).Str("msg", msg).Msg("Call failed")
		return &model.RecordState{
			Status: model.StatusUnknown,
		}, fmt.Errorf("response status code: %d, message: %s", res.StatusCode, msg)
	}
}

func (c *MetaLockerHTTPCaller) GetAssetHead(ctx context.Context, headID string) (*model.Record, error) {
	var lr model.Record
	err := c.client.LoadContents(ctx, http.MethodGet, fmt.Sprintf("/v1/head/%s", headID), nil, &lr)
	if err != nil {
		if errors.Is(err, httpsecure.ErrEntityNotFound) {
			return nil, model.ErrAssetHeadNotFound
		}
		return nil, err
	} else {
		return &lr, nil
	}
}
