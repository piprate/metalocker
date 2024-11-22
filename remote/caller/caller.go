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
	"io"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/golang-jwt/jwt/v5"
	"github.com/piprate/json-gold/ld"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/sdk/httpsecure"
	"github.com/piprate/metalocker/services/notification"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/wallet"
	"github.com/rs/zerolog/log"
)

var (
	ErrNotAuthorised = errors.New("not authorised to perform operation")
	ErrLoginFailed   = errors.New("login failed")
)

type MetaLockerHTTPCaller struct {
	client *httpsecure.Client

	authenticatedAccountID string
	currentAccountID       string

	cachedVaultMap map[string]*model.VaultProperties

	ns *notification.RemoteNotificationService
}

// check all interfaces the Caller should provide

var _ model.Ledger = (*MetaLockerHTTPCaller)(nil)
var _ model.OffChainStorage = (*MetaLockerHTTPCaller)(nil)
var _ model.BlobManager = (*MetaLockerHTTPCaller)(nil)
var _ wallet.NodeClient = (*MetaLockerHTTPCaller)(nil)

func NewMetaLockerHTTPCaller(url string, userAgent string) (*MetaLockerHTTPCaller, error) {
	httpClient, err := httpsecure.NewHTTPClient(url, userAgent, 0, nil)
	if err != nil {
		return nil, err
	}

	caller := &MetaLockerHTTPCaller{
		client: httpClient,
	}

	return caller, nil
}

func (c *MetaLockerHTTPCaller) Close() error {
	if c.ns != nil {
		if err := c.ns.Close(); err != nil {
			return err
		}
	}

	c.client.CloseIdleConnections()

	return nil
}

func (c *MetaLockerHTTPCaller) ConnectionURL() string {
	return c.client.GetConnectionURL()
}

func (c *MetaLockerHTTPCaller) AuthenticatedAccountID() string {
	return c.authenticatedAccountID
}

func (c *MetaLockerHTTPCaller) GetToken() string {
	return c.client.GetToken()
}

func (c *MetaLockerHTTPCaller) DIDProvider() model.DIDProvider {
	return c
}

func (c *MetaLockerHTTPCaller) OffChainStorage() model.OffChainStorage {
	return c
}

func (c *MetaLockerHTTPCaller) Ledger() model.Ledger {
	return c
}

func (c *MetaLockerHTTPCaller) BlobManager() model.BlobManager {
	return c
}

func (c *MetaLockerHTTPCaller) SecureClient() *httpsecure.Client {
	return c.client
}

func (c *MetaLockerHTTPCaller) NewInstance(ctx context.Context, email, passphrase string, isHash bool) (wallet.NodeClient, error) {
	if isHash {
		return nil, errors.New("can't acquire new caller using passphrase hash")
	}

	// make a shallow copy of the client
	newClient := *c.client

	newCaller := &MetaLockerHTTPCaller{
		client: &newClient,
	}

	err := newCaller.LoginWithCredentials(ctx, email, passphrase)
	if err != nil {
		log.Err(err).Str("user", email).Msg("Authentication failed")
		return nil, err
	}

	return newCaller, nil
}

func (c *MetaLockerHTTPCaller) SubAccountInstance(subAccountID string) (wallet.NodeClient, error) {
	newCaller := *c
	newCaller.currentAccountID = subAccountID

	return &newCaller, nil
}

type LoginResponse struct {
	Code   int       `json:"code"`
	Expire time.Time `json:"expire"`
	Token  string    `json:"token"`
}

func (c *MetaLockerHTTPCaller) LoginWithJWT(jwtToken string) error {
	// logout before trying to log in
	c.Logout()

	if err := httpsecure.WithJWTToken(jwtToken)(c.client); err != nil {
		return err
	}

	id, err := extractAccountID(jwtToken)
	if err != nil {
		return err
	}
	c.authenticatedAccountID = id
	c.currentAccountID = id

	return nil
}

func (c *MetaLockerHTTPCaller) LoginWithCredentials(ctx context.Context, email string, password string) error {
	// logout before trying to log in
	c.Logout()

	loginForm := LoginForm{
		Username: email,
		Password: account.HashUserPassword(password),
	}

	res, err := c.client.SendRequest(ctx, http.MethodPost, "/v1/authenticate",
		httpsecure.WithJSONBody(loginForm),
		httpsecure.SkipAuthentication())
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized {
			return ErrLoginFailed
		} else {
			return fmt.Errorf("bad response status code: %d", res.StatusCode)
		}
	}

	var resp *LoginResponse
	err = jsonw.Decode(res.Body, &resp)
	if err != nil {
		return errors.New("failed to decode response from authentication service")
	}

	sessionToken := resp.Token
	if sessionToken == "" {
		return fmt.Errorf("empty security token")
	}

	if err = c.LoginWithJWT(sessionToken); err != nil {
		return err
	}

	log.Debug().Str("userID", email).Str("id", c.authenticatedAccountID).Msg("Account login successful")

	return nil
}

func extractAccountID(token string) (string, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		return token, nil
	})
	// don't check the error, because it will be not nil due to failed validation
	if t == nil {
		return "", err
	}

	id, exists := t.Claims.(jwt.MapClaims)["id"]
	if !exists {
		return "", errors.New("jwt token doesn't include account ID")
	}

	return id.(string), nil
}

func (c *MetaLockerHTTPCaller) LoginWithAdminKeys(adminKey, adminSecret string) error {
	if err := httpsecure.WithAdminKeys(adminKey, adminSecret)(c.client); err != nil {
		return err
	}

	c.authenticatedAccountID = "admin-do-not-use"
	c.currentAccountID = "admin-do-not-use"

	return nil
}

func (c *MetaLockerHTTPCaller) LoginWithAccessKeys(ctx context.Context, apiKey, clientSecret string) error {
	if err := httpsecure.WithAccessKey(apiKey, clientSecret)(c.client); err != nil {
		return err
	}

	acct, err := c.GetOwnAccount(ctx)
	if err != nil {
		return err
	}
	c.authenticatedAccountID = acct.ID
	c.currentAccountID = acct.ID

	return nil
}

// Logout - clearing authorisation info
func (c *MetaLockerHTTPCaller) Logout() {
	_ = c.CloseNotificationService()
	c.client.Logout()

	c.authenticatedAccountID = ""
	c.currentAccountID = ""
}

type NewAccountForm struct {
	Account          *account.Account `json:"account"`
	RegistrationCode string           `json:"registrationCode"`
}

func (c *MetaLockerHTTPCaller) CreateAccount(ctx context.Context, acct *account.Account, registrationCode string) error {
	form := NewAccountForm{
		Account:          acct,
		RegistrationCode: registrationCode,
	}

	res, err := c.client.SendRequest(ctx, http.MethodPost, "/v1/register",
		httpsecure.WithJSONBody(form),
		httpsecure.SkipAuthentication())
	if err != nil {
		return err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		rspb, _ := io.ReadAll(res.Body)

		log.Debug().Str("body", string(rspb)).Msg("Registration response")

		d, err := sonic.Get(rspb)
		if err != nil {
			return err
		}

		val, err := d.Get("status").String()
		if err != nil {
			return fmt.Errorf("error reading registration status: %w", err)
		}
		if val != "ok" {
			return fmt.Errorf("bad registration status: %s", val)
		}

		tokenNode := d.Get("token")
		if tokenNode.Exists() {
			sessionToken, _ := tokenNode.String()
			if sessionToken != "" {
				if err = c.LoginWithJWT(sessionToken); err != nil {
					return err
				}

				log.Debug().Str("token", sessionToken).Msg("Authenticated")
			}
		}

	case http.StatusUnauthorized:
		return ErrNotAuthorised
	case http.StatusConflict:
		return fmt.Errorf("account already exists: %s", acct.Email)
	default:
		msg := apibase.ParseResponseMessage(res)
		return fmt.Errorf("account registration failed with status code %d. Message: %s", res.StatusCode, msg)
	}

	return nil
}

func (c *MetaLockerHTTPCaller) InitContextForwarding() {
	mapping := map[string]string{
		"https://piprate.org/context/piprate.jsonld": c.client.GetConnectionURL() + "/static/model/piprate.jsonld",
		"https://piprate.org/context/prov.jsonld":    c.client.GetConnectionURL() + "/static/model/prov.jsonld",
		"https://w3id.org/security/v1":               c.client.GetConnectionURL() + "/static/model/security-v1.jsonld",
		"https://w3id.org/security/v2":               c.client.GetConnectionURL() + "/static/model/security-v2.jsonld",
		"https://w3id.org/did/v1":                    c.client.GetConnectionURL() + "/static/model/did-v1.jsonld",
		"http://schema.org/":                         c.client.GetConnectionURL() + "/static/model/schema.jsonld",
	}
	model.SetDefaultDocumentLoader(
		ld.NewCachingDocumentLoader(
			NewProxyDocumentLoader(
				ld.NewDefaultDocumentLoader(c.client.GetHTTPClient()), mapping),
		),
	)
}

// ProxyDocumentLoader redirects 'well-known' context URLs to the given URLs.
// This is useful for redirecting Piprate context calls to a MetaLocker instance.
type ProxyDocumentLoader struct {
	nextLoader   ld.DocumentLoader
	proxyMapping map[string]string
}

// NewProxyDocumentLoader creates a new instance of ProxyDocumentLoader.
func NewProxyDocumentLoader(nextLoader ld.DocumentLoader, proxyMapping map[string]string) *ProxyDocumentLoader {
	rval := &ProxyDocumentLoader{
		nextLoader:   nextLoader,
		proxyMapping: proxyMapping,
	}

	return rval
}

// LoadDocument returns a RemoteDocument containing the contents of the JSON resource
// from the given URL.
func (pdl *ProxyDocumentLoader) LoadDocument(u string) (*ld.RemoteDocument, error) {
	targetURL := u
	if mappedURL, found := pdl.proxyMapping[u]; found {
		targetURL = mappedURL
	}

	log.Debug().Str("u", u).Str("target", targetURL).Msg("Mapped document URL")

	remoteDoc, err := pdl.nextLoader.LoadDocument(targetURL)
	if err != nil {
		log.Error().Err(err).Msg("Error when loading document")
		return nil, err
	}

	// substitute the context URL as if it came from the given address
	remoteDoc.DocumentURL = u

	return remoteDoc, nil
}
