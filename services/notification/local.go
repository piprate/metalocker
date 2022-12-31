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

package notification

import (
	"github.com/cskr/pubsub"
	"github.com/rs/zerolog/log"
)

type LocalNotificationService struct {
	ps *pubsub.PubSub
}

func NewLocalNotificationService(capacity int) *LocalNotificationService {
	ps := pubsub.New(capacity)

	return &LocalNotificationService{
		ps: ps,
	}
}

func (lns *LocalNotificationService) Publish(msg any, wait, broadcast bool, topics ...string) error {
	log.Debug().Strs("topics", topics).Interface("msg", msg).Msg("Publish notification message")
	if wait {
		lns.ps.Pub(msg, topics...)
	} else {
		lns.ps.TryPub(msg, topics...)
	}
	return nil
}

func (lns *LocalNotificationService) Subscribe(topics ...string) (chan any, error) {
	log.Debug().Strs("topics", topics).Msg("Subscribe to notification messages")
	return lns.ps.Sub(topics...), nil
}

func (lns *LocalNotificationService) Unsubscribe(ch chan any, topics ...string) error {
	log.Debug().Strs("topics", topics).Msg("Unsubscribe from notification messages")
	lns.ps.Unsub(ch, topics...)
	return nil
}

func (lns *LocalNotificationService) CloseTopics(topics ...string) error {
	log.Debug().Strs("topics", topics).Msg("Close notification topics")
	lns.ps.Close(topics...)
	return nil
}

func (lns *LocalNotificationService) Close() error {
	lns.ps.Shutdown()
	return nil
}
