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
	"fmt"
	"net/http"

	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/sdk/httpsecure"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog/log"
)

func (c *MetaLockerHTTPCaller) GetServerControls(ctx context.Context) (*ServerControls, error) {
	url := "/v1/status"
	res, err := c.client.SendRequest(ctx, http.MethodGet, url, httpsecure.SkipAuthentication())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		msg := apibase.ParseResponseMessage(res)
		log.Error().Str("url", url).Str("msg", msg).Msg("Call failed")
		return nil, fmt.Errorf("response status code: %d, message: %s", res.StatusCode, msg)
	}

	var controls ServerControls
	if err = jsonw.Decode(res.Body, &controls); err != nil {
		return nil, err
	}

	return &controls, nil
}
