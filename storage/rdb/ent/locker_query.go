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
	"github.com/piprate/metalocker/storage/rdb/ent/locker"
	"github.com/piprate/metalocker/storage/rdb/ent/predicate"

	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
)

// LockerQuery is the builder for querying Locker entities.
type LockerQuery struct {
	config
	ctx         *QueryContext
	order       []OrderFunc
	inters      []Interceptor
	predicates  []predicate.Locker
	withAccount *AccountQuery
	withFKs     bool
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the LockerQuery builder.
func (lq *LockerQuery) Where(ps ...predicate.Locker) *LockerQuery {
	lq.predicates = append(lq.predicates, ps...)
	return lq
}

// Limit the number of records to be returned by this query.
func (lq *LockerQuery) Limit(limit int) *LockerQuery {
	lq.ctx.Limit = &limit
	return lq
}

// Offset to start from.
func (lq *LockerQuery) Offset(offset int) *LockerQuery {
	lq.ctx.Offset = &offset
	return lq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (lq *LockerQuery) Unique(unique bool) *LockerQuery {
	lq.ctx.Unique = &unique
	return lq
}

// Order specifies how the records should be ordered.
func (lq *LockerQuery) Order(o ...OrderFunc) *LockerQuery {
	lq.order = append(lq.order, o...)
	return lq
}

// QueryAccount chains the current query on the "account" edge.
func (lq *LockerQuery) QueryAccount() *AccountQuery {
	query := (&AccountClient{config: lq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := lq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := lq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(locker.Table, locker.FieldID, selector),
			sqlgraph.To(entaccount.Table, entaccount.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, locker.AccountTable, locker.AccountColumn),
		)
		fromU = sqlgraph.SetNeighbors(lq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first Locker entity from the query.
// Returns a *NotFoundError when no Locker was found.
func (lq *LockerQuery) First(ctx context.Context) (*Locker, error) {
	nodes, err := lq.Limit(1).All(setContextOp(ctx, lq.ctx, "First"))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{locker.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (lq *LockerQuery) FirstX(ctx context.Context) *Locker {
	node, err := lq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first Locker ID from the query.
// Returns a *NotFoundError when no Locker ID was found.
func (lq *LockerQuery) FirstID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = lq.Limit(1).IDs(setContextOp(ctx, lq.ctx, "FirstID")); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{locker.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (lq *LockerQuery) FirstIDX(ctx context.Context) int {
	id, err := lq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single Locker entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one Locker entity is found.
// Returns a *NotFoundError when no Locker entities are found.
func (lq *LockerQuery) Only(ctx context.Context) (*Locker, error) {
	nodes, err := lq.Limit(2).All(setContextOp(ctx, lq.ctx, "Only"))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{locker.Label}
	default:
		return nil, &NotSingularError{locker.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (lq *LockerQuery) OnlyX(ctx context.Context) *Locker {
	node, err := lq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only Locker ID in the query.
// Returns a *NotSingularError when more than one Locker ID is found.
// Returns a *NotFoundError when no entities are found.
func (lq *LockerQuery) OnlyID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = lq.Limit(2).IDs(setContextOp(ctx, lq.ctx, "OnlyID")); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{locker.Label}
	default:
		err = &NotSingularError{locker.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (lq *LockerQuery) OnlyIDX(ctx context.Context) int {
	id, err := lq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of Lockers.
func (lq *LockerQuery) All(ctx context.Context) ([]*Locker, error) {
	ctx = setContextOp(ctx, lq.ctx, "All")
	if err := lq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*Locker, *LockerQuery]()
	return withInterceptors[[]*Locker](ctx, lq, qr, lq.inters)
}

// AllX is like All, but panics if an error occurs.
func (lq *LockerQuery) AllX(ctx context.Context) []*Locker {
	nodes, err := lq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of Locker IDs.
func (lq *LockerQuery) IDs(ctx context.Context) ([]int, error) {
	var ids []int
	ctx = setContextOp(ctx, lq.ctx, "IDs")
	if err := lq.Select(locker.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (lq *LockerQuery) IDsX(ctx context.Context) []int {
	ids, err := lq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (lq *LockerQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, lq.ctx, "Count")
	if err := lq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, lq, querierCount[*LockerQuery](), lq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (lq *LockerQuery) CountX(ctx context.Context) int {
	count, err := lq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (lq *LockerQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, lq.ctx, "Exist")
	switch _, err := lq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (lq *LockerQuery) ExistX(ctx context.Context) bool {
	exist, err := lq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the LockerQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (lq *LockerQuery) Clone() *LockerQuery {
	if lq == nil {
		return nil
	}
	return &LockerQuery{
		config:      lq.config,
		ctx:         lq.ctx.Clone(),
		order:       append([]OrderFunc{}, lq.order...),
		inters:      append([]Interceptor{}, lq.inters...),
		predicates:  append([]predicate.Locker{}, lq.predicates...),
		withAccount: lq.withAccount.Clone(),
		// clone intermediate query.
		sql:  lq.sql.Clone(),
		path: lq.path,
	}
}

// WithAccount tells the query-builder to eager-load the nodes that are connected to
// the "account" edge. The optional arguments are used to configure the query builder of the edge.
func (lq *LockerQuery) WithAccount(opts ...func(*AccountQuery)) *LockerQuery {
	query := (&AccountClient{config: lq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	lq.withAccount = query
	return lq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		Hash string `json:"hash,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.Locker.Query().
//		GroupBy(locker.FieldHash).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (lq *LockerQuery) GroupBy(field string, fields ...string) *LockerGroupBy {
	lq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &LockerGroupBy{build: lq}
	grbuild.flds = &lq.ctx.Fields
	grbuild.label = locker.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		Hash string `json:"hash,omitempty"`
//	}
//
//	client.Locker.Query().
//		Select(locker.FieldHash).
//		Scan(ctx, &v)
func (lq *LockerQuery) Select(fields ...string) *LockerSelect {
	lq.ctx.Fields = append(lq.ctx.Fields, fields...)
	sbuild := &LockerSelect{LockerQuery: lq}
	sbuild.label = locker.Label
	sbuild.flds, sbuild.scan = &lq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a LockerSelect configured with the given aggregations.
func (lq *LockerQuery) Aggregate(fns ...AggregateFunc) *LockerSelect {
	return lq.Select().Aggregate(fns...)
}

func (lq *LockerQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range lq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, lq); err != nil {
				return err
			}
		}
	}
	for _, f := range lq.ctx.Fields {
		if !locker.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if lq.path != nil {
		prev, err := lq.path(ctx)
		if err != nil {
			return err
		}
		lq.sql = prev
	}
	return nil
}

func (lq *LockerQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*Locker, error) {
	var (
		nodes       = []*Locker{}
		withFKs     = lq.withFKs
		_spec       = lq.querySpec()
		loadedTypes = [1]bool{
			lq.withAccount != nil,
		}
	)
	if lq.withAccount != nil {
		withFKs = true
	}
	if withFKs {
		_spec.Node.Columns = append(_spec.Node.Columns, locker.ForeignKeys...)
	}
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*Locker).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &Locker{config: lq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, lq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := lq.withAccount; query != nil {
		if err := lq.loadAccount(ctx, query, nodes, nil,
			func(n *Locker, e *Account) { n.Edges.Account = e }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (lq *LockerQuery) loadAccount(ctx context.Context, query *AccountQuery, nodes []*Locker, init func(*Locker), assign func(*Locker, *Account)) error {
	ids := make([]int, 0, len(nodes))
	nodeids := make(map[int][]*Locker)
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

func (lq *LockerQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := lq.querySpec()
	_spec.Node.Columns = lq.ctx.Fields
	if len(lq.ctx.Fields) > 0 {
		_spec.Unique = lq.ctx.Unique != nil && *lq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, lq.driver, _spec)
}

func (lq *LockerQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := &sqlgraph.QuerySpec{
		Node: &sqlgraph.NodeSpec{
			Table:   locker.Table,
			Columns: locker.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: locker.FieldID,
			},
		},
		From:   lq.sql,
		Unique: true,
	}
	if unique := lq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	}
	if fields := lq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, locker.FieldID)
		for i := range fields {
			if fields[i] != locker.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := lq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := lq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := lq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := lq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (lq *LockerQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(lq.driver.Dialect())
	t1 := builder.Table(locker.Table)
	columns := lq.ctx.Fields
	if len(columns) == 0 {
		columns = locker.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if lq.sql != nil {
		selector = lq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if lq.ctx.Unique != nil && *lq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range lq.predicates {
		p(selector)
	}
	for _, p := range lq.order {
		p(selector)
	}
	if offset := lq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := lq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// LockerGroupBy is the group-by builder for Locker entities.
type LockerGroupBy struct {
	selector
	build *LockerQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (lgb *LockerGroupBy) Aggregate(fns ...AggregateFunc) *LockerGroupBy {
	lgb.fns = append(lgb.fns, fns...)
	return lgb
}

// Scan applies the selector query and scans the result into the given value.
func (lgb *LockerGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, lgb.build.ctx, "GroupBy")
	if err := lgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*LockerQuery, *LockerGroupBy](ctx, lgb.build, lgb, lgb.build.inters, v)
}

func (lgb *LockerGroupBy) sqlScan(ctx context.Context, root *LockerQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(lgb.fns))
	for _, fn := range lgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*lgb.flds)+len(lgb.fns))
		for _, f := range *lgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*lgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := lgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// LockerSelect is the builder for selecting fields of Locker entities.
type LockerSelect struct {
	*LockerQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (ls *LockerSelect) Aggregate(fns ...AggregateFunc) *LockerSelect {
	ls.fns = append(ls.fns, fns...)
	return ls
}

// Scan applies the selector query and scans the result into the given value.
func (ls *LockerSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ls.ctx, "Select")
	if err := ls.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*LockerQuery, *LockerSelect](ctx, ls.LockerQuery, ls, ls.inters, v)
}

func (ls *LockerSelect) sqlScan(ctx context.Context, root *LockerQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(ls.fns))
	for _, fn := range ls.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*ls.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ls.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
