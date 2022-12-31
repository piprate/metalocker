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

package api

import (
	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/piprate/metalocker/sdk/apibase"
	"github.com/piprate/metalocker/services/notification"
	"github.com/piprate/metalocker/utils/jsonw"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NotificationChannelHandler(ns notification.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := apibase.CtxLogger(c)

		conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)

		if err != nil {
			log.Err(err).Msg("Failed to upgrade connection")
			return
		}

		defer conn.Close()

		writeCh := make(chan *notification.Message)

		go func() {
			for msg := range writeCh {
				log.Debug().Str("subID", msg.SubID).Msg("Writing new message")
				msgBytes, _ := jsonw.Marshal(msg)
				if err = conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
					log.Err(err).Msg("Error writing websocket message")
					break
				}
			}
		}()

		proxyMap := make(map[string]*notification.SubscriberProxy)

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
					log.Err(err).Msg("Error reading websocket message")
				}
				break
			}
			log.Debug().Str("msg", string(msg)).Msg("Received message via websockets")

			msgDoc, err := sonic.Get(msg)
			if err != nil {
				log.Err(err).Msg("Bad message, aborting...")
				break
			}

			var msgType string
			typeNode := msgDoc.Get("type")
			if typeNode.Valid() {
				msgType, _ = msgDoc.Get("type").String()
			}

			switch msgType {
			case "nsSubscribe":
				ch, _ := ns.Subscribe(extractStringArray(msgDoc.Get("topics"))...)

				subID, _ := msgDoc.Get("id").String()

				sub := notification.NewSubscriberProxy(subID, ns, ch, func(id string, msg any) error {
					log.Debug().Interface("msg", msg).Str("subID", id).Msg("Proxy received new message")
					writeCh <- &notification.Message{
						SubID: id,
						Msg:   msg,
					}
					log.Debug().Str("subID", id).Msg("Sent message for writing")
					return nil
				})

				proxyMap[subID] = sub

				defer sub.Close()
			case "nsUnsubscribe":
				subID, _ := msgDoc.Get("id").String()
				sub, found := proxyMap[subID]
				if found {
					sub.Close()
					delete(proxyMap, subID)
				} else {
					log.Warn().Str("subID", subID).Msg("Attempted tounsubscribe a non-existent client")
				}
			case "nsPublish":
				msgToPublish, _ := msgDoc.Get("msg").Interface()
				wait, _ := msgDoc.Get("wait").Bool()
				broadcast, _ := msgDoc.Get("broadcast").Bool()
				topics := extractStringArray(msgDoc.Get("topics"))
				err = ns.Publish(msgToPublish, wait, broadcast, topics...)
				if err != nil {
					log.Err(err).Msg("Error when publishing remote message")
				}
			}
		}
	}
}

func extractStringArray(d *ast.Node) []string {
	val, err := d.Array()
	if err != nil {
		return nil
	}
	res := make([]string, len(val))
	for i, v := range val {
		res[i] = v.(string)
	}
	return res
}
