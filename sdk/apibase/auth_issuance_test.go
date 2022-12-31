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
	"bytes"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/piprate/json-gold/ld"
	"github.com/piprate/metalocker/contexts"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	. "github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/sdk/testbase"
	"github.com/piprate/metalocker/storage/memory"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	authBody struct {
		Code   int       `json:"code"`
		Expire time.Time `json:"expire"`
		Token  string    `json:"token"`
	}
)

func init() {
	_ = contexts.PreloadContextsIntoMemory()
	testbase.SetupLogFormat()

	gin.SetMode(gin.ReleaseMode)
}

func invokeLoginHandler(t *testing.T, body string, authenticatorFn func(c *gin.Context) (any, error)) *httptest.ResponseRecorder {

	t.Helper()

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request, _ = http.NewRequest(http.MethodPost, "/v1/authenticate", bytes.NewReader([]byte(body)))

	privateKeyPath := "testdata/token.rsa"
	publicKeyPath := "testdata/token.rsa.pub"

	dummyPrivateKey := "8L0x196HUEQMCzOg5Y23jK7OcVQLOesCqkiA6ZnvPfzAtKdyqLIFGUiAIv5zyB8xLVWt1eerphva3lq+cVh7jQ=="

	mw, err := JWTMiddlewareWithTokenIssuance("Test Realm", "piprate.com", authenticatorFn,
		dummyPrivateKey, privateKeyPath, publicKeyPath, time.Hour, func() time.Time {
			return time.Unix(1000, 0).UTC()
		})
	require.NoError(t, err)
	mw.LoginHandler(c)

	return rec
}

func TestAuthenticationHandler_EmptyBody(t *testing.T) {
	identityBackend, _ := memory.CreateIdentityBackend(nil, nil)

	rec := invokeLoginHandler(t, "",
		AuthenticationHandler(identityBackend, nil, "", ""))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	checkResponseBody(t, `{
  "message": "missing Username or Password"
}`, rec.Body.Bytes())
}

func TestAuthenticationHandler_AccountSuspended(t *testing.T) {
	identityBackend, _ := memory.CreateIdentityBackend(nil, nil)

	err := identityBackend.CreateAccount(&account.Account{
		ID:                "did:piprate:F73f5Yr5iAxgfD8ZHrvKvrrpKbAR2M99TxeDcy1HXQAY",
		Email:             "test@example.com",
		State:             account.StateSuspended,
		EncryptedPassword: "$2a$10$IsBlCjQAZohKKu5JXpQMQeITWNvf.jsX9bMtqS1PZ24W3ho6LpAly",
		Name:              "John Doe",
	})
	require.NoError(t, err)

	rec := invokeLoginHandler(t,
		string(LoginForm{
			Username: "test@example.com",
			Password: "testpwd",
		}.Bytes()),
		AuthenticationHandler(identityBackend, nil, "", ""))

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	checkResponseBody(t, `{
  "message": "incorrect Username or Password"
}`, rec.Body.Bytes())
}

func TestAuthenticationHandler_UnknownAccount(t *testing.T) {
	identityBackend, _ := memory.CreateIdentityBackend(nil, nil)

	rec := invokeLoginHandler(t,
		string(LoginForm{
			Username: "other@example.com",
			Password: "testpwd",
		}.Bytes()),
		AuthenticationHandler(identityBackend, nil, "", ""))

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	checkResponseBody(t, `{
  "message": "incorrect Username or Password"
}`, rec.Body.Bytes())
}

func TestAuthenticationHandler_WrongPassword(t *testing.T) {
	identityBackend, _ := memory.CreateIdentityBackend(nil, nil)

	err := identityBackend.CreateAccount(&account.Account{
		ID:                "did:piprate:F73f5Yr5iAxgfD8ZHrvKvrrpKbAR2M99TxeDcy1HXQAY",
		Email:             "test@example.com",
		State:             account.StateActive,
		EncryptedPassword: "$2a$10$IsBlCjQAZohKKu5JXpQMQeITWNvf.jsX9bMtqS1PZ24W3ho6LpAly",
		Name:              "John Doe",
	})
	require.NoError(t, err)

	rec := invokeLoginHandler(t,
		string(LoginForm{
			Username: "test@example.com",
			Password: "wrong",
		}.Bytes()),
		AuthenticationHandler(identityBackend, nil, "", ""))

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	checkResponseBody(t, `{
  "message": "incorrect Username or Password"
}`, rec.Body.Bytes())
}

var testAccountV3 = &account.Account{
	ID:                "did:piprate:F73f5Yr5iAxgfD8ZHrvKvrrpKbAR2M99TxeDcy1HXQAY",
	Type:              account.Type,
	Version:           3,
	Email:             "test@example.com",
	State:             account.StateActive,
	EncryptedPassword: "$2a$10$SPGUMvT.dtyy53CEdVj1YOYuB510MeRB3j1R7ug/MPIz1Kug4/8Gu",
	Name:              "John Doe",
	AccessLevel:       model.AccessLevelManaged,
	RecoveryPublicKey: "ogffAbY4mXtgUsLZzyxdk12NQCYyUc7n8rc3GJaCmaY=",
	ManagedSecretStore: &account.SecretStore{
		AccessLevel:         model.AccessLevelManaged,
		MasterKeyParams:     "45yrzQ6zIS4nlIG1DF1dTz+aOnCP7t/mVvgy20hblpAht8r9vW3gS1OuLx7Pv454fhHAhJXowqrZGZs2q6PDiAAIAAAAAAAACAAAAAAAAAABAAAAAAAAAA==",
		EncryptedPayloadKey: "9VXund3DEfERb++DRRSJAUfwpvh5YzIWH1jfzNvBEgmhy1VOkj+J5pFuijJ76WP9URPbvVYZnNFFDN32jBFYSeLlYuZcr1iG",
		EncryptedPayload:    "PrHoNxVy25hkUyi97ToLj0FHx1YmseCweOLRKhA28svCBUg4/kFEgCe4uAHGezwx9QYXMsHXkuZicii6CeivpCT6WFJRpwGuc5MnFwK2VFvESlKA42ZvxyRO001eDMDkE4gjNruHgux0NGMSyWdYcimxIV/7MHAYKoDNdYU77qLmE5MOPZbtGxKWHeJ/eCmVVm0bpX8LZeBzrNShhT/IpBlhLgaVL+X4V4M8/ahgm9w2JKDyev8BLI3eocQCEIg6f22zz7nLC3QQMF3/8CFvzizZxw==",
	},
}

func TestAuthenticationHandler_Success_NoAudience_NoDefaultAudience(t *testing.T) {
	identityBackend, _ := memory.CreateIdentityBackend(nil, nil)

	err := identityBackend.CreateAccount(testAccountV3)
	require.NoError(t, err)

	rec := invokeLoginHandler(t,
		string(LoginForm{
			Username: "test@example.com",
			Password: "gm7FhMFxFD01wGd8dE1RqUAxx7noD8LvPQyBzK+27LA=",
		}.Bytes()),
		AuthenticationHandler(identityBackend, nil, "", ""),
	)

	bodyBytes := rec.Body.Bytes()

	require.Equal(t, http.StatusOK, rec.Code)

	var body authBody
	err = jsonw.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	token, _ := jwt.Parse(body.Token, func(token *jwt.Token) (any, error) {
		return token, nil
	})
	require.NotNil(t, token)

	claims := MapClaims(token.Claims.(jwt.MapClaims))

	assert.Equal(t, "piprate.com", claims[ClaimIssuer])
	assert.Empty(t, claims[ClaimAudienceKeyID])
	assert.Empty(t, claims[ClaimAudience])

	secretVal := claims[ClaimEncryptedAudienceSecret]
	assert.Empty(t, secretVal)
}

func TestAuthenticationHandler_Success_NoAudience_DefaultAudience(t *testing.T) {
	identityBackend, _ := memory.CreateIdentityBackend(nil, nil)

	err := identityBackend.CreateAccount(testAccountV3)
	require.NoError(t, err)

	defaultAudiencePublicKeyStr := "MCXWA73n/1g9gTfG6wR3mk1XLODzu9giissqL+Wg7Zs="
	defaultAudiencePrivateKey, _ := base64.StdEncoding.DecodeString("066+lRtJyQ9WAE8mnfIeWvHVkJtb0i/hO3Zk0SjOHyowJdYDvef/WD2BN8brBHeaTVcs4PO72CKKyyov5aDtmw==")

	rec := invokeLoginHandler(t,
		string(LoginForm{
			Username: "test@example.com",
			Password: "gm7FhMFxFD01wGd8dE1RqUAxx7noD8LvPQyBzK+27LA=",
		}.Bytes()),
		AuthenticationHandler(identityBackend, nil, "app0", defaultAudiencePublicKeyStr),
	)

	bodyBytes := rec.Body.Bytes()

	require.Equal(t, http.StatusOK, rec.Code)

	var body authBody
	err = jsonw.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	token, _ := jwt.Parse(body.Token, func(token *jwt.Token) (any, error) {
		return token, nil
	})
	require.NotNil(t, token)

	claims := MapClaims(token.Claims.(jwt.MapClaims))

	assert.Equal(t, "piprate.com", claims[ClaimIssuer])
	assert.Equal(t, "6h63tZ", claims[ClaimAudienceKeyID])
	assert.Equal(t, "app0", claims[ClaimAudience])

	secret, err := ExtractSecret(claims, defaultAudiencePrivateKey)
	require.NoError(t, err)

	assert.Equal(t, "MdySTIAiwWHrBm6jygYzcTjpHoFGoVSZcHmeOvMXsmM=", secret.Base64())
}

func TestAuthenticationHandler_Success_WithAudience_NotDefault(t *testing.T) {
	identityBackend, _ := memory.CreateIdentityBackend(nil, nil)

	err := identityBackend.CreateAccount(testAccountV3)
	require.NoError(t, err)

	defaultAudiencePublicKeyStr := "MCXWA73n/1g9gTfG6wR3mk1XLODzu9giissqL+Wg7Zs="

	publicKeyStr := "2CdGWNI7wuQZTRIy//cOKzRLr2IQZ+tqqSsvzvDJTo0="
	privateAudienceKey, _ := base64.StdEncoding.DecodeString("MDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwU3Rld2FyZDHYJ0ZY0jvC5BlNEjL/9w4rNEuvYhBn62qpKy/O8MlOjQ==")

	rec := invokeLoginHandler(t,
		string(LoginForm{
			Username:          "test@example.com",
			Password:          "gm7FhMFxFD01wGd8dE1RqUAxx7noD8LvPQyBzK+27LA=",
			Audience:          "app1",
			AudiencePublicKey: publicKeyStr,
		}.Bytes()),
		AuthenticationHandler(identityBackend, []string{"app0", "app1"}, "app0", defaultAudiencePublicKeyStr),
	)

	require.Equal(t, http.StatusOK, rec.Code)

	bodyBytes := rec.Body.Bytes()

	var body authBody
	err = jsonw.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	token, _ := jwt.Parse(body.Token, func(token *jwt.Token) (any, error) {
		return token, nil
	})
	require.NotNil(t, token)

	claims := MapClaims(token.Claims.(jwt.MapClaims))

	assert.Equal(t, "piprate.com", claims[ClaimIssuer])
	assert.Equal(t, "4PizJV", claims[ClaimAudienceKeyID])
	assert.Equal(t, "app1", claims[ClaimAudience])

	secret, err := ExtractSecret(claims, privateAudienceKey)
	require.NoError(t, err)

	assert.Equal(t, "MdySTIAiwWHrBm6jygYzcTjpHoFGoVSZcHmeOvMXsmM=", secret.Base64())
}

func TestAuthenticationHandler_Success_WithAudience_Default(t *testing.T) {
	identityBackend, _ := memory.CreateIdentityBackend(nil, nil)

	err := identityBackend.CreateAccount(testAccountV3)
	require.NoError(t, err)

	defaultAudiencePublicKeyStr := "MCXWA73n/1g9gTfG6wR3mk1XLODzu9giissqL+Wg7Zs="
	defaultAudiencePrivateKey, _ := base64.StdEncoding.DecodeString("066+lRtJyQ9WAE8mnfIeWvHVkJtb0i/hO3Zk0SjOHyowJdYDvef/WD2BN8brBHeaTVcs4PO72CKKyyov5aDtmw==")

	rec := invokeLoginHandler(t,
		string(LoginForm{
			Username: "test@example.com",
			Password: "gm7FhMFxFD01wGd8dE1RqUAxx7noD8LvPQyBzK+27LA=",
			Audience: "app0",
		}.Bytes()),
		AuthenticationHandler(identityBackend, []string{"app0", "app1"}, "app0", defaultAudiencePublicKeyStr),
	)

	require.Equal(t, http.StatusOK, rec.Code)

	bodyBytes := rec.Body.Bytes()

	var body authBody
	err = jsonw.Unmarshal(bodyBytes, &body)
	require.NoError(t, err)

	token, _ := jwt.Parse(body.Token, func(token *jwt.Token) (any, error) {
		return token, nil
	})
	require.NotNil(t, token)

	claims := MapClaims(token.Claims.(jwt.MapClaims))

	assert.Equal(t, "piprate.com", claims[ClaimIssuer])
	assert.Equal(t, "6h63tZ", claims[ClaimAudienceKeyID])
	assert.Equal(t, "app0", claims[ClaimAudience])

	secret, err := ExtractSecret(claims, defaultAudiencePrivateKey)
	require.NoError(t, err)

	assert.Equal(t, "MdySTIAiwWHrBm6jygYzcTjpHoFGoVSZcHmeOvMXsmM=", secret.Base64())
}

func checkResponseBody(t *testing.T, expectedBody string, actualBody []byte) {
	t.Helper()

	var expected any
	var actual any
	err := jsonw.Unmarshal([]byte(expectedBody), &expected)
	if err != nil {
		// Error when unmarshalling expected data. It will be treated as string.
		expected = expectedBody
	}
	err = jsonw.Unmarshal(actualBody, &actual)
	if err != nil {
		// Error when unmarshalling actual data. It will be treated as string.
		actual = string(actualBody)
	}

	if !assert.True(t, ld.DeepCompare(expected, actual, false), "Wrong body") {
		_, _ = os.Stdout.WriteString("==== ACTUAL ====\n")
		b, _ := jsonw.MarshalIndent(actual, "", "  ")
		_, _ = os.Stdout.Write(b)
		_, _ = os.Stdout.WriteString("\n")
		_, _ = os.Stdout.WriteString("==== EXPECTED ====\n")
		b, _ = jsonw.MarshalIndent(expected, "", "  ")
		_, _ = os.Stdout.Write(b)
		t.FailNow()
	}
}
