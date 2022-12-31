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

package prv21

import (
	"context"
	"time"

	"github.com/piprate/metalocker/model"
)

const (
	NoBlockNumber int64 = -1
)

type (
	// Rater is an interface to PRV21 rating algorithm. It can be used by a custom index
	// to algorithmically identify the "current" revision of the given variant, as well
	// as its current revision at a given block height.
	Rater interface {
		AddRevision(ctx context.Context, ds model.DataSet, effectiveBlockNumber int64) (bool, error)
		AddRevocation(ctx context.Context, rid string) error
		Head(ctx context.Context) string
		HeadAt(ctx context.Context, blockID int64) string
	}

	Revision interface {
		RecordID() string
		Status() model.RecordStatus
		Block() int64
		EffectiveBlock() int64
		Locker() string
		Participant() string
		AssetID() string
		VariantID() string
		ImpressionID() string
		RevisionNumber() int64
		CreatedAt() time.Time
		HeadFrom() int64
		HeadTo() int64
	}

	RevisionStore interface {
		Head(ctx context.Context, variantID string) (Revision, error)
		HeadAt(ctx context.Context, variantID string, blockNumber int64) (Revision, error)
		Revision(ctx context.Context, rid string) (Revision, error)
		CreateRevision(ctx context.Context, ds model.DataSet, effectiveBlockNumber int64,
			headFrom, headTo int64) error
		UpdateRevision(ctx context.Context, rid string, status model.RecordStatus,
			headFrom, headTo int64) error
		RevokeRevision(ctx context.Context, rid string) error
		SaveRevokedRevision(ctx context.Context, ds model.DataSet, effectiveBlockNumber int64) error
		DataSet(ctx context.Context, rid string) (model.DataSet, error)
	}
)
