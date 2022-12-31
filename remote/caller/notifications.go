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

package caller

import (
	"github.com/gorilla/websocket"
	"github.com/piprate/metalocker/services/notification"
)

func (c *MetaLockerHTTPCaller) NotificationService() (notification.Service, error) {
	if c.ns == nil {
		ns, err := notification.NewRemoteNotificationService(func() (*websocket.Conn, error) {
			return c.client.DialWebSocket("/v1/notifications")
		}, 0)
		if err != nil {
			return nil, err
		}
		c.ns = ns
	}
	return c.ns, nil
}

func (c *MetaLockerHTTPCaller) CloseNotificationService() error {
	if c.ns != nil {
		return c.ns.Close()
	}
	return nil
}
