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

package scanner

import (
	"context"
	"sort"
	"time"

	"github.com/piprate/metalocker/model"
	"github.com/rs/zerolog/log"
)

type (
	PartyLookup func(keyID int) (string, string, string, int64)

	IndexBlockConsumer interface {
		SetSubscription(sub Subscription)
		ConsumeBlock(ctx context.Context, indexID string, partyLookup PartyLookup, n BlockNotification) error
		NotifyScanCompleted(block int64) error
	}

	IndexSubscription struct {
		indexID       string
		status        int
		inactiveSince int64
		nextKeyID     int
		keys          map[int]*lockerParty
		keyOrder      []int
		configs       []*LockerConfig
		consumer      IndexBlockConsumer
	}

	lockerParty struct {
		LockerID        string
		ParticipantID   string
		SharedSecret    string
		PublicKeyStr    string
		AcceptedAtBlock int64
	}

	LockerEntry struct {
		Locker    *model.Locker
		LastBlock int64
	}
)

var _ Subscription = (*IndexSubscription)(nil)

func NewIndexSubscription(indexID string, consumer IndexBlockConsumer) *IndexSubscription {
	sub := &IndexSubscription{
		indexID:  indexID,
		status:   ScanStatusActive,
		keys:     map[int]*lockerParty{},
		consumer: consumer,
	}
	consumer.SetSubscription(sub)

	return sub
}

func (w *IndexSubscription) Close() error {
	w.keys = nil
	w.keyOrder = nil
	w.configs = nil
	w.consumer = nil

	return nil
}

func (w *IndexSubscription) IndexID() string {
	return w.indexID
}

func (w *IndexSubscription) LockerConfigs() []*LockerConfig {
	return w.configs
}

func (w *IndexSubscription) ConsumeBlock(ctx context.Context, n BlockNotification) error {
	return w.consumer.ConsumeBlock(ctx, w.indexID, w.partyLookup, n)
}

func (w *IndexSubscription) partyLookup(keyID int) (string, string, string, int64) {
	p, found := w.keys[keyID]
	if !found {
		log.Warn().Int("KeyID", keyID).Msg("Invalid key ID")
		return "", "", "", 0
	}
	return p.LockerID, p.ParticipantID, p.SharedSecret, p.AcceptedAtBlock
}

func (w *IndexSubscription) NotifyScanCompleted(topBlock int64) error {
	return w.consumer.NotifyScanCompleted(topBlock)
}

func (w *IndexSubscription) InactiveSince() int64 {
	return w.inactiveSince
}

func (w *IndexSubscription) Status() int {
	return w.status
}

func (w *IndexSubscription) SetStatus(status int) {
	log.Debug().Str("id", w.indexID).Int("status", status).Msg("Setting index subscription status")
	if w.status != status {
		var inactiveSince int64
		if status != ScanStatusActive {
			inactiveSince = time.Now().Unix()
		}
		w.status = status
		w.inactiveSince = inactiveSince
	}
}

func (w *IndexSubscription) AddLockers(lockers ...LockerEntry) error {
	for _, le := range lockers {
		for _, p := range le.Locker.Participants {
			keyID := w.nextKeyID
			w.keys[keyID] = &lockerParty{
				LockerID:        le.Locker.ID,
				ParticipantID:   p.ID,
				SharedSecret:    p.SharedSecret,
				PublicKeyStr:    p.RootPublicKey,
				AcceptedAtBlock: le.Locker.AcceptedAtBlock(),
			}

			w.keyOrder = append(w.keyOrder, keyID)
			sort.Ints(w.keyOrder)

			cfg := &LockerConfig{
				KeyID:        keyID,
				PublicKeyStr: p.RootPublicKey,
				LastBlock:    le.LastBlock,
				Subscription: w,
			}
			if err := cfg.Hydrate(); err != nil {
				return err
			}

			w.configs = append(w.configs, cfg)

			w.nextKeyID++
		}
	}

	return nil
}
