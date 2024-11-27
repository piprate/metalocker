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
	"github.com/piprate/metalocker/storage/rdb/ent/property"

	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
)

// Property is the model entity for the Property schema.
type Property struct {
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
	// The values are being populated by the PropertyQuery when eager-loading is set.
	Edges   PropertyEdges `json:"edges"`
	account *int
}

// PropertyEdges holds the relations/edges for other nodes in the graph.
type PropertyEdges struct {
	// Account holds the value of the account edge.
	Account *Account `json:"account,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// AccountOrErr returns the Account value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e PropertyEdges) AccountOrErr() (*Account, error) {
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
func (*Property) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case property.FieldID, property.FieldLevel:
			values[i] = new(sql.NullInt64)
		case property.FieldHash, property.FieldEncryptedID, property.FieldEncryptedBody:
			values[i] = new(sql.NullString)
		case property.ForeignKeys[0]: // account
			values[i] = new(sql.NullInt64)
		default:
			return nil, fmt.Errorf("unexpected column %q for type Property", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Property fields.
func (pr *Property) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case property.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			pr.ID = int(value.Int64)
		case property.FieldHash:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field hash", values[i])
			} else if value.Valid {
				pr.Hash = value.String
			}
		case property.FieldLevel:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field level", values[i])
			} else if value.Valid {
				pr.Level = int32(value.Int64)
			}
		case property.FieldEncryptedID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field encrypted_id", values[i])
			} else if value.Valid {
				pr.EncryptedID = value.String
			}
		case property.FieldEncryptedBody:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field encrypted_body", values[i])
			} else if value.Valid {
				pr.EncryptedBody = value.String
			}
		case property.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for edge-field account", value)
			} else if value.Valid {
				pr.account = new(int)
				*pr.account = int(value.Int64)
			}
		}
	}
	return nil
}

// QueryAccount queries the "account" edge of the Property entity.
func (pr *Property) QueryAccount() *AccountQuery {
	return NewPropertyClient(pr.config).QueryAccount(pr)
}

// Update returns a builder for updating this Property.
// Note that you need to call Property.Unwrap() before calling this method if this Property
// was returned from a transaction, and the transaction was committed or rolled back.
func (pr *Property) Update() *PropertyUpdateOne {
	return NewPropertyClient(pr.config).UpdateOne(pr)
}

// Unwrap unwraps the Property entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (pr *Property) Unwrap() *Property {
	_tx, ok := pr.config.driver.(*txDriver)
	if !ok {
		panic("ent: Property is not a transactional entity")
	}
	pr.config.driver = _tx.drv
	return pr
}

// String implements the fmt.Stringer.
func (pr *Property) String() string {
	var builder strings.Builder
	builder.WriteString("Property(")
	builder.WriteString(fmt.Sprintf("id=%v, ", pr.ID))
	builder.WriteString("hash=")
	builder.WriteString(pr.Hash)
	builder.WriteString(", ")
	builder.WriteString("level=")
	builder.WriteString(fmt.Sprintf("%v", pr.Level))
	builder.WriteString(", ")
	builder.WriteString("encrypted_id=")
	builder.WriteString(pr.EncryptedID)
	builder.WriteString(", ")
	builder.WriteString("encrypted_body=")
	builder.WriteString(pr.EncryptedBody)
	builder.WriteByte(')')
	return builder.String()
}

// Properties is a parsable slice of Property.
type Properties []*Property

func (pr Properties) config(cfg config) {
	for _i := range pr {
		pr[_i].config = cfg
	}
}
