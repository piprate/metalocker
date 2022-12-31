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

import "io"

// Service is an interface to Notification Service that enables
// different parts of the MetaLocker platform to publish and subscribe
// to events of interest, such as new ledger block generation, etc.
type Service interface {
	Publish(msg any, wait, broadcast bool, topics ...string) error
	Subscribe(topics ...string) (chan any, error)
	Unsubscribe(ch chan any, topics ...string) error
	CloseTopics(topics ...string) error
	io.Closer
}

type ServiceCreatorFn func() (Service, error)
