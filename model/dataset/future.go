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

package dataset

import (
	"context"
	"fmt"
	"time"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/services/notification"
	"github.com/piprate/metalocker/utils/measure"
)

type (
	recordFutureImpl struct {
		ledger   model.Ledger
		ns       notification.Service
		recordID string
		dataset  *DataSetImpl
		heads    map[string]string
		waitList []string
		err      error
		ready    bool
		ctx      context.Context
	}
)

var _ RecordFuture = (*recordFutureImpl)(nil)

var dummyContext = context.Background()

func RecordFutureWithError(err error) RecordFuture {
	return &recordFutureImpl{
		ready: true,
		err:   err,
		ctx:   dummyContext,
	}
}

func RecordFutureWithResult(ctx context.Context, ledger model.Ledger, ns notification.Service, recordID string, dataset *DataSetImpl, heads map[string]string, waitList []string) RecordFuture {
	return &recordFutureImpl{
		ledger:   ledger,
		ns:       ns,
		recordID: recordID,
		dataset:  dataset,
		heads:    heads,
		waitList: waitList,
		ctx:      ctx,
	}
}

func (f *recordFutureImpl) Wait(timeout time.Duration) error {
	defer measure.ExecTime("recordFuture.Wait")()

	if f.ready {
		return f.err
	}

	var blockNumber int64
	blockNumber, f.err = WaitForConfirmation(f.ctx, f.ledger, f.ns, time.Second, timeout, f.waitList...)
	f.ready = true

	if f.dataset == nil {
		// it's a lease revocation record. Check the record state explicitly.
		rs, err := f.ledger.GetRecordState(f.ctx, f.recordID)
		if err != nil {
			return err
		}

		if rs.Status != model.StatusPublished {
			f.err = fmt.Errorf("record not published, current state = %s", rs.Status)
		}
	} else {
		f.dataset.blockNumber = blockNumber
	}

	return f.err
}

func (f *recordFutureImpl) ID() string {
	return f.recordID
}

func (f *recordFutureImpl) Lease() *model.Lease {
	if f.dataset != nil {
		return f.dataset.lease
	} else {
		return nil
	}
}

func (f *recordFutureImpl) DataSet() model.DataSet {
	if f.ready && f.err == nil {
		return f.dataset
	} else {
		return nil
	}
}

func (f *recordFutureImpl) Heads() map[string]string {
	return f.heads
}

func (f *recordFutureImpl) Error() error {
	return f.err
}

func (f *recordFutureImpl) IsReady() bool {
	return f.ready
}

func (f *recordFutureImpl) WaitList() []string {
	return f.waitList
}
