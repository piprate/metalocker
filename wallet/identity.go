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

package wallet

import (
	"time"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
)

type (
	identityOptions struct {
		did          *model.DID
		identityType string
	}

	// IdentityOption is for defining parameters when creating new identities
	IdentityOption func(opts *identityOptions) error

	identityWrapper struct {
		wallet *LocalDataWallet
		raw    *account.Identity
	}
)

var _ Identity = (*identityWrapper)(nil)

func WithDID(did *model.DID) IdentityOption {
	return func(opts *identityOptions) error {
		opts.did = did
		return nil
	}
}

func WithType(identityType string) IdentityOption {
	return func(opts *identityOptions) error {
		opts.identityType = identityType
		return nil
	}
}

func newIdentityWrapper(dw *LocalDataWallet, idy *account.Identity) *identityWrapper {
	return &identityWrapper{
		wallet: dw,
		raw:    idy,
	}
}

func (iw *identityWrapper) ID() string {
	return iw.raw.ID()
}

func (iw *identityWrapper) DID() *model.DID {
	return iw.raw.DID
}

func (iw *identityWrapper) CreatedAt() *time.Time {
	return iw.raw.Created
}

func (iw *identityWrapper) Name() string {
	return iw.raw.Name
}

func (iw *identityWrapper) SetName(name string) error {
	panic("operation not implemented")
}

func (iw *identityWrapper) AccessLevel() model.AccessLevel {
	return iw.raw.AccessLevel
}

func (iw *identityWrapper) Raw() *account.Identity {
	return iw.raw
}

func (iw *identityWrapper) NewLocker(name string, options ...LockerOption) (Locker, error) {
	var opts lockerOptions

	for _, fn := range options {
		if err := fn(&opts); err != nil {
			return nil, err
		}
	}

	opts.parties = append(opts.parties, model.Us(iw.DID(), opts.ourSeed))

	tb, err := iw.wallet.nodeClient.Ledger().GetTopBlock()
	if err != nil {
		return nil, err
	}

	locker, err := model.GenerateLocker(
		iw.AccessLevel(),
		name,
		opts.expiresAt,
		tb.Number,
		opts.parties...)
	if err != nil {
		return nil, err
	}

	return iw.wallet.AddLocker(locker)
}

func (dw *LocalDataWallet) NewIdentity(accessLevel model.AccessLevel, name string, options ...IdentityOption) (Identity, error) {
	var opts identityOptions
	for _, fn := range options {
		if err := fn(&opts); err != nil {
			return nil, err
		}
	}

	if opts.did == nil {
		did, err := model.GenerateDID()
		if err != nil {
			panic(err)
		}
		opts.did = did
	}

	if opts.identityType == "" {
		opts.identityType = account.IdentityTypePersona
	}

	now := time.Now().UTC()
	idy := &account.Identity{
		DID:         opts.did,
		Created:     &now,
		Name:        name,
		Type:        opts.identityType,
		AccessLevel: accessLevel,
	}

	if err := dw.AddIdentity(idy); err != nil {
		return nil, err
	}

	return newIdentityWrapper(dw, idy), nil
}
