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
	"github.com/piprate/metalocker/model/account"

	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
)

// Account is the model entity for the Account schema.
type Account struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// Did holds the value of the "did" field.
	Did string `json:"did,omitempty"`
	// State holds the value of the "state" field.
	State string `json:"state,omitempty"`
	// Email holds the value of the "email" field.
	Email string `json:"email,omitempty"`
	// ParentAccount holds the value of the "parent_account" field.
	ParentAccount string `json:"parent_account,omitempty"`
	// Body holds the value of the "body" field.
	Body *account.Account `json:"body,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the AccountQuery when eager-loading is set.
	Edges AccountEdges `json:"edges"`
}

// AccountEdges holds the relations/edges for other nodes in the graph.
type AccountEdges struct {
	// RecoveryCodes holds the value of the recovery_codes edge.
	RecoveryCodes []*RecoveryCode `json:"recovery_codes,omitempty"`
	// AccessKeys holds the value of the access_keys edge.
	AccessKeys []*AccessKey `json:"access_keys,omitempty"`
	// Identities holds the value of the identities edge.
	Identities []*Identity `json:"identities,omitempty"`
	// Lockers holds the value of the lockers edge.
	Lockers []*Locker `json:"lockers,omitempty"`
	// Properties holds the value of the properties edge.
	Properties []*Property `json:"properties,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [5]bool
}

// RecoveryCodesOrErr returns the RecoveryCodes value or an error if the edge
// was not loaded in eager-loading.
func (e AccountEdges) RecoveryCodesOrErr() ([]*RecoveryCode, error) {
	if e.loadedTypes[0] {
		return e.RecoveryCodes, nil
	}
	return nil, &NotLoadedError{edge: "recovery_codes"}
}

// AccessKeysOrErr returns the AccessKeys value or an error if the edge
// was not loaded in eager-loading.
func (e AccountEdges) AccessKeysOrErr() ([]*AccessKey, error) {
	if e.loadedTypes[1] {
		return e.AccessKeys, nil
	}
	return nil, &NotLoadedError{edge: "access_keys"}
}

// IdentitiesOrErr returns the Identities value or an error if the edge
// was not loaded in eager-loading.
func (e AccountEdges) IdentitiesOrErr() ([]*Identity, error) {
	if e.loadedTypes[2] {
		return e.Identities, nil
	}
	return nil, &NotLoadedError{edge: "identities"}
}

// LockersOrErr returns the Lockers value or an error if the edge
// was not loaded in eager-loading.
func (e AccountEdges) LockersOrErr() ([]*Locker, error) {
	if e.loadedTypes[3] {
		return e.Lockers, nil
	}
	return nil, &NotLoadedError{edge: "lockers"}
}

// PropertiesOrErr returns the Properties value or an error if the edge
// was not loaded in eager-loading.
func (e AccountEdges) PropertiesOrErr() ([]*Property, error) {
	if e.loadedTypes[4] {
		return e.Properties, nil
	}
	return nil, &NotLoadedError{edge: "properties"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Account) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case entaccount.FieldBody:
			values[i] = new([]byte)
		case entaccount.FieldID:
			values[i] = new(sql.NullInt64)
		case entaccount.FieldDid, entaccount.FieldState, entaccount.FieldEmail, entaccount.FieldParentAccount:
			values[i] = new(sql.NullString)
		default:
			return nil, fmt.Errorf("unexpected column %q for type Account", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Account fields.
func (a *Account) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case entaccount.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			a.ID = int(value.Int64)
		case entaccount.FieldDid:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field did", values[i])
			} else if value.Valid {
				a.Did = value.String
			}
		case entaccount.FieldState:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field state", values[i])
			} else if value.Valid {
				a.State = value.String
			}
		case entaccount.FieldEmail:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field email", values[i])
			} else if value.Valid {
				a.Email = value.String
			}
		case entaccount.FieldParentAccount:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field parent_account", values[i])
			} else if value.Valid {
				a.ParentAccount = value.String
			}
		case entaccount.FieldBody:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field body", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &a.Body); err != nil {
					return fmt.Errorf("unmarshal field body: %w", err)
				}
			}
		}
	}
	return nil
}

// QueryRecoveryCodes queries the "recovery_codes" edge of the Account entity.
func (a *Account) QueryRecoveryCodes() *RecoveryCodeQuery {
	return NewAccountClient(a.config).QueryRecoveryCodes(a)
}

// QueryAccessKeys queries the "access_keys" edge of the Account entity.
func (a *Account) QueryAccessKeys() *AccessKeyQuery {
	return NewAccountClient(a.config).QueryAccessKeys(a)
}

// QueryIdentities queries the "identities" edge of the Account entity.
func (a *Account) QueryIdentities() *IdentityQuery {
	return NewAccountClient(a.config).QueryIdentities(a)
}

// QueryLockers queries the "lockers" edge of the Account entity.
func (a *Account) QueryLockers() *LockerQuery {
	return NewAccountClient(a.config).QueryLockers(a)
}

// QueryProperties queries the "properties" edge of the Account entity.
func (a *Account) QueryProperties() *PropertyQuery {
	return NewAccountClient(a.config).QueryProperties(a)
}

// Update returns a builder for updating this Account.
// Note that you need to call Account.Unwrap() before calling this method if this Account
// was returned from a transaction, and the transaction was committed or rolled back.
func (a *Account) Update() *AccountUpdateOne {
	return NewAccountClient(a.config).UpdateOne(a)
}

// Unwrap unwraps the Account entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (a *Account) Unwrap() *Account {
	_tx, ok := a.config.driver.(*txDriver)
	if !ok {
		panic("ent: Account is not a transactional entity")
	}
	a.config.driver = _tx.drv
	return a
}

// String implements the fmt.Stringer.
func (a *Account) String() string {
	var builder strings.Builder
	builder.WriteString("Account(")
	builder.WriteString(fmt.Sprintf("id=%v, ", a.ID))
	builder.WriteString("did=")
	builder.WriteString(a.Did)
	builder.WriteString(", ")
	builder.WriteString("state=")
	builder.WriteString(a.State)
	builder.WriteString(", ")
	builder.WriteString("email=")
	builder.WriteString(a.Email)
	builder.WriteString(", ")
	builder.WriteString("parent_account=")
	builder.WriteString(a.ParentAccount)
	builder.WriteString(", ")
	builder.WriteString("body=")
	builder.WriteString(fmt.Sprintf("%v", a.Body))
	builder.WriteByte(')')
	return builder.String()
}

// Accounts is a parsable slice of Account.
type Accounts []*Account

func (a Accounts) config(cfg config) {
	for _i := range a {
		a[_i].config = cfg
	}
}
