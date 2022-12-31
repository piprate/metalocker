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
	"crypto/ed25519"
	"errors"
	"fmt"
	"net/http"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/sdk/httpsecure"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog/log"
)

func (c *MetaLockerHTTPCaller) AdminGetAccountList() ([]account.Account, error) {
	if !c.client.IsAuthenticated() {
		return nil, errors.New("you need to log in before performing any operations")
	}

	var acctList []account.Account
	err := c.client.LoadContents(http.MethodGet, "/v1/admin/account", nil, &acctList)
	if err != nil {
		return nil, err
	}

	return acctList, nil
}

func (c *MetaLockerHTTPCaller) AdminStoreAccount(acc *account.Account) error {
	if !c.client.IsAuthenticated() {
		return errors.New("you need to log in before performing any operations")
	}

	res, err := c.client.SendRequest(http.MethodPost, "/v1/admin/account", httpsecure.WithJSONBody(acc))
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()

	switch res.StatusCode {
	case http.StatusOK:
	// do nothing
	case http.StatusUnauthorized:
		return ErrNotAuthorised
	default:
		return fmt.Errorf("account upload failed with status code %d", res.StatusCode)
	}

	return nil
}

// copied from admin package
type AccountAdminPatch struct {
	State string `json:"state,omitempty"`
}

func (c *MetaLockerHTTPCaller) AdminPatchAccount(id string, patch AccountAdminPatch) error {
	if !c.client.IsAuthenticated() {
		return errors.New("you need to log in before performing any operations")
	}

	url := fmt.Sprintf("/v1/admin/account/%s", id)
	res, err := c.client.SendRequest(http.MethodPatch, url, httpsecure.WithJSONBody(patch))
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()

	switch res.StatusCode {
	case http.StatusOK:
	// do nothing
	case http.StatusUnauthorized:
		return ErrNotAuthorised
	default:
		msg := apibase.ParseResponseMessage(res)
		return fmt.Errorf("response status code: %d, message: %s", res.StatusCode, msg)
	}

	return nil
}

func (c *MetaLockerHTTPCaller) GetOwnAccount() (*account.Account, error) {
	if !c.client.IsAuthenticated() {
		return nil, errors.New("you need to log in before performing any operations")
	}

	var acct account.Account
	err := c.client.LoadContents(http.MethodGet, "/v1/account", nil, &acct)
	if err != nil {
		return nil, err
	}

	return &acct, nil
}

func (c *MetaLockerHTTPCaller) GetAccount(id string) (*account.Account, error) {
	if !c.client.IsAuthenticated() {
		return nil, errors.New("you need to log in before performing any operations")
	}

	var acct account.Account
	err := c.client.LoadContents(http.MethodGet, "/v1/account/"+id, nil, &acct)
	if err != nil {
		return nil, err
	}

	return &acct, nil
}

func (c *MetaLockerHTTPCaller) UpdateAccount(acc *account.Account) error {
	if !c.client.IsAuthenticated() {
		return errors.New("you need to log in before performing any operations")
	}

	res, err := c.client.SendRequest(http.MethodPut, "/v1/account", httpsecure.WithJSONBody(acc))
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()

	switch res.StatusCode {
	case http.StatusOK:
	// do nothing
	case http.StatusUnauthorized:
		return ErrNotAuthorised
	default:
		return fmt.Errorf("account update failed with status code %d", res.StatusCode)
	}

	return nil
}

func (c *MetaLockerHTTPCaller) PatchAccount(email, oldEncryptedPassword, newEncryptedPassword, name, givenName, familyName string) error {
	if !c.client.IsAuthenticated() {
		return errors.New("you need to log in before performing any operations")
	}

	patch := &AccountPatch{
		Email:                email,
		OldEncryptedPassword: oldEncryptedPassword,
		NewEncryptedPassword: newEncryptedPassword,
		Name:                 name,
		GivenName:            givenName,
		FamilyName:           familyName,
	}

	res, err := c.client.SendRequest(http.MethodPatch, "/v1/account", httpsecure.WithJSONBody(patch))
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()

	switch res.StatusCode {
	case http.StatusOK:
	// do nothing
	case http.StatusUnauthorized:
		return ErrNotAuthorised
	default:
		return fmt.Errorf("account update failed with status code %d", res.StatusCode)
	}

	return nil
}

func (c *MetaLockerHTTPCaller) DeleteAccount(id string) error {
	if !c.client.IsAuthenticated() {
		return errors.New("you need to log in before performing any operations")
	}

	url := fmt.Sprintf("/v1/account/%s", id)

	res, err := c.client.SendRequest(http.MethodDelete, url)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()

	switch res.StatusCode {
	case http.StatusNoContent:
	// do nothing
	case http.StatusUnauthorized:
		return ErrNotAuthorised
	default:
		return fmt.Errorf("account deletion failed with status code %d", res.StatusCode)
	}

	return nil
}

func (c *MetaLockerHTTPCaller) GetAccountRecoveryCode(username string) (string, error) {
	url := fmt.Sprintf("/v1/recovery-code?email=%s", username)
	res, err := c.client.SendRequest(http.MethodGet, url, httpsecure.SkipAuthentication())
	if err != nil {
		return "", err
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		msg := apibase.ParseResponseMessage(res)
		log.Error().Str("url", url).Str("msg", msg).Msg("Call failed")
		return "", fmt.Errorf("response status code: %d, message: %s", res.StatusCode, msg)
	}

	var rsp GetRecoveryCodeResponse
	if err := jsonw.Decode(res.Body, &rsp); err != nil {
		return "", err
	}

	log.Debug().Str("code", rsp.Code).Msg("Recovery code response")

	return rsp.Code, nil
}

func (c *MetaLockerHTTPCaller) RecoverAccount(userID string, privKey ed25519.PrivateKey, recoveryCode, newPassphrase string) (*account.Account, error) {

	req := account.BuildRecoveryRequest(userID, recoveryCode, privKey, newPassphrase)

	res, err := c.client.SendRequest(http.MethodPost, "/v1/recover-account",
		httpsecure.WithJSONBody(req),
		httpsecure.SkipAuthentication())
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	switch res.StatusCode {
	case http.StatusOK:
		var rsp AccountRecoveryResponse
		if err := jsonw.Decode(res.Body, &rsp); err != nil {
			return nil, err
		}

		return rsp.Account, nil
	case http.StatusUnauthorized:
		msg := apibase.ParseResponseMessage(res)
		return nil, errors.New(msg)
	default:
		return nil, fmt.Errorf("account recovery failed with status code %d", res.StatusCode)
	}
}

func (c *MetaLockerHTTPCaller) CreateSubAccount(acct *account.Account) (*account.Account, error) {
	if !c.client.IsAuthenticated() {
		return nil, errors.New("you need to log in before performing any operations")
	}

	res, err := c.client.SendRequest(http.MethodPost, "/v1/account", httpsecure.WithJSONBody(acct))
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	switch res.StatusCode {
	case http.StatusCreated:
		var rsp account.Account
		if err := jsonw.Decode(res.Body, &rsp); err != nil {
			return nil, err
		}

		return &rsp, nil
	case http.StatusUnauthorized:
		return nil, ErrNotAuthorised
	default:
		msg := apibase.ParseResponseMessage(res)
		return nil, fmt.Errorf("sub-account submission failed with status code %d: %s", res.StatusCode, msg)
	}
}

func (c *MetaLockerHTTPCaller) ListSubAccounts(id string) ([]*account.Account, error) {
	if !c.client.IsAuthenticated() {
		return nil, errors.New("you need to log in before performing any operations")
	}

	var acctList []*account.Account
	err := c.client.LoadContents(http.MethodGet, "/v1/account/"+id+"/children", nil, &acctList)
	if err != nil {
		return nil, err
	}

	return acctList, nil
}

func (c *MetaLockerHTTPCaller) CreateAccessKey(key *model.AccessKey) (*model.AccessKey, error) {
	if !c.client.IsAuthenticated() {
		return nil, errors.New("you need to log in before performing any operations")
	}

	url := "/v1/account/" + c.currentAccountID + "/access-key"
	res, err := c.client.SendRequest(http.MethodPost, url, httpsecure.WithJSONBody(key))
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	switch res.StatusCode {
	case http.StatusCreated:
		var rsp model.AccessKey
		if err := jsonw.Decode(res.Body, &rsp); err != nil {
			return nil, err
		}

		key.ID = rsp.ID

		return key, nil
	case http.StatusUnauthorized:
		return nil, ErrNotAuthorised
	default:
		return nil, fmt.Errorf("access key creation failed with status code %d", res.StatusCode)
	}
}

func (c *MetaLockerHTTPCaller) DeleteAccessKey(keyID string) error {
	if !c.client.IsAuthenticated() {
		return errors.New("you need to log in before performing any operations")
	}

	url := "/v1/account/" + c.currentAccountID + "/access-key/" + keyID
	res, err := c.client.SendRequest(http.MethodDelete, url)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()

	switch res.StatusCode {
	case http.StatusNoContent:
	// do nothing
	case http.StatusUnauthorized:
		return ErrNotAuthorised
	default:
		return fmt.Errorf("access key deletion failed with status code %d", res.StatusCode)
	}

	return nil
}

func (c *MetaLockerHTTPCaller) GetAccessKey(keyID string) (*model.AccessKey, error) {
	if !c.client.IsAuthenticated() {
		return nil, errors.New("you need to log in before performing any operations")
	}

	var key model.AccessKey
	url := "/v1/account/" + c.currentAccountID + "/access-key/" + keyID
	err := c.client.LoadContents(http.MethodGet, url, nil, &key)
	if err != nil {
		return nil, err
	}

	return &key, nil
}

func (c *MetaLockerHTTPCaller) ListAccessKeys() ([]*model.AccessKey, error) {
	if !c.client.IsAuthenticated() {
		return nil, errors.New("you need to log in before performing any operations")
	}

	var keyList []*model.AccessKey
	url := "/v1/account/" + c.currentAccountID + "/access-key"
	err := c.client.LoadContents(http.MethodGet, url, nil, &keyList)
	if err != nil {
		return nil, err
	}

	return keyList, nil
}
