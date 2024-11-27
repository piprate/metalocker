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
	"github.com/piprate/metalocker/storage/rdb/ent/predicate"
	"github.com/piprate/metalocker/storage/rdb/ent/property"

	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
)

// PropertyUpdate is the builder for updating Property entities.
type PropertyUpdate struct {
	config
	hooks    []Hook
	mutation *PropertyMutation
}

// Where appends a list predicates to the PropertyUpdate builder.
func (pu *PropertyUpdate) Where(ps ...predicate.Property) *PropertyUpdate {
	pu.mutation.Where(ps...)
	return pu
}

// SetHash sets the "hash" field.
func (pu *PropertyUpdate) SetHash(s string) *PropertyUpdate {
	pu.mutation.SetHash(s)
	return pu
}

// SetLevel sets the "level" field.
func (pu *PropertyUpdate) SetLevel(i int32) *PropertyUpdate {
	pu.mutation.ResetLevel()
	pu.mutation.SetLevel(i)
	return pu
}

// AddLevel adds i to the "level" field.
func (pu *PropertyUpdate) AddLevel(i int32) *PropertyUpdate {
	pu.mutation.AddLevel(i)
	return pu
}

// SetEncryptedID sets the "encrypted_id" field.
func (pu *PropertyUpdate) SetEncryptedID(s string) *PropertyUpdate {
	pu.mutation.SetEncryptedID(s)
	return pu
}

// SetEncryptedBody sets the "encrypted_body" field.
func (pu *PropertyUpdate) SetEncryptedBody(s string) *PropertyUpdate {
	pu.mutation.SetEncryptedBody(s)
	return pu
}

// SetAccountID sets the "account" edge to the Account entity by ID.
func (pu *PropertyUpdate) SetAccountID(id int) *PropertyUpdate {
	pu.mutation.SetAccountID(id)
	return pu
}

// SetNillableAccountID sets the "account" edge to the Account entity by ID if the given value is not nil.
func (pu *PropertyUpdate) SetNillableAccountID(id *int) *PropertyUpdate {
	if id != nil {
		pu = pu.SetAccountID(*id)
	}
	return pu
}

// SetAccount sets the "account" edge to the Account entity.
func (pu *PropertyUpdate) SetAccount(a *Account) *PropertyUpdate {
	return pu.SetAccountID(a.ID)
}

// Mutation returns the PropertyMutation object of the builder.
func (pu *PropertyUpdate) Mutation() *PropertyMutation {
	return pu.mutation
}

// ClearAccount clears the "account" edge to the Account entity.
func (pu *PropertyUpdate) ClearAccount() *PropertyUpdate {
	pu.mutation.ClearAccount()
	return pu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (pu *PropertyUpdate) Save(ctx context.Context) (int, error) {
	return withHooks[int, PropertyMutation](ctx, pu.sqlSave, pu.mutation, pu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (pu *PropertyUpdate) SaveX(ctx context.Context) int {
	affected, err := pu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (pu *PropertyUpdate) Exec(ctx context.Context) error {
	_, err := pu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (pu *PropertyUpdate) ExecX(ctx context.Context) {
	if err := pu.Exec(ctx); err != nil {
		panic(err)
	}
}

func (pu *PropertyUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   property.Table,
			Columns: property.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: property.FieldID,
			},
		},
	}
	if ps := pu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := pu.mutation.Hash(); ok {
		_spec.SetField(property.FieldHash, field.TypeString, value)
	}
	if value, ok := pu.mutation.Level(); ok {
		_spec.SetField(property.FieldLevel, field.TypeInt32, value)
	}
	if value, ok := pu.mutation.AddedLevel(); ok {
		_spec.AddField(property.FieldLevel, field.TypeInt32, value)
	}
	if value, ok := pu.mutation.EncryptedID(); ok {
		_spec.SetField(property.FieldEncryptedID, field.TypeString, value)
	}
	if value, ok := pu.mutation.EncryptedBody(); ok {
		_spec.SetField(property.FieldEncryptedBody, field.TypeString, value)
	}
	if pu.mutation.AccountCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   property.AccountTable,
			Columns: []string{property.AccountColumn},
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
	if nodes := pu.mutation.AccountIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   property.AccountTable,
			Columns: []string{property.AccountColumn},
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
	if n, err = sqlgraph.UpdateNodes(ctx, pu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{property.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	pu.mutation.done = true
	return n, nil
}

// PropertyUpdateOne is the builder for updating a single Property entity.
type PropertyUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *PropertyMutation
}

// SetHash sets the "hash" field.
func (puo *PropertyUpdateOne) SetHash(s string) *PropertyUpdateOne {
	puo.mutation.SetHash(s)
	return puo
}

// SetLevel sets the "level" field.
func (puo *PropertyUpdateOne) SetLevel(i int32) *PropertyUpdateOne {
	puo.mutation.ResetLevel()
	puo.mutation.SetLevel(i)
	return puo
}

// AddLevel adds i to the "level" field.
func (puo *PropertyUpdateOne) AddLevel(i int32) *PropertyUpdateOne {
	puo.mutation.AddLevel(i)
	return puo
}

// SetEncryptedID sets the "encrypted_id" field.
func (puo *PropertyUpdateOne) SetEncryptedID(s string) *PropertyUpdateOne {
	puo.mutation.SetEncryptedID(s)
	return puo
}

// SetEncryptedBody sets the "encrypted_body" field.
func (puo *PropertyUpdateOne) SetEncryptedBody(s string) *PropertyUpdateOne {
	puo.mutation.SetEncryptedBody(s)
	return puo
}

// SetAccountID sets the "account" edge to the Account entity by ID.
func (puo *PropertyUpdateOne) SetAccountID(id int) *PropertyUpdateOne {
	puo.mutation.SetAccountID(id)
	return puo
}

// SetNillableAccountID sets the "account" edge to the Account entity by ID if the given value is not nil.
func (puo *PropertyUpdateOne) SetNillableAccountID(id *int) *PropertyUpdateOne {
	if id != nil {
		puo = puo.SetAccountID(*id)
	}
	return puo
}

// SetAccount sets the "account" edge to the Account entity.
func (puo *PropertyUpdateOne) SetAccount(a *Account) *PropertyUpdateOne {
	return puo.SetAccountID(a.ID)
}

// Mutation returns the PropertyMutation object of the builder.
func (puo *PropertyUpdateOne) Mutation() *PropertyMutation {
	return puo.mutation
}

// ClearAccount clears the "account" edge to the Account entity.
func (puo *PropertyUpdateOne) ClearAccount() *PropertyUpdateOne {
	puo.mutation.ClearAccount()
	return puo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (puo *PropertyUpdateOne) Select(field string, fields ...string) *PropertyUpdateOne {
	puo.fields = append([]string{field}, fields...)
	return puo
}

// Save executes the query and returns the updated Property entity.
func (puo *PropertyUpdateOne) Save(ctx context.Context) (*Property, error) {
	return withHooks[*Property, PropertyMutation](ctx, puo.sqlSave, puo.mutation, puo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (puo *PropertyUpdateOne) SaveX(ctx context.Context) *Property {
	node, err := puo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (puo *PropertyUpdateOne) Exec(ctx context.Context) error {
	_, err := puo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (puo *PropertyUpdateOne) ExecX(ctx context.Context) {
	if err := puo.Exec(ctx); err != nil {
		panic(err)
	}
}

func (puo *PropertyUpdateOne) sqlSave(ctx context.Context) (_node *Property, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   property.Table,
			Columns: property.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: property.FieldID,
			},
		},
	}
	id, ok := puo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Property.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := puo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, property.FieldID)
		for _, f := range fields {
			if !property.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != property.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := puo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := puo.mutation.Hash(); ok {
		_spec.SetField(property.FieldHash, field.TypeString, value)
	}
	if value, ok := puo.mutation.Level(); ok {
		_spec.SetField(property.FieldLevel, field.TypeInt32, value)
	}
	if value, ok := puo.mutation.AddedLevel(); ok {
		_spec.AddField(property.FieldLevel, field.TypeInt32, value)
	}
	if value, ok := puo.mutation.EncryptedID(); ok {
		_spec.SetField(property.FieldEncryptedID, field.TypeString, value)
	}
	if value, ok := puo.mutation.EncryptedBody(); ok {
		_spec.SetField(property.FieldEncryptedBody, field.TypeString, value)
	}
	if puo.mutation.AccountCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   property.AccountTable,
			Columns: []string{property.AccountColumn},
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
	if nodes := puo.mutation.AccountIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   property.AccountTable,
			Columns: []string{property.AccountColumn},
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
	_node = &Property{config: puo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, puo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{property.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	puo.mutation.done = true
	return _node, nil
}
