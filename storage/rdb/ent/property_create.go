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

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/piprate/metalocker/storage/rdb/ent/property"

	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
)

// PropertyCreate is the builder for creating a Property entity.
type PropertyCreate struct {
	config
	mutation *PropertyMutation
	hooks    []Hook
}

// SetHash sets the "hash" field.
func (pc *PropertyCreate) SetHash(s string) *PropertyCreate {
	pc.mutation.SetHash(s)
	return pc
}

// SetLevel sets the "level" field.
func (pc *PropertyCreate) SetLevel(i int32) *PropertyCreate {
	pc.mutation.SetLevel(i)
	return pc
}

// SetEncryptedID sets the "encrypted_id" field.
func (pc *PropertyCreate) SetEncryptedID(s string) *PropertyCreate {
	pc.mutation.SetEncryptedID(s)
	return pc
}

// SetEncryptedBody sets the "encrypted_body" field.
func (pc *PropertyCreate) SetEncryptedBody(s string) *PropertyCreate {
	pc.mutation.SetEncryptedBody(s)
	return pc
}

// SetAccountID sets the "account" edge to the Account entity by ID.
func (pc *PropertyCreate) SetAccountID(id int) *PropertyCreate {
	pc.mutation.SetAccountID(id)
	return pc
}

// SetNillableAccountID sets the "account" edge to the Account entity by ID if the given value is not nil.
func (pc *PropertyCreate) SetNillableAccountID(id *int) *PropertyCreate {
	if id != nil {
		pc = pc.SetAccountID(*id)
	}
	return pc
}

// SetAccount sets the "account" edge to the Account entity.
func (pc *PropertyCreate) SetAccount(a *Account) *PropertyCreate {
	return pc.SetAccountID(a.ID)
}

// Mutation returns the PropertyMutation object of the builder.
func (pc *PropertyCreate) Mutation() *PropertyMutation {
	return pc.mutation
}

// Save creates the Property in the database.
func (pc *PropertyCreate) Save(ctx context.Context) (*Property, error) {
	return withHooks[*Property, PropertyMutation](ctx, pc.sqlSave, pc.mutation, pc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (pc *PropertyCreate) SaveX(ctx context.Context) *Property {
	v, err := pc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (pc *PropertyCreate) Exec(ctx context.Context) error {
	_, err := pc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (pc *PropertyCreate) ExecX(ctx context.Context) {
	if err := pc.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (pc *PropertyCreate) check() error {
	if _, ok := pc.mutation.Hash(); !ok {
		return &ValidationError{Name: "hash", err: errors.New(`ent: missing required field "Property.hash"`)}
	}
	if _, ok := pc.mutation.Level(); !ok {
		return &ValidationError{Name: "level", err: errors.New(`ent: missing required field "Property.level"`)}
	}
	if _, ok := pc.mutation.EncryptedID(); !ok {
		return &ValidationError{Name: "encrypted_id", err: errors.New(`ent: missing required field "Property.encrypted_id"`)}
	}
	if _, ok := pc.mutation.EncryptedBody(); !ok {
		return &ValidationError{Name: "encrypted_body", err: errors.New(`ent: missing required field "Property.encrypted_body"`)}
	}
	return nil
}

func (pc *PropertyCreate) sqlSave(ctx context.Context) (*Property, error) {
	if err := pc.check(); err != nil {
		return nil, err
	}
	_node, _spec := pc.createSpec()
	if err := sqlgraph.CreateNode(ctx, pc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	pc.mutation.id = &_node.ID
	pc.mutation.done = true
	return _node, nil
}

func (pc *PropertyCreate) createSpec() (*Property, *sqlgraph.CreateSpec) {
	var (
		_node = &Property{config: pc.config}
		_spec = &sqlgraph.CreateSpec{
			Table: property.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: property.FieldID,
			},
		}
	)
	if value, ok := pc.mutation.Hash(); ok {
		_spec.SetField(property.FieldHash, field.TypeString, value)
		_node.Hash = value
	}
	if value, ok := pc.mutation.Level(); ok {
		_spec.SetField(property.FieldLevel, field.TypeInt32, value)
		_node.Level = value
	}
	if value, ok := pc.mutation.EncryptedID(); ok {
		_spec.SetField(property.FieldEncryptedID, field.TypeString, value)
		_node.EncryptedID = value
	}
	if value, ok := pc.mutation.EncryptedBody(); ok {
		_spec.SetField(property.FieldEncryptedBody, field.TypeString, value)
		_node.EncryptedBody = value
	}
	if nodes := pc.mutation.AccountIDs(); len(nodes) > 0 {
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
		_node.account = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// PropertyCreateBulk is the builder for creating many Property entities in bulk.
type PropertyCreateBulk struct {
	config
	builders []*PropertyCreate
}

// Save creates the Property entities in the database.
func (pcb *PropertyCreateBulk) Save(ctx context.Context) ([]*Property, error) {
	specs := make([]*sqlgraph.CreateSpec, len(pcb.builders))
	nodes := make([]*Property, len(pcb.builders))
	mutators := make([]Mutator, len(pcb.builders))
	for i := range pcb.builders {
		func(i int, root context.Context) {
			builder := pcb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*PropertyMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				nodes[i], specs[i] = builder.createSpec()
				var err error
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, pcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, pcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				if specs[i].ID.Value != nil {
					id := specs[i].ID.Value.(int64)
					nodes[i].ID = int(id)
				}
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, pcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (pcb *PropertyCreateBulk) SaveX(ctx context.Context) []*Property {
	v, err := pcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (pcb *PropertyCreateBulk) Exec(ctx context.Context) error {
	_, err := pcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (pcb *PropertyCreateBulk) ExecX(ctx context.Context) {
	if err := pcb.Exec(ctx); err != nil {
		panic(err)
	}
}
