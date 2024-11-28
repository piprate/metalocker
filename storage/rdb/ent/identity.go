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
	"github.com/piprate/metalocker/storage/rdb/ent/identity"

	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
)

// Identity is the model entity for the Identity schema.
type Identity struct {
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
	// The values are being populated by the IdentityQuery when eager-loading is set.
	Edges   IdentityEdges `json:"edges"`
	account *int
}

// IdentityEdges holds the relations/edges for other nodes in the graph.
type IdentityEdges struct {
	// Account holds the value of the account edge.
	Account *Account `json:"account,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// AccountOrErr returns the Account value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e IdentityEdges) AccountOrErr() (*Account, error) {
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
func (*Identity) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case identity.FieldID, identity.FieldLevel:
			values[i] = new(sql.NullInt64)
		case identity.FieldHash, identity.FieldEncryptedID, identity.FieldEncryptedBody:
			values[i] = new(sql.NullString)
		case identity.ForeignKeys[0]: // account
			values[i] = new(sql.NullInt64)
		default:
			return nil, fmt.Errorf("unexpected column %q for type Identity", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Identity fields.
func (i *Identity) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for j := range columns {
		switch columns[j] {
		case identity.FieldID:
			value, ok := values[j].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			i.ID = int(value.Int64)
		case identity.FieldHash:
			if value, ok := values[j].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field hash", values[j])
			} else if value.Valid {
				i.Hash = value.String
			}
		case identity.FieldLevel:
			if value, ok := values[j].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field level", values[j])
			} else if value.Valid {
				i.Level = int32(value.Int64)
			}
		case identity.FieldEncryptedID:
			if value, ok := values[j].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field encrypted_id", values[j])
			} else if value.Valid {
				i.EncryptedID = value.String
			}
		case identity.FieldEncryptedBody:
			if value, ok := values[j].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field encrypted_body", values[j])
			} else if value.Valid {
				i.EncryptedBody = value.String
			}
		case identity.ForeignKeys[0]:
			if value, ok := values[j].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for edge-field account", value)
			} else if value.Valid {
				i.account = new(int)
				*i.account = int(value.Int64)
			}
		}
	}
	return nil
}

// QueryAccount queries the "account" edge of the Identity entity.
func (i *Identity) QueryAccount() *AccountQuery {
	return NewIdentityClient(i.config).QueryAccount(i)
}

// Update returns a builder for updating this Identity.
// Note that you need to call Identity.Unwrap() before calling this method if this Identity
// was returned from a transaction, and the transaction was committed or rolled back.
func (i *Identity) Update() *IdentityUpdateOne {
	return NewIdentityClient(i.config).UpdateOne(i)
}

// Unwrap unwraps the Identity entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (i *Identity) Unwrap() *Identity {
	_tx, ok := i.config.driver.(*txDriver)
	if !ok {
		panic("ent: Identity is not a transactional entity")
	}
	i.config.driver = _tx.drv
	return i
}

// String implements the fmt.Stringer.
func (i *Identity) String() string {
	var builder strings.Builder
	builder.WriteString("Identity(")
	builder.WriteString(fmt.Sprintf("id=%v, ", i.ID))
	builder.WriteString("hash=")
	builder.WriteString(i.Hash)
	builder.WriteString(", ")
	builder.WriteString("level=")
	builder.WriteString(fmt.Sprintf("%v", i.Level))
	builder.WriteString(", ")
	builder.WriteString("encrypted_id=")
	builder.WriteString(i.EncryptedID)
	builder.WriteString(", ")
	builder.WriteString("encrypted_body=")
	builder.WriteString(i.EncryptedBody)
	builder.WriteByte(')')
	return builder.String()
}

// Identities is a parsable slice of Identity.
type Identities []*Identity

func (i Identities) config(cfg config) {
	for _i := range i {
		i[_i].config = cfg
	}
}
