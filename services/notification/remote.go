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
	"crypto/rand"
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/cskr/pubsub"
	"github.com/gorilla/websocket"
	"github.com/piprate/metalocker/utils/jsonw"
	"github.com/rs/zerolog/log"
)

type Message struct {
	SubID string
	Msg   any
}

type WSDialerCreator func() (*websocket.Conn, error)

type RemoteNotificationService struct {
	dc         WSDialerCreator
	maxRedials int
	startOnce  sync.Once
	started    bool

	ps         *pubsub.PubSub
	conn       *websocket.Conn
	serviceID  string
	nextSubID  uint64
	mapByID    map[string]chan any
	mapByCh    map[chan any]string
	configByCh map[chan any][]string

	redialCount int

	commMutex sync.Mutex
	closed    bool
}

func NewRemoteNotificationService(dc WSDialerCreator, maxRedials int) (*RemoteNotificationService, error) {
	rns := &RemoteNotificationService{
		dc:         dc,
		maxRedials: maxRedials,
	}

	return rns, nil
}

func (rns *RemoteNotificationService) checkStarted() error {
	if rns.started {
		if rns.closed {
			return errors.New("remote notification service closed")
		} else {
			return nil
		}
	} else {
		var err error
		rns.startOnce.Do(func() {
			err = rns.start()
		})
		return err
	}
}

func (rns *RemoteNotificationService) start() error {
	conn, err := rns.dc()
	if err != nil {
		return err
	}

	ps := pubsub.New(0)

	randBuffer := make([]byte, 8)
	_, err = rand.Read(randBuffer)
	if err != nil {
		return err
	}
	serviceID := base58.Encode(randBuffer)

	rns.ps = ps
	rns.conn = conn
	rns.serviceID = serviceID

	rns.mapByID = make(map[string]chan any)
	rns.mapByCh = make(map[chan any]string)
	rns.configByCh = make(map[chan any][]string)

	go func() {
		for {
			_, message, err := rns.conn.ReadMessage()
			if err != nil {
				if rns.closed {
					// stop quietly
					return
				}
				log.Err(err).Msg("Error when reading WS message")

				if err = rns.redial(); err != nil {
					_ = rns.Close()
					return
				}
				if err = rns.restoreSubs(); err != nil {
					log.Err(err).Msg("Error restoring subscriptions")
					_ = rns.Close()
					return
				}
				continue
			}

			var msg Message
			err = jsonw.Unmarshal(message, &msg)
			if err != nil {
				log.Err(err).Str("body", string(message)).Msg("Error when unmarshalling WS message")
			}

			log.Debug().Interface("body", msg).Msg("Received notification message")

			ch, found := rns.mapByID[msg.SubID]
			if found {
				select {
				case ch <- msg.Msg:
					log.Debug().Str("subID", msg.SubID).Msg("Forwarded notification message")
				default:
					log.Debug().Str("subID", msg.SubID).Msg("No receiver for notification message")
				}
			} else {
				log.Warn().Str("id", msg.SubID).Msg("Subscriber not found")
			}
		}
	}()

	log.Debug().Str("serviceID", serviceID).Msg("Remote notification service started")

	rns.started = true

	return nil
}

func (rns *RemoteNotificationService) nextID() string {
	idx := atomic.AddUint64(&rns.nextSubID, 1)
	return rns.serviceID + "_" + strconv.FormatUint(idx, 10)
}

func (rns *RemoteNotificationService) redial() error {
	rns.commMutex.Lock()
	defer rns.commMutex.Unlock()

	if rns.closed {
		return errors.New("trying to initiate redial for closed Notification Service")
	}

	log.Debug().Msg("Initiating WS redial")

	rns.redialCount = 0

	_ = rns.conn.Close()

	for {
		if rns.maxRedials != 0 && rns.redialCount >= rns.maxRedials {
			log.Error().Int("maxRedials", rns.maxRedials).Msg("Max notification service redial count exceeded")
			return errors.New("max notification service redial count exceeded")
		}
		conn, err := rns.dc()
		if err == nil {
			rns.conn = conn

			log.Debug().Msg("Notification Service connection restored")

			return nil
		}

		rns.redialCount++

		log.Debug().Msg("Waiting for 1 second before next redial")
		time.Sleep(time.Second)
	}
}

func (rns *RemoteNotificationService) restoreSubs() error {
	for ch, topics := range rns.configByCh {
		id := rns.mapByCh[ch]
		if err := rns.remoteSubscribe(id, topics...); err != nil {
			return err
		}
	}
	return nil
}

func (rns *RemoteNotificationService) Publish(msg any, wait, broadcast bool, topics ...string) error {
	if err := rns.checkStarted(); err != nil {
		return err
	}

	if !broadcast {
		if wait {
			rns.ps.Pub(msg, topics...)
		} else {
			rns.ps.TryPub(msg, topics...)
		}
		return nil
	} else {
		//return errors.New("remote publish not supported")
		pubMsg := map[string]any{
			"type":      "nsPublish",
			"msg":       msg,
			"wait":      wait,
			"broadcast": true,
			"topics":    topics,
		}
		pubMsgBytes, _ := jsonw.Marshal(pubMsg)
		err := rns.writeMessage(websocket.TextMessage, pubMsgBytes)
		if err != nil {
			log.Err(err).Msg("Error sending WS message")
			return err
		}
		return nil
	}
}

func (rns *RemoteNotificationService) remoteSubscribe(id string, topics ...string) error {
	subMsg := map[string]any{
		"type":   "nsSubscribe",
		"id":     id,
		"topics": topics,
	}
	subMsgBytes, _ := jsonw.Marshal(subMsg)
	err := rns.writeMessage(websocket.TextMessage, subMsgBytes)
	if err != nil {
		log.Err(err).Msg("Error sending WS message")
		return err
	}
	return nil
}

func (rns *RemoteNotificationService) remoteUnsubscribe(id string, topics ...string) error {
	subMsg := map[string]any{
		"type":   "nsUnsubscribe",
		"id":     id,
		"topics": topics,
	}
	subMsgBytes, _ := jsonw.Marshal(subMsg)
	err := rns.writeMessage(websocket.TextMessage, subMsgBytes)
	if err != nil {
		log.Err(err).Msg("Error sending WS message")
		return err
	}
	return nil
}

func (rns *RemoteNotificationService) writeMessage(messageType int, data []byte) error {
	rns.commMutex.Lock()
	defer rns.commMutex.Unlock()
	return rns.conn.WriteMessage(messageType, data)
}

func (rns *RemoteNotificationService) Subscribe(topics ...string) (chan any, error) {
	if err := rns.checkStarted(); err != nil {
		return nil, err
	}

	id := rns.nextID()

	if err := rns.remoteSubscribe(id, topics...); err != nil {
		return nil, err
	}

	ch := rns.ps.Sub(topics...)
	rns.mapByID[id] = ch
	rns.mapByCh[ch] = id
	rns.configByCh[ch] = topics

	log.Debug().Str("subID", id).Strs("topics", topics).Msg("New remote notification subscription")

	return ch, nil
}

func (rns *RemoteNotificationService) Unsubscribe(ch chan any, topics ...string) error {
	if err := rns.checkStarted(); err != nil {
		return err
	}

	id := rns.mapByCh[ch]
	delete(rns.mapByID, id)
	delete(rns.mapByCh, ch)
	delete(rns.configByCh, ch)

	rns.ps.Unsub(ch, topics...)

	return rns.remoteUnsubscribe(id, topics...)
}

func (rns *RemoteNotificationService) CloseTopics(topics ...string) error {
	return nil
}

func (rns *RemoteNotificationService) Close() error {
	if rns.started {
		log.Debug().Msg("Closing remote Notification Service")
		for ch := range rns.mapByCh {
			rns.ps.Unsub(ch)
		}

		rns.mapByID = make(map[string]chan any)
		rns.mapByCh = make(map[chan any]string)
		rns.configByCh = make(map[chan any][]string)

		rns.ps.Shutdown()

		rns.closed = true

		if rns.conn != nil {
			return rns.conn.Close()
		}
	}

	return nil
}
