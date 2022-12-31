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

package vaults

import (
	"io"

	"github.com/piprate/metalocker/model"
)

type (
	Params map[string]any

	// Config defines vault's configuration.
	Config struct {
		// ID is the vault's globally unique ID.
		ID string `json:"id"`
		// ID is the vault's locally (within the MetaLocker node) unique name.
		Name string `json:"name"`
		// Type is the registered vault algorithm type.
		Type string `json:"type"`
		// SSE is true if the vault uses Server Side Encryption
		SSE bool `json:"sse"`
		// CAS is true if the vault generates content addressable resource IDs
		CAS bool `json:"cas"`
		// Params are vault parameters that are specific to its Type.
		Params Params `json:"params"`
	}

	// Vault is a data storage facility for all user's datasets that are stored in MetaLocker.
	Vault interface {
		io.Closer

		// ID returns the vault's globally unique ID.
		ID() string
		// Name returns the vault's locally (within the node) unique name.
		Name() string
		// SSE returns true if server side encryption is enabled in this vault
		// If true, there may be no need to encrypt the blob before storing it in the vault.
		SSE() bool
		// CAS returns true if the vault produces content-addressable blob IDs. This means that
		// if the same blob is uploaded twice, it will receive a storage configuration with
		// the same ID and same parameters. This may not be desirable for private storage
		// as records can be correlated by its data asset IDs. However, it is essential
		// for off-chain operation storage to use content-addressable IDs.
		CAS() bool
		// CreateBlob stores a blob in the vault and returns a resource definition.
		CreateBlob(blob io.Reader) (*model.StoredResource, error)
		// ServeBlob returns a binary stream for the stored resource. Depending on the vault's
		// SSE property, it may be in cleartext or encrypted. The vault will check if
		// the caller can access the resource by checking the provided accessToken
		// against the ledger and other sources.
		ServeBlob(id string, params map[string]any, accessToken string) (io.ReadCloser, error)
		// PurgeBlob permanently purges the given resource from the vault. If will only
		// succeed in the resource is related to a revoked lease.
		PurgeBlob(id string, params map[string]any) error
	}
)
