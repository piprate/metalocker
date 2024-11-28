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
	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/storage/rdb/ent/did"
)

// DIDCreate is the builder for creating a DID entity.
type DIDCreate struct {
	config
	mutation *DIDMutation
	hooks    []Hook
}

// SetDid sets the "did" field.
func (dc *DIDCreate) SetDid(s string) *DIDCreate {
	dc.mutation.SetDid(s)
	return dc
}

// SetBody sets the "body" field.
func (dc *DIDCreate) SetBody(md *model.DIDDocument) *DIDCreate {
	dc.mutation.SetBody(md)
	return dc
}

// Mutation returns the DIDMutation object of the builder.
func (dc *DIDCreate) Mutation() *DIDMutation {
	return dc.mutation
}

// Save creates the DID in the database.
func (dc *DIDCreate) Save(ctx context.Context) (*DID, error) {
	return withHooks[*DID, DIDMutation](ctx, dc.sqlSave, dc.mutation, dc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (dc *DIDCreate) SaveX(ctx context.Context) *DID {
	v, err := dc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (dc *DIDCreate) Exec(ctx context.Context) error {
	_, err := dc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (dc *DIDCreate) ExecX(ctx context.Context) {
	if err := dc.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (dc *DIDCreate) check() error {
	if _, ok := dc.mutation.Did(); !ok {
		return &ValidationError{Name: "did", err: errors.New(`ent: missing required field "DID.did"`)}
	}
	if _, ok := dc.mutation.Body(); !ok {
		return &ValidationError{Name: "body", err: errors.New(`ent: missing required field "DID.body"`)}
	}
	return nil
}

func (dc *DIDCreate) sqlSave(ctx context.Context) (*DID, error) {
	if err := dc.check(); err != nil {
		return nil, err
	}
	_node, _spec := dc.createSpec()
	if err := sqlgraph.CreateNode(ctx, dc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	dc.mutation.id = &_node.ID
	dc.mutation.done = true
	return _node, nil
}

func (dc *DIDCreate) createSpec() (*DID, *sqlgraph.CreateSpec) {
	var (
		_node = &DID{config: dc.config}
		_spec = &sqlgraph.CreateSpec{
			Table: did.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: did.FieldID,
			},
		}
	)
	if value, ok := dc.mutation.Did(); ok {
		_spec.SetField(did.FieldDid, field.TypeString, value)
		_node.Did = value
	}
	if value, ok := dc.mutation.Body(); ok {
		_spec.SetField(did.FieldBody, field.TypeJSON, value)
		_node.Body = value
	}
	return _node, _spec
}

// DIDCreateBulk is the builder for creating many DID entities in bulk.
type DIDCreateBulk struct {
	config
	builders []*DIDCreate
}

// Save creates the DID entities in the database.
func (dcb *DIDCreateBulk) Save(ctx context.Context) ([]*DID, error) {
	specs := make([]*sqlgraph.CreateSpec, len(dcb.builders))
	nodes := make([]*DID, len(dcb.builders))
	mutators := make([]Mutator, len(dcb.builders))
	for i := range dcb.builders {
		func(i int, root context.Context) {
			builder := dcb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*DIDMutation)
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
					_, err = mutators[i+1].Mutate(root, dcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, dcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, dcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (dcb *DIDCreateBulk) SaveX(ctx context.Context) []*DID {
	v, err := dcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (dcb *DIDCreateBulk) Exec(ctx context.Context) error {
	_, err := dcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (dcb *DIDCreateBulk) ExecX(ctx context.Context) {
	if err := dcb.Exec(ctx); err != nil {
		panic(err)
	}
}
