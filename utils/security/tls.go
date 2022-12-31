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

package security

/*
	This module contains functions that allow https+insecure as a scheme in the service URL
	to indicate that we should skip TLS certificate verification. This is useful in
	development environments.
*/

import (
	"crypto/tls"
	"net/http"
	"net/url"
)

const InsecureHTTPSScheme = "https+insecure"

func CreateHTTPClient(serviceURL string) (string, *http.Client, *tls.Config, error) {
	urlStruct, err := url.Parse(serviceURL)
	if err != nil {
		return "", nil, nil, err
	}

	var tlsConfig *tls.Config
	if urlStruct.Scheme == InsecureHTTPSScheme {
		//skip tls verification for non-production request
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
		}
		urlStruct.Scheme = "https"
		serviceURL = urlStruct.String()
	}

	httpClient := &http.Client{}
	if tlsConfig != nil {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	return serviceURL, httpClient, tlsConfig, nil
}
