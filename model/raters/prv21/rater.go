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
	"errors"

	"github.com/piprate/metalocker/model"
)

type (
	RaterImpl struct {
		variantID string
		headID    string
		store     RevisionStore
	}
)

func NewRater(variantID string, store RevisionStore) Rater {
	r := &RaterImpl{
		variantID: variantID,
		store:     store,
	}

	return r
}

func (r *RaterImpl) AddRevision(ctx context.Context, ds model.DataSet, effectiveBlockNumber int64) (bool, error) {
	imp := ds.Impression()
	varID := imp.GetVariantID()
	revNum := imp.Revision()

	var headFrom int64

	var currentHead Revision
	var err error
	if r.headID != "" {
		currentHead, err = r.store.Revision(ctx, r.headID)
	} else {
		currentHead, err = r.store.Head(ctx, varID)
	}
	if err != nil {
		if errors.Is(err, model.ErrRecordNotFound) {
			// save the first revision. Nothing else to do
			r.headID = ds.ID()
			return true, r.store.CreateRevision(ctx, ds, effectiveBlockNumber, effectiveBlockNumber, NoBlockNumber)
		} else {
			return false, err
		}
	}

	newHead := false
	if revNum > currentHead.RevisionNumber() {
		newHead = true
	} else if revNum == currentHead.RevisionNumber() {
		if imp.GeneratedAtTime.After(currentHead.CreatedAt()) {
			newHead = true
		}
	}

	if newHead {
		err = r.store.UpdateRevision(ctx, currentHead.RecordID(), currentHead.Status(), currentHead.HeadFrom(), effectiveBlockNumber-1)
		if err != nil {
			return false, err
		}

		r.headID = ds.ID()
		headFrom = effectiveBlockNumber
	} else {
		// insert an orphan
		headFrom = NoBlockNumber
	}

	return false, r.store.CreateRevision(ctx, ds, effectiveBlockNumber, headFrom, NoBlockNumber)
}

func (r *RaterImpl) AddRevocation(ctx context.Context, rid string) error {
	rev, err := r.store.Revision(ctx, rid)
	if err != nil {
		if errors.Is(err, model.ErrRecordNotFound) {
			// nothing to do
			return nil
		} else {
			return err
		}
	}

	if rev.HeadFrom() != NoBlockNumber {
		prevHead, err := r.store.HeadAt(ctx, rev.VariantID(), rev.HeadFrom()-1)
		if err != nil && !errors.Is(err, model.ErrRecordNotFound) {
			return err
		}
		if prevHead != nil {
			newHeadTo := NoBlockNumber
			if rev.HeadTo() != NoBlockNumber {
				nextHead, err := r.store.HeadAt(ctx, rev.VariantID(), rev.HeadTo()+1)
				if err != nil && !errors.Is(err, model.ErrRecordNotFound) {
					return err
				}
				if nextHead != nil {
					newHeadTo = nextHead.HeadFrom() - 1
				}
			}
			err = r.store.UpdateRevision(ctx, prevHead.RecordID(), prevHead.Status(), prevHead.HeadFrom(), newHeadTo)
			if err != nil {
				return err
			}
		}
	}

	err = r.store.RevokeRevision(ctx, rid)
	if err != nil {
		return err
	}

	return nil
}

func (r *RaterImpl) HeadAt(ctx context.Context, blockID int64) string {
	head, err := r.store.HeadAt(ctx, r.variantID, blockID)
	if err != nil {
		return ""
	} else {
		return head.RecordID()
	}
}

func (r *RaterImpl) Head(ctx context.Context) string {
	if r.headID != "" {
		return r.headID
	}

	head, err := r.store.Head(ctx, r.variantID)
	if err != nil {
		return ""
	} else {
		r.headID = head.RecordID()
		return r.headID
	}
}
