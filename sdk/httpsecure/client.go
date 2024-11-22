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

package httpsecure

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/piprate/metalocker/utils/security"
	"github.com/rs/zerolog/log"
)

const (
	AuthMethodJWT       = "jwt"
	AuthMethodAdminKeys = "admin_keys"
	AuthMethodAccessKey = "access_key"
)

var (
	// ErrEntityNotFound indicates the API entity was not found
	ErrEntityNotFound = errors.New("entity not found")
	// ErrOperationTimedOut indicates the request has timed out
	ErrOperationTimedOut = errors.New("operation timed out")
)

type Client struct {
	httpClient    *http.Client
	tlsConfig     *tls.Config
	connectionURL string

	authMethod  string
	jwtToken    string
	adminKey    string
	adminSecret string
	apiKey      string
	apiSecret   *model.AESKey
	apiHMACKey  []byte

	userAgent string
}

type AuthenticationFunc func(pc *Client) error

// WithJWTToken - authentication with JWT token
func WithJWTToken(jwtToken string) AuthenticationFunc {
	return func(pc *Client) error {
		pc.authMethod = AuthMethodJWT
		pc.jwtToken = jwtToken
		return nil
	}
}

// WithAdminKeys - authentication with admin keys
func WithAdminKeys(adminKey, adminSecret string) AuthenticationFunc {
	return func(pc *Client) error {
		pc.authMethod = AuthMethodAdminKeys
		pc.adminKey = adminKey
		pc.adminSecret = adminSecret
		return nil
	}
}

// WithAccessKey - authentication with access keys
func WithAccessKey(apiKey, clientSecret string) AuthenticationFunc {
	return func(pc *Client) error {
		_, secret, hmacKey, err := model.SplitClientSecret(clientSecret)
		if err != nil {
			return err
		}

		pc.authMethod = AuthMethodAccessKey
		pc.apiKey = apiKey
		pc.apiSecret = secret
		pc.apiHMACKey = hmacKey
		return nil
	}
}

type (
	requestOptions struct {
		authMethod         string
		body               io.Reader
		bodyToSign         []byte
		skipAuthentication bool
		headers            map[string]string
	}

	Option func(opts *requestOptions) error
)

func WithBody(body io.Reader) Option {
	return func(opts *requestOptions) error {
		if opts.authMethod == AuthMethodAccessKey {
			// extract body for hashing
			var err error
			opts.bodyToSign, err = io.ReadAll(body)
			if err != nil {
				return err
			}
			opts.body = bytes.NewBuffer(opts.bodyToSign)
		} else {
			opts.body = body
		}

		return nil
	}
}

func WithJSONBody(val any) Option {
	return func(opts *requestOptions) error {
		bodyBytes, err := jsonw.Marshal(val)
		if err != nil {
			return err
		}

		opts.body = bytes.NewBuffer(bodyBytes)
		opts.bodyToSign = bodyBytes

		return nil
	}
}

func WithUnsignedBody(body io.Reader) Option {
	return func(opts *requestOptions) error {
		opts.body = body
		return nil
	}
}

func WithHeaders(headers map[string]string) Option {
	return func(opts *requestOptions) error {
		opts.headers = headers
		return nil
	}
}

func SkipAuthentication() Option {
	return func(opts *requestOptions) error {
		opts.skipAuthentication = true
		return nil
	}
}

// NewHTTPClient returns a client with the requested config
func NewHTTPClient(source, userAgent string, timeout time.Duration, authFunc AuthenticationFunc) (*Client, error) {
	source, httpClient, tlsConfig, err := security.CreateHTTPClient(source)
	if err != nil {
		return nil, err
	}

	client := &Client{
		httpClient:    httpClient,
		tlsConfig:     tlsConfig,
		connectionURL: source,
		userAgent:     userAgent,
	}

	if int(timeout) != 0 {
		client.httpClient.Timeout = timeout
	}

	if authFunc != nil {
		if err := authFunc(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}

func (c *Client) CloseIdleConnections() {
	c.httpClient.CloseIdleConnections()
}

func (c *Client) LoadContents(ctx context.Context, method, contentURL string, headerValues map[string]string, val any) error {
	res, err := c.SendRequest(ctx, method, contentURL, WithHeaders(headerValues))
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()

	switch res.StatusCode {
	case http.StatusOK:
		if dest, ok := val.(*[]byte); ok {
			var err error
			if *dest, err = io.ReadAll(res.Body); err != nil {
				return err
			}
		} else {
			if err := jsonw.Decode(res.Body, val); err != nil {
				return err
			}
		}
		return nil
	case http.StatusNotFound:
		return ErrEntityNotFound
	case http.StatusRequestTimeout:
		return ErrOperationTimedOut
	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorised call: %s %s", method, contentURL)
	default:
		msg := apibase.ParseResponseMessage(res)
		log.Error().Str("url", contentURL).Str("msg", msg).Msg("Call failed")
		return fmt.Errorf("response status code: %d, message: %s", res.StatusCode, msg)
	}
}

func (c *Client) SendRequest(ctx context.Context, method, requestURL string, opts ...Option) (*http.Response, error) {
	// process options
	var options requestOptions
	options.authMethod = c.authMethod
	for _, o := range opts {
		if err := o(&options); err != nil {
			return nil, err
		}
	}

	req, err := c.buildRequest(ctx, method, requestURL, options)
	if err != nil {
		return nil, err
	}

	for k, v := range options.headers {
		req.Header.Set(k, v)
	}

	return c.httpClient.Do(req)
}

func (c *Client) buildRequest(ctx context.Context, method, relativeURL string, opts requestOptions) (*http.Request, error) {

	fullURL := c.connectionURL + relativeURL

	log.Debug().Str("method", method).Str("url", fullURL).Msg("Sending HTTP request")

	req, err := http.NewRequestWithContext(ctx, method, fullURL, opts.body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)

	if !opts.skipAuthentication {
		switch c.authMethod {
		case AuthMethodJWT:
			c.setBearer(req.Header)
		case AuthMethodAdminKeys:
			c.setAdminKey(req.Header)
		case AuthMethodAccessKey:
			_, err = c.signRequest(req.Header, relativeURL, opts.bodyToSign)
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("not authenticated")
		}
	}

	return req, nil
}

func (c *Client) DialWebSocket(relativeURL string) (*websocket.Conn, error) {
	serverURLStruct, err := url.Parse(c.connectionURL)
	if err != nil {
		return nil, err
	}

	u := url.URL{Scheme: "wss", Host: serverURLStruct.Host, Path: relativeURL}
	fullURL := u.String()
	log.Debug().Str("url", fullURL).Msg("Connecting to a web socket")

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		TLSClientConfig:  c.tlsConfig,
	}

	hdr := http.Header{}

	switch c.authMethod {
	case AuthMethodJWT:
		c.setBearer(hdr)
	case AuthMethodAdminKeys:
		c.setAdminKey(hdr)
	case AuthMethodAccessKey:
		_, err = c.signRequest(hdr, relativeURL, nil)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("not authenticated")
	}

	conn, _, err := dialer.Dial(fullURL, hdr) //nolint:bodyclose
	return conn, err
}

func (c *Client) signRequest(hdr http.Header, requestURL string, body []byte) (string, error) {
	return model.SignRequest(hdr, c.apiKey, c.apiSecret, c.apiHMACKey, time.Now(), requestURL, body)
}

// setBearer - set header for JWT authentication
func (c *Client) setBearer(hdr http.Header) {
	hdr.Set("Authorization", "Bearer "+c.jwtToken)
}

// setAdminKey - set header for admin keys authentication
func (c *Client) setAdminKey(hdr http.Header) {
	hdr.Set("X-Auth-Key", c.adminKey)
	hdr.Set("X-Auth-Secret", c.adminSecret)
}

// Logout - clearing authorisation info
func (c *Client) Logout() {
	c.jwtToken = ""
	c.adminKey = ""
	c.adminSecret = ""
	c.apiSecret = nil
	c.apiHMACKey = nil
	c.authMethod = ""
}

// GetConnectionURL - returning client connection url
func (c *Client) GetConnectionURL() string {
	return c.connectionURL
}

// GetUserAgent - returning client user agent
func (c *Client) GetUserAgent() string {
	return c.userAgent
}

// GetHTTPClient - returning http client
func (c *Client) GetHTTPClient() *http.Client {
	return c.httpClient
}

// GetToken - returning JWT token
func (c *Client) GetToken() string {
	return c.jwtToken
}

// IsAuthenticated - check authentificaion
func (c *Client) IsAuthenticated() bool {
	return c.authMethod != ""
}

// GetAuthMethod - returning authorization method
func (c *Client) GetAuthMethod() string {
	return c.authMethod
}

func (c *Client) Copy(authFunc AuthenticationFunc) (*Client, error) {
	cpy := *c
	if authFunc != nil {
		c.Logout()
		if err := authFunc(&cpy); err != nil {
			return nil, err
		}
	}
	return &cpy, nil
}
