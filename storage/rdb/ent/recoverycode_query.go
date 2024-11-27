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
	"github.com/piprate/metalocker/storage/rdb/ent/predicate"
	"github.com/piprate/metalocker/storage/rdb/ent/recoverycode"

	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
)

// RecoveryCodeQuery is the builder for querying RecoveryCode entities.
type RecoveryCodeQuery struct {
	config
	ctx         *QueryContext
	order       []OrderFunc
	inters      []Interceptor
	predicates  []predicate.RecoveryCode
	withAccount *AccountQuery
	withFKs     bool
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the RecoveryCodeQuery builder.
func (rcq *RecoveryCodeQuery) Where(ps ...predicate.RecoveryCode) *RecoveryCodeQuery {
	rcq.predicates = append(rcq.predicates, ps...)
	return rcq
}

// Limit the number of records to be returned by this query.
func (rcq *RecoveryCodeQuery) Limit(limit int) *RecoveryCodeQuery {
	rcq.ctx.Limit = &limit
	return rcq
}

// Offset to start from.
func (rcq *RecoveryCodeQuery) Offset(offset int) *RecoveryCodeQuery {
	rcq.ctx.Offset = &offset
	return rcq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (rcq *RecoveryCodeQuery) Unique(unique bool) *RecoveryCodeQuery {
	rcq.ctx.Unique = &unique
	return rcq
}

// Order specifies how the records should be ordered.
func (rcq *RecoveryCodeQuery) Order(o ...OrderFunc) *RecoveryCodeQuery {
	rcq.order = append(rcq.order, o...)
	return rcq
}

// QueryAccount chains the current query on the "account" edge.
func (rcq *RecoveryCodeQuery) QueryAccount() *AccountQuery {
	query := (&AccountClient{config: rcq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := rcq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := rcq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(recoverycode.Table, recoverycode.FieldID, selector),
			sqlgraph.To(entaccount.Table, entaccount.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, recoverycode.AccountTable, recoverycode.AccountColumn),
		)
		fromU = sqlgraph.SetNeighbors(rcq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first RecoveryCode entity from the query.
// Returns a *NotFoundError when no RecoveryCode was found.
func (rcq *RecoveryCodeQuery) First(ctx context.Context) (*RecoveryCode, error) {
	nodes, err := rcq.Limit(1).All(setContextOp(ctx, rcq.ctx, "First"))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{recoverycode.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (rcq *RecoveryCodeQuery) FirstX(ctx context.Context) *RecoveryCode {
	node, err := rcq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first RecoveryCode ID from the query.
// Returns a *NotFoundError when no RecoveryCode ID was found.
func (rcq *RecoveryCodeQuery) FirstID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = rcq.Limit(1).IDs(setContextOp(ctx, rcq.ctx, "FirstID")); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{recoverycode.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (rcq *RecoveryCodeQuery) FirstIDX(ctx context.Context) int {
	id, err := rcq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single RecoveryCode entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one RecoveryCode entity is found.
// Returns a *NotFoundError when no RecoveryCode entities are found.
func (rcq *RecoveryCodeQuery) Only(ctx context.Context) (*RecoveryCode, error) {
	nodes, err := rcq.Limit(2).All(setContextOp(ctx, rcq.ctx, "Only"))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{recoverycode.Label}
	default:
		return nil, &NotSingularError{recoverycode.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (rcq *RecoveryCodeQuery) OnlyX(ctx context.Context) *RecoveryCode {
	node, err := rcq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only RecoveryCode ID in the query.
// Returns a *NotSingularError when more than one RecoveryCode ID is found.
// Returns a *NotFoundError when no entities are found.
func (rcq *RecoveryCodeQuery) OnlyID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = rcq.Limit(2).IDs(setContextOp(ctx, rcq.ctx, "OnlyID")); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{recoverycode.Label}
	default:
		err = &NotSingularError{recoverycode.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (rcq *RecoveryCodeQuery) OnlyIDX(ctx context.Context) int {
	id, err := rcq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of RecoveryCodes.
func (rcq *RecoveryCodeQuery) All(ctx context.Context) ([]*RecoveryCode, error) {
	ctx = setContextOp(ctx, rcq.ctx, "All")
	if err := rcq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*RecoveryCode, *RecoveryCodeQuery]()
	return withInterceptors[[]*RecoveryCode](ctx, rcq, qr, rcq.inters)
}

// AllX is like All, but panics if an error occurs.
func (rcq *RecoveryCodeQuery) AllX(ctx context.Context) []*RecoveryCode {
	nodes, err := rcq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of RecoveryCode IDs.
func (rcq *RecoveryCodeQuery) IDs(ctx context.Context) ([]int, error) {
	var ids []int
	ctx = setContextOp(ctx, rcq.ctx, "IDs")
	if err := rcq.Select(recoverycode.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (rcq *RecoveryCodeQuery) IDsX(ctx context.Context) []int {
	ids, err := rcq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (rcq *RecoveryCodeQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, rcq.ctx, "Count")
	if err := rcq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, rcq, querierCount[*RecoveryCodeQuery](), rcq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (rcq *RecoveryCodeQuery) CountX(ctx context.Context) int {
	count, err := rcq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (rcq *RecoveryCodeQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, rcq.ctx, "Exist")
	switch _, err := rcq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (rcq *RecoveryCodeQuery) ExistX(ctx context.Context) bool {
	exist, err := rcq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the RecoveryCodeQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (rcq *RecoveryCodeQuery) Clone() *RecoveryCodeQuery {
	if rcq == nil {
		return nil
	}
	return &RecoveryCodeQuery{
		config:      rcq.config,
		ctx:         rcq.ctx.Clone(),
		order:       append([]OrderFunc{}, rcq.order...),
		inters:      append([]Interceptor{}, rcq.inters...),
		predicates:  append([]predicate.RecoveryCode{}, rcq.predicates...),
		withAccount: rcq.withAccount.Clone(),
		// clone intermediate query.
		sql:  rcq.sql.Clone(),
		path: rcq.path,
	}
}

// WithAccount tells the query-builder to eager-load the nodes that are connected to
// the "account" edge. The optional arguments are used to configure the query builder of the edge.
func (rcq *RecoveryCodeQuery) WithAccount(opts ...func(*AccountQuery)) *RecoveryCodeQuery {
	query := (&AccountClient{config: rcq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	rcq.withAccount = query
	return rcq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		Code string `json:"code,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.RecoveryCode.Query().
//		GroupBy(recoverycode.FieldCode).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (rcq *RecoveryCodeQuery) GroupBy(field string, fields ...string) *RecoveryCodeGroupBy {
	rcq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &RecoveryCodeGroupBy{build: rcq}
	grbuild.flds = &rcq.ctx.Fields
	grbuild.label = recoverycode.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		Code string `json:"code,omitempty"`
//	}
//
//	client.RecoveryCode.Query().
//		Select(recoverycode.FieldCode).
//		Scan(ctx, &v)
func (rcq *RecoveryCodeQuery) Select(fields ...string) *RecoveryCodeSelect {
	rcq.ctx.Fields = append(rcq.ctx.Fields, fields...)
	sbuild := &RecoveryCodeSelect{RecoveryCodeQuery: rcq}
	sbuild.label = recoverycode.Label
	sbuild.flds, sbuild.scan = &rcq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a RecoveryCodeSelect configured with the given aggregations.
func (rcq *RecoveryCodeQuery) Aggregate(fns ...AggregateFunc) *RecoveryCodeSelect {
	return rcq.Select().Aggregate(fns...)
}

func (rcq *RecoveryCodeQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range rcq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, rcq); err != nil {
				return err
			}
		}
	}
	for _, f := range rcq.ctx.Fields {
		if !recoverycode.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if rcq.path != nil {
		prev, err := rcq.path(ctx)
		if err != nil {
			return err
		}
		rcq.sql = prev
	}
	return nil
}

func (rcq *RecoveryCodeQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*RecoveryCode, error) {
	var (
		nodes       = []*RecoveryCode{}
		withFKs     = rcq.withFKs
		_spec       = rcq.querySpec()
		loadedTypes = [1]bool{
			rcq.withAccount != nil,
		}
	)
	if rcq.withAccount != nil {
		withFKs = true
	}
	if withFKs {
		_spec.Node.Columns = append(_spec.Node.Columns, recoverycode.ForeignKeys...)
	}
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*RecoveryCode).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &RecoveryCode{config: rcq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, rcq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := rcq.withAccount; query != nil {
		if err := rcq.loadAccount(ctx, query, nodes, nil,
			func(n *RecoveryCode, e *Account) { n.Edges.Account = e }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (rcq *RecoveryCodeQuery) loadAccount(ctx context.Context, query *AccountQuery, nodes []*RecoveryCode, init func(*RecoveryCode), assign func(*RecoveryCode, *Account)) error {
	ids := make([]int, 0, len(nodes))
	nodeids := make(map[int][]*RecoveryCode)
	for i := range nodes {
		if nodes[i].account == nil {
			continue
		}
		fk := *nodes[i].account
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(entaccount.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "account" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}

func (rcq *RecoveryCodeQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := rcq.querySpec()
	_spec.Node.Columns = rcq.ctx.Fields
	if len(rcq.ctx.Fields) > 0 {
		_spec.Unique = rcq.ctx.Unique != nil && *rcq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, rcq.driver, _spec)
}

func (rcq *RecoveryCodeQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := &sqlgraph.QuerySpec{
		Node: &sqlgraph.NodeSpec{
			Table:   recoverycode.Table,
			Columns: recoverycode.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: recoverycode.FieldID,
			},
		},
		From:   rcq.sql,
		Unique: true,
	}
	if unique := rcq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	}
	if fields := rcq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, recoverycode.FieldID)
		for i := range fields {
			if fields[i] != recoverycode.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := rcq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := rcq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := rcq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := rcq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (rcq *RecoveryCodeQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(rcq.driver.Dialect())
	t1 := builder.Table(recoverycode.Table)
	columns := rcq.ctx.Fields
	if len(columns) == 0 {
		columns = recoverycode.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if rcq.sql != nil {
		selector = rcq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if rcq.ctx.Unique != nil && *rcq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range rcq.predicates {
		p(selector)
	}
	for _, p := range rcq.order {
		p(selector)
	}
	if offset := rcq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := rcq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// RecoveryCodeGroupBy is the group-by builder for RecoveryCode entities.
type RecoveryCodeGroupBy struct {
	selector
	build *RecoveryCodeQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (rcgb *RecoveryCodeGroupBy) Aggregate(fns ...AggregateFunc) *RecoveryCodeGroupBy {
	rcgb.fns = append(rcgb.fns, fns...)
	return rcgb
}

// Scan applies the selector query and scans the result into the given value.
func (rcgb *RecoveryCodeGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, rcgb.build.ctx, "GroupBy")
	if err := rcgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*RecoveryCodeQuery, *RecoveryCodeGroupBy](ctx, rcgb.build, rcgb, rcgb.build.inters, v)
}

func (rcgb *RecoveryCodeGroupBy) sqlScan(ctx context.Context, root *RecoveryCodeQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(rcgb.fns))
	for _, fn := range rcgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*rcgb.flds)+len(rcgb.fns))
		for _, f := range *rcgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*rcgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := rcgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// RecoveryCodeSelect is the builder for selecting fields of RecoveryCode entities.
type RecoveryCodeSelect struct {
	*RecoveryCodeQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (rcs *RecoveryCodeSelect) Aggregate(fns ...AggregateFunc) *RecoveryCodeSelect {
	rcs.fns = append(rcs.fns, fns...)
	return rcs
}

// Scan applies the selector query and scans the result into the given value.
func (rcs *RecoveryCodeSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, rcs.ctx, "Select")
	if err := rcs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*RecoveryCodeQuery, *RecoveryCodeSelect](ctx, rcs.RecoveryCodeQuery, rcs, rcs.inters, v)
}

func (rcs *RecoveryCodeSelect) sqlScan(ctx context.Context, root *RecoveryCodeQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(rcs.fns))
	for _, fn := range rcs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*rcs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := rcs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
