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

package apibase_test

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	. "github.com/piprate/metalocker/sdk/apibase"
	"github.com/stretchr/testify/require"
)

func TestExtractSecret_NoSecret(t *testing.T) {
	//nolint:gosec
	tokenLegacySecret := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiIiLCJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20iLCJleHAiOjQ2MDAsImlhdCI6MTAwMCwiaWQiOiJkaWQ6cGlwcmF0ZTpGNzNmNVlyNWlBeGdmRDhaSHJ2S3ZycnBLYkFSMk05OVR4ZURjeTFIWFFBWSIsImlzcyI6InBpcHJhdGUuY29tIn0.clT2rgTZmOXFxYKrlCi9ewRVc0Zt7k2RZnM1h9ATWUhQ_bD9_Bd9MLJgnjxbZGi3GviksuHsTMyez-JZJnmnoQYTkebp9HPSCp1KXMsOCmM-AZB17Tv4fbNCHJbIg28Y7ud3aSinGGNyYbn1A3wIcRi-L8aQcxj9SJt0Wu3SBLg"

	token, _ := jwt.Parse(tokenLegacySecret, func(token *jwt.Token) (any, error) {
		return token, nil
	})
	require.NotNil(t, token)

	secret, err := ExtractSecret(MapClaims(token.Claims.(jwt.MapClaims)), nil)
	require.NoError(t, err)
	require.Nil(t, secret)
}

func TestExtractSecret_WithSecret(t *testing.T) {
	//nolint:gosec
	tokenStr := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJha2lkIjoiNmg2M3RaIiwiYXVkIjoiYXBwMCIsImVhcyI6IlorWERTVTJ1SVY3aGxxRTNLVCtiY3BPemZyM3VqL2FaMnI3VXRkYU1mRW1jaU5aNldRbjgyUGpSMlJKNFM2UkhQUzFNSmVyQUlWc3VuU1U3WGU1UVVEZld6aEZVbTBTaUM2cExuaGJ1NUZJPSIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6NDYwMCwiaWF0IjoxMDAwLCJpZCI6ImRpZDpwaXByYXRlOkY3M2Y1WXI1aUF4Z2ZEOFpIcnZLdnJycEtiQVIyTTk5VHhlRGN5MUhYUUFZIiwiaXNzIjoicGlwcmF0ZS5jb20ifQ.TDA1QPLl2xdLC9oe0KTENec21tRR-2X3LdMIt8OKf679kyPVsSCN-tD2PqjohSYb6CPOf7x0_WSnbmfCG-ylvYf8TOVy7dSf1_fpBebIuwamzfw9ooe9quXY9BGta1yrAb_hZudnP2vE3hcRxxXEj9eHCBGu-r0M78YLX7sr5GM"

	audiencePrivateKey, _ := ReadAudiencePrivateKeyFromString("066+lRtJyQ9WAE8mnfIeWvHVkJtb0i/hO3Zk0SjOHyowJdYDvef/WD2BN8brBHeaTVcs4PO72CKKyyov5aDtmw==")

	token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		return token, nil
	})
	require.NotNil(t, token)

	secret, err := ExtractSecret(MapClaims(token.Claims.(jwt.MapClaims)), audiencePrivateKey)
	require.NoError(t, err)
	require.Equal(t, "MdySTIAiwWHrBm6jygYzcTjpHoFGoVSZcHmeOvMXsmM=", secret.Base64())
}
