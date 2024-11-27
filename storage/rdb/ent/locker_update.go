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

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/piprate/metalocker/storage/rdb/ent/locker"
	"github.com/piprate/metalocker/storage/rdb/ent/predicate"

	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
)

// LockerUpdate is the builder for updating Locker entities.
type LockerUpdate struct {
	config
	hooks    []Hook
	mutation *LockerMutation
}

// Where appends a list predicates to the LockerUpdate builder.
func (lu *LockerUpdate) Where(ps ...predicate.Locker) *LockerUpdate {
	lu.mutation.Where(ps...)
	return lu
}

// SetHash sets the "hash" field.
func (lu *LockerUpdate) SetHash(s string) *LockerUpdate {
	lu.mutation.SetHash(s)
	return lu
}

// SetLevel sets the "level" field.
func (lu *LockerUpdate) SetLevel(i int32) *LockerUpdate {
	lu.mutation.ResetLevel()
	lu.mutation.SetLevel(i)
	return lu
}

// AddLevel adds i to the "level" field.
func (lu *LockerUpdate) AddLevel(i int32) *LockerUpdate {
	lu.mutation.AddLevel(i)
	return lu
}

// SetEncryptedID sets the "encrypted_id" field.
func (lu *LockerUpdate) SetEncryptedID(s string) *LockerUpdate {
	lu.mutation.SetEncryptedID(s)
	return lu
}

// SetEncryptedBody sets the "encrypted_body" field.
func (lu *LockerUpdate) SetEncryptedBody(s string) *LockerUpdate {
	lu.mutation.SetEncryptedBody(s)
	return lu
}

// SetAccountID sets the "account" edge to the Account entity by ID.
func (lu *LockerUpdate) SetAccountID(id int) *LockerUpdate {
	lu.mutation.SetAccountID(id)
	return lu
}

// SetNillableAccountID sets the "account" edge to the Account entity by ID if the given value is not nil.
func (lu *LockerUpdate) SetNillableAccountID(id *int) *LockerUpdate {
	if id != nil {
		lu = lu.SetAccountID(*id)
	}
	return lu
}

// SetAccount sets the "account" edge to the Account entity.
func (lu *LockerUpdate) SetAccount(a *Account) *LockerUpdate {
	return lu.SetAccountID(a.ID)
}

// Mutation returns the LockerMutation object of the builder.
func (lu *LockerUpdate) Mutation() *LockerMutation {
	return lu.mutation
}

// ClearAccount clears the "account" edge to the Account entity.
func (lu *LockerUpdate) ClearAccount() *LockerUpdate {
	lu.mutation.ClearAccount()
	return lu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (lu *LockerUpdate) Save(ctx context.Context) (int, error) {
	return withHooks[int, LockerMutation](ctx, lu.sqlSave, lu.mutation, lu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (lu *LockerUpdate) SaveX(ctx context.Context) int {
	affected, err := lu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (lu *LockerUpdate) Exec(ctx context.Context) error {
	_, err := lu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (lu *LockerUpdate) ExecX(ctx context.Context) {
	if err := lu.Exec(ctx); err != nil {
		panic(err)
	}
}

func (lu *LockerUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   locker.Table,
			Columns: locker.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: locker.FieldID,
			},
		},
	}
	if ps := lu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := lu.mutation.Hash(); ok {
		_spec.SetField(locker.FieldHash, field.TypeString, value)
	}
	if value, ok := lu.mutation.Level(); ok {
		_spec.SetField(locker.FieldLevel, field.TypeInt32, value)
	}
	if value, ok := lu.mutation.AddedLevel(); ok {
		_spec.AddField(locker.FieldLevel, field.TypeInt32, value)
	}
	if value, ok := lu.mutation.EncryptedID(); ok {
		_spec.SetField(locker.FieldEncryptedID, field.TypeString, value)
	}
	if value, ok := lu.mutation.EncryptedBody(); ok {
		_spec.SetField(locker.FieldEncryptedBody, field.TypeString, value)
	}
	if lu.mutation.AccountCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   locker.AccountTable,
			Columns: []string{locker.AccountColumn},
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
	if nodes := lu.mutation.AccountIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   locker.AccountTable,
			Columns: []string{locker.AccountColumn},
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
	if n, err = sqlgraph.UpdateNodes(ctx, lu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{locker.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	lu.mutation.done = true
	return n, nil
}

// LockerUpdateOne is the builder for updating a single Locker entity.
type LockerUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *LockerMutation
}

// SetHash sets the "hash" field.
func (luo *LockerUpdateOne) SetHash(s string) *LockerUpdateOne {
	luo.mutation.SetHash(s)
	return luo
}

// SetLevel sets the "level" field.
func (luo *LockerUpdateOne) SetLevel(i int32) *LockerUpdateOne {
	luo.mutation.ResetLevel()
	luo.mutation.SetLevel(i)
	return luo
}

// AddLevel adds i to the "level" field.
func (luo *LockerUpdateOne) AddLevel(i int32) *LockerUpdateOne {
	luo.mutation.AddLevel(i)
	return luo
}

// SetEncryptedID sets the "encrypted_id" field.
func (luo *LockerUpdateOne) SetEncryptedID(s string) *LockerUpdateOne {
	luo.mutation.SetEncryptedID(s)
	return luo
}

// SetEncryptedBody sets the "encrypted_body" field.
func (luo *LockerUpdateOne) SetEncryptedBody(s string) *LockerUpdateOne {
	luo.mutation.SetEncryptedBody(s)
	return luo
}

// SetAccountID sets the "account" edge to the Account entity by ID.
func (luo *LockerUpdateOne) SetAccountID(id int) *LockerUpdateOne {
	luo.mutation.SetAccountID(id)
	return luo
}

// SetNillableAccountID sets the "account" edge to the Account entity by ID if the given value is not nil.
func (luo *LockerUpdateOne) SetNillableAccountID(id *int) *LockerUpdateOne {
	if id != nil {
		luo = luo.SetAccountID(*id)
	}
	return luo
}

// SetAccount sets the "account" edge to the Account entity.
func (luo *LockerUpdateOne) SetAccount(a *Account) *LockerUpdateOne {
	return luo.SetAccountID(a.ID)
}

// Mutation returns the LockerMutation object of the builder.
func (luo *LockerUpdateOne) Mutation() *LockerMutation {
	return luo.mutation
}

// ClearAccount clears the "account" edge to the Account entity.
func (luo *LockerUpdateOne) ClearAccount() *LockerUpdateOne {
	luo.mutation.ClearAccount()
	return luo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (luo *LockerUpdateOne) Select(field string, fields ...string) *LockerUpdateOne {
	luo.fields = append([]string{field}, fields...)
	return luo
}

// Save executes the query and returns the updated Locker entity.
func (luo *LockerUpdateOne) Save(ctx context.Context) (*Locker, error) {
	return withHooks[*Locker, LockerMutation](ctx, luo.sqlSave, luo.mutation, luo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (luo *LockerUpdateOne) SaveX(ctx context.Context) *Locker {
	node, err := luo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (luo *LockerUpdateOne) Exec(ctx context.Context) error {
	_, err := luo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (luo *LockerUpdateOne) ExecX(ctx context.Context) {
	if err := luo.Exec(ctx); err != nil {
		panic(err)
	}
}

func (luo *LockerUpdateOne) sqlSave(ctx context.Context) (_node *Locker, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   locker.Table,
			Columns: locker.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: locker.FieldID,
			},
		},
	}
	id, ok := luo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Locker.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := luo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, locker.FieldID)
		for _, f := range fields {
			if !locker.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != locker.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := luo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := luo.mutation.Hash(); ok {
		_spec.SetField(locker.FieldHash, field.TypeString, value)
	}
	if value, ok := luo.mutation.Level(); ok {
		_spec.SetField(locker.FieldLevel, field.TypeInt32, value)
	}
	if value, ok := luo.mutation.AddedLevel(); ok {
		_spec.AddField(locker.FieldLevel, field.TypeInt32, value)
	}
	if value, ok := luo.mutation.EncryptedID(); ok {
		_spec.SetField(locker.FieldEncryptedID, field.TypeString, value)
	}
	if value, ok := luo.mutation.EncryptedBody(); ok {
		_spec.SetField(locker.FieldEncryptedBody, field.TypeString, value)
	}
	if luo.mutation.AccountCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   locker.AccountTable,
			Columns: []string{locker.AccountColumn},
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
	if nodes := luo.mutation.AccountIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   locker.AccountTable,
			Columns: []string{locker.AccountColumn},
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
	_node = &Locker{config: luo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, luo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{locker.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	luo.mutation.done = true
	return _node, nil
}
