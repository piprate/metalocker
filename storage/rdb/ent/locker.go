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

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/piprate/metalocker/storage/rdb/ent/locker"

	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
)

// Locker is the model entity for the Locker schema.
type Locker struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// Hash holds the value of the "hash" field.
	Hash string `json:"hash,omitempty"`
	// Level holds the value of the "level" field.
	Level int32 `json:"level,omitempty"`
	// EncryptedID holds the value of the "encrypted_id" field.
	EncryptedID string `json:"encrypted_id,omitempty"`
	// EncryptedBody holds the value of the "encrypted_body" field.
	EncryptedBody string `json:"encrypted_body,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the LockerQuery when eager-loading is set.
	Edges   LockerEdges `json:"edges"`
	account *int
}

// LockerEdges holds the relations/edges for other nodes in the graph.
type LockerEdges struct {
	// Account holds the value of the account edge.
	Account *Account `json:"account,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// AccountOrErr returns the Account value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e LockerEdges) AccountOrErr() (*Account, error) {
	if e.loadedTypes[0] {
		if e.Account == nil {
			// Edge was loaded but was not found.
			return nil, &NotFoundError{label: entaccount.Label}
		}
		return e.Account, nil
	}
	return nil, &NotLoadedError{edge: "account"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Locker) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case locker.FieldID, locker.FieldLevel:
			values[i] = new(sql.NullInt64)
		case locker.FieldHash, locker.FieldEncryptedID, locker.FieldEncryptedBody:
			values[i] = new(sql.NullString)
		case locker.ForeignKeys[0]: // account
			values[i] = new(sql.NullInt64)
		default:
			return nil, fmt.Errorf("unexpected column %q for type Locker", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Locker fields.
func (l *Locker) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case locker.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			l.ID = int(value.Int64)
		case locker.FieldHash:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field hash", values[i])
			} else if value.Valid {
				l.Hash = value.String
			}
		case locker.FieldLevel:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field level", values[i])
			} else if value.Valid {
				l.Level = int32(value.Int64)
			}
		case locker.FieldEncryptedID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field encrypted_id", values[i])
			} else if value.Valid {
				l.EncryptedID = value.String
			}
		case locker.FieldEncryptedBody:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field encrypted_body", values[i])
			} else if value.Valid {
				l.EncryptedBody = value.String
			}
		case locker.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for edge-field account", value)
			} else if value.Valid {
				l.account = new(int)
				*l.account = int(value.Int64)
			}
		}
	}
	return nil
}

// QueryAccount queries the "account" edge of the Locker entity.
func (l *Locker) QueryAccount() *AccountQuery {
	return NewLockerClient(l.config).QueryAccount(l)
}

// Update returns a builder for updating this Locker.
// Note that you need to call Locker.Unwrap() before calling this method if this Locker
// was returned from a transaction, and the transaction was committed or rolled back.
func (l *Locker) Update() *LockerUpdateOne {
	return NewLockerClient(l.config).UpdateOne(l)
}

// Unwrap unwraps the Locker entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (l *Locker) Unwrap() *Locker {
	_tx, ok := l.config.driver.(*txDriver)
	if !ok {
		panic("ent: Locker is not a transactional entity")
	}
	l.config.driver = _tx.drv
	return l
}

// String implements the fmt.Stringer.
func (l *Locker) String() string {
	var builder strings.Builder
	builder.WriteString("Locker(")
	builder.WriteString(fmt.Sprintf("id=%v, ", l.ID))
	builder.WriteString("hash=")
	builder.WriteString(l.Hash)
	builder.WriteString(", ")
	builder.WriteString("level=")
	builder.WriteString(fmt.Sprintf("%v", l.Level))
	builder.WriteString(", ")
	builder.WriteString("encrypted_id=")
	builder.WriteString(l.EncryptedID)
	builder.WriteString(", ")
	builder.WriteString("encrypted_body=")
	builder.WriteString(l.EncryptedBody)
	builder.WriteByte(')')
	return builder.String()
}

// Lockers is a parsable slice of Locker.
type Lockers []*Locker

func (l Lockers) config(cfg config) {
	for _i := range l {
		l[_i].config = cfg
	}
}
