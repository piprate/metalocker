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

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/piprate/metalocker/storage/rdb/ent/predicate"
	"github.com/piprate/metalocker/storage/rdb/ent/recoverycode"
)

// RecoveryCodeDelete is the builder for deleting a RecoveryCode entity.
type RecoveryCodeDelete struct {
	config
	hooks    []Hook
	mutation *RecoveryCodeMutation
}

// Where appends a list predicates to the RecoveryCodeDelete builder.
func (rcd *RecoveryCodeDelete) Where(ps ...predicate.RecoveryCode) *RecoveryCodeDelete {
	rcd.mutation.Where(ps...)
	return rcd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (rcd *RecoveryCodeDelete) Exec(ctx context.Context) (int, error) {
	return withHooks[int, RecoveryCodeMutation](ctx, rcd.sqlExec, rcd.mutation, rcd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (rcd *RecoveryCodeDelete) ExecX(ctx context.Context) int {
	n, err := rcd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (rcd *RecoveryCodeDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := &sqlgraph.DeleteSpec{
		Node: &sqlgraph.NodeSpec{
			Table: recoverycode.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: recoverycode.FieldID,
			},
		},
	}
	if ps := rcd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, rcd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	rcd.mutation.done = true
	return affected, err
}

// RecoveryCodeDeleteOne is the builder for deleting a single RecoveryCode entity.
type RecoveryCodeDeleteOne struct {
	rcd *RecoveryCodeDelete
}

// Exec executes the deletion query.
func (rcdo *RecoveryCodeDeleteOne) Exec(ctx context.Context) error {
	n, err := rcdo.rcd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{recoverycode.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (rcdo *RecoveryCodeDeleteOne) ExecX(ctx context.Context) {
	rcdo.rcd.ExecX(ctx)
}
