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
	"github.com/piprate/metalocker/model/account"

	"github.com/piprate/metalocker/storage/rdb/ent/accesskey"
	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
	"github.com/piprate/metalocker/storage/rdb/ent/identity"
	"github.com/piprate/metalocker/storage/rdb/ent/locker"
	"github.com/piprate/metalocker/storage/rdb/ent/property"
	"github.com/piprate/metalocker/storage/rdb/ent/recoverycode"
)

// AccountCreate is the builder for creating a Account entity.
type AccountCreate struct {
	config
	mutation *AccountMutation
	hooks    []Hook
}

// SetDid sets the "did" field.
func (ac *AccountCreate) SetDid(s string) *AccountCreate {
	ac.mutation.SetDid(s)
	return ac
}

// SetState sets the "state" field.
func (ac *AccountCreate) SetState(s string) *AccountCreate {
	ac.mutation.SetState(s)
	return ac
}

// SetEmail sets the "email" field.
func (ac *AccountCreate) SetEmail(s string) *AccountCreate {
	ac.mutation.SetEmail(s)
	return ac
}

// SetNillableEmail sets the "email" field if the given value is not nil.
func (ac *AccountCreate) SetNillableEmail(s *string) *AccountCreate {
	if s != nil {
		ac.SetEmail(*s)
	}
	return ac
}

// SetParentAccount sets the "parent_account" field.
func (ac *AccountCreate) SetParentAccount(s string) *AccountCreate {
	ac.mutation.SetParentAccount(s)
	return ac
}

// SetNillableParentAccount sets the "parent_account" field if the given value is not nil.
func (ac *AccountCreate) SetNillableParentAccount(s *string) *AccountCreate {
	if s != nil {
		ac.SetParentAccount(*s)
	}
	return ac
}

// SetBody sets the "body" field.
func (ac *AccountCreate) SetBody(a *account.Account) *AccountCreate {
	ac.mutation.SetBody(a)
	return ac
}

// AddRecoveryCodeIDs adds the "recovery_codes" edge to the RecoveryCode entity by IDs.
func (ac *AccountCreate) AddRecoveryCodeIDs(ids ...int) *AccountCreate {
	ac.mutation.AddRecoveryCodeIDs(ids...)
	return ac
}

// AddRecoveryCodes adds the "recovery_codes" edges to the RecoveryCode entity.
func (ac *AccountCreate) AddRecoveryCodes(r ...*RecoveryCode) *AccountCreate {
	ids := make([]int, len(r))
	for i := range r {
		ids[i] = r[i].ID
	}
	return ac.AddRecoveryCodeIDs(ids...)
}

// AddAccessKeyIDs adds the "access_keys" edge to the AccessKey entity by IDs.
func (ac *AccountCreate) AddAccessKeyIDs(ids ...int) *AccountCreate {
	ac.mutation.AddAccessKeyIDs(ids...)
	return ac
}

// AddAccessKeys adds the "access_keys" edges to the AccessKey entity.
func (ac *AccountCreate) AddAccessKeys(a ...*AccessKey) *AccountCreate {
	ids := make([]int, len(a))
	for i := range a {
		ids[i] = a[i].ID
	}
	return ac.AddAccessKeyIDs(ids...)
}

// AddIdentityIDs adds the "identities" edge to the Identity entity by IDs.
func (ac *AccountCreate) AddIdentityIDs(ids ...int) *AccountCreate {
	ac.mutation.AddIdentityIDs(ids...)
	return ac
}

// AddIdentities adds the "identities" edges to the Identity entity.
func (ac *AccountCreate) AddIdentities(i ...*Identity) *AccountCreate {
	ids := make([]int, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return ac.AddIdentityIDs(ids...)
}

// AddLockerIDs adds the "lockers" edge to the Locker entity by IDs.
func (ac *AccountCreate) AddLockerIDs(ids ...int) *AccountCreate {
	ac.mutation.AddLockerIDs(ids...)
	return ac
}

// AddLockers adds the "lockers" edges to the Locker entity.
func (ac *AccountCreate) AddLockers(l ...*Locker) *AccountCreate {
	ids := make([]int, len(l))
	for i := range l {
		ids[i] = l[i].ID
	}
	return ac.AddLockerIDs(ids...)
}

// AddPropertyIDs adds the "properties" edge to the Property entity by IDs.
func (ac *AccountCreate) AddPropertyIDs(ids ...int) *AccountCreate {
	ac.mutation.AddPropertyIDs(ids...)
	return ac
}

// AddProperties adds the "properties" edges to the Property entity.
func (ac *AccountCreate) AddProperties(p ...*Property) *AccountCreate {
	ids := make([]int, len(p))
	for i := range p {
		ids[i] = p[i].ID
	}
	return ac.AddPropertyIDs(ids...)
}

// Mutation returns the AccountMutation object of the builder.
func (ac *AccountCreate) Mutation() *AccountMutation {
	return ac.mutation
}

// Save creates the Account in the database.
func (ac *AccountCreate) Save(ctx context.Context) (*Account, error) {
	return withHooks[*Account, AccountMutation](ctx, ac.sqlSave, ac.mutation, ac.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (ac *AccountCreate) SaveX(ctx context.Context) *Account {
	v, err := ac.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ac *AccountCreate) Exec(ctx context.Context) error {
	_, err := ac.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ac *AccountCreate) ExecX(ctx context.Context) {
	if err := ac.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (ac *AccountCreate) check() error {
	if _, ok := ac.mutation.Did(); !ok {
		return &ValidationError{Name: "did", err: errors.New(`ent: missing required field "Account.did"`)}
	}
	if _, ok := ac.mutation.State(); !ok {
		return &ValidationError{Name: "state", err: errors.New(`ent: missing required field "Account.state"`)}
	}
	if _, ok := ac.mutation.Body(); !ok {
		return &ValidationError{Name: "body", err: errors.New(`ent: missing required field "Account.body"`)}
	}
	return nil
}

func (ac *AccountCreate) sqlSave(ctx context.Context) (*Account, error) {
	if err := ac.check(); err != nil {
		return nil, err
	}
	_node, _spec := ac.createSpec()
	if err := sqlgraph.CreateNode(ctx, ac.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	ac.mutation.id = &_node.ID
	ac.mutation.done = true
	return _node, nil
}

func (ac *AccountCreate) createSpec() (*Account, *sqlgraph.CreateSpec) {
	var (
		_node = &Account{config: ac.config}
		_spec = &sqlgraph.CreateSpec{
			Table: entaccount.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: entaccount.FieldID,
			},
		}
	)
	if value, ok := ac.mutation.Did(); ok {
		_spec.SetField(entaccount.FieldDid, field.TypeString, value)
		_node.Did = value
	}
	if value, ok := ac.mutation.State(); ok {
		_spec.SetField(entaccount.FieldState, field.TypeString, value)
		_node.State = value
	}
	if value, ok := ac.mutation.Email(); ok {
		_spec.SetField(entaccount.FieldEmail, field.TypeString, value)
		_node.Email = value
	}
	if value, ok := ac.mutation.ParentAccount(); ok {
		_spec.SetField(entaccount.FieldParentAccount, field.TypeString, value)
		_node.ParentAccount = value
	}
	if value, ok := ac.mutation.Body(); ok {
		_spec.SetField(entaccount.FieldBody, field.TypeJSON, value)
		_node.Body = value
	}
	if nodes := ac.mutation.RecoveryCodesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   entaccount.RecoveryCodesTable,
			Columns: []string{entaccount.RecoveryCodesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeInt,
					Column: recoverycode.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ac.mutation.AccessKeysIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   entaccount.AccessKeysTable,
			Columns: []string{entaccount.AccessKeysColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeInt,
					Column: accesskey.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ac.mutation.IdentitiesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   entaccount.IdentitiesTable,
			Columns: []string{entaccount.IdentitiesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeInt,
					Column: identity.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ac.mutation.LockersIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   entaccount.LockersTable,
			Columns: []string{entaccount.LockersColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeInt,
					Column: locker.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := ac.mutation.PropertiesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   entaccount.PropertiesTable,
			Columns: []string{entaccount.PropertiesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeInt,
					Column: property.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// AccountCreateBulk is the builder for creating many Account entities in bulk.
type AccountCreateBulk struct {
	config
	builders []*AccountCreate
}

// Save creates the Account entities in the database.
func (acb *AccountCreateBulk) Save(ctx context.Context) ([]*Account, error) {
	specs := make([]*sqlgraph.CreateSpec, len(acb.builders))
	nodes := make([]*Account, len(acb.builders))
	mutators := make([]Mutator, len(acb.builders))
	for i := range acb.builders {
		func(i int, root context.Context) {
			builder := acb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*AccountMutation)
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
					_, err = mutators[i+1].Mutate(root, acb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, acb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, acb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (acb *AccountCreateBulk) SaveX(ctx context.Context) []*Account {
	v, err := acb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (acb *AccountCreateBulk) Exec(ctx context.Context) error {
	_, err := acb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (acb *AccountCreateBulk) ExecX(ctx context.Context) {
	if err := acb.Exec(ctx); err != nil {
		panic(err)
	}
}
