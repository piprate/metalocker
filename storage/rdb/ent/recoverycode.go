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
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/piprate/metalocker/storage/rdb/ent/recoverycode"

	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
)

// RecoveryCode is the model entity for the RecoveryCode schema.
type RecoveryCode struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// Code holds the value of the "code" field.
	Code string `json:"code,omitempty"`
	// ExpiresAt holds the value of the "expires_at" field.
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the RecoveryCodeQuery when eager-loading is set.
	Edges   RecoveryCodeEdges `json:"edges"`
	account *int
}

// RecoveryCodeEdges holds the relations/edges for other nodes in the graph.
type RecoveryCodeEdges struct {
	// Account holds the value of the account edge.
	Account *Account `json:"account,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// AccountOrErr returns the Account value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e RecoveryCodeEdges) AccountOrErr() (*Account, error) {
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
func (*RecoveryCode) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case recoverycode.FieldID:
			values[i] = new(sql.NullInt64)
		case recoverycode.FieldCode:
			values[i] = new(sql.NullString)
		case recoverycode.FieldExpiresAt:
			values[i] = new(sql.NullTime)
		case recoverycode.ForeignKeys[0]: // account
			values[i] = new(sql.NullInt64)
		default:
			return nil, fmt.Errorf("unexpected column %q for type RecoveryCode", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the RecoveryCode fields.
func (rc *RecoveryCode) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case recoverycode.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			rc.ID = int(value.Int64)
		case recoverycode.FieldCode:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field code", values[i])
			} else if value.Valid {
				rc.Code = value.String
			}
		case recoverycode.FieldExpiresAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field expires_at", values[i])
			} else if value.Valid {
				rc.ExpiresAt = new(time.Time)
				*rc.ExpiresAt = value.Time
			}
		case recoverycode.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for edge-field account", value)
			} else if value.Valid {
				rc.account = new(int)
				*rc.account = int(value.Int64)
			}
		}
	}
	return nil
}

// QueryAccount queries the "account" edge of the RecoveryCode entity.
func (rc *RecoveryCode) QueryAccount() *AccountQuery {
	return NewRecoveryCodeClient(rc.config).QueryAccount(rc)
}

// Update returns a builder for updating this RecoveryCode.
// Note that you need to call RecoveryCode.Unwrap() before calling this method if this RecoveryCode
// was returned from a transaction, and the transaction was committed or rolled back.
func (rc *RecoveryCode) Update() *RecoveryCodeUpdateOne {
	return NewRecoveryCodeClient(rc.config).UpdateOne(rc)
}

// Unwrap unwraps the RecoveryCode entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (rc *RecoveryCode) Unwrap() *RecoveryCode {
	_tx, ok := rc.config.driver.(*txDriver)
	if !ok {
		panic("ent: RecoveryCode is not a transactional entity")
	}
	rc.config.driver = _tx.drv
	return rc
}

// String implements the fmt.Stringer.
func (rc *RecoveryCode) String() string {
	var builder strings.Builder
	builder.WriteString("RecoveryCode(")
	builder.WriteString(fmt.Sprintf("id=%v, ", rc.ID))
	builder.WriteString("code=")
	builder.WriteString(rc.Code)
	builder.WriteString(", ")
	if v := rc.ExpiresAt; v != nil {
		builder.WriteString("expires_at=")
		builder.WriteString(v.Format(time.ANSIC))
	}
	builder.WriteByte(')')
	return builder.String()
}

// RecoveryCodes is a parsable slice of RecoveryCode.
type RecoveryCodes []*RecoveryCode

func (rc RecoveryCodes) config(cfg config) {
	for _i := range rc {
		rc[_i].config = cfg
	}
}
