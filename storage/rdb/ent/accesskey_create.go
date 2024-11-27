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
	"github.com/piprate/metalocker/storage/rdb/ent/accesskey"

	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
)

// AccessKeyCreate is the builder for creating a AccessKey entity.
type AccessKeyCreate struct {
	config
	mutation *AccessKeyMutation
	hooks    []Hook
}

// SetDid sets the "did" field.
func (akc *AccessKeyCreate) SetDid(s string) *AccessKeyCreate {
	akc.mutation.SetDid(s)
	return akc
}

// SetBody sets the "body" field.
func (akc *AccessKeyCreate) SetBody(mk *model.AccessKey) *AccessKeyCreate {
	akc.mutation.SetBody(mk)
	return akc
}

// SetAccountID sets the "account" edge to the Account entity by ID.
func (akc *AccessKeyCreate) SetAccountID(id int) *AccessKeyCreate {
	akc.mutation.SetAccountID(id)
	return akc
}

// SetNillableAccountID sets the "account" edge to the Account entity by ID if the given value is not nil.
func (akc *AccessKeyCreate) SetNillableAccountID(id *int) *AccessKeyCreate {
	if id != nil {
		akc = akc.SetAccountID(*id)
	}
	return akc
}

// SetAccount sets the "account" edge to the Account entity.
func (akc *AccessKeyCreate) SetAccount(a *Account) *AccessKeyCreate {
	return akc.SetAccountID(a.ID)
}

// Mutation returns the AccessKeyMutation object of the builder.
func (akc *AccessKeyCreate) Mutation() *AccessKeyMutation {
	return akc.mutation
}

// Save creates the AccessKey in the database.
func (akc *AccessKeyCreate) Save(ctx context.Context) (*AccessKey, error) {
	return withHooks[*AccessKey, AccessKeyMutation](ctx, akc.sqlSave, akc.mutation, akc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (akc *AccessKeyCreate) SaveX(ctx context.Context) *AccessKey {
	v, err := akc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (akc *AccessKeyCreate) Exec(ctx context.Context) error {
	_, err := akc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (akc *AccessKeyCreate) ExecX(ctx context.Context) {
	if err := akc.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (akc *AccessKeyCreate) check() error {
	if _, ok := akc.mutation.Did(); !ok {
		return &ValidationError{Name: "did", err: errors.New(`ent: missing required field "AccessKey.did"`)}
	}
	if _, ok := akc.mutation.Body(); !ok {
		return &ValidationError{Name: "body", err: errors.New(`ent: missing required field "AccessKey.body"`)}
	}
	return nil
}

func (akc *AccessKeyCreate) sqlSave(ctx context.Context) (*AccessKey, error) {
	if err := akc.check(); err != nil {
		return nil, err
	}
	_node, _spec := akc.createSpec()
	if err := sqlgraph.CreateNode(ctx, akc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	akc.mutation.id = &_node.ID
	akc.mutation.done = true
	return _node, nil
}

func (akc *AccessKeyCreate) createSpec() (*AccessKey, *sqlgraph.CreateSpec) {
	var (
		_node = &AccessKey{config: akc.config}
		_spec = &sqlgraph.CreateSpec{
			Table: accesskey.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: accesskey.FieldID,
			},
		}
	)
	if value, ok := akc.mutation.Did(); ok {
		_spec.SetField(accesskey.FieldDid, field.TypeString, value)
		_node.Did = value
	}
	if value, ok := akc.mutation.Body(); ok {
		_spec.SetField(accesskey.FieldBody, field.TypeJSON, value)
		_node.Body = value
	}
	if nodes := akc.mutation.AccountIDs(); len(nodes) > 0 {
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
		_node.account = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// AccessKeyCreateBulk is the builder for creating many AccessKey entities in bulk.
type AccessKeyCreateBulk struct {
	config
	builders []*AccessKeyCreate
}

// Save creates the AccessKey entities in the database.
func (akcb *AccessKeyCreateBulk) Save(ctx context.Context) ([]*AccessKey, error) {
	specs := make([]*sqlgraph.CreateSpec, len(akcb.builders))
	nodes := make([]*AccessKey, len(akcb.builders))
	mutators := make([]Mutator, len(akcb.builders))
	for i := range akcb.builders {
		func(i int, root context.Context) {
			builder := akcb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*AccessKeyMutation)
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
					_, err = mutators[i+1].Mutate(root, akcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, akcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, akcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (akcb *AccessKeyCreateBulk) SaveX(ctx context.Context) []*AccessKey {
	v, err := akcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (akcb *AccessKeyCreateBulk) Exec(ctx context.Context) error {
	_, err := akcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (akcb *AccessKeyCreateBulk) ExecX(ctx context.Context) {
	if err := akcb.Exec(ctx); err != nil {
		panic(err)
	}
}
