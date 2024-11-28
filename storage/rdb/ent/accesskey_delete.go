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
	"github.com/piprate/metalocker/storage/rdb/ent/accesskey"
	"github.com/piprate/metalocker/storage/rdb/ent/predicate"
)

// AccessKeyDelete is the builder for deleting a AccessKey entity.
type AccessKeyDelete struct {
	config
	hooks    []Hook
	mutation *AccessKeyMutation
}

// Where appends a list predicates to the AccessKeyDelete builder.
func (akd *AccessKeyDelete) Where(ps ...predicate.AccessKey) *AccessKeyDelete {
	akd.mutation.Where(ps...)
	return akd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (akd *AccessKeyDelete) Exec(ctx context.Context) (int, error) {
	return withHooks[int, AccessKeyMutation](ctx, akd.sqlExec, akd.mutation, akd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (akd *AccessKeyDelete) ExecX(ctx context.Context) int {
	n, err := akd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (akd *AccessKeyDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := &sqlgraph.DeleteSpec{
		Node: &sqlgraph.NodeSpec{
			Table: accesskey.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: accesskey.FieldID,
			},
		},
	}
	if ps := akd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, akd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	akd.mutation.done = true
	return affected, err
}

// AccessKeyDeleteOne is the builder for deleting a single AccessKey entity.
type AccessKeyDeleteOne struct {
	akd *AccessKeyDelete
}

// Exec executes the deletion query.
func (akdo *AccessKeyDeleteOne) Exec(ctx context.Context) error {
	n, err := akdo.akd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{accesskey.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (akdo *AccessKeyDeleteOne) ExecX(ctx context.Context) {
	akdo.akd.ExecX(ctx)
}
