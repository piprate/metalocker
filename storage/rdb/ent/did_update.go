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
	"github.com/piprate/metalocker/storage/rdb/ent/did"
	"github.com/piprate/metalocker/storage/rdb/ent/predicate"
)

// DIDUpdate is the builder for updating DID entities.
type DIDUpdate struct {
	config
	hooks    []Hook
	mutation *DIDMutation
}

// Where appends a list predicates to the DIDUpdate builder.
func (du *DIDUpdate) Where(ps ...predicate.DID) *DIDUpdate {
	du.mutation.Where(ps...)
	return du
}

// SetDid sets the "did" field.
func (du *DIDUpdate) SetDid(s string) *DIDUpdate {
	du.mutation.SetDid(s)
	return du
}

// SetBody sets the "body" field.
func (du *DIDUpdate) SetBody(md *model.DIDDocument) *DIDUpdate {
	du.mutation.SetBody(md)
	return du
}

// Mutation returns the DIDMutation object of the builder.
func (du *DIDUpdate) Mutation() *DIDMutation {
	return du.mutation
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (du *DIDUpdate) Save(ctx context.Context) (int, error) {
	return withHooks[int, DIDMutation](ctx, du.sqlSave, du.mutation, du.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (du *DIDUpdate) SaveX(ctx context.Context) int {
	affected, err := du.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (du *DIDUpdate) Exec(ctx context.Context) error {
	_, err := du.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (du *DIDUpdate) ExecX(ctx context.Context) {
	if err := du.Exec(ctx); err != nil {
		panic(err)
	}
}

func (du *DIDUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   did.Table,
			Columns: did.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: did.FieldID,
			},
		},
	}
	if ps := du.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := du.mutation.Did(); ok {
		_spec.SetField(did.FieldDid, field.TypeString, value)
	}
	if value, ok := du.mutation.Body(); ok {
		_spec.SetField(did.FieldBody, field.TypeJSON, value)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, du.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{did.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	du.mutation.done = true
	return n, nil
}

// DIDUpdateOne is the builder for updating a single DID entity.
type DIDUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *DIDMutation
}

// SetDid sets the "did" field.
func (duo *DIDUpdateOne) SetDid(s string) *DIDUpdateOne {
	duo.mutation.SetDid(s)
	return duo
}

// SetBody sets the "body" field.
func (duo *DIDUpdateOne) SetBody(md *model.DIDDocument) *DIDUpdateOne {
	duo.mutation.SetBody(md)
	return duo
}

// Mutation returns the DIDMutation object of the builder.
func (duo *DIDUpdateOne) Mutation() *DIDMutation {
	return duo.mutation
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (duo *DIDUpdateOne) Select(field string, fields ...string) *DIDUpdateOne {
	duo.fields = append([]string{field}, fields...)
	return duo
}

// Save executes the query and returns the updated DID entity.
func (duo *DIDUpdateOne) Save(ctx context.Context) (*DID, error) {
	return withHooks[*DID, DIDMutation](ctx, duo.sqlSave, duo.mutation, duo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (duo *DIDUpdateOne) SaveX(ctx context.Context) *DID {
	node, err := duo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (duo *DIDUpdateOne) Exec(ctx context.Context) error {
	_, err := duo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (duo *DIDUpdateOne) ExecX(ctx context.Context) {
	if err := duo.Exec(ctx); err != nil {
		panic(err)
	}
}

func (duo *DIDUpdateOne) sqlSave(ctx context.Context) (_node *DID, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   did.Table,
			Columns: did.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: did.FieldID,
			},
		},
	}
	id, ok := duo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "DID.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := duo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, did.FieldID)
		for _, f := range fields {
			if !did.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != did.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := duo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := duo.mutation.Did(); ok {
		_spec.SetField(did.FieldDid, field.TypeString, value)
	}
	if value, ok := duo.mutation.Body(); ok {
		_spec.SetField(did.FieldBody, field.TypeJSON, value)
	}
	_node = &DID{config: duo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, duo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{did.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	duo.mutation.done = true
	return _node, nil
}
