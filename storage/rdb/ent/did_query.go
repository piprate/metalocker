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
	"fmt"
	"math"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/piprate/metalocker/storage/rdb/ent/did"
	"github.com/piprate/metalocker/storage/rdb/ent/predicate"
)

// DIDQuery is the builder for querying DID entities.
type DIDQuery struct {
	config
	ctx        *QueryContext
	order      []OrderFunc
	inters     []Interceptor
	predicates []predicate.DID
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the DIDQuery builder.
func (dq *DIDQuery) Where(ps ...predicate.DID) *DIDQuery {
	dq.predicates = append(dq.predicates, ps...)
	return dq
}

// Limit the number of records to be returned by this query.
func (dq *DIDQuery) Limit(limit int) *DIDQuery {
	dq.ctx.Limit = &limit
	return dq
}

// Offset to start from.
func (dq *DIDQuery) Offset(offset int) *DIDQuery {
	dq.ctx.Offset = &offset
	return dq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (dq *DIDQuery) Unique(unique bool) *DIDQuery {
	dq.ctx.Unique = &unique
	return dq
}

// Order specifies how the records should be ordered.
func (dq *DIDQuery) Order(o ...OrderFunc) *DIDQuery {
	dq.order = append(dq.order, o...)
	return dq
}

// First returns the first DID entity from the query.
// Returns a *NotFoundError when no DID was found.
func (dq *DIDQuery) First(ctx context.Context) (*DID, error) {
	nodes, err := dq.Limit(1).All(setContextOp(ctx, dq.ctx, "First"))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{did.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (dq *DIDQuery) FirstX(ctx context.Context) *DID {
	node, err := dq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first DID ID from the query.
// Returns a *NotFoundError when no DID ID was found.
func (dq *DIDQuery) FirstID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = dq.Limit(1).IDs(setContextOp(ctx, dq.ctx, "FirstID")); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{did.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (dq *DIDQuery) FirstIDX(ctx context.Context) int {
	id, err := dq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single DID entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one DID entity is found.
// Returns a *NotFoundError when no DID entities are found.
func (dq *DIDQuery) Only(ctx context.Context) (*DID, error) {
	nodes, err := dq.Limit(2).All(setContextOp(ctx, dq.ctx, "Only"))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{did.Label}
	default:
		return nil, &NotSingularError{did.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (dq *DIDQuery) OnlyX(ctx context.Context) *DID {
	node, err := dq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only DID ID in the query.
// Returns a *NotSingularError when more than one DID ID is found.
// Returns a *NotFoundError when no entities are found.
func (dq *DIDQuery) OnlyID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = dq.Limit(2).IDs(setContextOp(ctx, dq.ctx, "OnlyID")); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{did.Label}
	default:
		err = &NotSingularError{did.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (dq *DIDQuery) OnlyIDX(ctx context.Context) int {
	id, err := dq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of DIDs.
func (dq *DIDQuery) All(ctx context.Context) ([]*DID, error) {
	ctx = setContextOp(ctx, dq.ctx, "All")
	if err := dq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*DID, *DIDQuery]()
	return withInterceptors[[]*DID](ctx, dq, qr, dq.inters)
}

// AllX is like All, but panics if an error occurs.
func (dq *DIDQuery) AllX(ctx context.Context) []*DID {
	nodes, err := dq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of DID IDs.
func (dq *DIDQuery) IDs(ctx context.Context) ([]int, error) {
	var ids []int
	ctx = setContextOp(ctx, dq.ctx, "IDs")
	if err := dq.Select(did.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (dq *DIDQuery) IDsX(ctx context.Context) []int {
	ids, err := dq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (dq *DIDQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, dq.ctx, "Count")
	if err := dq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, dq, querierCount[*DIDQuery](), dq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (dq *DIDQuery) CountX(ctx context.Context) int {
	count, err := dq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (dq *DIDQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, dq.ctx, "Exist")
	switch _, err := dq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (dq *DIDQuery) ExistX(ctx context.Context) bool {
	exist, err := dq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the DIDQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (dq *DIDQuery) Clone() *DIDQuery {
	if dq == nil {
		return nil
	}
	return &DIDQuery{
		config:     dq.config,
		ctx:        dq.ctx.Clone(),
		order:      append([]OrderFunc{}, dq.order...),
		inters:     append([]Interceptor{}, dq.inters...),
		predicates: append([]predicate.DID{}, dq.predicates...),
		// clone intermediate query.
		sql:  dq.sql.Clone(),
		path: dq.path,
	}
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		Did string `json:"did,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.DID.Query().
//		GroupBy(did.FieldDid).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (dq *DIDQuery) GroupBy(field string, fields ...string) *DIDGroupBy {
	dq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &DIDGroupBy{build: dq}
	grbuild.flds = &dq.ctx.Fields
	grbuild.label = did.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		Did string `json:"did,omitempty"`
//	}
//
//	client.DID.Query().
//		Select(did.FieldDid).
//		Scan(ctx, &v)
func (dq *DIDQuery) Select(fields ...string) *DIDSelect {
	dq.ctx.Fields = append(dq.ctx.Fields, fields...)
	sbuild := &DIDSelect{DIDQuery: dq}
	sbuild.label = did.Label
	sbuild.flds, sbuild.scan = &dq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a DIDSelect configured with the given aggregations.
func (dq *DIDQuery) Aggregate(fns ...AggregateFunc) *DIDSelect {
	return dq.Select().Aggregate(fns...)
}

func (dq *DIDQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range dq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, dq); err != nil {
				return err
			}
		}
	}
	for _, f := range dq.ctx.Fields {
		if !did.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if dq.path != nil {
		prev, err := dq.path(ctx)
		if err != nil {
			return err
		}
		dq.sql = prev
	}
	return nil
}

func (dq *DIDQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*DID, error) {
	var (
		nodes = []*DID{}
		_spec = dq.querySpec()
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*DID).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &DID{config: dq.config}
		nodes = append(nodes, node)
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, dq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	return nodes, nil
}

func (dq *DIDQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := dq.querySpec()
	_spec.Node.Columns = dq.ctx.Fields
	if len(dq.ctx.Fields) > 0 {
		_spec.Unique = dq.ctx.Unique != nil && *dq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, dq.driver, _spec)
}

func (dq *DIDQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := &sqlgraph.QuerySpec{
		Node: &sqlgraph.NodeSpec{
			Table:   did.Table,
			Columns: did.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: did.FieldID,
			},
		},
		From:   dq.sql,
		Unique: true,
	}
	if unique := dq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	}
	if fields := dq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, did.FieldID)
		for i := range fields {
			if fields[i] != did.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := dq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := dq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := dq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := dq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (dq *DIDQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(dq.driver.Dialect())
	t1 := builder.Table(did.Table)
	columns := dq.ctx.Fields
	if len(columns) == 0 {
		columns = did.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if dq.sql != nil {
		selector = dq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if dq.ctx.Unique != nil && *dq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range dq.predicates {
		p(selector)
	}
	for _, p := range dq.order {
		p(selector)
	}
	if offset := dq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := dq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// DIDGroupBy is the group-by builder for DID entities.
type DIDGroupBy struct {
	selector
	build *DIDQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (dgb *DIDGroupBy) Aggregate(fns ...AggregateFunc) *DIDGroupBy {
	dgb.fns = append(dgb.fns, fns...)
	return dgb
}

// Scan applies the selector query and scans the result into the given value.
func (dgb *DIDGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, dgb.build.ctx, "GroupBy")
	if err := dgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*DIDQuery, *DIDGroupBy](ctx, dgb.build, dgb, dgb.build.inters, v)
}

func (dgb *DIDGroupBy) sqlScan(ctx context.Context, root *DIDQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(dgb.fns))
	for _, fn := range dgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*dgb.flds)+len(dgb.fns))
		for _, f := range *dgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*dgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := dgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// DIDSelect is the builder for selecting fields of DID entities.
type DIDSelect struct {
	*DIDQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (ds *DIDSelect) Aggregate(fns ...AggregateFunc) *DIDSelect {
	ds.fns = append(ds.fns, fns...)
	return ds
}

// Scan applies the selector query and scans the result into the given value.
func (ds *DIDSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ds.ctx, "Select")
	if err := ds.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*DIDQuery, *DIDSelect](ctx, ds.DIDQuery, ds, ds.inters, v)
}

func (ds *DIDSelect) sqlScan(ctx context.Context, root *DIDQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(ds.fns))
	for _, fn := range ds.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*ds.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ds.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
