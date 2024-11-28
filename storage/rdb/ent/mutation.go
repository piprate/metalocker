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
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/storage/rdb/ent/accesskey"
	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
	"github.com/piprate/metalocker/storage/rdb/ent/did"
	"github.com/piprate/metalocker/storage/rdb/ent/identity"
	"github.com/piprate/metalocker/storage/rdb/ent/locker"
	"github.com/piprate/metalocker/storage/rdb/ent/predicate"
	"github.com/piprate/metalocker/storage/rdb/ent/property"
	"github.com/piprate/metalocker/storage/rdb/ent/recoverycode"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
)

const (
	// Operation types.
	OpCreate    = ent.OpCreate
	OpDelete    = ent.OpDelete
	OpDeleteOne = ent.OpDeleteOne
	OpUpdate    = ent.OpUpdate
	OpUpdateOne = ent.OpUpdateOne

	// Node types.
	TypeAccessKey    = "AccessKey"
	TypeAccount      = "Account"
	TypeDID          = "DID"
	TypeIdentity     = "Identity"
	TypeLocker       = "Locker"
	TypeProperty     = "Property"
	TypeRecoveryCode = "RecoveryCode"
)

// AccessKeyMutation represents an operation that mutates the AccessKey nodes in the graph.
type AccessKeyMutation struct {
	config
	op             Op
	typ            string
	id             *int
	did            *string
	body           **model.AccessKey
	clearedFields  map[string]struct{}
	account        *int
	clearedaccount bool
	done           bool
	oldValue       func(context.Context) (*AccessKey, error)
	predicates     []predicate.AccessKey
}

var _ ent.Mutation = (*AccessKeyMutation)(nil)

// accesskeyOption allows management of the mutation configuration using functional options.
type accesskeyOption func(*AccessKeyMutation)

// newAccessKeyMutation creates new mutation for the AccessKey entity.
func newAccessKeyMutation(c config, op Op, opts ...accesskeyOption) *AccessKeyMutation {
	m := &AccessKeyMutation{
		config:        c,
		op:            op,
		typ:           TypeAccessKey,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withAccessKeyID sets the ID field of the mutation.
func withAccessKeyID(id int) accesskeyOption {
	return func(m *AccessKeyMutation) {
		var (
			err   error
			once  sync.Once
			value *AccessKey
		)
		m.oldValue = func(ctx context.Context) (*AccessKey, error) {
			once.Do(func() {
				if m.done {
					err = errors.New("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().AccessKey.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withAccessKey sets the old AccessKey of the mutation.
func withAccessKey(node *AccessKey) accesskeyOption {
	return func(m *AccessKeyMutation) {
		m.oldValue = func(context.Context) (*AccessKey, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m AccessKeyMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m AccessKeyMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, errors.New("ent: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// ID returns the ID value in the mutation. Note that the ID is only available
// if it was provided to the builder or after it was returned from the database.
func (m *AccessKeyMutation) ID() (id int, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// IDs queries the database and returns the entity ids that match the mutation's predicate.
// That means, if the mutation is applied within a transaction with an isolation level such
// as sql.LevelSerializable, the returned ids match the ids of the rows that will be updated
// or updated by the mutation.
func (m *AccessKeyMutation) IDs(ctx context.Context) ([]int, error) {
	switch {
	case m.op.Is(OpUpdateOne | OpDeleteOne):
		id, exists := m.ID()
		if exists {
			return []int{id}, nil
		}
		fallthrough
	case m.op.Is(OpUpdate | OpDelete):
		return m.Client().AccessKey.Query().Where(m.predicates...).IDs(ctx)
	default:
		return nil, fmt.Errorf("IDs is not allowed on %s operations", m.op)
	}
}

// SetDid sets the "did" field.
func (m *AccessKeyMutation) SetDid(s string) {
	m.did = &s
}

// Did returns the value of the "did" field in the mutation.
func (m *AccessKeyMutation) Did() (r string, exists bool) {
	v := m.did
	if v == nil {
		return
	}
	return *v, true
}

// OldDid returns the old "did" field's value of the AccessKey entity.
// If the AccessKey object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *AccessKeyMutation) OldDid(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldDid is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldDid requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldDid: %w", err)
	}
	return oldValue.Did, nil
}

// ResetDid resets all changes to the "did" field.
func (m *AccessKeyMutation) ResetDid() {
	m.did = nil
}

// SetBody sets the "body" field.
func (m *AccessKeyMutation) SetBody(mk *model.AccessKey) {
	m.body = &mk
}

// Body returns the value of the "body" field in the mutation.
func (m *AccessKeyMutation) Body() (r *model.AccessKey, exists bool) {
	v := m.body
	if v == nil {
		return
	}
	return *v, true
}

// OldBody returns the old "body" field's value of the AccessKey entity.
// If the AccessKey object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *AccessKeyMutation) OldBody(ctx context.Context) (v *model.AccessKey, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldBody is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldBody requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldBody: %w", err)
	}
	return oldValue.Body, nil
}

// ResetBody resets all changes to the "body" field.
func (m *AccessKeyMutation) ResetBody() {
	m.body = nil
}

// SetAccountID sets the "account" edge to the Account entity by id.
func (m *AccessKeyMutation) SetAccountID(id int) {
	m.account = &id
}

// ClearAccount clears the "account" edge to the Account entity.
func (m *AccessKeyMutation) ClearAccount() {
	m.clearedaccount = true
}

// AccountCleared reports if the "account" edge to the Account entity was cleared.
func (m *AccessKeyMutation) AccountCleared() bool {
	return m.clearedaccount
}

// AccountID returns the "account" edge ID in the mutation.
func (m *AccessKeyMutation) AccountID() (id int, exists bool) {
	if m.account != nil {
		return *m.account, true
	}
	return
}

// AccountIDs returns the "account" edge IDs in the mutation.
// Note that IDs always returns len(IDs) <= 1 for unique edges, and you should use
// AccountID instead. It exists only for internal usage by the builders.
func (m *AccessKeyMutation) AccountIDs() (ids []int) {
	if id := m.account; id != nil {
		ids = append(ids, *id)
	}
	return
}

// ResetAccount resets all changes to the "account" edge.
func (m *AccessKeyMutation) ResetAccount() {
	m.account = nil
	m.clearedaccount = false
}

// Where appends a list predicates to the AccessKeyMutation builder.
func (m *AccessKeyMutation) Where(ps ...predicate.AccessKey) {
	m.predicates = append(m.predicates, ps...)
}

// WhereP appends storage-level predicates to the AccessKeyMutation builder. Using this method,
// users can use type-assertion to append predicates that do not depend on any generated package.
func (m *AccessKeyMutation) WhereP(ps ...func(*sql.Selector)) {
	p := make([]predicate.AccessKey, len(ps))
	for i := range ps {
		p[i] = ps[i]
	}
	m.Where(p...)
}

// Op returns the operation name.
func (m *AccessKeyMutation) Op() Op {
	return m.op
}

// SetOp allows setting the mutation operation.
func (m *AccessKeyMutation) SetOp(op Op) {
	m.op = op
}

// Type returns the node type of this mutation (AccessKey).
func (m *AccessKeyMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *AccessKeyMutation) Fields() []string {
	fields := make([]string, 0, 2)
	if m.did != nil {
		fields = append(fields, accesskey.FieldDid)
	}
	if m.body != nil {
		fields = append(fields, accesskey.FieldBody)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *AccessKeyMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case accesskey.FieldDid:
		return m.Did()
	case accesskey.FieldBody:
		return m.Body()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *AccessKeyMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case accesskey.FieldDid:
		return m.OldDid(ctx)
	case accesskey.FieldBody:
		return m.OldBody(ctx)
	}
	return nil, fmt.Errorf("unknown AccessKey field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *AccessKeyMutation) SetField(name string, value ent.Value) error {
	switch name {
	case accesskey.FieldDid:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetDid(v)
		return nil
	case accesskey.FieldBody:
		v, ok := value.(*model.AccessKey)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetBody(v)
		return nil
	}
	return fmt.Errorf("unknown AccessKey field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *AccessKeyMutation) AddedFields() []string {
	return nil
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *AccessKeyMutation) AddedField(name string) (ent.Value, bool) {
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *AccessKeyMutation) AddField(name string, value ent.Value) error {
	switch name {
	}
	return fmt.Errorf("unknown AccessKey numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *AccessKeyMutation) ClearedFields() []string {
	return nil
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *AccessKeyMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *AccessKeyMutation) ClearField(name string) error {
	return fmt.Errorf("unknown AccessKey nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *AccessKeyMutation) ResetField(name string) error {
	switch name {
	case accesskey.FieldDid:
		m.ResetDid()
		return nil
	case accesskey.FieldBody:
		m.ResetBody()
		return nil
	}
	return fmt.Errorf("unknown AccessKey field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *AccessKeyMutation) AddedEdges() []string {
	edges := make([]string, 0, 1)
	if m.account != nil {
		edges = append(edges, accesskey.EdgeAccount)
	}
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *AccessKeyMutation) AddedIDs(name string) []ent.Value {
	switch name {
	case accesskey.EdgeAccount:
		if id := m.account; id != nil {
			return []ent.Value{*id}
		}
	}
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *AccessKeyMutation) RemovedEdges() []string {
	edges := make([]string, 0, 1)
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *AccessKeyMutation) RemovedIDs(name string) []ent.Value {
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *AccessKeyMutation) ClearedEdges() []string {
	edges := make([]string, 0, 1)
	if m.clearedaccount {
		edges = append(edges, accesskey.EdgeAccount)
	}
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *AccessKeyMutation) EdgeCleared(name string) bool {
	switch name {
	case accesskey.EdgeAccount:
		return m.clearedaccount
	}
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *AccessKeyMutation) ClearEdge(name string) error {
	switch name {
	case accesskey.EdgeAccount:
		m.ClearAccount()
		return nil
	}
	return fmt.Errorf("unknown AccessKey unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *AccessKeyMutation) ResetEdge(name string) error {
	switch name {
	case accesskey.EdgeAccount:
		m.ResetAccount()
		return nil
	}
	return fmt.Errorf("unknown AccessKey edge %s", name)
}

// AccountMutation represents an operation that mutates the Account nodes in the graph.
type AccountMutation struct {
	config
	op                    Op
	typ                   string
	id                    *int
	did                   *string
	state                 *string
	email                 *string
	parent_account        *string
	body                  **account.Account
	clearedFields         map[string]struct{}
	recovery_codes        map[int]struct{}
	removedrecovery_codes map[int]struct{}
	clearedrecovery_codes bool
	access_keys           map[int]struct{}
	removedaccess_keys    map[int]struct{}
	clearedaccess_keys    bool
	identities            map[int]struct{}
	removedidentities     map[int]struct{}
	clearedidentities     bool
	lockers               map[int]struct{}
	removedlockers        map[int]struct{}
	clearedlockers        bool
	properties            map[int]struct{}
	removedproperties     map[int]struct{}
	clearedproperties     bool
	done                  bool
	oldValue              func(context.Context) (*Account, error)
	predicates            []predicate.Account
}

var _ ent.Mutation = (*AccountMutation)(nil)

// accountOption allows management of the mutation configuration using functional options.
type accountOption func(*AccountMutation)

// newAccountMutation creates new mutation for the Account entity.
func newAccountMutation(c config, op Op, opts ...accountOption) *AccountMutation {
	m := &AccountMutation{
		config:        c,
		op:            op,
		typ:           TypeAccount,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withAccountID sets the ID field of the mutation.
func withAccountID(id int) accountOption {
	return func(m *AccountMutation) {
		var (
			err   error
			once  sync.Once
			value *Account
		)
		m.oldValue = func(ctx context.Context) (*Account, error) {
			once.Do(func() {
				if m.done {
					err = errors.New("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().Account.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withAccount sets the old Account of the mutation.
func withAccount(node *Account) accountOption {
	return func(m *AccountMutation) {
		m.oldValue = func(context.Context) (*Account, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m AccountMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m AccountMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, errors.New("ent: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// ID returns the ID value in the mutation. Note that the ID is only available
// if it was provided to the builder or after it was returned from the database.
func (m *AccountMutation) ID() (id int, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// IDs queries the database and returns the entity ids that match the mutation's predicate.
// That means, if the mutation is applied within a transaction with an isolation level such
// as sql.LevelSerializable, the returned ids match the ids of the rows that will be updated
// or updated by the mutation.
func (m *AccountMutation) IDs(ctx context.Context) ([]int, error) {
	switch {
	case m.op.Is(OpUpdateOne | OpDeleteOne):
		id, exists := m.ID()
		if exists {
			return []int{id}, nil
		}
		fallthrough
	case m.op.Is(OpUpdate | OpDelete):
		return m.Client().Account.Query().Where(m.predicates...).IDs(ctx)
	default:
		return nil, fmt.Errorf("IDs is not allowed on %s operations", m.op)
	}
}

// SetDid sets the "did" field.
func (m *AccountMutation) SetDid(s string) {
	m.did = &s
}

// Did returns the value of the "did" field in the mutation.
func (m *AccountMutation) Did() (r string, exists bool) {
	v := m.did
	if v == nil {
		return
	}
	return *v, true
}

// OldDid returns the old "did" field's value of the Account entity.
// If the Account object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *AccountMutation) OldDid(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldDid is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldDid requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldDid: %w", err)
	}
	return oldValue.Did, nil
}

// ResetDid resets all changes to the "did" field.
func (m *AccountMutation) ResetDid() {
	m.did = nil
}

// SetState sets the "state" field.
func (m *AccountMutation) SetState(s string) {
	m.state = &s
}

// State returns the value of the "state" field in the mutation.
func (m *AccountMutation) State() (r string, exists bool) {
	v := m.state
	if v == nil {
		return
	}
	return *v, true
}

// OldState returns the old "state" field's value of the Account entity.
// If the Account object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *AccountMutation) OldState(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldState is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldState requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldState: %w", err)
	}
	return oldValue.State, nil
}

// ResetState resets all changes to the "state" field.
func (m *AccountMutation) ResetState() {
	m.state = nil
}

// SetEmail sets the "email" field.
func (m *AccountMutation) SetEmail(s string) {
	m.email = &s
}

// Email returns the value of the "email" field in the mutation.
func (m *AccountMutation) Email() (r string, exists bool) {
	v := m.email
	if v == nil {
		return
	}
	return *v, true
}

// OldEmail returns the old "email" field's value of the Account entity.
// If the Account object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *AccountMutation) OldEmail(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldEmail is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldEmail requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldEmail: %w", err)
	}
	return oldValue.Email, nil
}

// ClearEmail clears the value of the "email" field.
func (m *AccountMutation) ClearEmail() {
	m.email = nil
	m.clearedFields[entaccount.FieldEmail] = struct{}{}
}

// EmailCleared returns if the "email" field was cleared in this mutation.
func (m *AccountMutation) EmailCleared() bool {
	_, ok := m.clearedFields[entaccount.FieldEmail]
	return ok
}

// ResetEmail resets all changes to the "email" field.
func (m *AccountMutation) ResetEmail() {
	m.email = nil
	delete(m.clearedFields, entaccount.FieldEmail)
}

// SetParentAccount sets the "parent_account" field.
func (m *AccountMutation) SetParentAccount(s string) {
	m.parent_account = &s
}

// ParentAccount returns the value of the "parent_account" field in the mutation.
func (m *AccountMutation) ParentAccount() (r string, exists bool) {
	v := m.parent_account
	if v == nil {
		return
	}
	return *v, true
}

// OldParentAccount returns the old "parent_account" field's value of the Account entity.
// If the Account object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *AccountMutation) OldParentAccount(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldParentAccount is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldParentAccount requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldParentAccount: %w", err)
	}
	return oldValue.ParentAccount, nil
}

// ClearParentAccount clears the value of the "parent_account" field.
func (m *AccountMutation) ClearParentAccount() {
	m.parent_account = nil
	m.clearedFields[entaccount.FieldParentAccount] = struct{}{}
}

// ParentAccountCleared returns if the "parent_account" field was cleared in this mutation.
func (m *AccountMutation) ParentAccountCleared() bool {
	_, ok := m.clearedFields[entaccount.FieldParentAccount]
	return ok
}

// ResetParentAccount resets all changes to the "parent_account" field.
func (m *AccountMutation) ResetParentAccount() {
	m.parent_account = nil
	delete(m.clearedFields, entaccount.FieldParentAccount)
}

// SetBody sets the "body" field.
func (m *AccountMutation) SetBody(a *account.Account) {
	m.body = &a
}

// Body returns the value of the "body" field in the mutation.
func (m *AccountMutation) Body() (r *account.Account, exists bool) {
	v := m.body
	if v == nil {
		return
	}
	return *v, true
}

// OldBody returns the old "body" field's value of the Account entity.
// If the Account object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *AccountMutation) OldBody(ctx context.Context) (v *account.Account, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldBody is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldBody requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldBody: %w", err)
	}
	return oldValue.Body, nil
}

// ResetBody resets all changes to the "body" field.
func (m *AccountMutation) ResetBody() {
	m.body = nil
}

// AddRecoveryCodeIDs adds the "recovery_codes" edge to the RecoveryCode entity by ids.
func (m *AccountMutation) AddRecoveryCodeIDs(ids ...int) {
	if m.recovery_codes == nil {
		m.recovery_codes = make(map[int]struct{})
	}
	for i := range ids {
		m.recovery_codes[ids[i]] = struct{}{}
	}
}

// ClearRecoveryCodes clears the "recovery_codes" edge to the RecoveryCode entity.
func (m *AccountMutation) ClearRecoveryCodes() {
	m.clearedrecovery_codes = true
}

// RecoveryCodesCleared reports if the "recovery_codes" edge to the RecoveryCode entity was cleared.
func (m *AccountMutation) RecoveryCodesCleared() bool {
	return m.clearedrecovery_codes
}

// RemoveRecoveryCodeIDs removes the "recovery_codes" edge to the RecoveryCode entity by IDs.
func (m *AccountMutation) RemoveRecoveryCodeIDs(ids ...int) {
	if m.removedrecovery_codes == nil {
		m.removedrecovery_codes = make(map[int]struct{})
	}
	for i := range ids {
		delete(m.recovery_codes, ids[i])
		m.removedrecovery_codes[ids[i]] = struct{}{}
	}
}

// RemovedRecoveryCodes returns the removed IDs of the "recovery_codes" edge to the RecoveryCode entity.
func (m *AccountMutation) RemovedRecoveryCodesIDs() (ids []int) {
	for id := range m.removedrecovery_codes {
		ids = append(ids, id)
	}
	return
}

// RecoveryCodesIDs returns the "recovery_codes" edge IDs in the mutation.
func (m *AccountMutation) RecoveryCodesIDs() (ids []int) {
	for id := range m.recovery_codes {
		ids = append(ids, id)
	}
	return
}

// ResetRecoveryCodes resets all changes to the "recovery_codes" edge.
func (m *AccountMutation) ResetRecoveryCodes() {
	m.recovery_codes = nil
	m.clearedrecovery_codes = false
	m.removedrecovery_codes = nil
}

// AddAccessKeyIDs adds the "access_keys" edge to the AccessKey entity by ids.
func (m *AccountMutation) AddAccessKeyIDs(ids ...int) {
	if m.access_keys == nil {
		m.access_keys = make(map[int]struct{})
	}
	for i := range ids {
		m.access_keys[ids[i]] = struct{}{}
	}
}

// ClearAccessKeys clears the "access_keys" edge to the AccessKey entity.
func (m *AccountMutation) ClearAccessKeys() {
	m.clearedaccess_keys = true
}

// AccessKeysCleared reports if the "access_keys" edge to the AccessKey entity was cleared.
func (m *AccountMutation) AccessKeysCleared() bool {
	return m.clearedaccess_keys
}

// RemoveAccessKeyIDs removes the "access_keys" edge to the AccessKey entity by IDs.
func (m *AccountMutation) RemoveAccessKeyIDs(ids ...int) {
	if m.removedaccess_keys == nil {
		m.removedaccess_keys = make(map[int]struct{})
	}
	for i := range ids {
		delete(m.access_keys, ids[i])
		m.removedaccess_keys[ids[i]] = struct{}{}
	}
}

// RemovedAccessKeys returns the removed IDs of the "access_keys" edge to the AccessKey entity.
func (m *AccountMutation) RemovedAccessKeysIDs() (ids []int) {
	for id := range m.removedaccess_keys {
		ids = append(ids, id)
	}
	return
}

// AccessKeysIDs returns the "access_keys" edge IDs in the mutation.
func (m *AccountMutation) AccessKeysIDs() (ids []int) {
	for id := range m.access_keys {
		ids = append(ids, id)
	}
	return
}

// ResetAccessKeys resets all changes to the "access_keys" edge.
func (m *AccountMutation) ResetAccessKeys() {
	m.access_keys = nil
	m.clearedaccess_keys = false
	m.removedaccess_keys = nil
}

// AddIdentityIDs adds the "identities" edge to the Identity entity by ids.
func (m *AccountMutation) AddIdentityIDs(ids ...int) {
	if m.identities == nil {
		m.identities = make(map[int]struct{})
	}
	for i := range ids {
		m.identities[ids[i]] = struct{}{}
	}
}

// ClearIdentities clears the "identities" edge to the Identity entity.
func (m *AccountMutation) ClearIdentities() {
	m.clearedidentities = true
}

// IdentitiesCleared reports if the "identities" edge to the Identity entity was cleared.
func (m *AccountMutation) IdentitiesCleared() bool {
	return m.clearedidentities
}

// RemoveIdentityIDs removes the "identities" edge to the Identity entity by IDs.
func (m *AccountMutation) RemoveIdentityIDs(ids ...int) {
	if m.removedidentities == nil {
		m.removedidentities = make(map[int]struct{})
	}
	for i := range ids {
		delete(m.identities, ids[i])
		m.removedidentities[ids[i]] = struct{}{}
	}
}

// RemovedIdentities returns the removed IDs of the "identities" edge to the Identity entity.
func (m *AccountMutation) RemovedIdentitiesIDs() (ids []int) {
	for id := range m.removedidentities {
		ids = append(ids, id)
	}
	return
}

// IdentitiesIDs returns the "identities" edge IDs in the mutation.
func (m *AccountMutation) IdentitiesIDs() (ids []int) {
	for id := range m.identities {
		ids = append(ids, id)
	}
	return
}

// ResetIdentities resets all changes to the "identities" edge.
func (m *AccountMutation) ResetIdentities() {
	m.identities = nil
	m.clearedidentities = false
	m.removedidentities = nil
}

// AddLockerIDs adds the "lockers" edge to the Locker entity by ids.
func (m *AccountMutation) AddLockerIDs(ids ...int) {
	if m.lockers == nil {
		m.lockers = make(map[int]struct{})
	}
	for i := range ids {
		m.lockers[ids[i]] = struct{}{}
	}
}

// ClearLockers clears the "lockers" edge to the Locker entity.
func (m *AccountMutation) ClearLockers() {
	m.clearedlockers = true
}

// LockersCleared reports if the "lockers" edge to the Locker entity was cleared.
func (m *AccountMutation) LockersCleared() bool {
	return m.clearedlockers
}

// RemoveLockerIDs removes the "lockers" edge to the Locker entity by IDs.
func (m *AccountMutation) RemoveLockerIDs(ids ...int) {
	if m.removedlockers == nil {
		m.removedlockers = make(map[int]struct{})
	}
	for i := range ids {
		delete(m.lockers, ids[i])
		m.removedlockers[ids[i]] = struct{}{}
	}
}

// RemovedLockers returns the removed IDs of the "lockers" edge to the Locker entity.
func (m *AccountMutation) RemovedLockersIDs() (ids []int) {
	for id := range m.removedlockers {
		ids = append(ids, id)
	}
	return
}

// LockersIDs returns the "lockers" edge IDs in the mutation.
func (m *AccountMutation) LockersIDs() (ids []int) {
	for id := range m.lockers {
		ids = append(ids, id)
	}
	return
}

// ResetLockers resets all changes to the "lockers" edge.
func (m *AccountMutation) ResetLockers() {
	m.lockers = nil
	m.clearedlockers = false
	m.removedlockers = nil
}

// AddPropertyIDs adds the "properties" edge to the Property entity by ids.
func (m *AccountMutation) AddPropertyIDs(ids ...int) {
	if m.properties == nil {
		m.properties = make(map[int]struct{})
	}
	for i := range ids {
		m.properties[ids[i]] = struct{}{}
	}
}

// ClearProperties clears the "properties" edge to the Property entity.
func (m *AccountMutation) ClearProperties() {
	m.clearedproperties = true
}

// PropertiesCleared reports if the "properties" edge to the Property entity was cleared.
func (m *AccountMutation) PropertiesCleared() bool {
	return m.clearedproperties
}

// RemovePropertyIDs removes the "properties" edge to the Property entity by IDs.
func (m *AccountMutation) RemovePropertyIDs(ids ...int) {
	if m.removedproperties == nil {
		m.removedproperties = make(map[int]struct{})
	}
	for i := range ids {
		delete(m.properties, ids[i])
		m.removedproperties[ids[i]] = struct{}{}
	}
}

// RemovedProperties returns the removed IDs of the "properties" edge to the Property entity.
func (m *AccountMutation) RemovedPropertiesIDs() (ids []int) {
	for id := range m.removedproperties {
		ids = append(ids, id)
	}
	return
}

// PropertiesIDs returns the "properties" edge IDs in the mutation.
func (m *AccountMutation) PropertiesIDs() (ids []int) {
	for id := range m.properties {
		ids = append(ids, id)
	}
	return
}

// ResetProperties resets all changes to the "properties" edge.
func (m *AccountMutation) ResetProperties() {
	m.properties = nil
	m.clearedproperties = false
	m.removedproperties = nil
}

// Where appends a list predicates to the AccountMutation builder.
func (m *AccountMutation) Where(ps ...predicate.Account) {
	m.predicates = append(m.predicates, ps...)
}

// WhereP appends storage-level predicates to the AccountMutation builder. Using this method,
// users can use type-assertion to append predicates that do not depend on any generated package.
func (m *AccountMutation) WhereP(ps ...func(*sql.Selector)) {
	p := make([]predicate.Account, len(ps))
	for i := range ps {
		p[i] = ps[i]
	}
	m.Where(p...)
}

// Op returns the operation name.
func (m *AccountMutation) Op() Op {
	return m.op
}

// SetOp allows setting the mutation operation.
func (m *AccountMutation) SetOp(op Op) {
	m.op = op
}

// Type returns the node type of this mutation (Account).
func (m *AccountMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *AccountMutation) Fields() []string {
	fields := make([]string, 0, 5)
	if m.did != nil {
		fields = append(fields, entaccount.FieldDid)
	}
	if m.state != nil {
		fields = append(fields, entaccount.FieldState)
	}
	if m.email != nil {
		fields = append(fields, entaccount.FieldEmail)
	}
	if m.parent_account != nil {
		fields = append(fields, entaccount.FieldParentAccount)
	}
	if m.body != nil {
		fields = append(fields, entaccount.FieldBody)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *AccountMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case entaccount.FieldDid:
		return m.Did()
	case entaccount.FieldState:
		return m.State()
	case entaccount.FieldEmail:
		return m.Email()
	case entaccount.FieldParentAccount:
		return m.ParentAccount()
	case entaccount.FieldBody:
		return m.Body()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *AccountMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case entaccount.FieldDid:
		return m.OldDid(ctx)
	case entaccount.FieldState:
		return m.OldState(ctx)
	case entaccount.FieldEmail:
		return m.OldEmail(ctx)
	case entaccount.FieldParentAccount:
		return m.OldParentAccount(ctx)
	case entaccount.FieldBody:
		return m.OldBody(ctx)
	}
	return nil, fmt.Errorf("unknown Account field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *AccountMutation) SetField(name string, value ent.Value) error {
	switch name {
	case entaccount.FieldDid:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetDid(v)
		return nil
	case entaccount.FieldState:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetState(v)
		return nil
	case entaccount.FieldEmail:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetEmail(v)
		return nil
	case entaccount.FieldParentAccount:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetParentAccount(v)
		return nil
	case entaccount.FieldBody:
		v, ok := value.(*account.Account)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetBody(v)
		return nil
	}
	return fmt.Errorf("unknown Account field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *AccountMutation) AddedFields() []string {
	return nil
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *AccountMutation) AddedField(name string) (ent.Value, bool) {
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *AccountMutation) AddField(name string, value ent.Value) error {
	switch name {
	}
	return fmt.Errorf("unknown Account numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *AccountMutation) ClearedFields() []string {
	var fields []string
	if m.FieldCleared(entaccount.FieldEmail) {
		fields = append(fields, entaccount.FieldEmail)
	}
	if m.FieldCleared(entaccount.FieldParentAccount) {
		fields = append(fields, entaccount.FieldParentAccount)
	}
	return fields
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *AccountMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *AccountMutation) ClearField(name string) error {
	switch name {
	case entaccount.FieldEmail:
		m.ClearEmail()
		return nil
	case entaccount.FieldParentAccount:
		m.ClearParentAccount()
		return nil
	}
	return fmt.Errorf("unknown Account nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *AccountMutation) ResetField(name string) error {
	switch name {
	case entaccount.FieldDid:
		m.ResetDid()
		return nil
	case entaccount.FieldState:
		m.ResetState()
		return nil
	case entaccount.FieldEmail:
		m.ResetEmail()
		return nil
	case entaccount.FieldParentAccount:
		m.ResetParentAccount()
		return nil
	case entaccount.FieldBody:
		m.ResetBody()
		return nil
	}
	return fmt.Errorf("unknown Account field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *AccountMutation) AddedEdges() []string {
	edges := make([]string, 0, 5)
	if m.recovery_codes != nil {
		edges = append(edges, entaccount.EdgeRecoveryCodes)
	}
	if m.access_keys != nil {
		edges = append(edges, entaccount.EdgeAccessKeys)
	}
	if m.identities != nil {
		edges = append(edges, entaccount.EdgeIdentities)
	}
	if m.lockers != nil {
		edges = append(edges, entaccount.EdgeLockers)
	}
	if m.properties != nil {
		edges = append(edges, entaccount.EdgeProperties)
	}
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *AccountMutation) AddedIDs(name string) []ent.Value {
	switch name {
	case entaccount.EdgeRecoveryCodes:
		ids := make([]ent.Value, 0, len(m.recovery_codes))
		for id := range m.recovery_codes {
			ids = append(ids, id)
		}
		return ids
	case entaccount.EdgeAccessKeys:
		ids := make([]ent.Value, 0, len(m.access_keys))
		for id := range m.access_keys {
			ids = append(ids, id)
		}
		return ids
	case entaccount.EdgeIdentities:
		ids := make([]ent.Value, 0, len(m.identities))
		for id := range m.identities {
			ids = append(ids, id)
		}
		return ids
	case entaccount.EdgeLockers:
		ids := make([]ent.Value, 0, len(m.lockers))
		for id := range m.lockers {
			ids = append(ids, id)
		}
		return ids
	case entaccount.EdgeProperties:
		ids := make([]ent.Value, 0, len(m.properties))
		for id := range m.properties {
			ids = append(ids, id)
		}
		return ids
	}
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *AccountMutation) RemovedEdges() []string {
	edges := make([]string, 0, 5)
	if m.removedrecovery_codes != nil {
		edges = append(edges, entaccount.EdgeRecoveryCodes)
	}
	if m.removedaccess_keys != nil {
		edges = append(edges, entaccount.EdgeAccessKeys)
	}
	if m.removedidentities != nil {
		edges = append(edges, entaccount.EdgeIdentities)
	}
	if m.removedlockers != nil {
		edges = append(edges, entaccount.EdgeLockers)
	}
	if m.removedproperties != nil {
		edges = append(edges, entaccount.EdgeProperties)
	}
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *AccountMutation) RemovedIDs(name string) []ent.Value {
	switch name {
	case entaccount.EdgeRecoveryCodes:
		ids := make([]ent.Value, 0, len(m.removedrecovery_codes))
		for id := range m.removedrecovery_codes {
			ids = append(ids, id)
		}
		return ids
	case entaccount.EdgeAccessKeys:
		ids := make([]ent.Value, 0, len(m.removedaccess_keys))
		for id := range m.removedaccess_keys {
			ids = append(ids, id)
		}
		return ids
	case entaccount.EdgeIdentities:
		ids := make([]ent.Value, 0, len(m.removedidentities))
		for id := range m.removedidentities {
			ids = append(ids, id)
		}
		return ids
	case entaccount.EdgeLockers:
		ids := make([]ent.Value, 0, len(m.removedlockers))
		for id := range m.removedlockers {
			ids = append(ids, id)
		}
		return ids
	case entaccount.EdgeProperties:
		ids := make([]ent.Value, 0, len(m.removedproperties))
		for id := range m.removedproperties {
			ids = append(ids, id)
		}
		return ids
	}
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *AccountMutation) ClearedEdges() []string {
	edges := make([]string, 0, 5)
	if m.clearedrecovery_codes {
		edges = append(edges, entaccount.EdgeRecoveryCodes)
	}
	if m.clearedaccess_keys {
		edges = append(edges, entaccount.EdgeAccessKeys)
	}
	if m.clearedidentities {
		edges = append(edges, entaccount.EdgeIdentities)
	}
	if m.clearedlockers {
		edges = append(edges, entaccount.EdgeLockers)
	}
	if m.clearedproperties {
		edges = append(edges, entaccount.EdgeProperties)
	}
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *AccountMutation) EdgeCleared(name string) bool {
	switch name {
	case entaccount.EdgeRecoveryCodes:
		return m.clearedrecovery_codes
	case entaccount.EdgeAccessKeys:
		return m.clearedaccess_keys
	case entaccount.EdgeIdentities:
		return m.clearedidentities
	case entaccount.EdgeLockers:
		return m.clearedlockers
	case entaccount.EdgeProperties:
		return m.clearedproperties
	}
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *AccountMutation) ClearEdge(name string) error {
	switch name {
	}
	return fmt.Errorf("unknown Account unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *AccountMutation) ResetEdge(name string) error {
	switch name {
	case entaccount.EdgeRecoveryCodes:
		m.ResetRecoveryCodes()
		return nil
	case entaccount.EdgeAccessKeys:
		m.ResetAccessKeys()
		return nil
	case entaccount.EdgeIdentities:
		m.ResetIdentities()
		return nil
	case entaccount.EdgeLockers:
		m.ResetLockers()
		return nil
	case entaccount.EdgeProperties:
		m.ResetProperties()
		return nil
	}
	return fmt.Errorf("unknown Account edge %s", name)
}

// DIDMutation represents an operation that mutates the DID nodes in the graph.
type DIDMutation struct {
	config
	op            Op
	typ           string
	id            *int
	did           *string
	body          **model.DIDDocument
	clearedFields map[string]struct{}
	done          bool
	oldValue      func(context.Context) (*DID, error)
	predicates    []predicate.DID
}

var _ ent.Mutation = (*DIDMutation)(nil)

// didOption allows management of the mutation configuration using functional options.
type didOption func(*DIDMutation)

// newDIDMutation creates new mutation for the DID entity.
func newDIDMutation(c config, op Op, opts ...didOption) *DIDMutation {
	m := &DIDMutation{
		config:        c,
		op:            op,
		typ:           TypeDID,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withDIDID sets the ID field of the mutation.
func withDIDID(id int) didOption {
	return func(m *DIDMutation) {
		var (
			err   error
			once  sync.Once
			value *DID
		)
		m.oldValue = func(ctx context.Context) (*DID, error) {
			once.Do(func() {
				if m.done {
					err = errors.New("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().DID.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withDID sets the old DID of the mutation.
func withDID(node *DID) didOption {
	return func(m *DIDMutation) {
		m.oldValue = func(context.Context) (*DID, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m DIDMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m DIDMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, errors.New("ent: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// ID returns the ID value in the mutation. Note that the ID is only available
// if it was provided to the builder or after it was returned from the database.
func (m *DIDMutation) ID() (id int, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// IDs queries the database and returns the entity ids that match the mutation's predicate.
// That means, if the mutation is applied within a transaction with an isolation level such
// as sql.LevelSerializable, the returned ids match the ids of the rows that will be updated
// or updated by the mutation.
func (m *DIDMutation) IDs(ctx context.Context) ([]int, error) {
	switch {
	case m.op.Is(OpUpdateOne | OpDeleteOne):
		id, exists := m.ID()
		if exists {
			return []int{id}, nil
		}
		fallthrough
	case m.op.Is(OpUpdate | OpDelete):
		return m.Client().DID.Query().Where(m.predicates...).IDs(ctx)
	default:
		return nil, fmt.Errorf("IDs is not allowed on %s operations", m.op)
	}
}

// SetDid sets the "did" field.
func (m *DIDMutation) SetDid(s string) {
	m.did = &s
}

// Did returns the value of the "did" field in the mutation.
func (m *DIDMutation) Did() (r string, exists bool) {
	v := m.did
	if v == nil {
		return
	}
	return *v, true
}

// OldDid returns the old "did" field's value of the DID entity.
// If the DID object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *DIDMutation) OldDid(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldDid is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldDid requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldDid: %w", err)
	}
	return oldValue.Did, nil
}

// ResetDid resets all changes to the "did" field.
func (m *DIDMutation) ResetDid() {
	m.did = nil
}

// SetBody sets the "body" field.
func (m *DIDMutation) SetBody(md *model.DIDDocument) {
	m.body = &md
}

// Body returns the value of the "body" field in the mutation.
func (m *DIDMutation) Body() (r *model.DIDDocument, exists bool) {
	v := m.body
	if v == nil {
		return
	}
	return *v, true
}

// OldBody returns the old "body" field's value of the DID entity.
// If the DID object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *DIDMutation) OldBody(ctx context.Context) (v *model.DIDDocument, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldBody is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldBody requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldBody: %w", err)
	}
	return oldValue.Body, nil
}

// ResetBody resets all changes to the "body" field.
func (m *DIDMutation) ResetBody() {
	m.body = nil
}

// Where appends a list predicates to the DIDMutation builder.
func (m *DIDMutation) Where(ps ...predicate.DID) {
	m.predicates = append(m.predicates, ps...)
}

// WhereP appends storage-level predicates to the DIDMutation builder. Using this method,
// users can use type-assertion to append predicates that do not depend on any generated package.
func (m *DIDMutation) WhereP(ps ...func(*sql.Selector)) {
	p := make([]predicate.DID, len(ps))
	for i := range ps {
		p[i] = ps[i]
	}
	m.Where(p...)
}

// Op returns the operation name.
func (m *DIDMutation) Op() Op {
	return m.op
}

// SetOp allows setting the mutation operation.
func (m *DIDMutation) SetOp(op Op) {
	m.op = op
}

// Type returns the node type of this mutation (DID).
func (m *DIDMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *DIDMutation) Fields() []string {
	fields := make([]string, 0, 2)
	if m.did != nil {
		fields = append(fields, did.FieldDid)
	}
	if m.body != nil {
		fields = append(fields, did.FieldBody)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *DIDMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case did.FieldDid:
		return m.Did()
	case did.FieldBody:
		return m.Body()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *DIDMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case did.FieldDid:
		return m.OldDid(ctx)
	case did.FieldBody:
		return m.OldBody(ctx)
	}
	return nil, fmt.Errorf("unknown DID field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *DIDMutation) SetField(name string, value ent.Value) error {
	switch name {
	case did.FieldDid:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetDid(v)
		return nil
	case did.FieldBody:
		v, ok := value.(*model.DIDDocument)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetBody(v)
		return nil
	}
	return fmt.Errorf("unknown DID field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *DIDMutation) AddedFields() []string {
	return nil
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *DIDMutation) AddedField(name string) (ent.Value, bool) {
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *DIDMutation) AddField(name string, value ent.Value) error {
	switch name {
	}
	return fmt.Errorf("unknown DID numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *DIDMutation) ClearedFields() []string {
	return nil
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *DIDMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *DIDMutation) ClearField(name string) error {
	return fmt.Errorf("unknown DID nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *DIDMutation) ResetField(name string) error {
	switch name {
	case did.FieldDid:
		m.ResetDid()
		return nil
	case did.FieldBody:
		m.ResetBody()
		return nil
	}
	return fmt.Errorf("unknown DID field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *DIDMutation) AddedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *DIDMutation) AddedIDs(name string) []ent.Value {
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *DIDMutation) RemovedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *DIDMutation) RemovedIDs(name string) []ent.Value {
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *DIDMutation) ClearedEdges() []string {
	edges := make([]string, 0, 0)
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *DIDMutation) EdgeCleared(name string) bool {
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *DIDMutation) ClearEdge(name string) error {
	return fmt.Errorf("unknown DID unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *DIDMutation) ResetEdge(name string) error {
	return fmt.Errorf("unknown DID edge %s", name)
}

// IdentityMutation represents an operation that mutates the Identity nodes in the graph.
type IdentityMutation struct {
	config
	op             Op
	typ            string
	id             *int
	hash           *string
	level          *int32
	addlevel       *int32
	encrypted_id   *string
	encrypted_body *string
	clearedFields  map[string]struct{}
	account        *int
	clearedaccount bool
	done           bool
	oldValue       func(context.Context) (*Identity, error)
	predicates     []predicate.Identity
}

var _ ent.Mutation = (*IdentityMutation)(nil)

// identityOption allows management of the mutation configuration using functional options.
type identityOption func(*IdentityMutation)

// newIdentityMutation creates new mutation for the Identity entity.
func newIdentityMutation(c config, op Op, opts ...identityOption) *IdentityMutation {
	m := &IdentityMutation{
		config:        c,
		op:            op,
		typ:           TypeIdentity,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withIdentityID sets the ID field of the mutation.
func withIdentityID(id int) identityOption {
	return func(m *IdentityMutation) {
		var (
			err   error
			once  sync.Once
			value *Identity
		)
		m.oldValue = func(ctx context.Context) (*Identity, error) {
			once.Do(func() {
				if m.done {
					err = errors.New("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().Identity.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withIdentity sets the old Identity of the mutation.
func withIdentity(node *Identity) identityOption {
	return func(m *IdentityMutation) {
		m.oldValue = func(context.Context) (*Identity, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m IdentityMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m IdentityMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, errors.New("ent: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// ID returns the ID value in the mutation. Note that the ID is only available
// if it was provided to the builder or after it was returned from the database.
func (m *IdentityMutation) ID() (id int, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// IDs queries the database and returns the entity ids that match the mutation's predicate.
// That means, if the mutation is applied within a transaction with an isolation level such
// as sql.LevelSerializable, the returned ids match the ids of the rows that will be updated
// or updated by the mutation.
func (m *IdentityMutation) IDs(ctx context.Context) ([]int, error) {
	switch {
	case m.op.Is(OpUpdateOne | OpDeleteOne):
		id, exists := m.ID()
		if exists {
			return []int{id}, nil
		}
		fallthrough
	case m.op.Is(OpUpdate | OpDelete):
		return m.Client().Identity.Query().Where(m.predicates...).IDs(ctx)
	default:
		return nil, fmt.Errorf("IDs is not allowed on %s operations", m.op)
	}
}

// SetHash sets the "hash" field.
func (m *IdentityMutation) SetHash(s string) {
	m.hash = &s
}

// Hash returns the value of the "hash" field in the mutation.
func (m *IdentityMutation) Hash() (r string, exists bool) {
	v := m.hash
	if v == nil {
		return
	}
	return *v, true
}

// OldHash returns the old "hash" field's value of the Identity entity.
// If the Identity object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *IdentityMutation) OldHash(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldHash is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldHash requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldHash: %w", err)
	}
	return oldValue.Hash, nil
}

// ResetHash resets all changes to the "hash" field.
func (m *IdentityMutation) ResetHash() {
	m.hash = nil
}

// SetLevel sets the "level" field.
func (m *IdentityMutation) SetLevel(i int32) {
	m.level = &i
	m.addlevel = nil
}

// Level returns the value of the "level" field in the mutation.
func (m *IdentityMutation) Level() (r int32, exists bool) {
	v := m.level
	if v == nil {
		return
	}
	return *v, true
}

// OldLevel returns the old "level" field's value of the Identity entity.
// If the Identity object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *IdentityMutation) OldLevel(ctx context.Context) (v int32, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldLevel is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldLevel requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldLevel: %w", err)
	}
	return oldValue.Level, nil
}

// AddLevel adds i to the "level" field.
func (m *IdentityMutation) AddLevel(i int32) {
	if m.addlevel != nil {
		*m.addlevel += i
	} else {
		m.addlevel = &i
	}
}

// AddedLevel returns the value that was added to the "level" field in this mutation.
func (m *IdentityMutation) AddedLevel() (r int32, exists bool) {
	v := m.addlevel
	if v == nil {
		return
	}
	return *v, true
}

// ResetLevel resets all changes to the "level" field.
func (m *IdentityMutation) ResetLevel() {
	m.level = nil
	m.addlevel = nil
}

// SetEncryptedID sets the "encrypted_id" field.
func (m *IdentityMutation) SetEncryptedID(s string) {
	m.encrypted_id = &s
}

// EncryptedID returns the value of the "encrypted_id" field in the mutation.
func (m *IdentityMutation) EncryptedID() (r string, exists bool) {
	v := m.encrypted_id
	if v == nil {
		return
	}
	return *v, true
}

// OldEncryptedID returns the old "encrypted_id" field's value of the Identity entity.
// If the Identity object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *IdentityMutation) OldEncryptedID(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldEncryptedID is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldEncryptedID requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldEncryptedID: %w", err)
	}
	return oldValue.EncryptedID, nil
}

// ResetEncryptedID resets all changes to the "encrypted_id" field.
func (m *IdentityMutation) ResetEncryptedID() {
	m.encrypted_id = nil
}

// SetEncryptedBody sets the "encrypted_body" field.
func (m *IdentityMutation) SetEncryptedBody(s string) {
	m.encrypted_body = &s
}

// EncryptedBody returns the value of the "encrypted_body" field in the mutation.
func (m *IdentityMutation) EncryptedBody() (r string, exists bool) {
	v := m.encrypted_body
	if v == nil {
		return
	}
	return *v, true
}

// OldEncryptedBody returns the old "encrypted_body" field's value of the Identity entity.
// If the Identity object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *IdentityMutation) OldEncryptedBody(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldEncryptedBody is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldEncryptedBody requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldEncryptedBody: %w", err)
	}
	return oldValue.EncryptedBody, nil
}

// ResetEncryptedBody resets all changes to the "encrypted_body" field.
func (m *IdentityMutation) ResetEncryptedBody() {
	m.encrypted_body = nil
}

// SetAccountID sets the "account" edge to the Account entity by id.
func (m *IdentityMutation) SetAccountID(id int) {
	m.account = &id
}

// ClearAccount clears the "account" edge to the Account entity.
func (m *IdentityMutation) ClearAccount() {
	m.clearedaccount = true
}

// AccountCleared reports if the "account" edge to the Account entity was cleared.
func (m *IdentityMutation) AccountCleared() bool {
	return m.clearedaccount
}

// AccountID returns the "account" edge ID in the mutation.
func (m *IdentityMutation) AccountID() (id int, exists bool) {
	if m.account != nil {
		return *m.account, true
	}
	return
}

// AccountIDs returns the "account" edge IDs in the mutation.
// Note that IDs always returns len(IDs) <= 1 for unique edges, and you should use
// AccountID instead. It exists only for internal usage by the builders.
func (m *IdentityMutation) AccountIDs() (ids []int) {
	if id := m.account; id != nil {
		ids = append(ids, *id)
	}
	return
}

// ResetAccount resets all changes to the "account" edge.
func (m *IdentityMutation) ResetAccount() {
	m.account = nil
	m.clearedaccount = false
}

// Where appends a list predicates to the IdentityMutation builder.
func (m *IdentityMutation) Where(ps ...predicate.Identity) {
	m.predicates = append(m.predicates, ps...)
}

// WhereP appends storage-level predicates to the IdentityMutation builder. Using this method,
// users can use type-assertion to append predicates that do not depend on any generated package.
func (m *IdentityMutation) WhereP(ps ...func(*sql.Selector)) {
	p := make([]predicate.Identity, len(ps))
	for i := range ps {
		p[i] = ps[i]
	}
	m.Where(p...)
}

// Op returns the operation name.
func (m *IdentityMutation) Op() Op {
	return m.op
}

// SetOp allows setting the mutation operation.
func (m *IdentityMutation) SetOp(op Op) {
	m.op = op
}

// Type returns the node type of this mutation (Identity).
func (m *IdentityMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *IdentityMutation) Fields() []string {
	fields := make([]string, 0, 4)
	if m.hash != nil {
		fields = append(fields, identity.FieldHash)
	}
	if m.level != nil {
		fields = append(fields, identity.FieldLevel)
	}
	if m.encrypted_id != nil {
		fields = append(fields, identity.FieldEncryptedID)
	}
	if m.encrypted_body != nil {
		fields = append(fields, identity.FieldEncryptedBody)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *IdentityMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case identity.FieldHash:
		return m.Hash()
	case identity.FieldLevel:
		return m.Level()
	case identity.FieldEncryptedID:
		return m.EncryptedID()
	case identity.FieldEncryptedBody:
		return m.EncryptedBody()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *IdentityMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case identity.FieldHash:
		return m.OldHash(ctx)
	case identity.FieldLevel:
		return m.OldLevel(ctx)
	case identity.FieldEncryptedID:
		return m.OldEncryptedID(ctx)
	case identity.FieldEncryptedBody:
		return m.OldEncryptedBody(ctx)
	}
	return nil, fmt.Errorf("unknown Identity field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *IdentityMutation) SetField(name string, value ent.Value) error {
	switch name {
	case identity.FieldHash:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetHash(v)
		return nil
	case identity.FieldLevel:
		v, ok := value.(int32)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetLevel(v)
		return nil
	case identity.FieldEncryptedID:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetEncryptedID(v)
		return nil
	case identity.FieldEncryptedBody:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetEncryptedBody(v)
		return nil
	}
	return fmt.Errorf("unknown Identity field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *IdentityMutation) AddedFields() []string {
	var fields []string
	if m.addlevel != nil {
		fields = append(fields, identity.FieldLevel)
	}
	return fields
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *IdentityMutation) AddedField(name string) (ent.Value, bool) {
	switch name {
	case identity.FieldLevel:
		return m.AddedLevel()
	}
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *IdentityMutation) AddField(name string, value ent.Value) error {
	switch name {
	case identity.FieldLevel:
		v, ok := value.(int32)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.AddLevel(v)
		return nil
	}
	return fmt.Errorf("unknown Identity numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *IdentityMutation) ClearedFields() []string {
	return nil
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *IdentityMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *IdentityMutation) ClearField(name string) error {
	return fmt.Errorf("unknown Identity nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *IdentityMutation) ResetField(name string) error {
	switch name {
	case identity.FieldHash:
		m.ResetHash()
		return nil
	case identity.FieldLevel:
		m.ResetLevel()
		return nil
	case identity.FieldEncryptedID:
		m.ResetEncryptedID()
		return nil
	case identity.FieldEncryptedBody:
		m.ResetEncryptedBody()
		return nil
	}
	return fmt.Errorf("unknown Identity field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *IdentityMutation) AddedEdges() []string {
	edges := make([]string, 0, 1)
	if m.account != nil {
		edges = append(edges, identity.EdgeAccount)
	}
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *IdentityMutation) AddedIDs(name string) []ent.Value {
	switch name {
	case identity.EdgeAccount:
		if id := m.account; id != nil {
			return []ent.Value{*id}
		}
	}
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *IdentityMutation) RemovedEdges() []string {
	edges := make([]string, 0, 1)
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *IdentityMutation) RemovedIDs(name string) []ent.Value {
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *IdentityMutation) ClearedEdges() []string {
	edges := make([]string, 0, 1)
	if m.clearedaccount {
		edges = append(edges, identity.EdgeAccount)
	}
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *IdentityMutation) EdgeCleared(name string) bool {
	switch name {
	case identity.EdgeAccount:
		return m.clearedaccount
	}
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *IdentityMutation) ClearEdge(name string) error {
	switch name {
	case identity.EdgeAccount:
		m.ClearAccount()
		return nil
	}
	return fmt.Errorf("unknown Identity unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *IdentityMutation) ResetEdge(name string) error {
	switch name {
	case identity.EdgeAccount:
		m.ResetAccount()
		return nil
	}
	return fmt.Errorf("unknown Identity edge %s", name)
}

// LockerMutation represents an operation that mutates the Locker nodes in the graph.
type LockerMutation struct {
	config
	op             Op
	typ            string
	id             *int
	hash           *string
	level          *int32
	addlevel       *int32
	encrypted_id   *string
	encrypted_body *string
	clearedFields  map[string]struct{}
	account        *int
	clearedaccount bool
	done           bool
	oldValue       func(context.Context) (*Locker, error)
	predicates     []predicate.Locker
}

var _ ent.Mutation = (*LockerMutation)(nil)

// lockerOption allows management of the mutation configuration using functional options.
type lockerOption func(*LockerMutation)

// newLockerMutation creates new mutation for the Locker entity.
func newLockerMutation(c config, op Op, opts ...lockerOption) *LockerMutation {
	m := &LockerMutation{
		config:        c,
		op:            op,
		typ:           TypeLocker,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withLockerID sets the ID field of the mutation.
func withLockerID(id int) lockerOption {
	return func(m *LockerMutation) {
		var (
			err   error
			once  sync.Once
			value *Locker
		)
		m.oldValue = func(ctx context.Context) (*Locker, error) {
			once.Do(func() {
				if m.done {
					err = errors.New("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().Locker.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withLocker sets the old Locker of the mutation.
func withLocker(node *Locker) lockerOption {
	return func(m *LockerMutation) {
		m.oldValue = func(context.Context) (*Locker, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m LockerMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m LockerMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, errors.New("ent: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// ID returns the ID value in the mutation. Note that the ID is only available
// if it was provided to the builder or after it was returned from the database.
func (m *LockerMutation) ID() (id int, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// IDs queries the database and returns the entity ids that match the mutation's predicate.
// That means, if the mutation is applied within a transaction with an isolation level such
// as sql.LevelSerializable, the returned ids match the ids of the rows that will be updated
// or updated by the mutation.
func (m *LockerMutation) IDs(ctx context.Context) ([]int, error) {
	switch {
	case m.op.Is(OpUpdateOne | OpDeleteOne):
		id, exists := m.ID()
		if exists {
			return []int{id}, nil
		}
		fallthrough
	case m.op.Is(OpUpdate | OpDelete):
		return m.Client().Locker.Query().Where(m.predicates...).IDs(ctx)
	default:
		return nil, fmt.Errorf("IDs is not allowed on %s operations", m.op)
	}
}

// SetHash sets the "hash" field.
func (m *LockerMutation) SetHash(s string) {
	m.hash = &s
}

// Hash returns the value of the "hash" field in the mutation.
func (m *LockerMutation) Hash() (r string, exists bool) {
	v := m.hash
	if v == nil {
		return
	}
	return *v, true
}

// OldHash returns the old "hash" field's value of the Locker entity.
// If the Locker object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *LockerMutation) OldHash(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldHash is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldHash requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldHash: %w", err)
	}
	return oldValue.Hash, nil
}

// ResetHash resets all changes to the "hash" field.
func (m *LockerMutation) ResetHash() {
	m.hash = nil
}

// SetLevel sets the "level" field.
func (m *LockerMutation) SetLevel(i int32) {
	m.level = &i
	m.addlevel = nil
}

// Level returns the value of the "level" field in the mutation.
func (m *LockerMutation) Level() (r int32, exists bool) {
	v := m.level
	if v == nil {
		return
	}
	return *v, true
}

// OldLevel returns the old "level" field's value of the Locker entity.
// If the Locker object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *LockerMutation) OldLevel(ctx context.Context) (v int32, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldLevel is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldLevel requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldLevel: %w", err)
	}
	return oldValue.Level, nil
}

// AddLevel adds i to the "level" field.
func (m *LockerMutation) AddLevel(i int32) {
	if m.addlevel != nil {
		*m.addlevel += i
	} else {
		m.addlevel = &i
	}
}

// AddedLevel returns the value that was added to the "level" field in this mutation.
func (m *LockerMutation) AddedLevel() (r int32, exists bool) {
	v := m.addlevel
	if v == nil {
		return
	}
	return *v, true
}

// ResetLevel resets all changes to the "level" field.
func (m *LockerMutation) ResetLevel() {
	m.level = nil
	m.addlevel = nil
}

// SetEncryptedID sets the "encrypted_id" field.
func (m *LockerMutation) SetEncryptedID(s string) {
	m.encrypted_id = &s
}

// EncryptedID returns the value of the "encrypted_id" field in the mutation.
func (m *LockerMutation) EncryptedID() (r string, exists bool) {
	v := m.encrypted_id
	if v == nil {
		return
	}
	return *v, true
}

// OldEncryptedID returns the old "encrypted_id" field's value of the Locker entity.
// If the Locker object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *LockerMutation) OldEncryptedID(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldEncryptedID is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldEncryptedID requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldEncryptedID: %w", err)
	}
	return oldValue.EncryptedID, nil
}

// ResetEncryptedID resets all changes to the "encrypted_id" field.
func (m *LockerMutation) ResetEncryptedID() {
	m.encrypted_id = nil
}

// SetEncryptedBody sets the "encrypted_body" field.
func (m *LockerMutation) SetEncryptedBody(s string) {
	m.encrypted_body = &s
}

// EncryptedBody returns the value of the "encrypted_body" field in the mutation.
func (m *LockerMutation) EncryptedBody() (r string, exists bool) {
	v := m.encrypted_body
	if v == nil {
		return
	}
	return *v, true
}

// OldEncryptedBody returns the old "encrypted_body" field's value of the Locker entity.
// If the Locker object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *LockerMutation) OldEncryptedBody(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldEncryptedBody is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldEncryptedBody requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldEncryptedBody: %w", err)
	}
	return oldValue.EncryptedBody, nil
}

// ResetEncryptedBody resets all changes to the "encrypted_body" field.
func (m *LockerMutation) ResetEncryptedBody() {
	m.encrypted_body = nil
}

// SetAccountID sets the "account" edge to the Account entity by id.
func (m *LockerMutation) SetAccountID(id int) {
	m.account = &id
}

// ClearAccount clears the "account" edge to the Account entity.
func (m *LockerMutation) ClearAccount() {
	m.clearedaccount = true
}

// AccountCleared reports if the "account" edge to the Account entity was cleared.
func (m *LockerMutation) AccountCleared() bool {
	return m.clearedaccount
}

// AccountID returns the "account" edge ID in the mutation.
func (m *LockerMutation) AccountID() (id int, exists bool) {
	if m.account != nil {
		return *m.account, true
	}
	return
}

// AccountIDs returns the "account" edge IDs in the mutation.
// Note that IDs always returns len(IDs) <= 1 for unique edges, and you should use
// AccountID instead. It exists only for internal usage by the builders.
func (m *LockerMutation) AccountIDs() (ids []int) {
	if id := m.account; id != nil {
		ids = append(ids, *id)
	}
	return
}

// ResetAccount resets all changes to the "account" edge.
func (m *LockerMutation) ResetAccount() {
	m.account = nil
	m.clearedaccount = false
}

// Where appends a list predicates to the LockerMutation builder.
func (m *LockerMutation) Where(ps ...predicate.Locker) {
	m.predicates = append(m.predicates, ps...)
}

// WhereP appends storage-level predicates to the LockerMutation builder. Using this method,
// users can use type-assertion to append predicates that do not depend on any generated package.
func (m *LockerMutation) WhereP(ps ...func(*sql.Selector)) {
	p := make([]predicate.Locker, len(ps))
	for i := range ps {
		p[i] = ps[i]
	}
	m.Where(p...)
}

// Op returns the operation name.
func (m *LockerMutation) Op() Op {
	return m.op
}

// SetOp allows setting the mutation operation.
func (m *LockerMutation) SetOp(op Op) {
	m.op = op
}

// Type returns the node type of this mutation (Locker).
func (m *LockerMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *LockerMutation) Fields() []string {
	fields := make([]string, 0, 4)
	if m.hash != nil {
		fields = append(fields, locker.FieldHash)
	}
	if m.level != nil {
		fields = append(fields, locker.FieldLevel)
	}
	if m.encrypted_id != nil {
		fields = append(fields, locker.FieldEncryptedID)
	}
	if m.encrypted_body != nil {
		fields = append(fields, locker.FieldEncryptedBody)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *LockerMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case locker.FieldHash:
		return m.Hash()
	case locker.FieldLevel:
		return m.Level()
	case locker.FieldEncryptedID:
		return m.EncryptedID()
	case locker.FieldEncryptedBody:
		return m.EncryptedBody()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *LockerMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case locker.FieldHash:
		return m.OldHash(ctx)
	case locker.FieldLevel:
		return m.OldLevel(ctx)
	case locker.FieldEncryptedID:
		return m.OldEncryptedID(ctx)
	case locker.FieldEncryptedBody:
		return m.OldEncryptedBody(ctx)
	}
	return nil, fmt.Errorf("unknown Locker field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *LockerMutation) SetField(name string, value ent.Value) error {
	switch name {
	case locker.FieldHash:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetHash(v)
		return nil
	case locker.FieldLevel:
		v, ok := value.(int32)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetLevel(v)
		return nil
	case locker.FieldEncryptedID:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetEncryptedID(v)
		return nil
	case locker.FieldEncryptedBody:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetEncryptedBody(v)
		return nil
	}
	return fmt.Errorf("unknown Locker field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *LockerMutation) AddedFields() []string {
	var fields []string
	if m.addlevel != nil {
		fields = append(fields, locker.FieldLevel)
	}
	return fields
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *LockerMutation) AddedField(name string) (ent.Value, bool) {
	switch name {
	case locker.FieldLevel:
		return m.AddedLevel()
	}
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *LockerMutation) AddField(name string, value ent.Value) error {
	switch name {
	case locker.FieldLevel:
		v, ok := value.(int32)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.AddLevel(v)
		return nil
	}
	return fmt.Errorf("unknown Locker numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *LockerMutation) ClearedFields() []string {
	return nil
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *LockerMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *LockerMutation) ClearField(name string) error {
	return fmt.Errorf("unknown Locker nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *LockerMutation) ResetField(name string) error {
	switch name {
	case locker.FieldHash:
		m.ResetHash()
		return nil
	case locker.FieldLevel:
		m.ResetLevel()
		return nil
	case locker.FieldEncryptedID:
		m.ResetEncryptedID()
		return nil
	case locker.FieldEncryptedBody:
		m.ResetEncryptedBody()
		return nil
	}
	return fmt.Errorf("unknown Locker field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *LockerMutation) AddedEdges() []string {
	edges := make([]string, 0, 1)
	if m.account != nil {
		edges = append(edges, locker.EdgeAccount)
	}
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *LockerMutation) AddedIDs(name string) []ent.Value {
	switch name {
	case locker.EdgeAccount:
		if id := m.account; id != nil {
			return []ent.Value{*id}
		}
	}
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *LockerMutation) RemovedEdges() []string {
	edges := make([]string, 0, 1)
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *LockerMutation) RemovedIDs(name string) []ent.Value {
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *LockerMutation) ClearedEdges() []string {
	edges := make([]string, 0, 1)
	if m.clearedaccount {
		edges = append(edges, locker.EdgeAccount)
	}
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *LockerMutation) EdgeCleared(name string) bool {
	switch name {
	case locker.EdgeAccount:
		return m.clearedaccount
	}
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *LockerMutation) ClearEdge(name string) error {
	switch name {
	case locker.EdgeAccount:
		m.ClearAccount()
		return nil
	}
	return fmt.Errorf("unknown Locker unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *LockerMutation) ResetEdge(name string) error {
	switch name {
	case locker.EdgeAccount:
		m.ResetAccount()
		return nil
	}
	return fmt.Errorf("unknown Locker edge %s", name)
}

// PropertyMutation represents an operation that mutates the Property nodes in the graph.
type PropertyMutation struct {
	config
	op             Op
	typ            string
	id             *int
	hash           *string
	level          *int32
	addlevel       *int32
	encrypted_id   *string
	encrypted_body *string
	clearedFields  map[string]struct{}
	account        *int
	clearedaccount bool
	done           bool
	oldValue       func(context.Context) (*Property, error)
	predicates     []predicate.Property
}

var _ ent.Mutation = (*PropertyMutation)(nil)

// propertyOption allows management of the mutation configuration using functional options.
type propertyOption func(*PropertyMutation)

// newPropertyMutation creates new mutation for the Property entity.
func newPropertyMutation(c config, op Op, opts ...propertyOption) *PropertyMutation {
	m := &PropertyMutation{
		config:        c,
		op:            op,
		typ:           TypeProperty,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withPropertyID sets the ID field of the mutation.
func withPropertyID(id int) propertyOption {
	return func(m *PropertyMutation) {
		var (
			err   error
			once  sync.Once
			value *Property
		)
		m.oldValue = func(ctx context.Context) (*Property, error) {
			once.Do(func() {
				if m.done {
					err = errors.New("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().Property.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withProperty sets the old Property of the mutation.
func withProperty(node *Property) propertyOption {
	return func(m *PropertyMutation) {
		m.oldValue = func(context.Context) (*Property, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m PropertyMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m PropertyMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, errors.New("ent: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// ID returns the ID value in the mutation. Note that the ID is only available
// if it was provided to the builder or after it was returned from the database.
func (m *PropertyMutation) ID() (id int, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// IDs queries the database and returns the entity ids that match the mutation's predicate.
// That means, if the mutation is applied within a transaction with an isolation level such
// as sql.LevelSerializable, the returned ids match the ids of the rows that will be updated
// or updated by the mutation.
func (m *PropertyMutation) IDs(ctx context.Context) ([]int, error) {
	switch {
	case m.op.Is(OpUpdateOne | OpDeleteOne):
		id, exists := m.ID()
		if exists {
			return []int{id}, nil
		}
		fallthrough
	case m.op.Is(OpUpdate | OpDelete):
		return m.Client().Property.Query().Where(m.predicates...).IDs(ctx)
	default:
		return nil, fmt.Errorf("IDs is not allowed on %s operations", m.op)
	}
}

// SetHash sets the "hash" field.
func (m *PropertyMutation) SetHash(s string) {
	m.hash = &s
}

// Hash returns the value of the "hash" field in the mutation.
func (m *PropertyMutation) Hash() (r string, exists bool) {
	v := m.hash
	if v == nil {
		return
	}
	return *v, true
}

// OldHash returns the old "hash" field's value of the Property entity.
// If the Property object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *PropertyMutation) OldHash(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldHash is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldHash requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldHash: %w", err)
	}
	return oldValue.Hash, nil
}

// ResetHash resets all changes to the "hash" field.
func (m *PropertyMutation) ResetHash() {
	m.hash = nil
}

// SetLevel sets the "level" field.
func (m *PropertyMutation) SetLevel(i int32) {
	m.level = &i
	m.addlevel = nil
}

// Level returns the value of the "level" field in the mutation.
func (m *PropertyMutation) Level() (r int32, exists bool) {
	v := m.level
	if v == nil {
		return
	}
	return *v, true
}

// OldLevel returns the old "level" field's value of the Property entity.
// If the Property object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *PropertyMutation) OldLevel(ctx context.Context) (v int32, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldLevel is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldLevel requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldLevel: %w", err)
	}
	return oldValue.Level, nil
}

// AddLevel adds i to the "level" field.
func (m *PropertyMutation) AddLevel(i int32) {
	if m.addlevel != nil {
		*m.addlevel += i
	} else {
		m.addlevel = &i
	}
}

// AddedLevel returns the value that was added to the "level" field in this mutation.
func (m *PropertyMutation) AddedLevel() (r int32, exists bool) {
	v := m.addlevel
	if v == nil {
		return
	}
	return *v, true
}

// ResetLevel resets all changes to the "level" field.
func (m *PropertyMutation) ResetLevel() {
	m.level = nil
	m.addlevel = nil
}

// SetEncryptedID sets the "encrypted_id" field.
func (m *PropertyMutation) SetEncryptedID(s string) {
	m.encrypted_id = &s
}

// EncryptedID returns the value of the "encrypted_id" field in the mutation.
func (m *PropertyMutation) EncryptedID() (r string, exists bool) {
	v := m.encrypted_id
	if v == nil {
		return
	}
	return *v, true
}

// OldEncryptedID returns the old "encrypted_id" field's value of the Property entity.
// If the Property object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *PropertyMutation) OldEncryptedID(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldEncryptedID is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldEncryptedID requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldEncryptedID: %w", err)
	}
	return oldValue.EncryptedID, nil
}

// ResetEncryptedID resets all changes to the "encrypted_id" field.
func (m *PropertyMutation) ResetEncryptedID() {
	m.encrypted_id = nil
}

// SetEncryptedBody sets the "encrypted_body" field.
func (m *PropertyMutation) SetEncryptedBody(s string) {
	m.encrypted_body = &s
}

// EncryptedBody returns the value of the "encrypted_body" field in the mutation.
func (m *PropertyMutation) EncryptedBody() (r string, exists bool) {
	v := m.encrypted_body
	if v == nil {
		return
	}
	return *v, true
}

// OldEncryptedBody returns the old "encrypted_body" field's value of the Property entity.
// If the Property object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *PropertyMutation) OldEncryptedBody(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldEncryptedBody is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldEncryptedBody requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldEncryptedBody: %w", err)
	}
	return oldValue.EncryptedBody, nil
}

// ResetEncryptedBody resets all changes to the "encrypted_body" field.
func (m *PropertyMutation) ResetEncryptedBody() {
	m.encrypted_body = nil
}

// SetAccountID sets the "account" edge to the Account entity by id.
func (m *PropertyMutation) SetAccountID(id int) {
	m.account = &id
}

// ClearAccount clears the "account" edge to the Account entity.
func (m *PropertyMutation) ClearAccount() {
	m.clearedaccount = true
}

// AccountCleared reports if the "account" edge to the Account entity was cleared.
func (m *PropertyMutation) AccountCleared() bool {
	return m.clearedaccount
}

// AccountID returns the "account" edge ID in the mutation.
func (m *PropertyMutation) AccountID() (id int, exists bool) {
	if m.account != nil {
		return *m.account, true
	}
	return
}

// AccountIDs returns the "account" edge IDs in the mutation.
// Note that IDs always returns len(IDs) <= 1 for unique edges, and you should use
// AccountID instead. It exists only for internal usage by the builders.
func (m *PropertyMutation) AccountIDs() (ids []int) {
	if id := m.account; id != nil {
		ids = append(ids, *id)
	}
	return
}

// ResetAccount resets all changes to the "account" edge.
func (m *PropertyMutation) ResetAccount() {
	m.account = nil
	m.clearedaccount = false
}

// Where appends a list predicates to the PropertyMutation builder.
func (m *PropertyMutation) Where(ps ...predicate.Property) {
	m.predicates = append(m.predicates, ps...)
}

// WhereP appends storage-level predicates to the PropertyMutation builder. Using this method,
// users can use type-assertion to append predicates that do not depend on any generated package.
func (m *PropertyMutation) WhereP(ps ...func(*sql.Selector)) {
	p := make([]predicate.Property, len(ps))
	for i := range ps {
		p[i] = ps[i]
	}
	m.Where(p...)
}

// Op returns the operation name.
func (m *PropertyMutation) Op() Op {
	return m.op
}

// SetOp allows setting the mutation operation.
func (m *PropertyMutation) SetOp(op Op) {
	m.op = op
}

// Type returns the node type of this mutation (Property).
func (m *PropertyMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *PropertyMutation) Fields() []string {
	fields := make([]string, 0, 4)
	if m.hash != nil {
		fields = append(fields, property.FieldHash)
	}
	if m.level != nil {
		fields = append(fields, property.FieldLevel)
	}
	if m.encrypted_id != nil {
		fields = append(fields, property.FieldEncryptedID)
	}
	if m.encrypted_body != nil {
		fields = append(fields, property.FieldEncryptedBody)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *PropertyMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case property.FieldHash:
		return m.Hash()
	case property.FieldLevel:
		return m.Level()
	case property.FieldEncryptedID:
		return m.EncryptedID()
	case property.FieldEncryptedBody:
		return m.EncryptedBody()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *PropertyMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case property.FieldHash:
		return m.OldHash(ctx)
	case property.FieldLevel:
		return m.OldLevel(ctx)
	case property.FieldEncryptedID:
		return m.OldEncryptedID(ctx)
	case property.FieldEncryptedBody:
		return m.OldEncryptedBody(ctx)
	}
	return nil, fmt.Errorf("unknown Property field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *PropertyMutation) SetField(name string, value ent.Value) error {
	switch name {
	case property.FieldHash:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetHash(v)
		return nil
	case property.FieldLevel:
		v, ok := value.(int32)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetLevel(v)
		return nil
	case property.FieldEncryptedID:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetEncryptedID(v)
		return nil
	case property.FieldEncryptedBody:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetEncryptedBody(v)
		return nil
	}
	return fmt.Errorf("unknown Property field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *PropertyMutation) AddedFields() []string {
	var fields []string
	if m.addlevel != nil {
		fields = append(fields, property.FieldLevel)
	}
	return fields
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *PropertyMutation) AddedField(name string) (ent.Value, bool) {
	switch name {
	case property.FieldLevel:
		return m.AddedLevel()
	}
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *PropertyMutation) AddField(name string, value ent.Value) error {
	switch name {
	case property.FieldLevel:
		v, ok := value.(int32)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.AddLevel(v)
		return nil
	}
	return fmt.Errorf("unknown Property numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *PropertyMutation) ClearedFields() []string {
	return nil
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *PropertyMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *PropertyMutation) ClearField(name string) error {
	return fmt.Errorf("unknown Property nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *PropertyMutation) ResetField(name string) error {
	switch name {
	case property.FieldHash:
		m.ResetHash()
		return nil
	case property.FieldLevel:
		m.ResetLevel()
		return nil
	case property.FieldEncryptedID:
		m.ResetEncryptedID()
		return nil
	case property.FieldEncryptedBody:
		m.ResetEncryptedBody()
		return nil
	}
	return fmt.Errorf("unknown Property field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *PropertyMutation) AddedEdges() []string {
	edges := make([]string, 0, 1)
	if m.account != nil {
		edges = append(edges, property.EdgeAccount)
	}
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *PropertyMutation) AddedIDs(name string) []ent.Value {
	switch name {
	case property.EdgeAccount:
		if id := m.account; id != nil {
			return []ent.Value{*id}
		}
	}
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *PropertyMutation) RemovedEdges() []string {
	edges := make([]string, 0, 1)
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *PropertyMutation) RemovedIDs(name string) []ent.Value {
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *PropertyMutation) ClearedEdges() []string {
	edges := make([]string, 0, 1)
	if m.clearedaccount {
		edges = append(edges, property.EdgeAccount)
	}
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *PropertyMutation) EdgeCleared(name string) bool {
	switch name {
	case property.EdgeAccount:
		return m.clearedaccount
	}
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *PropertyMutation) ClearEdge(name string) error {
	switch name {
	case property.EdgeAccount:
		m.ClearAccount()
		return nil
	}
	return fmt.Errorf("unknown Property unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *PropertyMutation) ResetEdge(name string) error {
	switch name {
	case property.EdgeAccount:
		m.ResetAccount()
		return nil
	}
	return fmt.Errorf("unknown Property edge %s", name)
}

// RecoveryCodeMutation represents an operation that mutates the RecoveryCode nodes in the graph.
type RecoveryCodeMutation struct {
	config
	op             Op
	typ            string
	id             *int
	code           *string
	expires_at     *time.Time
	clearedFields  map[string]struct{}
	account        *int
	clearedaccount bool
	done           bool
	oldValue       func(context.Context) (*RecoveryCode, error)
	predicates     []predicate.RecoveryCode
}

var _ ent.Mutation = (*RecoveryCodeMutation)(nil)

// recoverycodeOption allows management of the mutation configuration using functional options.
type recoverycodeOption func(*RecoveryCodeMutation)

// newRecoveryCodeMutation creates new mutation for the RecoveryCode entity.
func newRecoveryCodeMutation(c config, op Op, opts ...recoverycodeOption) *RecoveryCodeMutation {
	m := &RecoveryCodeMutation{
		config:        c,
		op:            op,
		typ:           TypeRecoveryCode,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withRecoveryCodeID sets the ID field of the mutation.
func withRecoveryCodeID(id int) recoverycodeOption {
	return func(m *RecoveryCodeMutation) {
		var (
			err   error
			once  sync.Once
			value *RecoveryCode
		)
		m.oldValue = func(ctx context.Context) (*RecoveryCode, error) {
			once.Do(func() {
				if m.done {
					err = errors.New("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().RecoveryCode.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withRecoveryCode sets the old RecoveryCode of the mutation.
func withRecoveryCode(node *RecoveryCode) recoverycodeOption {
	return func(m *RecoveryCodeMutation) {
		m.oldValue = func(context.Context) (*RecoveryCode, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m RecoveryCodeMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m RecoveryCodeMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, errors.New("ent: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// ID returns the ID value in the mutation. Note that the ID is only available
// if it was provided to the builder or after it was returned from the database.
func (m *RecoveryCodeMutation) ID() (id int, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// IDs queries the database and returns the entity ids that match the mutation's predicate.
// That means, if the mutation is applied within a transaction with an isolation level such
// as sql.LevelSerializable, the returned ids match the ids of the rows that will be updated
// or updated by the mutation.
func (m *RecoveryCodeMutation) IDs(ctx context.Context) ([]int, error) {
	switch {
	case m.op.Is(OpUpdateOne | OpDeleteOne):
		id, exists := m.ID()
		if exists {
			return []int{id}, nil
		}
		fallthrough
	case m.op.Is(OpUpdate | OpDelete):
		return m.Client().RecoveryCode.Query().Where(m.predicates...).IDs(ctx)
	default:
		return nil, fmt.Errorf("IDs is not allowed on %s operations", m.op)
	}
}

// SetCode sets the "code" field.
func (m *RecoveryCodeMutation) SetCode(s string) {
	m.code = &s
}

// Code returns the value of the "code" field in the mutation.
func (m *RecoveryCodeMutation) Code() (r string, exists bool) {
	v := m.code
	if v == nil {
		return
	}
	return *v, true
}

// OldCode returns the old "code" field's value of the RecoveryCode entity.
// If the RecoveryCode object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *RecoveryCodeMutation) OldCode(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldCode is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldCode requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldCode: %w", err)
	}
	return oldValue.Code, nil
}

// ResetCode resets all changes to the "code" field.
func (m *RecoveryCodeMutation) ResetCode() {
	m.code = nil
}

// SetExpiresAt sets the "expires_at" field.
func (m *RecoveryCodeMutation) SetExpiresAt(t time.Time) {
	m.expires_at = &t
}

// ExpiresAt returns the value of the "expires_at" field in the mutation.
func (m *RecoveryCodeMutation) ExpiresAt() (r time.Time, exists bool) {
	v := m.expires_at
	if v == nil {
		return
	}
	return *v, true
}

// OldExpiresAt returns the old "expires_at" field's value of the RecoveryCode entity.
// If the RecoveryCode object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *RecoveryCodeMutation) OldExpiresAt(ctx context.Context) (v *time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldExpiresAt is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldExpiresAt requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldExpiresAt: %w", err)
	}
	return oldValue.ExpiresAt, nil
}

// ClearExpiresAt clears the value of the "expires_at" field.
func (m *RecoveryCodeMutation) ClearExpiresAt() {
	m.expires_at = nil
	m.clearedFields[recoverycode.FieldExpiresAt] = struct{}{}
}

// ExpiresAtCleared returns if the "expires_at" field was cleared in this mutation.
func (m *RecoveryCodeMutation) ExpiresAtCleared() bool {
	_, ok := m.clearedFields[recoverycode.FieldExpiresAt]
	return ok
}

// ResetExpiresAt resets all changes to the "expires_at" field.
func (m *RecoveryCodeMutation) ResetExpiresAt() {
	m.expires_at = nil
	delete(m.clearedFields, recoverycode.FieldExpiresAt)
}

// SetAccountID sets the "account" edge to the Account entity by id.
func (m *RecoveryCodeMutation) SetAccountID(id int) {
	m.account = &id
}

// ClearAccount clears the "account" edge to the Account entity.
func (m *RecoveryCodeMutation) ClearAccount() {
	m.clearedaccount = true
}

// AccountCleared reports if the "account" edge to the Account entity was cleared.
func (m *RecoveryCodeMutation) AccountCleared() bool {
	return m.clearedaccount
}

// AccountID returns the "account" edge ID in the mutation.
func (m *RecoveryCodeMutation) AccountID() (id int, exists bool) {
	if m.account != nil {
		return *m.account, true
	}
	return
}

// AccountIDs returns the "account" edge IDs in the mutation.
// Note that IDs always returns len(IDs) <= 1 for unique edges, and you should use
// AccountID instead. It exists only for internal usage by the builders.
func (m *RecoveryCodeMutation) AccountIDs() (ids []int) {
	if id := m.account; id != nil {
		ids = append(ids, *id)
	}
	return
}

// ResetAccount resets all changes to the "account" edge.
func (m *RecoveryCodeMutation) ResetAccount() {
	m.account = nil
	m.clearedaccount = false
}

// Where appends a list predicates to the RecoveryCodeMutation builder.
func (m *RecoveryCodeMutation) Where(ps ...predicate.RecoveryCode) {
	m.predicates = append(m.predicates, ps...)
}

// WhereP appends storage-level predicates to the RecoveryCodeMutation builder. Using this method,
// users can use type-assertion to append predicates that do not depend on any generated package.
func (m *RecoveryCodeMutation) WhereP(ps ...func(*sql.Selector)) {
	p := make([]predicate.RecoveryCode, len(ps))
	for i := range ps {
		p[i] = ps[i]
	}
	m.Where(p...)
}

// Op returns the operation name.
func (m *RecoveryCodeMutation) Op() Op {
	return m.op
}

// SetOp allows setting the mutation operation.
func (m *RecoveryCodeMutation) SetOp(op Op) {
	m.op = op
}

// Type returns the node type of this mutation (RecoveryCode).
func (m *RecoveryCodeMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *RecoveryCodeMutation) Fields() []string {
	fields := make([]string, 0, 2)
	if m.code != nil {
		fields = append(fields, recoverycode.FieldCode)
	}
	if m.expires_at != nil {
		fields = append(fields, recoverycode.FieldExpiresAt)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *RecoveryCodeMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case recoverycode.FieldCode:
		return m.Code()
	case recoverycode.FieldExpiresAt:
		return m.ExpiresAt()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *RecoveryCodeMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case recoverycode.FieldCode:
		return m.OldCode(ctx)
	case recoverycode.FieldExpiresAt:
		return m.OldExpiresAt(ctx)
	}
	return nil, fmt.Errorf("unknown RecoveryCode field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *RecoveryCodeMutation) SetField(name string, value ent.Value) error {
	switch name {
	case recoverycode.FieldCode:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetCode(v)
		return nil
	case recoverycode.FieldExpiresAt:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetExpiresAt(v)
		return nil
	}
	return fmt.Errorf("unknown RecoveryCode field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *RecoveryCodeMutation) AddedFields() []string {
	return nil
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *RecoveryCodeMutation) AddedField(name string) (ent.Value, bool) {
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *RecoveryCodeMutation) AddField(name string, value ent.Value) error {
	switch name {
	}
	return fmt.Errorf("unknown RecoveryCode numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *RecoveryCodeMutation) ClearedFields() []string {
	var fields []string
	if m.FieldCleared(recoverycode.FieldExpiresAt) {
		fields = append(fields, recoverycode.FieldExpiresAt)
	}
	return fields
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *RecoveryCodeMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *RecoveryCodeMutation) ClearField(name string) error {
	switch name {
	case recoverycode.FieldExpiresAt:
		m.ClearExpiresAt()
		return nil
	}
	return fmt.Errorf("unknown RecoveryCode nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *RecoveryCodeMutation) ResetField(name string) error {
	switch name {
	case recoverycode.FieldCode:
		m.ResetCode()
		return nil
	case recoverycode.FieldExpiresAt:
		m.ResetExpiresAt()
		return nil
	}
	return fmt.Errorf("unknown RecoveryCode field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *RecoveryCodeMutation) AddedEdges() []string {
	edges := make([]string, 0, 1)
	if m.account != nil {
		edges = append(edges, recoverycode.EdgeAccount)
	}
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *RecoveryCodeMutation) AddedIDs(name string) []ent.Value {
	switch name {
	case recoverycode.EdgeAccount:
		if id := m.account; id != nil {
			return []ent.Value{*id}
		}
	}
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *RecoveryCodeMutation) RemovedEdges() []string {
	edges := make([]string, 0, 1)
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *RecoveryCodeMutation) RemovedIDs(name string) []ent.Value {
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *RecoveryCodeMutation) ClearedEdges() []string {
	edges := make([]string, 0, 1)
	if m.clearedaccount {
		edges = append(edges, recoverycode.EdgeAccount)
	}
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *RecoveryCodeMutation) EdgeCleared(name string) bool {
	switch name {
	case recoverycode.EdgeAccount:
		return m.clearedaccount
	}
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *RecoveryCodeMutation) ClearEdge(name string) error {
	switch name {
	case recoverycode.EdgeAccount:
		m.ClearAccount()
		return nil
	}
	return fmt.Errorf("unknown RecoveryCode unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *RecoveryCodeMutation) ResetEdge(name string) error {
	switch name {
	case recoverycode.EdgeAccount:
		m.ResetAccount()
		return nil
	}
	return fmt.Errorf("unknown RecoveryCode edge %s", name)
}
