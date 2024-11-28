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
	"encoding/json"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/storage/rdb/ent/accesskey"

	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
)

// AccessKey is the model entity for the AccessKey schema.
type AccessKey struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// Did holds the value of the "did" field.
	Did string `json:"did,omitempty"`
	// Body holds the value of the "body" field.
	Body *model.AccessKey `json:"body,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the AccessKeyQuery when eager-loading is set.
	Edges   AccessKeyEdges `json:"edges"`
	account *int
}

// AccessKeyEdges holds the relations/edges for other nodes in the graph.
type AccessKeyEdges struct {
	// Account holds the value of the account edge.
	Account *Account `json:"account,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// AccountOrErr returns the Account value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e AccessKeyEdges) AccountOrErr() (*Account, error) {
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
func (*AccessKey) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case accesskey.FieldBody:
			values[i] = new([]byte)
		case accesskey.FieldID:
			values[i] = new(sql.NullInt64)
		case accesskey.FieldDid:
			values[i] = new(sql.NullString)
		case accesskey.ForeignKeys[0]: // account
			values[i] = new(sql.NullInt64)
		default:
			return nil, fmt.Errorf("unexpected column %q for type AccessKey", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the AccessKey fields.
func (ak *AccessKey) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case accesskey.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			ak.ID = int(value.Int64)
		case accesskey.FieldDid:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field did", values[i])
			} else if value.Valid {
				ak.Did = value.String
			}
		case accesskey.FieldBody:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field body", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &ak.Body); err != nil {
					return fmt.Errorf("unmarshal field body: %w", err)
				}
			}
		case accesskey.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for edge-field account", value)
			} else if value.Valid {
				ak.account = new(int)
				*ak.account = int(value.Int64)
			}
		}
	}
	return nil
}

// QueryAccount queries the "account" edge of the AccessKey entity.
func (ak *AccessKey) QueryAccount() *AccountQuery {
	return NewAccessKeyClient(ak.config).QueryAccount(ak)
}

// Update returns a builder for updating this AccessKey.
// Note that you need to call AccessKey.Unwrap() before calling this method if this AccessKey
// was returned from a transaction, and the transaction was committed or rolled back.
func (ak *AccessKey) Update() *AccessKeyUpdateOne {
	return NewAccessKeyClient(ak.config).UpdateOne(ak)
}

// Unwrap unwraps the AccessKey entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (ak *AccessKey) Unwrap() *AccessKey {
	_tx, ok := ak.config.driver.(*txDriver)
	if !ok {
		panic("ent: AccessKey is not a transactional entity")
	}
	ak.config.driver = _tx.drv
	return ak
}

// String implements the fmt.Stringer.
func (ak *AccessKey) String() string {
	var builder strings.Builder
	builder.WriteString("AccessKey(")
	builder.WriteString(fmt.Sprintf("id=%v, ", ak.ID))
	builder.WriteString("did=")
	builder.WriteString(ak.Did)
	builder.WriteString(", ")
	builder.WriteString("body=")
	builder.WriteString(fmt.Sprintf("%v", ak.Body))
	builder.WriteByte(')')
	return builder.String()
}

// AccessKeys is a parsable slice of AccessKey.
type AccessKeys []*AccessKey

func (ak AccessKeys) config(cfg config) {
	for _i := range ak {
		ak[_i].config = cfg
	}
}
