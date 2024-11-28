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
	"github.com/piprate/metalocker/storage/rdb/ent/property"

	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
)

// PropertyQuery is the builder for querying Property entities.
type PropertyQuery struct {
	config
	ctx         *QueryContext
	order       []OrderFunc
	inters      []Interceptor
	predicates  []predicate.Property
	withAccount *AccountQuery
	withFKs     bool
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the PropertyQuery builder.
func (pq *PropertyQuery) Where(ps ...predicate.Property) *PropertyQuery {
	pq.predicates = append(pq.predicates, ps...)
	return pq
}

// Limit the number of records to be returned by this query.
func (pq *PropertyQuery) Limit(limit int) *PropertyQuery {
	pq.ctx.Limit = &limit
	return pq
}

// Offset to start from.
func (pq *PropertyQuery) Offset(offset int) *PropertyQuery {
	pq.ctx.Offset = &offset
	return pq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (pq *PropertyQuery) Unique(unique bool) *PropertyQuery {
	pq.ctx.Unique = &unique
	return pq
}

// Order specifies how the records should be ordered.
func (pq *PropertyQuery) Order(o ...OrderFunc) *PropertyQuery {
	pq.order = append(pq.order, o...)
	return pq
}

// QueryAccount chains the current query on the "account" edge.
func (pq *PropertyQuery) QueryAccount() *AccountQuery {
	query := (&AccountClient{config: pq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := pq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := pq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(property.Table, property.FieldID, selector),
			sqlgraph.To(entaccount.Table, entaccount.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, property.AccountTable, property.AccountColumn),
		)
		fromU = sqlgraph.SetNeighbors(pq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first Property entity from the query.
// Returns a *NotFoundError when no Property was found.
func (pq *PropertyQuery) First(ctx context.Context) (*Property, error) {
	nodes, err := pq.Limit(1).All(setContextOp(ctx, pq.ctx, "First"))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{property.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (pq *PropertyQuery) FirstX(ctx context.Context) *Property {
	node, err := pq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first Property ID from the query.
// Returns a *NotFoundError when no Property ID was found.
func (pq *PropertyQuery) FirstID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = pq.Limit(1).IDs(setContextOp(ctx, pq.ctx, "FirstID")); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{property.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (pq *PropertyQuery) FirstIDX(ctx context.Context) int {
	id, err := pq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single Property entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one Property entity is found.
// Returns a *NotFoundError when no Property entities are found.
func (pq *PropertyQuery) Only(ctx context.Context) (*Property, error) {
	nodes, err := pq.Limit(2).All(setContextOp(ctx, pq.ctx, "Only"))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{property.Label}
	default:
		return nil, &NotSingularError{property.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (pq *PropertyQuery) OnlyX(ctx context.Context) *Property {
	node, err := pq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only Property ID in the query.
// Returns a *NotSingularError when more than one Property ID is found.
// Returns a *NotFoundError when no entities are found.
func (pq *PropertyQuery) OnlyID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = pq.Limit(2).IDs(setContextOp(ctx, pq.ctx, "OnlyID")); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{property.Label}
	default:
		err = &NotSingularError{property.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (pq *PropertyQuery) OnlyIDX(ctx context.Context) int {
	id, err := pq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of Properties.
func (pq *PropertyQuery) All(ctx context.Context) ([]*Property, error) {
	ctx = setContextOp(ctx, pq.ctx, "All")
	if err := pq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*Property, *PropertyQuery]()
	return withInterceptors[[]*Property](ctx, pq, qr, pq.inters)
}

// AllX is like All, but panics if an error occurs.
func (pq *PropertyQuery) AllX(ctx context.Context) []*Property {
	nodes, err := pq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of Property IDs.
func (pq *PropertyQuery) IDs(ctx context.Context) ([]int, error) {
	var ids []int
	ctx = setContextOp(ctx, pq.ctx, "IDs")
	if err := pq.Select(property.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (pq *PropertyQuery) IDsX(ctx context.Context) []int {
	ids, err := pq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (pq *PropertyQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, pq.ctx, "Count")
	if err := pq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, pq, querierCount[*PropertyQuery](), pq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (pq *PropertyQuery) CountX(ctx context.Context) int {
	count, err := pq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (pq *PropertyQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, pq.ctx, "Exist")
	switch _, err := pq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (pq *PropertyQuery) ExistX(ctx context.Context) bool {
	exist, err := pq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the PropertyQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (pq *PropertyQuery) Clone() *PropertyQuery {
	if pq == nil {
		return nil
	}
	return &PropertyQuery{
		config:      pq.config,
		ctx:         pq.ctx.Clone(),
		order:       append([]OrderFunc{}, pq.order...),
		inters:      append([]Interceptor{}, pq.inters...),
		predicates:  append([]predicate.Property{}, pq.predicates...),
		withAccount: pq.withAccount.Clone(),
		// clone intermediate query.
		sql:  pq.sql.Clone(),
		path: pq.path,
	}
}

// WithAccount tells the query-builder to eager-load the nodes that are connected to
// the "account" edge. The optional arguments are used to configure the query builder of the edge.
func (pq *PropertyQuery) WithAccount(opts ...func(*AccountQuery)) *PropertyQuery {
	query := (&AccountClient{config: pq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	pq.withAccount = query
	return pq
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
//	client.Property.Query().
//		GroupBy(property.FieldHash).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (pq *PropertyQuery) GroupBy(field string, fields ...string) *PropertyGroupBy {
	pq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &PropertyGroupBy{build: pq}
	grbuild.flds = &pq.ctx.Fields
	grbuild.label = property.Label
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
//	client.Property.Query().
//		Select(property.FieldHash).
//		Scan(ctx, &v)
func (pq *PropertyQuery) Select(fields ...string) *PropertySelect {
	pq.ctx.Fields = append(pq.ctx.Fields, fields...)
	sbuild := &PropertySelect{PropertyQuery: pq}
	sbuild.label = property.Label
	sbuild.flds, sbuild.scan = &pq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a PropertySelect configured with the given aggregations.
func (pq *PropertyQuery) Aggregate(fns ...AggregateFunc) *PropertySelect {
	return pq.Select().Aggregate(fns...)
}

func (pq *PropertyQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range pq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, pq); err != nil {
				return err
			}
		}
	}
	for _, f := range pq.ctx.Fields {
		if !property.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if pq.path != nil {
		prev, err := pq.path(ctx)
		if err != nil {
			return err
		}
		pq.sql = prev
	}
	return nil
}

func (pq *PropertyQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*Property, error) {
	var (
		nodes       = []*Property{}
		withFKs     = pq.withFKs
		_spec       = pq.querySpec()
		loadedTypes = [1]bool{
			pq.withAccount != nil,
		}
	)
	if pq.withAccount != nil {
		withFKs = true
	}
	if withFKs {
		_spec.Node.Columns = append(_spec.Node.Columns, property.ForeignKeys...)
	}
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*Property).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &Property{config: pq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, pq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := pq.withAccount; query != nil {
		if err := pq.loadAccount(ctx, query, nodes, nil,
			func(n *Property, e *Account) { n.Edges.Account = e }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (pq *PropertyQuery) loadAccount(ctx context.Context, query *AccountQuery, nodes []*Property, init func(*Property), assign func(*Property, *Account)) error {
	ids := make([]int, 0, len(nodes))
	nodeids := make(map[int][]*Property)
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

func (pq *PropertyQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := pq.querySpec()
	_spec.Node.Columns = pq.ctx.Fields
	if len(pq.ctx.Fields) > 0 {
		_spec.Unique = pq.ctx.Unique != nil && *pq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, pq.driver, _spec)
}

func (pq *PropertyQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := &sqlgraph.QuerySpec{
		Node: &sqlgraph.NodeSpec{
			Table:   property.Table,
			Columns: property.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: property.FieldID,
			},
		},
		From:   pq.sql,
		Unique: true,
	}
	if unique := pq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	}
	if fields := pq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, property.FieldID)
		for i := range fields {
			if fields[i] != property.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := pq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := pq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := pq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := pq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (pq *PropertyQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(pq.driver.Dialect())
	t1 := builder.Table(property.Table)
	columns := pq.ctx.Fields
	if len(columns) == 0 {
		columns = property.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if pq.sql != nil {
		selector = pq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if pq.ctx.Unique != nil && *pq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range pq.predicates {
		p(selector)
	}
	for _, p := range pq.order {
		p(selector)
	}
	if offset := pq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := pq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// PropertyGroupBy is the group-by builder for Property entities.
type PropertyGroupBy struct {
	selector
	build *PropertyQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (pgb *PropertyGroupBy) Aggregate(fns ...AggregateFunc) *PropertyGroupBy {
	pgb.fns = append(pgb.fns, fns...)
	return pgb
}

// Scan applies the selector query and scans the result into the given value.
func (pgb *PropertyGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, pgb.build.ctx, "GroupBy")
	if err := pgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*PropertyQuery, *PropertyGroupBy](ctx, pgb.build, pgb, pgb.build.inters, v)
}

func (pgb *PropertyGroupBy) sqlScan(ctx context.Context, root *PropertyQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(pgb.fns))
	for _, fn := range pgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*pgb.flds)+len(pgb.fns))
		for _, f := range *pgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*pgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := pgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// PropertySelect is the builder for selecting fields of Property entities.
type PropertySelect struct {
	*PropertyQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (ps *PropertySelect) Aggregate(fns ...AggregateFunc) *PropertySelect {
	ps.fns = append(ps.fns, fns...)
	return ps
}

// Scan applies the selector query and scans the result into the given value.
func (ps *PropertySelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ps.ctx, "Select")
	if err := ps.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*PropertyQuery, *PropertySelect](ctx, ps.PropertyQuery, ps, ps.inters, v)
}

func (ps *PropertySelect) sqlScan(ctx context.Context, root *PropertyQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(ps.fns))
	for _, fn := range ps.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*ps.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ps.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
