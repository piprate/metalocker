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
	"github.com/piprate/metalocker/storage/rdb/ent/identity"
	"github.com/piprate/metalocker/storage/rdb/ent/predicate"

	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
)

// IdentityUpdate is the builder for updating Identity entities.
type IdentityUpdate struct {
	config
	hooks    []Hook
	mutation *IdentityMutation
}

// Where appends a list predicates to the IdentityUpdate builder.
func (iu *IdentityUpdate) Where(ps ...predicate.Identity) *IdentityUpdate {
	iu.mutation.Where(ps...)
	return iu
}

// SetHash sets the "hash" field.
func (iu *IdentityUpdate) SetHash(s string) *IdentityUpdate {
	iu.mutation.SetHash(s)
	return iu
}

// SetLevel sets the "level" field.
func (iu *IdentityUpdate) SetLevel(i int32) *IdentityUpdate {
	iu.mutation.ResetLevel()
	iu.mutation.SetLevel(i)
	return iu
}

// AddLevel adds i to the "level" field.
func (iu *IdentityUpdate) AddLevel(i int32) *IdentityUpdate {
	iu.mutation.AddLevel(i)
	return iu
}

// SetEncryptedID sets the "encrypted_id" field.
func (iu *IdentityUpdate) SetEncryptedID(s string) *IdentityUpdate {
	iu.mutation.SetEncryptedID(s)
	return iu
}

// SetEncryptedBody sets the "encrypted_body" field.
func (iu *IdentityUpdate) SetEncryptedBody(s string) *IdentityUpdate {
	iu.mutation.SetEncryptedBody(s)
	return iu
}

// SetAccountID sets the "account" edge to the Account entity by ID.
func (iu *IdentityUpdate) SetAccountID(id int) *IdentityUpdate {
	iu.mutation.SetAccountID(id)
	return iu
}

// SetNillableAccountID sets the "account" edge to the Account entity by ID if the given value is not nil.
func (iu *IdentityUpdate) SetNillableAccountID(id *int) *IdentityUpdate {
	if id != nil {
		iu = iu.SetAccountID(*id)
	}
	return iu
}

// SetAccount sets the "account" edge to the Account entity.
func (iu *IdentityUpdate) SetAccount(a *Account) *IdentityUpdate {
	return iu.SetAccountID(a.ID)
}

// Mutation returns the IdentityMutation object of the builder.
func (iu *IdentityUpdate) Mutation() *IdentityMutation {
	return iu.mutation
}

// ClearAccount clears the "account" edge to the Account entity.
func (iu *IdentityUpdate) ClearAccount() *IdentityUpdate {
	iu.mutation.ClearAccount()
	return iu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (iu *IdentityUpdate) Save(ctx context.Context) (int, error) {
	return withHooks[int, IdentityMutation](ctx, iu.sqlSave, iu.mutation, iu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (iu *IdentityUpdate) SaveX(ctx context.Context) int {
	affected, err := iu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (iu *IdentityUpdate) Exec(ctx context.Context) error {
	_, err := iu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (iu *IdentityUpdate) ExecX(ctx context.Context) {
	if err := iu.Exec(ctx); err != nil {
		panic(err)
	}
}

func (iu *IdentityUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   identity.Table,
			Columns: identity.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: identity.FieldID,
			},
		},
	}
	if ps := iu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := iu.mutation.Hash(); ok {
		_spec.SetField(identity.FieldHash, field.TypeString, value)
	}
	if value, ok := iu.mutation.Level(); ok {
		_spec.SetField(identity.FieldLevel, field.TypeInt32, value)
	}
	if value, ok := iu.mutation.AddedLevel(); ok {
		_spec.AddField(identity.FieldLevel, field.TypeInt32, value)
	}
	if value, ok := iu.mutation.EncryptedID(); ok {
		_spec.SetField(identity.FieldEncryptedID, field.TypeString, value)
	}
	if value, ok := iu.mutation.EncryptedBody(); ok {
		_spec.SetField(identity.FieldEncryptedBody, field.TypeString, value)
	}
	if iu.mutation.AccountCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   identity.AccountTable,
			Columns: []string{identity.AccountColumn},
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
	if nodes := iu.mutation.AccountIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   identity.AccountTable,
			Columns: []string{identity.AccountColumn},
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
	if n, err = sqlgraph.UpdateNodes(ctx, iu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{identity.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	iu.mutation.done = true
	return n, nil
}

// IdentityUpdateOne is the builder for updating a single Identity entity.
type IdentityUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *IdentityMutation
}

// SetHash sets the "hash" field.
func (iuo *IdentityUpdateOne) SetHash(s string) *IdentityUpdateOne {
	iuo.mutation.SetHash(s)
	return iuo
}

// SetLevel sets the "level" field.
func (iuo *IdentityUpdateOne) SetLevel(i int32) *IdentityUpdateOne {
	iuo.mutation.ResetLevel()
	iuo.mutation.SetLevel(i)
	return iuo
}

// AddLevel adds i to the "level" field.
func (iuo *IdentityUpdateOne) AddLevel(i int32) *IdentityUpdateOne {
	iuo.mutation.AddLevel(i)
	return iuo
}

// SetEncryptedID sets the "encrypted_id" field.
func (iuo *IdentityUpdateOne) SetEncryptedID(s string) *IdentityUpdateOne {
	iuo.mutation.SetEncryptedID(s)
	return iuo
}

// SetEncryptedBody sets the "encrypted_body" field.
func (iuo *IdentityUpdateOne) SetEncryptedBody(s string) *IdentityUpdateOne {
	iuo.mutation.SetEncryptedBody(s)
	return iuo
}

// SetAccountID sets the "account" edge to the Account entity by ID.
func (iuo *IdentityUpdateOne) SetAccountID(id int) *IdentityUpdateOne {
	iuo.mutation.SetAccountID(id)
	return iuo
}

// SetNillableAccountID sets the "account" edge to the Account entity by ID if the given value is not nil.
func (iuo *IdentityUpdateOne) SetNillableAccountID(id *int) *IdentityUpdateOne {
	if id != nil {
		iuo = iuo.SetAccountID(*id)
	}
	return iuo
}

// SetAccount sets the "account" edge to the Account entity.
func (iuo *IdentityUpdateOne) SetAccount(a *Account) *IdentityUpdateOne {
	return iuo.SetAccountID(a.ID)
}

// Mutation returns the IdentityMutation object of the builder.
func (iuo *IdentityUpdateOne) Mutation() *IdentityMutation {
	return iuo.mutation
}

// ClearAccount clears the "account" edge to the Account entity.
func (iuo *IdentityUpdateOne) ClearAccount() *IdentityUpdateOne {
	iuo.mutation.ClearAccount()
	return iuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (iuo *IdentityUpdateOne) Select(field string, fields ...string) *IdentityUpdateOne {
	iuo.fields = append([]string{field}, fields...)
	return iuo
}

// Save executes the query and returns the updated Identity entity.
func (iuo *IdentityUpdateOne) Save(ctx context.Context) (*Identity, error) {
	return withHooks[*Identity, IdentityMutation](ctx, iuo.sqlSave, iuo.mutation, iuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (iuo *IdentityUpdateOne) SaveX(ctx context.Context) *Identity {
	node, err := iuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (iuo *IdentityUpdateOne) Exec(ctx context.Context) error {
	_, err := iuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (iuo *IdentityUpdateOne) ExecX(ctx context.Context) {
	if err := iuo.Exec(ctx); err != nil {
		panic(err)
	}
}

func (iuo *IdentityUpdateOne) sqlSave(ctx context.Context) (_node *Identity, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   identity.Table,
			Columns: identity.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: identity.FieldID,
			},
		},
	}
	id, ok := iuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Identity.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := iuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, identity.FieldID)
		for _, f := range fields {
			if !identity.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != identity.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := iuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := iuo.mutation.Hash(); ok {
		_spec.SetField(identity.FieldHash, field.TypeString, value)
	}
	if value, ok := iuo.mutation.Level(); ok {
		_spec.SetField(identity.FieldLevel, field.TypeInt32, value)
	}
	if value, ok := iuo.mutation.AddedLevel(); ok {
		_spec.AddField(identity.FieldLevel, field.TypeInt32, value)
	}
	if value, ok := iuo.mutation.EncryptedID(); ok {
		_spec.SetField(identity.FieldEncryptedID, field.TypeString, value)
	}
	if value, ok := iuo.mutation.EncryptedBody(); ok {
		_spec.SetField(identity.FieldEncryptedBody, field.TypeString, value)
	}
	if iuo.mutation.AccountCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   identity.AccountTable,
			Columns: []string{identity.AccountColumn},
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
	if nodes := iuo.mutation.AccountIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   identity.AccountTable,
			Columns: []string{identity.AccountColumn},
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
	_node = &Identity{config: iuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, iuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{identity.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	iuo.mutation.done = true
	return _node, nil
}
