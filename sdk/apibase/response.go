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

package apibase

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

const (
	StatusOK     = "ok"
	StatusError  = "error"
	StatusFailed = "failed"
)

type (
	Response struct {
		Status  string   `json:"status,omitempty"`
		Message string   `json:"message,omitempty"`
		Errors  []string `json:"errors,omitempty"`
	}
)

func ParseResponseMessage(res *http.Response) string {
	var rsp Response
	err := json.NewDecoder(res.Body).Decode(&rsp)
	if err != nil {
		log.Warn().Msg("Failed to parse response message")
		return ""
	}
	if rsp.Message != "" {
		return rsp.Message
	} else {
		log.Warn().Msg("Failed to parse response message")
		return ""
	}
}
