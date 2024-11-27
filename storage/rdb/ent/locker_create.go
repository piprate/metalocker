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
	"github.com/piprate/metalocker/storage/rdb/ent/locker"

	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
)

// LockerCreate is the builder for creating a Locker entity.
type LockerCreate struct {
	config
	mutation *LockerMutation
	hooks    []Hook
}

// SetHash sets the "hash" field.
func (lc *LockerCreate) SetHash(s string) *LockerCreate {
	lc.mutation.SetHash(s)
	return lc
}

// SetLevel sets the "level" field.
func (lc *LockerCreate) SetLevel(i int32) *LockerCreate {
	lc.mutation.SetLevel(i)
	return lc
}

// SetEncryptedID sets the "encrypted_id" field.
func (lc *LockerCreate) SetEncryptedID(s string) *LockerCreate {
	lc.mutation.SetEncryptedID(s)
	return lc
}

// SetEncryptedBody sets the "encrypted_body" field.
func (lc *LockerCreate) SetEncryptedBody(s string) *LockerCreate {
	lc.mutation.SetEncryptedBody(s)
	return lc
}

// SetAccountID sets the "account" edge to the Account entity by ID.
func (lc *LockerCreate) SetAccountID(id int) *LockerCreate {
	lc.mutation.SetAccountID(id)
	return lc
}

// SetNillableAccountID sets the "account" edge to the Account entity by ID if the given value is not nil.
func (lc *LockerCreate) SetNillableAccountID(id *int) *LockerCreate {
	if id != nil {
		lc = lc.SetAccountID(*id)
	}
	return lc
}

// SetAccount sets the "account" edge to the Account entity.
func (lc *LockerCreate) SetAccount(a *Account) *LockerCreate {
	return lc.SetAccountID(a.ID)
}

// Mutation returns the LockerMutation object of the builder.
func (lc *LockerCreate) Mutation() *LockerMutation {
	return lc.mutation
}

// Save creates the Locker in the database.
func (lc *LockerCreate) Save(ctx context.Context) (*Locker, error) {
	return withHooks[*Locker, LockerMutation](ctx, lc.sqlSave, lc.mutation, lc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (lc *LockerCreate) SaveX(ctx context.Context) *Locker {
	v, err := lc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (lc *LockerCreate) Exec(ctx context.Context) error {
	_, err := lc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (lc *LockerCreate) ExecX(ctx context.Context) {
	if err := lc.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (lc *LockerCreate) check() error {
	if _, ok := lc.mutation.Hash(); !ok {
		return &ValidationError{Name: "hash", err: errors.New(`ent: missing required field "Locker.hash"`)}
	}
	if _, ok := lc.mutation.Level(); !ok {
		return &ValidationError{Name: "level", err: errors.New(`ent: missing required field "Locker.level"`)}
	}
	if _, ok := lc.mutation.EncryptedID(); !ok {
		return &ValidationError{Name: "encrypted_id", err: errors.New(`ent: missing required field "Locker.encrypted_id"`)}
	}
	if _, ok := lc.mutation.EncryptedBody(); !ok {
		return &ValidationError{Name: "encrypted_body", err: errors.New(`ent: missing required field "Locker.encrypted_body"`)}
	}
	return nil
}

func (lc *LockerCreate) sqlSave(ctx context.Context) (*Locker, error) {
	if err := lc.check(); err != nil {
		return nil, err
	}
	_node, _spec := lc.createSpec()
	if err := sqlgraph.CreateNode(ctx, lc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	lc.mutation.id = &_node.ID
	lc.mutation.done = true
	return _node, nil
}

func (lc *LockerCreate) createSpec() (*Locker, *sqlgraph.CreateSpec) {
	var (
		_node = &Locker{config: lc.config}
		_spec = &sqlgraph.CreateSpec{
			Table: locker.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: locker.FieldID,
			},
		}
	)
	if value, ok := lc.mutation.Hash(); ok {
		_spec.SetField(locker.FieldHash, field.TypeString, value)
		_node.Hash = value
	}
	if value, ok := lc.mutation.Level(); ok {
		_spec.SetField(locker.FieldLevel, field.TypeInt32, value)
		_node.Level = value
	}
	if value, ok := lc.mutation.EncryptedID(); ok {
		_spec.SetField(locker.FieldEncryptedID, field.TypeString, value)
		_node.EncryptedID = value
	}
	if value, ok := lc.mutation.EncryptedBody(); ok {
		_spec.SetField(locker.FieldEncryptedBody, field.TypeString, value)
		_node.EncryptedBody = value
	}
	if nodes := lc.mutation.AccountIDs(); len(nodes) > 0 {
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
		_node.account = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// LockerCreateBulk is the builder for creating many Locker entities in bulk.
type LockerCreateBulk struct {
	config
	builders []*LockerCreate
}

// Save creates the Locker entities in the database.
func (lcb *LockerCreateBulk) Save(ctx context.Context) ([]*Locker, error) {
	specs := make([]*sqlgraph.CreateSpec, len(lcb.builders))
	nodes := make([]*Locker, len(lcb.builders))
	mutators := make([]Mutator, len(lcb.builders))
	for i := range lcb.builders {
		func(i int, root context.Context) {
			builder := lcb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*LockerMutation)
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
					_, err = mutators[i+1].Mutate(root, lcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, lcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, lcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (lcb *LockerCreateBulk) SaveX(ctx context.Context) []*Locker {
	v, err := lcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (lcb *LockerCreateBulk) Exec(ctx context.Context) error {
	_, err := lcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (lcb *LockerCreateBulk) ExecX(ctx context.Context) {
	if err := lcb.Exec(ctx); err != nil {
		panic(err)
	}
}
