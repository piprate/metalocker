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
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/piprate/metalocker/storage/rdb/ent/recoverycode"

	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
)

// RecoveryCodeCreate is the builder for creating a RecoveryCode entity.
type RecoveryCodeCreate struct {
	config
	mutation *RecoveryCodeMutation
	hooks    []Hook
}

// SetCode sets the "code" field.
func (rcc *RecoveryCodeCreate) SetCode(s string) *RecoveryCodeCreate {
	rcc.mutation.SetCode(s)
	return rcc
}

// SetExpiresAt sets the "expires_at" field.
func (rcc *RecoveryCodeCreate) SetExpiresAt(t time.Time) *RecoveryCodeCreate {
	rcc.mutation.SetExpiresAt(t)
	return rcc
}

// SetNillableExpiresAt sets the "expires_at" field if the given value is not nil.
func (rcc *RecoveryCodeCreate) SetNillableExpiresAt(t *time.Time) *RecoveryCodeCreate {
	if t != nil {
		rcc.SetExpiresAt(*t)
	}
	return rcc
}

// SetAccountID sets the "account" edge to the Account entity by ID.
func (rcc *RecoveryCodeCreate) SetAccountID(id int) *RecoveryCodeCreate {
	rcc.mutation.SetAccountID(id)
	return rcc
}

// SetNillableAccountID sets the "account" edge to the Account entity by ID if the given value is not nil.
func (rcc *RecoveryCodeCreate) SetNillableAccountID(id *int) *RecoveryCodeCreate {
	if id != nil {
		rcc = rcc.SetAccountID(*id)
	}
	return rcc
}

// SetAccount sets the "account" edge to the Account entity.
func (rcc *RecoveryCodeCreate) SetAccount(a *Account) *RecoveryCodeCreate {
	return rcc.SetAccountID(a.ID)
}

// Mutation returns the RecoveryCodeMutation object of the builder.
func (rcc *RecoveryCodeCreate) Mutation() *RecoveryCodeMutation {
	return rcc.mutation
}

// Save creates the RecoveryCode in the database.
func (rcc *RecoveryCodeCreate) Save(ctx context.Context) (*RecoveryCode, error) {
	return withHooks[*RecoveryCode, RecoveryCodeMutation](ctx, rcc.sqlSave, rcc.mutation, rcc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (rcc *RecoveryCodeCreate) SaveX(ctx context.Context) *RecoveryCode {
	v, err := rcc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (rcc *RecoveryCodeCreate) Exec(ctx context.Context) error {
	_, err := rcc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (rcc *RecoveryCodeCreate) ExecX(ctx context.Context) {
	if err := rcc.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (rcc *RecoveryCodeCreate) check() error {
	if _, ok := rcc.mutation.Code(); !ok {
		return &ValidationError{Name: "code", err: errors.New(`ent: missing required field "RecoveryCode.code"`)}
	}
	return nil
}

func (rcc *RecoveryCodeCreate) sqlSave(ctx context.Context) (*RecoveryCode, error) {
	if err := rcc.check(); err != nil {
		return nil, err
	}
	_node, _spec := rcc.createSpec()
	if err := sqlgraph.CreateNode(ctx, rcc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	rcc.mutation.id = &_node.ID
	rcc.mutation.done = true
	return _node, nil
}

func (rcc *RecoveryCodeCreate) createSpec() (*RecoveryCode, *sqlgraph.CreateSpec) {
	var (
		_node = &RecoveryCode{config: rcc.config}
		_spec = &sqlgraph.CreateSpec{
			Table: recoverycode.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: recoverycode.FieldID,
			},
		}
	)
	if value, ok := rcc.mutation.Code(); ok {
		_spec.SetField(recoverycode.FieldCode, field.TypeString, value)
		_node.Code = value
	}
	if value, ok := rcc.mutation.ExpiresAt(); ok {
		_spec.SetField(recoverycode.FieldExpiresAt, field.TypeTime, value)
		_node.ExpiresAt = &value
	}
	if nodes := rcc.mutation.AccountIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   recoverycode.AccountTable,
			Columns: []string{recoverycode.AccountColumn},
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

// RecoveryCodeCreateBulk is the builder for creating many RecoveryCode entities in bulk.
type RecoveryCodeCreateBulk struct {
	config
	builders []*RecoveryCodeCreate
}

// Save creates the RecoveryCode entities in the database.
func (rccb *RecoveryCodeCreateBulk) Save(ctx context.Context) ([]*RecoveryCode, error) {
	specs := make([]*sqlgraph.CreateSpec, len(rccb.builders))
	nodes := make([]*RecoveryCode, len(rccb.builders))
	mutators := make([]Mutator, len(rccb.builders))
	for i := range rccb.builders {
		func(i int, root context.Context) {
			builder := rccb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*RecoveryCodeMutation)
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
					_, err = mutators[i+1].Mutate(root, rccb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, rccb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, rccb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (rccb *RecoveryCodeCreateBulk) SaveX(ctx context.Context) []*RecoveryCode {
	v, err := rccb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (rccb *RecoveryCodeCreateBulk) Exec(ctx context.Context) error {
	_, err := rccb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (rccb *RecoveryCodeCreateBulk) ExecX(ctx context.Context) {
	if err := rccb.Exec(ctx); err != nil {
		panic(err)
	}
}
