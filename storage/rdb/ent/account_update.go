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
	"github.com/piprate/metalocker/model/account"
	"github.com/piprate/metalocker/storage/rdb/ent/predicate"

	"github.com/piprate/metalocker/storage/rdb/ent/accesskey"
	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
	"github.com/piprate/metalocker/storage/rdb/ent/identity"
	"github.com/piprate/metalocker/storage/rdb/ent/locker"
	"github.com/piprate/metalocker/storage/rdb/ent/property"
	"github.com/piprate/metalocker/storage/rdb/ent/recoverycode"
)

// AccountUpdate is the builder for updating Account entities.
type AccountUpdate struct {
	config
	hooks    []Hook
	mutation *AccountMutation
}

// Where appends a list predicates to the AccountUpdate builder.
func (au *AccountUpdate) Where(ps ...predicate.Account) *AccountUpdate {
	au.mutation.Where(ps...)
	return au
}

// SetDid sets the "did" field.
func (au *AccountUpdate) SetDid(s string) *AccountUpdate {
	au.mutation.SetDid(s)
	return au
}

// SetState sets the "state" field.
func (au *AccountUpdate) SetState(s string) *AccountUpdate {
	au.mutation.SetState(s)
	return au
}

// SetEmail sets the "email" field.
func (au *AccountUpdate) SetEmail(s string) *AccountUpdate {
	au.mutation.SetEmail(s)
	return au
}

// SetNillableEmail sets the "email" field if the given value is not nil.
func (au *AccountUpdate) SetNillableEmail(s *string) *AccountUpdate {
	if s != nil {
		au.SetEmail(*s)
	}
	return au
}

// ClearEmail clears the value of the "email" field.
func (au *AccountUpdate) ClearEmail() *AccountUpdate {
	au.mutation.ClearEmail()
	return au
}

// SetParentAccount sets the "parent_account" field.
func (au *AccountUpdate) SetParentAccount(s string) *AccountUpdate {
	au.mutation.SetParentAccount(s)
	return au
}

// SetNillableParentAccount sets the "parent_account" field if the given value is not nil.
func (au *AccountUpdate) SetNillableParentAccount(s *string) *AccountUpdate {
	if s != nil {
		au.SetParentAccount(*s)
	}
	return au
}

// ClearParentAccount clears the value of the "parent_account" field.
func (au *AccountUpdate) ClearParentAccount() *AccountUpdate {
	au.mutation.ClearParentAccount()
	return au
}

// SetBody sets the "body" field.
func (au *AccountUpdate) SetBody(a *account.Account) *AccountUpdate {
	au.mutation.SetBody(a)
	return au
}

// AddRecoveryCodeIDs adds the "recovery_codes" edge to the RecoveryCode entity by IDs.
func (au *AccountUpdate) AddRecoveryCodeIDs(ids ...int) *AccountUpdate {
	au.mutation.AddRecoveryCodeIDs(ids...)
	return au
}

// AddRecoveryCodes adds the "recovery_codes" edges to the RecoveryCode entity.
func (au *AccountUpdate) AddRecoveryCodes(r ...*RecoveryCode) *AccountUpdate {
	ids := make([]int, len(r))
	for i := range r {
		ids[i] = r[i].ID
	}
	return au.AddRecoveryCodeIDs(ids...)
}

// AddAccessKeyIDs adds the "access_keys" edge to the AccessKey entity by IDs.
func (au *AccountUpdate) AddAccessKeyIDs(ids ...int) *AccountUpdate {
	au.mutation.AddAccessKeyIDs(ids...)
	return au
}

// AddAccessKeys adds the "access_keys" edges to the AccessKey entity.
func (au *AccountUpdate) AddAccessKeys(a ...*AccessKey) *AccountUpdate {
	ids := make([]int, len(a))
	for i := range a {
		ids[i] = a[i].ID
	}
	return au.AddAccessKeyIDs(ids...)
}

// AddIdentityIDs adds the "identities" edge to the Identity entity by IDs.
func (au *AccountUpdate) AddIdentityIDs(ids ...int) *AccountUpdate {
	au.mutation.AddIdentityIDs(ids...)
	return au
}

// AddIdentities adds the "identities" edges to the Identity entity.
func (au *AccountUpdate) AddIdentities(i ...*Identity) *AccountUpdate {
	ids := make([]int, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return au.AddIdentityIDs(ids...)
}

// AddLockerIDs adds the "lockers" edge to the Locker entity by IDs.
func (au *AccountUpdate) AddLockerIDs(ids ...int) *AccountUpdate {
	au.mutation.AddLockerIDs(ids...)
	return au
}

// AddLockers adds the "lockers" edges to the Locker entity.
func (au *AccountUpdate) AddLockers(l ...*Locker) *AccountUpdate {
	ids := make([]int, len(l))
	for i := range l {
		ids[i] = l[i].ID
	}
	return au.AddLockerIDs(ids...)
}

// AddPropertyIDs adds the "properties" edge to the Property entity by IDs.
func (au *AccountUpdate) AddPropertyIDs(ids ...int) *AccountUpdate {
	au.mutation.AddPropertyIDs(ids...)
	return au
}

// AddProperties adds the "properties" edges to the Property entity.
func (au *AccountUpdate) AddProperties(p ...*Property) *AccountUpdate {
	ids := make([]int, len(p))
	for i := range p {
		ids[i] = p[i].ID
	}
	return au.AddPropertyIDs(ids...)
}

// Mutation returns the AccountMutation object of the builder.
func (au *AccountUpdate) Mutation() *AccountMutation {
	return au.mutation
}

// ClearRecoveryCodes clears all "recovery_codes" edges to the RecoveryCode entity.
func (au *AccountUpdate) ClearRecoveryCodes() *AccountUpdate {
	au.mutation.ClearRecoveryCodes()
	return au
}

// RemoveRecoveryCodeIDs removes the "recovery_codes" edge to RecoveryCode entities by IDs.
func (au *AccountUpdate) RemoveRecoveryCodeIDs(ids ...int) *AccountUpdate {
	au.mutation.RemoveRecoveryCodeIDs(ids...)
	return au
}

// RemoveRecoveryCodes removes "recovery_codes" edges to RecoveryCode entities.
func (au *AccountUpdate) RemoveRecoveryCodes(r ...*RecoveryCode) *AccountUpdate {
	ids := make([]int, len(r))
	for i := range r {
		ids[i] = r[i].ID
	}
	return au.RemoveRecoveryCodeIDs(ids...)
}

// ClearAccessKeys clears all "access_keys" edges to the AccessKey entity.
func (au *AccountUpdate) ClearAccessKeys() *AccountUpdate {
	au.mutation.ClearAccessKeys()
	return au
}

// RemoveAccessKeyIDs removes the "access_keys" edge to AccessKey entities by IDs.
func (au *AccountUpdate) RemoveAccessKeyIDs(ids ...int) *AccountUpdate {
	au.mutation.RemoveAccessKeyIDs(ids...)
	return au
}

// RemoveAccessKeys removes "access_keys" edges to AccessKey entities.
func (au *AccountUpdate) RemoveAccessKeys(a ...*AccessKey) *AccountUpdate {
	ids := make([]int, len(a))
	for i := range a {
		ids[i] = a[i].ID
	}
	return au.RemoveAccessKeyIDs(ids...)
}

// ClearIdentities clears all "identities" edges to the Identity entity.
func (au *AccountUpdate) ClearIdentities() *AccountUpdate {
	au.mutation.ClearIdentities()
	return au
}

// RemoveIdentityIDs removes the "identities" edge to Identity entities by IDs.
func (au *AccountUpdate) RemoveIdentityIDs(ids ...int) *AccountUpdate {
	au.mutation.RemoveIdentityIDs(ids...)
	return au
}

// RemoveIdentities removes "identities" edges to Identity entities.
func (au *AccountUpdate) RemoveIdentities(i ...*Identity) *AccountUpdate {
	ids := make([]int, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return au.RemoveIdentityIDs(ids...)
}

// ClearLockers clears all "lockers" edges to the Locker entity.
func (au *AccountUpdate) ClearLockers() *AccountUpdate {
	au.mutation.ClearLockers()
	return au
}

// RemoveLockerIDs removes the "lockers" edge to Locker entities by IDs.
func (au *AccountUpdate) RemoveLockerIDs(ids ...int) *AccountUpdate {
	au.mutation.RemoveLockerIDs(ids...)
	return au
}

// RemoveLockers removes "lockers" edges to Locker entities.
func (au *AccountUpdate) RemoveLockers(l ...*Locker) *AccountUpdate {
	ids := make([]int, len(l))
	for i := range l {
		ids[i] = l[i].ID
	}
	return au.RemoveLockerIDs(ids...)
}

// ClearProperties clears all "properties" edges to the Property entity.
func (au *AccountUpdate) ClearProperties() *AccountUpdate {
	au.mutation.ClearProperties()
	return au
}

// RemovePropertyIDs removes the "properties" edge to Property entities by IDs.
func (au *AccountUpdate) RemovePropertyIDs(ids ...int) *AccountUpdate {
	au.mutation.RemovePropertyIDs(ids...)
	return au
}

// RemoveProperties removes "properties" edges to Property entities.
func (au *AccountUpdate) RemoveProperties(p ...*Property) *AccountUpdate {
	ids := make([]int, len(p))
	for i := range p {
		ids[i] = p[i].ID
	}
	return au.RemovePropertyIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (au *AccountUpdate) Save(ctx context.Context) (int, error) {
	return withHooks[int, AccountMutation](ctx, au.sqlSave, au.mutation, au.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (au *AccountUpdate) SaveX(ctx context.Context) int {
	affected, err := au.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (au *AccountUpdate) Exec(ctx context.Context) error {
	_, err := au.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (au *AccountUpdate) ExecX(ctx context.Context) {
	if err := au.Exec(ctx); err != nil {
		panic(err)
	}
}

func (au *AccountUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   entaccount.Table,
			Columns: entaccount.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: entaccount.FieldID,
			},
		},
	}
	if ps := au.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := au.mutation.Did(); ok {
		_spec.SetField(entaccount.FieldDid, field.TypeString, value)
	}
	if value, ok := au.mutation.State(); ok {
		_spec.SetField(entaccount.FieldState, field.TypeString, value)
	}
	if value, ok := au.mutation.Email(); ok {
		_spec.SetField(entaccount.FieldEmail, field.TypeString, value)
	}
	if au.mutation.EmailCleared() {
		_spec.ClearField(entaccount.FieldEmail, field.TypeString)
	}
	if value, ok := au.mutation.ParentAccount(); ok {
		_spec.SetField(entaccount.FieldParentAccount, field.TypeString, value)
	}
	if au.mutation.ParentAccountCleared() {
		_spec.ClearField(entaccount.FieldParentAccount, field.TypeString)
	}
	if value, ok := au.mutation.Body(); ok {
		_spec.SetField(entaccount.FieldBody, field.TypeJSON, value)
	}
	if au.mutation.RecoveryCodesCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := au.mutation.RemovedRecoveryCodesIDs(); len(nodes) > 0 && !au.mutation.RecoveryCodesCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := au.mutation.RecoveryCodesIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if au.mutation.AccessKeysCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := au.mutation.RemovedAccessKeysIDs(); len(nodes) > 0 && !au.mutation.AccessKeysCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := au.mutation.AccessKeysIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if au.mutation.IdentitiesCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := au.mutation.RemovedIdentitiesIDs(); len(nodes) > 0 && !au.mutation.IdentitiesCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := au.mutation.IdentitiesIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if au.mutation.LockersCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := au.mutation.RemovedLockersIDs(); len(nodes) > 0 && !au.mutation.LockersCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := au.mutation.LockersIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if au.mutation.PropertiesCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := au.mutation.RemovedPropertiesIDs(); len(nodes) > 0 && !au.mutation.PropertiesCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := au.mutation.PropertiesIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, au.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{entaccount.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	au.mutation.done = true
	return n, nil
}

// AccountUpdateOne is the builder for updating a single Account entity.
type AccountUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *AccountMutation
}

// SetDid sets the "did" field.
func (auo *AccountUpdateOne) SetDid(s string) *AccountUpdateOne {
	auo.mutation.SetDid(s)
	return auo
}

// SetState sets the "state" field.
func (auo *AccountUpdateOne) SetState(s string) *AccountUpdateOne {
	auo.mutation.SetState(s)
	return auo
}

// SetEmail sets the "email" field.
func (auo *AccountUpdateOne) SetEmail(s string) *AccountUpdateOne {
	auo.mutation.SetEmail(s)
	return auo
}

// SetNillableEmail sets the "email" field if the given value is not nil.
func (auo *AccountUpdateOne) SetNillableEmail(s *string) *AccountUpdateOne {
	if s != nil {
		auo.SetEmail(*s)
	}
	return auo
}

// ClearEmail clears the value of the "email" field.
func (auo *AccountUpdateOne) ClearEmail() *AccountUpdateOne {
	auo.mutation.ClearEmail()
	return auo
}

// SetParentAccount sets the "parent_account" field.
func (auo *AccountUpdateOne) SetParentAccount(s string) *AccountUpdateOne {
	auo.mutation.SetParentAccount(s)
	return auo
}

// SetNillableParentAccount sets the "parent_account" field if the given value is not nil.
func (auo *AccountUpdateOne) SetNillableParentAccount(s *string) *AccountUpdateOne {
	if s != nil {
		auo.SetParentAccount(*s)
	}
	return auo
}

// ClearParentAccount clears the value of the "parent_account" field.
func (auo *AccountUpdateOne) ClearParentAccount() *AccountUpdateOne {
	auo.mutation.ClearParentAccount()
	return auo
}

// SetBody sets the "body" field.
func (auo *AccountUpdateOne) SetBody(a *account.Account) *AccountUpdateOne {
	auo.mutation.SetBody(a)
	return auo
}

// AddRecoveryCodeIDs adds the "recovery_codes" edge to the RecoveryCode entity by IDs.
func (auo *AccountUpdateOne) AddRecoveryCodeIDs(ids ...int) *AccountUpdateOne {
	auo.mutation.AddRecoveryCodeIDs(ids...)
	return auo
}

// AddRecoveryCodes adds the "recovery_codes" edges to the RecoveryCode entity.
func (auo *AccountUpdateOne) AddRecoveryCodes(r ...*RecoveryCode) *AccountUpdateOne {
	ids := make([]int, len(r))
	for i := range r {
		ids[i] = r[i].ID
	}
	return auo.AddRecoveryCodeIDs(ids...)
}

// AddAccessKeyIDs adds the "access_keys" edge to the AccessKey entity by IDs.
func (auo *AccountUpdateOne) AddAccessKeyIDs(ids ...int) *AccountUpdateOne {
	auo.mutation.AddAccessKeyIDs(ids...)
	return auo
}

// AddAccessKeys adds the "access_keys" edges to the AccessKey entity.
func (auo *AccountUpdateOne) AddAccessKeys(a ...*AccessKey) *AccountUpdateOne {
	ids := make([]int, len(a))
	for i := range a {
		ids[i] = a[i].ID
	}
	return auo.AddAccessKeyIDs(ids...)
}

// AddIdentityIDs adds the "identities" edge to the Identity entity by IDs.
func (auo *AccountUpdateOne) AddIdentityIDs(ids ...int) *AccountUpdateOne {
	auo.mutation.AddIdentityIDs(ids...)
	return auo
}

// AddIdentities adds the "identities" edges to the Identity entity.
func (auo *AccountUpdateOne) AddIdentities(i ...*Identity) *AccountUpdateOne {
	ids := make([]int, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return auo.AddIdentityIDs(ids...)
}

// AddLockerIDs adds the "lockers" edge to the Locker entity by IDs.
func (auo *AccountUpdateOne) AddLockerIDs(ids ...int) *AccountUpdateOne {
	auo.mutation.AddLockerIDs(ids...)
	return auo
}

// AddLockers adds the "lockers" edges to the Locker entity.
func (auo *AccountUpdateOne) AddLockers(l ...*Locker) *AccountUpdateOne {
	ids := make([]int, len(l))
	for i := range l {
		ids[i] = l[i].ID
	}
	return auo.AddLockerIDs(ids...)
}

// AddPropertyIDs adds the "properties" edge to the Property entity by IDs.
func (auo *AccountUpdateOne) AddPropertyIDs(ids ...int) *AccountUpdateOne {
	auo.mutation.AddPropertyIDs(ids...)
	return auo
}

// AddProperties adds the "properties" edges to the Property entity.
func (auo *AccountUpdateOne) AddProperties(p ...*Property) *AccountUpdateOne {
	ids := make([]int, len(p))
	for i := range p {
		ids[i] = p[i].ID
	}
	return auo.AddPropertyIDs(ids...)
}

// Mutation returns the AccountMutation object of the builder.
func (auo *AccountUpdateOne) Mutation() *AccountMutation {
	return auo.mutation
}

// ClearRecoveryCodes clears all "recovery_codes" edges to the RecoveryCode entity.
func (auo *AccountUpdateOne) ClearRecoveryCodes() *AccountUpdateOne {
	auo.mutation.ClearRecoveryCodes()
	return auo
}

// RemoveRecoveryCodeIDs removes the "recovery_codes" edge to RecoveryCode entities by IDs.
func (auo *AccountUpdateOne) RemoveRecoveryCodeIDs(ids ...int) *AccountUpdateOne {
	auo.mutation.RemoveRecoveryCodeIDs(ids...)
	return auo
}

// RemoveRecoveryCodes removes "recovery_codes" edges to RecoveryCode entities.
func (auo *AccountUpdateOne) RemoveRecoveryCodes(r ...*RecoveryCode) *AccountUpdateOne {
	ids := make([]int, len(r))
	for i := range r {
		ids[i] = r[i].ID
	}
	return auo.RemoveRecoveryCodeIDs(ids...)
}

// ClearAccessKeys clears all "access_keys" edges to the AccessKey entity.
func (auo *AccountUpdateOne) ClearAccessKeys() *AccountUpdateOne {
	auo.mutation.ClearAccessKeys()
	return auo
}

// RemoveAccessKeyIDs removes the "access_keys" edge to AccessKey entities by IDs.
func (auo *AccountUpdateOne) RemoveAccessKeyIDs(ids ...int) *AccountUpdateOne {
	auo.mutation.RemoveAccessKeyIDs(ids...)
	return auo
}

// RemoveAccessKeys removes "access_keys" edges to AccessKey entities.
func (auo *AccountUpdateOne) RemoveAccessKeys(a ...*AccessKey) *AccountUpdateOne {
	ids := make([]int, len(a))
	for i := range a {
		ids[i] = a[i].ID
	}
	return auo.RemoveAccessKeyIDs(ids...)
}

// ClearIdentities clears all "identities" edges to the Identity entity.
func (auo *AccountUpdateOne) ClearIdentities() *AccountUpdateOne {
	auo.mutation.ClearIdentities()
	return auo
}

// RemoveIdentityIDs removes the "identities" edge to Identity entities by IDs.
func (auo *AccountUpdateOne) RemoveIdentityIDs(ids ...int) *AccountUpdateOne {
	auo.mutation.RemoveIdentityIDs(ids...)
	return auo
}

// RemoveIdentities removes "identities" edges to Identity entities.
func (auo *AccountUpdateOne) RemoveIdentities(i ...*Identity) *AccountUpdateOne {
	ids := make([]int, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return auo.RemoveIdentityIDs(ids...)
}

// ClearLockers clears all "lockers" edges to the Locker entity.
func (auo *AccountUpdateOne) ClearLockers() *AccountUpdateOne {
	auo.mutation.ClearLockers()
	return auo
}

// RemoveLockerIDs removes the "lockers" edge to Locker entities by IDs.
func (auo *AccountUpdateOne) RemoveLockerIDs(ids ...int) *AccountUpdateOne {
	auo.mutation.RemoveLockerIDs(ids...)
	return auo
}

// RemoveLockers removes "lockers" edges to Locker entities.
func (auo *AccountUpdateOne) RemoveLockers(l ...*Locker) *AccountUpdateOne {
	ids := make([]int, len(l))
	for i := range l {
		ids[i] = l[i].ID
	}
	return auo.RemoveLockerIDs(ids...)
}

// ClearProperties clears all "properties" edges to the Property entity.
func (auo *AccountUpdateOne) ClearProperties() *AccountUpdateOne {
	auo.mutation.ClearProperties()
	return auo
}

// RemovePropertyIDs removes the "properties" edge to Property entities by IDs.
func (auo *AccountUpdateOne) RemovePropertyIDs(ids ...int) *AccountUpdateOne {
	auo.mutation.RemovePropertyIDs(ids...)
	return auo
}

// RemoveProperties removes "properties" edges to Property entities.
func (auo *AccountUpdateOne) RemoveProperties(p ...*Property) *AccountUpdateOne {
	ids := make([]int, len(p))
	for i := range p {
		ids[i] = p[i].ID
	}
	return auo.RemovePropertyIDs(ids...)
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (auo *AccountUpdateOne) Select(field string, fields ...string) *AccountUpdateOne {
	auo.fields = append([]string{field}, fields...)
	return auo
}

// Save executes the query and returns the updated Account entity.
func (auo *AccountUpdateOne) Save(ctx context.Context) (*Account, error) {
	return withHooks[*Account, AccountMutation](ctx, auo.sqlSave, auo.mutation, auo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (auo *AccountUpdateOne) SaveX(ctx context.Context) *Account {
	node, err := auo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (auo *AccountUpdateOne) Exec(ctx context.Context) error {
	_, err := auo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (auo *AccountUpdateOne) ExecX(ctx context.Context) {
	if err := auo.Exec(ctx); err != nil {
		panic(err)
	}
}

func (auo *AccountUpdateOne) sqlSave(ctx context.Context) (_node *Account, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   entaccount.Table,
			Columns: entaccount.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: entaccount.FieldID,
			},
		},
	}
	id, ok := auo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Account.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := auo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, entaccount.FieldID)
		for _, f := range fields {
			if !entaccount.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != entaccount.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := auo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := auo.mutation.Did(); ok {
		_spec.SetField(entaccount.FieldDid, field.TypeString, value)
	}
	if value, ok := auo.mutation.State(); ok {
		_spec.SetField(entaccount.FieldState, field.TypeString, value)
	}
	if value, ok := auo.mutation.Email(); ok {
		_spec.SetField(entaccount.FieldEmail, field.TypeString, value)
	}
	if auo.mutation.EmailCleared() {
		_spec.ClearField(entaccount.FieldEmail, field.TypeString)
	}
	if value, ok := auo.mutation.ParentAccount(); ok {
		_spec.SetField(entaccount.FieldParentAccount, field.TypeString, value)
	}
	if auo.mutation.ParentAccountCleared() {
		_spec.ClearField(entaccount.FieldParentAccount, field.TypeString)
	}
	if value, ok := auo.mutation.Body(); ok {
		_spec.SetField(entaccount.FieldBody, field.TypeJSON, value)
	}
	if auo.mutation.RecoveryCodesCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := auo.mutation.RemovedRecoveryCodesIDs(); len(nodes) > 0 && !auo.mutation.RecoveryCodesCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := auo.mutation.RecoveryCodesIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if auo.mutation.AccessKeysCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := auo.mutation.RemovedAccessKeysIDs(); len(nodes) > 0 && !auo.mutation.AccessKeysCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := auo.mutation.AccessKeysIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if auo.mutation.IdentitiesCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := auo.mutation.RemovedIdentitiesIDs(); len(nodes) > 0 && !auo.mutation.IdentitiesCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := auo.mutation.IdentitiesIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if auo.mutation.LockersCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := auo.mutation.RemovedLockersIDs(); len(nodes) > 0 && !auo.mutation.LockersCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := auo.mutation.LockersIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if auo.mutation.PropertiesCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := auo.mutation.RemovedPropertiesIDs(); len(nodes) > 0 && !auo.mutation.PropertiesCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := auo.mutation.PropertiesIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &Account{config: auo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, auo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{entaccount.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	auo.mutation.done = true
	return _node, nil
}
