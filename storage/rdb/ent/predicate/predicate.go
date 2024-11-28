// Copyright 2024 Piprate Limited
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

package predicate

import (
	"entgo.io/ent/dialect/sql"
)

// AccessKey is the predicate function for accesskey builders.
type AccessKey func(*sql.Selector)

// Account is the predicate function for entaccount builders.
type Account func(*sql.Selector)

// DID is the predicate function for did builders.
type DID func(*sql.Selector)

// Identity is the predicate function for identity builders.
type Identity func(*sql.Selector)

// Locker is the predicate function for locker builders.
type Locker func(*sql.Selector)

// Property is the predicate function for property builders.
type Property func(*sql.Selector)

// RecoveryCode is the predicate function for recoverycode builders.
type RecoveryCode func(*sql.Selector)
