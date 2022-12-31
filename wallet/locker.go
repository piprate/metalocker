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
	"github.com/piprate/metalocker/model/dataset"
)

type (
	lockerOptions struct {
		expiresAt *time.Time
		ourSeed   []byte
		parties   []model.PartyOption
	}

	// LockerOption is for defining parameters when creating new lockers
	LockerOption func(opts *lockerOptions) error

	lockerWrapper struct {
		wallet        *LocalDataWallet
		raw           *model.Locker
		us            *model.LockerParticipant
		themExtracted bool
		them          []*model.LockerParticipant
	}
)

var _ Locker = (*lockerWrapper)(nil)

func ExpiresAt(expiresAt time.Time) LockerOption {
	return func(opts *lockerOptions) error {
		opts.expiresAt = &expiresAt
		return nil
	}
}

func FixedSeed(seed []byte) LockerOption {
	return func(opts *lockerOptions) error {
		opts.ourSeed = seed
		return nil
	}
}

func Participant(did *model.DID, seed []byte) LockerOption {
	return func(opts *lockerOptions) error {
		opts.parties = append(opts.parties, model.Them(did, seed))
		return nil
	}
}

func newLockerWrapper(dw *LocalDataWallet, locker *model.Locker) *lockerWrapper {
	lw := &lockerWrapper{
		wallet: dw,
		raw:    locker,
		us:     locker.Us(),
	}

	return lw
}

func (lw *lockerWrapper) ID() string {
	return lw.raw.ID
}

func (lw *lockerWrapper) CreatedAt() *time.Time {
	return lw.raw.Created
}

func (lw *lockerWrapper) Name() string {
	return lw.raw.Name
}

func (lw *lockerWrapper) SetName(name string) error {
	panic("operation not implemented")
}

func (lw *lockerWrapper) AccessLevel() model.AccessLevel {
	return lw.raw.AccessLevel
}

func (lw *lockerWrapper) Raw() *model.Locker {
	return lw.raw
}

func (lw *lockerWrapper) IsUniLocker() bool {
	return len(lw.them) == 0
}

func (lw *lockerWrapper) IsThirdParty() bool {
	return lw.us == nil
}

func (lw *lockerWrapper) Us() *model.LockerParticipant {
	return lw.us
}

func (lw *lockerWrapper) Them() []*model.LockerParticipant {
	if !lw.themExtracted {
		// lazy-load 'them' relationship
		lw.them = lw.extractThem()
		lw.themExtracted = true
	}
	return lw.them
}

func (lw *lockerWrapper) extractThem() []*model.LockerParticipant {
	var them []*model.LockerParticipant
	for _, p := range lw.raw.Participants {
		if !p.Self {
			them = append(them, p)
		}
	}
	return them
}

func (lw *lockerWrapper) NewDataSetBuilder(opts ...dataset.BuilderOption) (dataset.Builder, error) {
	return lw.wallet.DataStore().NewDataSetBuilder(lw.ID(), opts...)
}

func (lw *lockerWrapper) Store(meta any, expiryTime time.Time, opts ...dataset.BuilderOption) dataset.RecordFuture {
	b, err := lw.wallet.DataStore().NewDataSetBuilder(lw.ID(), opts...)
	if err != nil {
		return dataset.RecordFutureWithError(err)
	}
	_, err = b.AddMetaResource(meta)
	if err != nil {
		return dataset.RecordFutureWithError(err)
	}
	return b.Submit(expiryTime)
}

func (lw *lockerWrapper) Share(id, vaultName string, expiryTime time.Time) dataset.RecordFuture {
	ds, err := lw.wallet.DataStore().Load(id)
	if err != nil {
		return dataset.RecordFutureWithError(err)
	}

	return lw.wallet.DataStore().Share(ds, lw, vaultName, expiryTime)
}

func (lw *lockerWrapper) HeadID(assetID string, headName string) string {
	return model.HeadID(assetID, lw.raw.ID, lw.raw.Us(), headName)
}

func (lw *lockerWrapper) SetAssetHead(assetID, headName, recordID string) dataset.RecordFuture {
	return lw.wallet.DataStore().SetAssetHead(assetID, lw.raw, headName, recordID)
}

func (lw *lockerWrapper) Seal() error {
	panic("operation not implemented")
}
