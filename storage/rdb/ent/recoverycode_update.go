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
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/piprate/metalocker/storage/rdb/ent/predicate"
	"github.com/piprate/metalocker/storage/rdb/ent/recoverycode"

	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
)

// RecoveryCodeUpdate is the builder for updating RecoveryCode entities.
type RecoveryCodeUpdate struct {
	config
	hooks    []Hook
	mutation *RecoveryCodeMutation
}

// Where appends a list predicates to the RecoveryCodeUpdate builder.
func (rcu *RecoveryCodeUpdate) Where(ps ...predicate.RecoveryCode) *RecoveryCodeUpdate {
	rcu.mutation.Where(ps...)
	return rcu
}

// SetCode sets the "code" field.
func (rcu *RecoveryCodeUpdate) SetCode(s string) *RecoveryCodeUpdate {
	rcu.mutation.SetCode(s)
	return rcu
}

// SetExpiresAt sets the "expires_at" field.
func (rcu *RecoveryCodeUpdate) SetExpiresAt(t time.Time) *RecoveryCodeUpdate {
	rcu.mutation.SetExpiresAt(t)
	return rcu
}

// SetNillableExpiresAt sets the "expires_at" field if the given value is not nil.
func (rcu *RecoveryCodeUpdate) SetNillableExpiresAt(t *time.Time) *RecoveryCodeUpdate {
	if t != nil {
		rcu.SetExpiresAt(*t)
	}
	return rcu
}

// ClearExpiresAt clears the value of the "expires_at" field.
func (rcu *RecoveryCodeUpdate) ClearExpiresAt() *RecoveryCodeUpdate {
	rcu.mutation.ClearExpiresAt()
	return rcu
}

// SetAccountID sets the "account" edge to the Account entity by ID.
func (rcu *RecoveryCodeUpdate) SetAccountID(id int) *RecoveryCodeUpdate {
	rcu.mutation.SetAccountID(id)
	return rcu
}

// SetNillableAccountID sets the "account" edge to the Account entity by ID if the given value is not nil.
func (rcu *RecoveryCodeUpdate) SetNillableAccountID(id *int) *RecoveryCodeUpdate {
	if id != nil {
		rcu = rcu.SetAccountID(*id)
	}
	return rcu
}

// SetAccount sets the "account" edge to the Account entity.
func (rcu *RecoveryCodeUpdate) SetAccount(a *Account) *RecoveryCodeUpdate {
	return rcu.SetAccountID(a.ID)
}

// Mutation returns the RecoveryCodeMutation object of the builder.
func (rcu *RecoveryCodeUpdate) Mutation() *RecoveryCodeMutation {
	return rcu.mutation
}

// ClearAccount clears the "account" edge to the Account entity.
func (rcu *RecoveryCodeUpdate) ClearAccount() *RecoveryCodeUpdate {
	rcu.mutation.ClearAccount()
	return rcu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (rcu *RecoveryCodeUpdate) Save(ctx context.Context) (int, error) {
	return withHooks[int, RecoveryCodeMutation](ctx, rcu.sqlSave, rcu.mutation, rcu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (rcu *RecoveryCodeUpdate) SaveX(ctx context.Context) int {
	affected, err := rcu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (rcu *RecoveryCodeUpdate) Exec(ctx context.Context) error {
	_, err := rcu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (rcu *RecoveryCodeUpdate) ExecX(ctx context.Context) {
	if err := rcu.Exec(ctx); err != nil {
		panic(err)
	}
}

func (rcu *RecoveryCodeUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   recoverycode.Table,
			Columns: recoverycode.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: recoverycode.FieldID,
			},
		},
	}
	if ps := rcu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := rcu.mutation.Code(); ok {
		_spec.SetField(recoverycode.FieldCode, field.TypeString, value)
	}
	if value, ok := rcu.mutation.ExpiresAt(); ok {
		_spec.SetField(recoverycode.FieldExpiresAt, field.TypeTime, value)
	}
	if rcu.mutation.ExpiresAtCleared() {
		_spec.ClearField(recoverycode.FieldExpiresAt, field.TypeTime)
	}
	if rcu.mutation.AccountCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   recoverycode.AccountTable,
			Columns: []string{recoverycode.AccountColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeInt,
					Column: entaccount.FieldID,
				},
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := rcu.mutation.AccountIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   recoverycode.AccountTable,
			Columns: []string{recoverycode.AccountColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeInt,
					Column: entaccount.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, rcu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{recoverycode.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	rcu.mutation.done = true
	return n, nil
}

// RecoveryCodeUpdateOne is the builder for updating a single RecoveryCode entity.
type RecoveryCodeUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *RecoveryCodeMutation
}

// SetCode sets the "code" field.
func (rcuo *RecoveryCodeUpdateOne) SetCode(s string) *RecoveryCodeUpdateOne {
	rcuo.mutation.SetCode(s)
	return rcuo
}

// SetExpiresAt sets the "expires_at" field.
func (rcuo *RecoveryCodeUpdateOne) SetExpiresAt(t time.Time) *RecoveryCodeUpdateOne {
	rcuo.mutation.SetExpiresAt(t)
	return rcuo
}

// SetNillableExpiresAt sets the "expires_at" field if the given value is not nil.
func (rcuo *RecoveryCodeUpdateOne) SetNillableExpiresAt(t *time.Time) *RecoveryCodeUpdateOne {
	if t != nil {
		rcuo.SetExpiresAt(*t)
	}
	return rcuo
}

// ClearExpiresAt clears the value of the "expires_at" field.
func (rcuo *RecoveryCodeUpdateOne) ClearExpiresAt() *RecoveryCodeUpdateOne {
	rcuo.mutation.ClearExpiresAt()
	return rcuo
}

// SetAccountID sets the "account" edge to the Account entity by ID.
func (rcuo *RecoveryCodeUpdateOne) SetAccountID(id int) *RecoveryCodeUpdateOne {
	rcuo.mutation.SetAccountID(id)
	return rcuo
}

// SetNillableAccountID sets the "account" edge to the Account entity by ID if the given value is not nil.
func (rcuo *RecoveryCodeUpdateOne) SetNillableAccountID(id *int) *RecoveryCodeUpdateOne {
	if id != nil {
		rcuo = rcuo.SetAccountID(*id)
	}
	return rcuo
}

// SetAccount sets the "account" edge to the Account entity.
func (rcuo *RecoveryCodeUpdateOne) SetAccount(a *Account) *RecoveryCodeUpdateOne {
	return rcuo.SetAccountID(a.ID)
}

// Mutation returns the RecoveryCodeMutation object of the builder.
func (rcuo *RecoveryCodeUpdateOne) Mutation() *RecoveryCodeMutation {
	return rcuo.mutation
}

// ClearAccount clears the "account" edge to the Account entity.
func (rcuo *RecoveryCodeUpdateOne) ClearAccount() *RecoveryCodeUpdateOne {
	rcuo.mutation.ClearAccount()
	return rcuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (rcuo *RecoveryCodeUpdateOne) Select(field string, fields ...string) *RecoveryCodeUpdateOne {
	rcuo.fields = append([]string{field}, fields...)
	return rcuo
}

// Save executes the query and returns the updated RecoveryCode entity.
func (rcuo *RecoveryCodeUpdateOne) Save(ctx context.Context) (*RecoveryCode, error) {
	return withHooks[*RecoveryCode, RecoveryCodeMutation](ctx, rcuo.sqlSave, rcuo.mutation, rcuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (rcuo *RecoveryCodeUpdateOne) SaveX(ctx context.Context) *RecoveryCode {
	node, err := rcuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (rcuo *RecoveryCodeUpdateOne) Exec(ctx context.Context) error {
	_, err := rcuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (rcuo *RecoveryCodeUpdateOne) ExecX(ctx context.Context) {
	if err := rcuo.Exec(ctx); err != nil {
		panic(err)
	}
}

func (rcuo *RecoveryCodeUpdateOne) sqlSave(ctx context.Context) (_node *RecoveryCode, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   recoverycode.Table,
			Columns: recoverycode.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: recoverycode.FieldID,
			},
		},
	}
	id, ok := rcuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "RecoveryCode.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := rcuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, recoverycode.FieldID)
		for _, f := range fields {
			if !recoverycode.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != recoverycode.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := rcuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := rcuo.mutation.Code(); ok {
		_spec.SetField(recoverycode.FieldCode, field.TypeString, value)
	}
	if value, ok := rcuo.mutation.ExpiresAt(); ok {
		_spec.SetField(recoverycode.FieldExpiresAt, field.TypeTime, value)
	}
	if rcuo.mutation.ExpiresAtCleared() {
		_spec.ClearField(recoverycode.FieldExpiresAt, field.TypeTime)
	}
	if rcuo.mutation.AccountCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   recoverycode.AccountTable,
			Columns: []string{recoverycode.AccountColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeInt,
					Column: entaccount.FieldID,
				},
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := rcuo.mutation.AccountIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   recoverycode.AccountTable,
			Columns: []string{recoverycode.AccountColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeInt,
					Column: entaccount.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &RecoveryCode{config: rcuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, rcuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{recoverycode.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	rcuo.mutation.done = true
	return _node, nil
}
