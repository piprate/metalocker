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
	"sync"

	"github.com/rs/zerolog/log"
)

type MessageProcessorFn func(subID string, msg any) error

type SubscriberProxy struct {
	id        string
	ns        Service
	ch        chan any
	onceClose sync.Once
	fn        MessageProcessorFn
}

func NewSubscriberProxy(id string, ns Service, ch chan any, fn MessageProcessorFn) *SubscriberProxy {
	sp := &SubscriberProxy{
		id: id,
		ns: ns,
		ch: ch,
		fn: fn,
	}

	go func() {
		for msg := range sp.ch {
			if msg == nil {
				log.Debug().Msg("Received nil message, shutting down subscriber proxy")
				return
			}
			_ = sp.fn(id, msg)
		}
	}()
	return sp
}

// Close will close the internal channel and stop receiving messages
func (sp *SubscriberProxy) Close() {
	log.Debug().Str("id", sp.id).Msg("Closing subscriber proxy")
	sp.onceClose.Do(func() {
		_ = sp.ns.Unsubscribe(sp.ch)
		//close(sp.ch)
	})
}
