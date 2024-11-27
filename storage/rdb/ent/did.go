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
	"github.com/piprate/metalocker/storage/rdb/ent/did"
)

// DID is the model entity for the DID schema.
type DID struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// Did holds the value of the "did" field.
	Did string `json:"did,omitempty"`
	// Body holds the value of the "body" field.
	Body *model.DIDDocument `json:"body,omitempty"`
}

// scanValues returns the types for scanning values from sql.Rows.
func (*DID) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case did.FieldBody:
			values[i] = new([]byte)
		case did.FieldID:
			values[i] = new(sql.NullInt64)
		case did.FieldDid:
			values[i] = new(sql.NullString)
		default:
			return nil, fmt.Errorf("unexpected column %q for type DID", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the DID fields.
func (d *DID) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case did.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			d.ID = int(value.Int64)
		case did.FieldDid:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field did", values[i])
			} else if value.Valid {
				d.Did = value.String
			}
		case did.FieldBody:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field body", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &d.Body); err != nil {
					return fmt.Errorf("unmarshal field body: %w", err)
				}
			}
		}
	}
	return nil
}

// Update returns a builder for updating this DID.
// Note that you need to call DID.Unwrap() before calling this method if this DID
// was returned from a transaction, and the transaction was committed or rolled back.
func (d *DID) Update() *DIDUpdateOne {
	return NewDIDClient(d.config).UpdateOne(d)
}

// Unwrap unwraps the DID entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (d *DID) Unwrap() *DID {
	_tx, ok := d.config.driver.(*txDriver)
	if !ok {
		panic("ent: DID is not a transactional entity")
	}
	d.config.driver = _tx.drv
	return d
}

// String implements the fmt.Stringer.
func (d *DID) String() string {
	var builder strings.Builder
	builder.WriteString("DID(")
	builder.WriteString(fmt.Sprintf("id=%v, ", d.ID))
	builder.WriteString("did=")
	builder.WriteString(d.Did)
	builder.WriteString(", ")
	builder.WriteString("body=")
	builder.WriteString(fmt.Sprintf("%v", d.Body))
	builder.WriteByte(')')
	return builder.String()
}

// DIDs is a parsable slice of DID.
type DIDs []*DID

func (d DIDs) config(cfg config) {
	for _i := range d {
		d[_i].config = cfg
	}
}
