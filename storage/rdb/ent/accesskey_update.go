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
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/storage/rdb/ent/accesskey"
	"github.com/piprate/metalocker/storage/rdb/ent/predicate"

	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
)

// AccessKeyUpdate is the builder for updating AccessKey entities.
type AccessKeyUpdate struct {
	config
	hooks    []Hook
	mutation *AccessKeyMutation
}

// Where appends a list predicates to the AccessKeyUpdate builder.
func (aku *AccessKeyUpdate) Where(ps ...predicate.AccessKey) *AccessKeyUpdate {
	aku.mutation.Where(ps...)
	return aku
}

// SetDid sets the "did" field.
func (aku *AccessKeyUpdate) SetDid(s string) *AccessKeyUpdate {
	aku.mutation.SetDid(s)
	return aku
}

// SetBody sets the "body" field.
func (aku *AccessKeyUpdate) SetBody(mk *model.AccessKey) *AccessKeyUpdate {
	aku.mutation.SetBody(mk)
	return aku
}

// SetAccountID sets the "account" edge to the Account entity by ID.
func (aku *AccessKeyUpdate) SetAccountID(id int) *AccessKeyUpdate {
	aku.mutation.SetAccountID(id)
	return aku
}

// SetNillableAccountID sets the "account" edge to the Account entity by ID if the given value is not nil.
func (aku *AccessKeyUpdate) SetNillableAccountID(id *int) *AccessKeyUpdate {
	if id != nil {
		aku = aku.SetAccountID(*id)
	}
	return aku
}

// SetAccount sets the "account" edge to the Account entity.
func (aku *AccessKeyUpdate) SetAccount(a *Account) *AccessKeyUpdate {
	return aku.SetAccountID(a.ID)
}

// Mutation returns the AccessKeyMutation object of the builder.
func (aku *AccessKeyUpdate) Mutation() *AccessKeyMutation {
	return aku.mutation
}

// ClearAccount clears the "account" edge to the Account entity.
func (aku *AccessKeyUpdate) ClearAccount() *AccessKeyUpdate {
	aku.mutation.ClearAccount()
	return aku
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (aku *AccessKeyUpdate) Save(ctx context.Context) (int, error) {
	return withHooks[int, AccessKeyMutation](ctx, aku.sqlSave, aku.mutation, aku.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (aku *AccessKeyUpdate) SaveX(ctx context.Context) int {
	affected, err := aku.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (aku *AccessKeyUpdate) Exec(ctx context.Context) error {
	_, err := aku.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (aku *AccessKeyUpdate) ExecX(ctx context.Context) {
	if err := aku.Exec(ctx); err != nil {
		panic(err)
	}
}

func (aku *AccessKeyUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   accesskey.Table,
			Columns: accesskey.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: accesskey.FieldID,
			},
		},
	}
	if ps := aku.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := aku.mutation.Did(); ok {
		_spec.SetField(accesskey.FieldDid, field.TypeString, value)
	}
	if value, ok := aku.mutation.Body(); ok {
		_spec.SetField(accesskey.FieldBody, field.TypeJSON, value)
	}
	if aku.mutation.AccountCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   accesskey.AccountTable,
			Columns: []string{accesskey.AccountColumn},
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
	if nodes := aku.mutation.AccountIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   accesskey.AccountTable,
			Columns: []string{accesskey.AccountColumn},
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
	if n, err = sqlgraph.UpdateNodes(ctx, aku.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{accesskey.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	aku.mutation.done = true
	return n, nil
}

// AccessKeyUpdateOne is the builder for updating a single AccessKey entity.
type AccessKeyUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *AccessKeyMutation
}

// SetDid sets the "did" field.
func (akuo *AccessKeyUpdateOne) SetDid(s string) *AccessKeyUpdateOne {
	akuo.mutation.SetDid(s)
	return akuo
}

// SetBody sets the "body" field.
func (akuo *AccessKeyUpdateOne) SetBody(mk *model.AccessKey) *AccessKeyUpdateOne {
	akuo.mutation.SetBody(mk)
	return akuo
}

// SetAccountID sets the "account" edge to the Account entity by ID.
func (akuo *AccessKeyUpdateOne) SetAccountID(id int) *AccessKeyUpdateOne {
	akuo.mutation.SetAccountID(id)
	return akuo
}

// SetNillableAccountID sets the "account" edge to the Account entity by ID if the given value is not nil.
func (akuo *AccessKeyUpdateOne) SetNillableAccountID(id *int) *AccessKeyUpdateOne {
	if id != nil {
		akuo = akuo.SetAccountID(*id)
	}
	return akuo
}

// SetAccount sets the "account" edge to the Account entity.
func (akuo *AccessKeyUpdateOne) SetAccount(a *Account) *AccessKeyUpdateOne {
	return akuo.SetAccountID(a.ID)
}

// Mutation returns the AccessKeyMutation object of the builder.
func (akuo *AccessKeyUpdateOne) Mutation() *AccessKeyMutation {
	return akuo.mutation
}

// ClearAccount clears the "account" edge to the Account entity.
func (akuo *AccessKeyUpdateOne) ClearAccount() *AccessKeyUpdateOne {
	akuo.mutation.ClearAccount()
	return akuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (akuo *AccessKeyUpdateOne) Select(field string, fields ...string) *AccessKeyUpdateOne {
	akuo.fields = append([]string{field}, fields...)
	return akuo
}

// Save executes the query and returns the updated AccessKey entity.
func (akuo *AccessKeyUpdateOne) Save(ctx context.Context) (*AccessKey, error) {
	return withHooks[*AccessKey, AccessKeyMutation](ctx, akuo.sqlSave, akuo.mutation, akuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (akuo *AccessKeyUpdateOne) SaveX(ctx context.Context) *AccessKey {
	node, err := akuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (akuo *AccessKeyUpdateOne) Exec(ctx context.Context) error {
	_, err := akuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (akuo *AccessKeyUpdateOne) ExecX(ctx context.Context) {
	if err := akuo.Exec(ctx); err != nil {
		panic(err)
	}
}

func (akuo *AccessKeyUpdateOne) sqlSave(ctx context.Context) (_node *AccessKey, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   accesskey.Table,
			Columns: accesskey.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: accesskey.FieldID,
			},
		},
	}
	id, ok := akuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "AccessKey.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := akuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, accesskey.FieldID)
		for _, f := range fields {
			if !accesskey.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != accesskey.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := akuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := akuo.mutation.Did(); ok {
		_spec.SetField(accesskey.FieldDid, field.TypeString, value)
	}
	if value, ok := akuo.mutation.Body(); ok {
		_spec.SetField(accesskey.FieldBody, field.TypeJSON, value)
	}
	if akuo.mutation.AccountCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   accesskey.AccountTable,
			Columns: []string{accesskey.AccountColumn},
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
	if nodes := akuo.mutation.AccountIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   accesskey.AccountTable,
			Columns: []string{accesskey.AccountColumn},
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
	_node = &AccessKey{config: akuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, akuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{accesskey.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	akuo.mutation.done = true
	return _node, nil
}
