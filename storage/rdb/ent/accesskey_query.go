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
	"github.com/piprate/metalocker/storage/rdb/ent/accesskey"
	"github.com/piprate/metalocker/storage/rdb/ent/predicate"

	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
)

// AccessKeyQuery is the builder for querying AccessKey entities.
type AccessKeyQuery struct {
	config
	ctx         *QueryContext
	order       []OrderFunc
	inters      []Interceptor
	predicates  []predicate.AccessKey
	withAccount *AccountQuery
	withFKs     bool
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the AccessKeyQuery builder.
func (akq *AccessKeyQuery) Where(ps ...predicate.AccessKey) *AccessKeyQuery {
	akq.predicates = append(akq.predicates, ps...)
	return akq
}

// Limit the number of records to be returned by this query.
func (akq *AccessKeyQuery) Limit(limit int) *AccessKeyQuery {
	akq.ctx.Limit = &limit
	return akq
}

// Offset to start from.
func (akq *AccessKeyQuery) Offset(offset int) *AccessKeyQuery {
	akq.ctx.Offset = &offset
	return akq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (akq *AccessKeyQuery) Unique(unique bool) *AccessKeyQuery {
	akq.ctx.Unique = &unique
	return akq
}

// Order specifies how the records should be ordered.
func (akq *AccessKeyQuery) Order(o ...OrderFunc) *AccessKeyQuery {
	akq.order = append(akq.order, o...)
	return akq
}

// QueryAccount chains the current query on the "account" edge.
func (akq *AccessKeyQuery) QueryAccount() *AccountQuery {
	query := (&AccountClient{config: akq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := akq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := akq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(accesskey.Table, accesskey.FieldID, selector),
			sqlgraph.To(entaccount.Table, entaccount.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, accesskey.AccountTable, accesskey.AccountColumn),
		)
		fromU = sqlgraph.SetNeighbors(akq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first AccessKey entity from the query.
// Returns a *NotFoundError when no AccessKey was found.
func (akq *AccessKeyQuery) First(ctx context.Context) (*AccessKey, error) {
	nodes, err := akq.Limit(1).All(setContextOp(ctx, akq.ctx, "First"))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{accesskey.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (akq *AccessKeyQuery) FirstX(ctx context.Context) *AccessKey {
	node, err := akq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first AccessKey ID from the query.
// Returns a *NotFoundError when no AccessKey ID was found.
func (akq *AccessKeyQuery) FirstID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = akq.Limit(1).IDs(setContextOp(ctx, akq.ctx, "FirstID")); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{accesskey.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (akq *AccessKeyQuery) FirstIDX(ctx context.Context) int {
	id, err := akq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single AccessKey entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one AccessKey entity is found.
// Returns a *NotFoundError when no AccessKey entities are found.
func (akq *AccessKeyQuery) Only(ctx context.Context) (*AccessKey, error) {
	nodes, err := akq.Limit(2).All(setContextOp(ctx, akq.ctx, "Only"))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{accesskey.Label}
	default:
		return nil, &NotSingularError{accesskey.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (akq *AccessKeyQuery) OnlyX(ctx context.Context) *AccessKey {
	node, err := akq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only AccessKey ID in the query.
// Returns a *NotSingularError when more than one AccessKey ID is found.
// Returns a *NotFoundError when no entities are found.
func (akq *AccessKeyQuery) OnlyID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = akq.Limit(2).IDs(setContextOp(ctx, akq.ctx, "OnlyID")); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{accesskey.Label}
	default:
		err = &NotSingularError{accesskey.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (akq *AccessKeyQuery) OnlyIDX(ctx context.Context) int {
	id, err := akq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of AccessKeys.
func (akq *AccessKeyQuery) All(ctx context.Context) ([]*AccessKey, error) {
	ctx = setContextOp(ctx, akq.ctx, "All")
	if err := akq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*AccessKey, *AccessKeyQuery]()
	return withInterceptors[[]*AccessKey](ctx, akq, qr, akq.inters)
}

// AllX is like All, but panics if an error occurs.
func (akq *AccessKeyQuery) AllX(ctx context.Context) []*AccessKey {
	nodes, err := akq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of AccessKey IDs.
func (akq *AccessKeyQuery) IDs(ctx context.Context) ([]int, error) {
	var ids []int
	ctx = setContextOp(ctx, akq.ctx, "IDs")
	if err := akq.Select(accesskey.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (akq *AccessKeyQuery) IDsX(ctx context.Context) []int {
	ids, err := akq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (akq *AccessKeyQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, akq.ctx, "Count")
	if err := akq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, akq, querierCount[*AccessKeyQuery](), akq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (akq *AccessKeyQuery) CountX(ctx context.Context) int {
	count, err := akq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (akq *AccessKeyQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, akq.ctx, "Exist")
	switch _, err := akq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (akq *AccessKeyQuery) ExistX(ctx context.Context) bool {
	exist, err := akq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the AccessKeyQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (akq *AccessKeyQuery) Clone() *AccessKeyQuery {
	if akq == nil {
		return nil
	}
	return &AccessKeyQuery{
		config:      akq.config,
		ctx:         akq.ctx.Clone(),
		order:       append([]OrderFunc{}, akq.order...),
		inters:      append([]Interceptor{}, akq.inters...),
		predicates:  append([]predicate.AccessKey{}, akq.predicates...),
		withAccount: akq.withAccount.Clone(),
		// clone intermediate query.
		sql:  akq.sql.Clone(),
		path: akq.path,
	}
}

// WithAccount tells the query-builder to eager-load the nodes that are connected to
// the "account" edge. The optional arguments are used to configure the query builder of the edge.
func (akq *AccessKeyQuery) WithAccount(opts ...func(*AccountQuery)) *AccessKeyQuery {
	query := (&AccountClient{config: akq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	akq.withAccount = query
	return akq
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
//	client.AccessKey.Query().
//		GroupBy(accesskey.FieldDid).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (akq *AccessKeyQuery) GroupBy(field string, fields ...string) *AccessKeyGroupBy {
	akq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &AccessKeyGroupBy{build: akq}
	grbuild.flds = &akq.ctx.Fields
	grbuild.label = accesskey.Label
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
//	client.AccessKey.Query().
//		Select(accesskey.FieldDid).
//		Scan(ctx, &v)
func (akq *AccessKeyQuery) Select(fields ...string) *AccessKeySelect {
	akq.ctx.Fields = append(akq.ctx.Fields, fields...)
	sbuild := &AccessKeySelect{AccessKeyQuery: akq}
	sbuild.label = accesskey.Label
	sbuild.flds, sbuild.scan = &akq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a AccessKeySelect configured with the given aggregations.
func (akq *AccessKeyQuery) Aggregate(fns ...AggregateFunc) *AccessKeySelect {
	return akq.Select().Aggregate(fns...)
}

func (akq *AccessKeyQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range akq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, akq); err != nil {
				return err
			}
		}
	}
	for _, f := range akq.ctx.Fields {
		if !accesskey.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if akq.path != nil {
		prev, err := akq.path(ctx)
		if err != nil {
			return err
		}
		akq.sql = prev
	}
	return nil
}

func (akq *AccessKeyQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*AccessKey, error) {
	var (
		nodes       = []*AccessKey{}
		withFKs     = akq.withFKs
		_spec       = akq.querySpec()
		loadedTypes = [1]bool{
			akq.withAccount != nil,
		}
	)
	if akq.withAccount != nil {
		withFKs = true
	}
	if withFKs {
		_spec.Node.Columns = append(_spec.Node.Columns, accesskey.ForeignKeys...)
	}
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*AccessKey).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &AccessKey{config: akq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, akq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := akq.withAccount; query != nil {
		if err := akq.loadAccount(ctx, query, nodes, nil,
			func(n *AccessKey, e *Account) { n.Edges.Account = e }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (akq *AccessKeyQuery) loadAccount(ctx context.Context, query *AccountQuery, nodes []*AccessKey, init func(*AccessKey), assign func(*AccessKey, *Account)) error {
	ids := make([]int, 0, len(nodes))
	nodeids := make(map[int][]*AccessKey)
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

func (akq *AccessKeyQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := akq.querySpec()
	_spec.Node.Columns = akq.ctx.Fields
	if len(akq.ctx.Fields) > 0 {
		_spec.Unique = akq.ctx.Unique != nil && *akq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, akq.driver, _spec)
}

func (akq *AccessKeyQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := &sqlgraph.QuerySpec{
		Node: &sqlgraph.NodeSpec{
			Table:   accesskey.Table,
			Columns: accesskey.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: accesskey.FieldID,
			},
		},
		From:   akq.sql,
		Unique: true,
	}
	if unique := akq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	}
	if fields := akq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, accesskey.FieldID)
		for i := range fields {
			if fields[i] != accesskey.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := akq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := akq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := akq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := akq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (akq *AccessKeyQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(akq.driver.Dialect())
	t1 := builder.Table(accesskey.Table)
	columns := akq.ctx.Fields
	if len(columns) == 0 {
		columns = accesskey.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if akq.sql != nil {
		selector = akq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if akq.ctx.Unique != nil && *akq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range akq.predicates {
		p(selector)
	}
	for _, p := range akq.order {
		p(selector)
	}
	if offset := akq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := akq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// AccessKeyGroupBy is the group-by builder for AccessKey entities.
type AccessKeyGroupBy struct {
	selector
	build *AccessKeyQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (akgb *AccessKeyGroupBy) Aggregate(fns ...AggregateFunc) *AccessKeyGroupBy {
	akgb.fns = append(akgb.fns, fns...)
	return akgb
}

// Scan applies the selector query and scans the result into the given value.
func (akgb *AccessKeyGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, akgb.build.ctx, "GroupBy")
	if err := akgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*AccessKeyQuery, *AccessKeyGroupBy](ctx, akgb.build, akgb, akgb.build.inters, v)
}

func (akgb *AccessKeyGroupBy) sqlScan(ctx context.Context, root *AccessKeyQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(akgb.fns))
	for _, fn := range akgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*akgb.flds)+len(akgb.fns))
		for _, f := range *akgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*akgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := akgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// AccessKeySelect is the builder for selecting fields of AccessKey entities.
type AccessKeySelect struct {
	*AccessKeyQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (aks *AccessKeySelect) Aggregate(fns ...AggregateFunc) *AccessKeySelect {
	aks.fns = append(aks.fns, fns...)
	return aks
}

// Scan applies the selector query and scans the result into the given value.
func (aks *AccessKeySelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, aks.ctx, "Select")
	if err := aks.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*AccessKeyQuery, *AccessKeySelect](ctx, aks.AccessKeyQuery, aks, aks.inters, v)
}

func (aks *AccessKeySelect) sqlScan(ctx context.Context, root *AccessKeyQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(aks.fns))
	for _, fn := range aks.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*aks.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := aks.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
