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
	"errors"
	"fmt"
	"time"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/services/notification"
	"github.com/rs/zerolog/log"
)

func WaitForConfirmation(ctx context.Context, ledger model.Ledger, ns notification.Service, interval, timeout time.Duration, recordID ...string) (int64, error) {

	if len(recordID) == 0 {
		// fast exit if there's nothing to wait for
		return 0, nil
	}

	// wait for the last record

	blockNumber, err := waitForOneRecord(ctx, ledger, ns, interval, timeout, recordID[len(recordID)-1])
	if err != nil {
		return 0, err
	}

	// check if all the previous record got published
	if len(recordID) > 1 {
		blockNumber = 0
		for _, rid := range recordID[0 : len(recordID)-1] {
			currentState, err := ledger.GetRecordState(ctx, rid)
			if err != nil {
				return 0, err
			}

			if currentState == nil {
				return 0, fmt.Errorf("queued record not found: %s", rid)
			}

			if currentState.Status != model.StatusPublished {
				return 0, fmt.Errorf("record %s not published, status=%s", rid, currentState.Status)
			}

			if blockNumber == 0 {
				blockNumber = currentState.BlockNumber
			}
		}
	}

	return blockNumber, nil
}

func waitForOneRecord(ctx context.Context, ledger model.Ledger, ns notification.Service, interval, timeout time.Duration, recordID string) (int64, error) {

	var blockCh chan any

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	timeoutTicker := time.NewTicker(timeout)
	defer timeoutTicker.Stop()

	for {
		currentState, err := ledger.GetRecordState(ctx, recordID)
		if err != nil {
			return 0, err
		}

		if currentState == nil {
			continue
		}

		if currentState.Status == model.StatusPublished {
			// record published
			return currentState.BlockNumber, nil
		}
		if currentState.Status == model.StatusFailed {
			return 0, errors.New("record submission failed")
		}
		if currentState.Status == model.StatusRevoked {
			return 0, errors.New("record revoked")
		}

		log.Debug().Str("currentStatus", string(currentState.Status)).
			Msg("Waiting for the record to appear on the ledger")

		if blockCh == nil && ns != nil {
			blockCh, err = ns.Subscribe(model.NTopicNewBlock)
			if err != nil {
				return 0, err
			}
			defer func() { _ = ns.Unsubscribe(blockCh) }()
		}
		select {
		case <-blockCh:
			log.Debug().Msg("Record wait: received new block notification")
		case <-ticker.C:
			log.Debug().Dur("interval", interval).Msg("Waking up to check the record state")
		case <-timeoutTicker.C:
			return 0, errors.New("message confirmation timed out")
		}
	}
}
