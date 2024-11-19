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

package model

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
)

const (
	TypeResource = "Resource"
)

var (
	ErrDataAssetAccessDenied = errors.New("access to data asset denied")
	ErrBlobNotFound          = errors.New("blob not found")
)

type (
	// StoredResource contains details about location and the way to access a specific data asset.
	StoredResource struct {
		// ID is the resource ID.
		ID string `json:"id,omitempty"`
		// Type is always equal to TypeResource.
		Type string `json:"type"`
		// Asset is the data asset's content-addressable ID.
		Asset string `json:"asset,omitempty"`
		// Vault is the ID of the vault where the data asset is stored.
		Vault string `json:"vault"`
		// Method is the vault's method of storage. This field defines the meaning of Params field.
		Method string `json:"method"`
		// Params is key/value pairs that are specific to the selected Method. These parameters should be
		// sufficient to locate the resource blob in the vault.
		Params map[string]any `json:"params,omitempty"`
		// EncryptionKey is a Base64-encoded client side encryption key (if the asset was encrypted on the client side).
		EncryptionKey string `json:"encryptionKey,omitempty"`
		// MIMEType is the data asset's MIME type (if known).
		MIMEType string `json:"mimeType,omitempty"`
		// Size is the resource's size in bytes.
		Size int64 `json:"size,omitempty"`
	}

	// VaultProperties defines basic properties of a MetaLocker vault.
	VaultProperties struct {
		// ID is the vault's ID.
		ID string
		// Name is the vault's name. Vault names should be unique within each instance of MetaLocker.
		Name string
		// Type is the vault's type. It defines the underlying technology.
		Type string
		// SSE is true if the vault provides Server Side Encryption. If it does, data sent to the vault
		// should not be encrypted on the client side.
		SSE bool
		// CAS is true if the vault generates content addressable IDs
		CAS bool
	}

	// BlobManager is a trusted component that reads and writes binary data to MetaLocker vaults.
	// BlobManager manages client-side encryption for blobs. There is no need to encrypt of decrypt data
	// that comes from BlobManager.
	BlobManager interface {
		GetBlob(ctx context.Context, res *StoredResource, accessToken string) (io.ReadCloser, error)
		SendBlob(ctx context.Context, data io.Reader, cleartext bool, vaultID string) (*StoredResource, error)
		PurgeBlob(ctx context.Context, res *StoredResource) error

		GetVaultMap() (map[string]*VaultProperties, error)
	}
)

func (sc *StoredResource) StorageID() string {
	switch {
	case sc.ID != "":
		return sc.ID
	case sc.Params != nil:
		val, _ := sc.Params["id"].(string)
		return val
	default:
		return ""
	}
}

func (sc *StoredResource) GetEncryptionKey() *AESKey {
	if sc.EncryptionKey != "" {
		key := &AESKey{}
		keyBytes, _ := base64.StdEncoding.DecodeString(sc.EncryptionKey)
		copy(key[:], keyBytes)
		return key
	} else {
		return nil
	}
}
