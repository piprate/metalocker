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
	"github.com/piprate/metalocker/storage/rdb/ent/identity"

	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
)

// IdentityCreate is the builder for creating a Identity entity.
type IdentityCreate struct {
	config
	mutation *IdentityMutation
	hooks    []Hook
}

// SetHash sets the "hash" field.
func (ic *IdentityCreate) SetHash(s string) *IdentityCreate {
	ic.mutation.SetHash(s)
	return ic
}

// SetLevel sets the "level" field.
func (ic *IdentityCreate) SetLevel(i int32) *IdentityCreate {
	ic.mutation.SetLevel(i)
	return ic
}

// SetEncryptedID sets the "encrypted_id" field.
func (ic *IdentityCreate) SetEncryptedID(s string) *IdentityCreate {
	ic.mutation.SetEncryptedID(s)
	return ic
}

// SetEncryptedBody sets the "encrypted_body" field.
func (ic *IdentityCreate) SetEncryptedBody(s string) *IdentityCreate {
	ic.mutation.SetEncryptedBody(s)
	return ic
}

// SetAccountID sets the "account" edge to the Account entity by ID.
func (ic *IdentityCreate) SetAccountID(id int) *IdentityCreate {
	ic.mutation.SetAccountID(id)
	return ic
}

// SetNillableAccountID sets the "account" edge to the Account entity by ID if the given value is not nil.
func (ic *IdentityCreate) SetNillableAccountID(id *int) *IdentityCreate {
	if id != nil {
		ic = ic.SetAccountID(*id)
	}
	return ic
}

// SetAccount sets the "account" edge to the Account entity.
func (ic *IdentityCreate) SetAccount(a *Account) *IdentityCreate {
	return ic.SetAccountID(a.ID)
}

// Mutation returns the IdentityMutation object of the builder.
func (ic *IdentityCreate) Mutation() *IdentityMutation {
	return ic.mutation
}

// Save creates the Identity in the database.
func (ic *IdentityCreate) Save(ctx context.Context) (*Identity, error) {
	return withHooks[*Identity, IdentityMutation](ctx, ic.sqlSave, ic.mutation, ic.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (ic *IdentityCreate) SaveX(ctx context.Context) *Identity {
	v, err := ic.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ic *IdentityCreate) Exec(ctx context.Context) error {
	_, err := ic.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ic *IdentityCreate) ExecX(ctx context.Context) {
	if err := ic.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (ic *IdentityCreate) check() error {
	if _, ok := ic.mutation.Hash(); !ok {
		return &ValidationError{Name: "hash", err: errors.New(`ent: missing required field "Identity.hash"`)}
	}
	if _, ok := ic.mutation.Level(); !ok {
		return &ValidationError{Name: "level", err: errors.New(`ent: missing required field "Identity.level"`)}
	}
	if _, ok := ic.mutation.EncryptedID(); !ok {
		return &ValidationError{Name: "encrypted_id", err: errors.New(`ent: missing required field "Identity.encrypted_id"`)}
	}
	if _, ok := ic.mutation.EncryptedBody(); !ok {
		return &ValidationError{Name: "encrypted_body", err: errors.New(`ent: missing required field "Identity.encrypted_body"`)}
	}
	return nil
}

func (ic *IdentityCreate) sqlSave(ctx context.Context) (*Identity, error) {
	if err := ic.check(); err != nil {
		return nil, err
	}
	_node, _spec := ic.createSpec()
	if err := sqlgraph.CreateNode(ctx, ic.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	ic.mutation.id = &_node.ID
	ic.mutation.done = true
	return _node, nil
}

func (ic *IdentityCreate) createSpec() (*Identity, *sqlgraph.CreateSpec) {
	var (
		_node = &Identity{config: ic.config}
		_spec = &sqlgraph.CreateSpec{
			Table: identity.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: identity.FieldID,
			},
		}
	)
	if value, ok := ic.mutation.Hash(); ok {
		_spec.SetField(identity.FieldHash, field.TypeString, value)
		_node.Hash = value
	}
	if value, ok := ic.mutation.Level(); ok {
		_spec.SetField(identity.FieldLevel, field.TypeInt32, value)
		_node.Level = value
	}
	if value, ok := ic.mutation.EncryptedID(); ok {
		_spec.SetField(identity.FieldEncryptedID, field.TypeString, value)
		_node.EncryptedID = value
	}
	if value, ok := ic.mutation.EncryptedBody(); ok {
		_spec.SetField(identity.FieldEncryptedBody, field.TypeString, value)
		_node.EncryptedBody = value
	}
	if nodes := ic.mutation.AccountIDs(); len(nodes) > 0 {
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
		_node.account = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// IdentityCreateBulk is the builder for creating many Identity entities in bulk.
type IdentityCreateBulk struct {
	config
	builders []*IdentityCreate
}

// Save creates the Identity entities in the database.
func (icb *IdentityCreateBulk) Save(ctx context.Context) ([]*Identity, error) {
	specs := make([]*sqlgraph.CreateSpec, len(icb.builders))
	nodes := make([]*Identity, len(icb.builders))
	mutators := make([]Mutator, len(icb.builders))
	for i := range icb.builders {
		func(i int, root context.Context) {
			builder := icb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*IdentityMutation)
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
					_, err = mutators[i+1].Mutate(root, icb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, icb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, icb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (icb *IdentityCreateBulk) SaveX(ctx context.Context) []*Identity {
	v, err := icb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (icb *IdentityCreateBulk) Exec(ctx context.Context) error {
	_, err := icb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (icb *IdentityCreateBulk) ExecX(ctx context.Context) {
	if err := icb.Exec(ctx); err != nil {
		panic(err)
	}
}
