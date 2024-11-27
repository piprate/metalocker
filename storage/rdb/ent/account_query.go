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
	"database/sql/driver"
	"fmt"
	"math"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/piprate/metalocker/storage/rdb/ent/predicate"

	"github.com/piprate/metalocker/storage/rdb/ent/accesskey"
	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
	"github.com/piprate/metalocker/storage/rdb/ent/identity"
	"github.com/piprate/metalocker/storage/rdb/ent/locker"
	"github.com/piprate/metalocker/storage/rdb/ent/property"
	"github.com/piprate/metalocker/storage/rdb/ent/recoverycode"
)

// AccountQuery is the builder for querying Account entities.
type AccountQuery struct {
	config
	ctx               *QueryContext
	order             []OrderFunc
	inters            []Interceptor
	predicates        []predicate.Account
	withRecoveryCodes *RecoveryCodeQuery
	withAccessKeys    *AccessKeyQuery
	withIdentities    *IdentityQuery
	withLockers       *LockerQuery
	withProperties    *PropertyQuery
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the AccountQuery builder.
func (aq *AccountQuery) Where(ps ...predicate.Account) *AccountQuery {
	aq.predicates = append(aq.predicates, ps...)
	return aq
}

// Limit the number of records to be returned by this query.
func (aq *AccountQuery) Limit(limit int) *AccountQuery {
	aq.ctx.Limit = &limit
	return aq
}

// Offset to start from.
func (aq *AccountQuery) Offset(offset int) *AccountQuery {
	aq.ctx.Offset = &offset
	return aq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (aq *AccountQuery) Unique(unique bool) *AccountQuery {
	aq.ctx.Unique = &unique
	return aq
}

// Order specifies how the records should be ordered.
func (aq *AccountQuery) Order(o ...OrderFunc) *AccountQuery {
	aq.order = append(aq.order, o...)
	return aq
}

// QueryRecoveryCodes chains the current query on the "recovery_codes" edge.
func (aq *AccountQuery) QueryRecoveryCodes() *RecoveryCodeQuery {
	query := (&RecoveryCodeClient{config: aq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := aq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := aq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(entaccount.Table, entaccount.FieldID, selector),
			sqlgraph.To(recoverycode.Table, recoverycode.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, entaccount.RecoveryCodesTable, entaccount.RecoveryCodesColumn),
		)
		fromU = sqlgraph.SetNeighbors(aq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryAccessKeys chains the current query on the "access_keys" edge.
func (aq *AccountQuery) QueryAccessKeys() *AccessKeyQuery {
	query := (&AccessKeyClient{config: aq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := aq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := aq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(entaccount.Table, entaccount.FieldID, selector),
			sqlgraph.To(accesskey.Table, accesskey.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, entaccount.AccessKeysTable, entaccount.AccessKeysColumn),
		)
		fromU = sqlgraph.SetNeighbors(aq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryIdentities chains the current query on the "identities" edge.
func (aq *AccountQuery) QueryIdentities() *IdentityQuery {
	query := (&IdentityClient{config: aq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := aq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := aq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(entaccount.Table, entaccount.FieldID, selector),
			sqlgraph.To(identity.Table, identity.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, entaccount.IdentitiesTable, entaccount.IdentitiesColumn),
		)
		fromU = sqlgraph.SetNeighbors(aq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryLockers chains the current query on the "lockers" edge.
func (aq *AccountQuery) QueryLockers() *LockerQuery {
	query := (&LockerClient{config: aq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := aq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := aq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(entaccount.Table, entaccount.FieldID, selector),
			sqlgraph.To(locker.Table, locker.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, entaccount.LockersTable, entaccount.LockersColumn),
		)
		fromU = sqlgraph.SetNeighbors(aq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryProperties chains the current query on the "properties" edge.
func (aq *AccountQuery) QueryProperties() *PropertyQuery {
	query := (&PropertyClient{config: aq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := aq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := aq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(entaccount.Table, entaccount.FieldID, selector),
			sqlgraph.To(property.Table, property.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, entaccount.PropertiesTable, entaccount.PropertiesColumn),
		)
		fromU = sqlgraph.SetNeighbors(aq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first Account entity from the query.
// Returns a *NotFoundError when no Account was found.
func (aq *AccountQuery) First(ctx context.Context) (*Account, error) {
	nodes, err := aq.Limit(1).All(setContextOp(ctx, aq.ctx, "First"))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{entaccount.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (aq *AccountQuery) FirstX(ctx context.Context) *Account {
	node, err := aq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first Account ID from the query.
// Returns a *NotFoundError when no Account ID was found.
func (aq *AccountQuery) FirstID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = aq.Limit(1).IDs(setContextOp(ctx, aq.ctx, "FirstID")); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{entaccount.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (aq *AccountQuery) FirstIDX(ctx context.Context) int {
	id, err := aq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single Account entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one Account entity is found.
// Returns a *NotFoundError when no Account entities are found.
func (aq *AccountQuery) Only(ctx context.Context) (*Account, error) {
	nodes, err := aq.Limit(2).All(setContextOp(ctx, aq.ctx, "Only"))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{entaccount.Label}
	default:
		return nil, &NotSingularError{entaccount.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (aq *AccountQuery) OnlyX(ctx context.Context) *Account {
	node, err := aq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only Account ID in the query.
// Returns a *NotSingularError when more than one Account ID is found.
// Returns a *NotFoundError when no entities are found.
func (aq *AccountQuery) OnlyID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = aq.Limit(2).IDs(setContextOp(ctx, aq.ctx, "OnlyID")); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{entaccount.Label}
	default:
		err = &NotSingularError{entaccount.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (aq *AccountQuery) OnlyIDX(ctx context.Context) int {
	id, err := aq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of Accounts.
func (aq *AccountQuery) All(ctx context.Context) ([]*Account, error) {
	ctx = setContextOp(ctx, aq.ctx, "All")
	if err := aq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*Account, *AccountQuery]()
	return withInterceptors[[]*Account](ctx, aq, qr, aq.inters)
}

// AllX is like All, but panics if an error occurs.
func (aq *AccountQuery) AllX(ctx context.Context) []*Account {
	nodes, err := aq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of Account IDs.
func (aq *AccountQuery) IDs(ctx context.Context) ([]int, error) {
	var ids []int
	ctx = setContextOp(ctx, aq.ctx, "IDs")
	if err := aq.Select(entaccount.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (aq *AccountQuery) IDsX(ctx context.Context) []int {
	ids, err := aq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (aq *AccountQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, aq.ctx, "Count")
	if err := aq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, aq, querierCount[*AccountQuery](), aq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (aq *AccountQuery) CountX(ctx context.Context) int {
	count, err := aq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (aq *AccountQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, aq.ctx, "Exist")
	switch _, err := aq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (aq *AccountQuery) ExistX(ctx context.Context) bool {
	exist, err := aq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the AccountQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (aq *AccountQuery) Clone() *AccountQuery {
	if aq == nil {
		return nil
	}
	return &AccountQuery{
		config:            aq.config,
		ctx:               aq.ctx.Clone(),
		order:             append([]OrderFunc{}, aq.order...),
		inters:            append([]Interceptor{}, aq.inters...),
		predicates:        append([]predicate.Account{}, aq.predicates...),
		withRecoveryCodes: aq.withRecoveryCodes.Clone(),
		withAccessKeys:    aq.withAccessKeys.Clone(),
		withIdentities:    aq.withIdentities.Clone(),
		withLockers:       aq.withLockers.Clone(),
		withProperties:    aq.withProperties.Clone(),
		// clone intermediate query.
		sql:  aq.sql.Clone(),
		path: aq.path,
	}
}

// WithRecoveryCodes tells the query-builder to eager-load the nodes that are connected to
// the "recovery_codes" edge. The optional arguments are used to configure the query builder of the edge.
func (aq *AccountQuery) WithRecoveryCodes(opts ...func(*RecoveryCodeQuery)) *AccountQuery {
	query := (&RecoveryCodeClient{config: aq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	aq.withRecoveryCodes = query
	return aq
}

// WithAccessKeys tells the query-builder to eager-load the nodes that are connected to
// the "access_keys" edge. The optional arguments are used to configure the query builder of the edge.
func (aq *AccountQuery) WithAccessKeys(opts ...func(*AccessKeyQuery)) *AccountQuery {
	query := (&AccessKeyClient{config: aq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	aq.withAccessKeys = query
	return aq
}

// WithIdentities tells the query-builder to eager-load the nodes that are connected to
// the "identities" edge. The optional arguments are used to configure the query builder of the edge.
func (aq *AccountQuery) WithIdentities(opts ...func(*IdentityQuery)) *AccountQuery {
	query := (&IdentityClient{config: aq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	aq.withIdentities = query
	return aq
}

// WithLockers tells the query-builder to eager-load the nodes that are connected to
// the "lockers" edge. The optional arguments are used to configure the query builder of the edge.
func (aq *AccountQuery) WithLockers(opts ...func(*LockerQuery)) *AccountQuery {
	query := (&LockerClient{config: aq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	aq.withLockers = query
	return aq
}

// WithProperties tells the query-builder to eager-load the nodes that are connected to
// the "properties" edge. The optional arguments are used to configure the query builder of the edge.
func (aq *AccountQuery) WithProperties(opts ...func(*PropertyQuery)) *AccountQuery {
	query := (&PropertyClient{config: aq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	aq.withProperties = query
	return aq
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
//	client.Account.Query().
//		GroupBy(entaccount.FieldDid).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (aq *AccountQuery) GroupBy(field string, fields ...string) *AccountGroupBy {
	aq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &AccountGroupBy{build: aq}
	grbuild.flds = &aq.ctx.Fields
	grbuild.label = entaccount.Label
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
//	client.Account.Query().
//		Select(entaccount.FieldDid).
//		Scan(ctx, &v)
func (aq *AccountQuery) Select(fields ...string) *AccountSelect {
	aq.ctx.Fields = append(aq.ctx.Fields, fields...)
	sbuild := &AccountSelect{AccountQuery: aq}
	sbuild.label = entaccount.Label
	sbuild.flds, sbuild.scan = &aq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a AccountSelect configured with the given aggregations.
func (aq *AccountQuery) Aggregate(fns ...AggregateFunc) *AccountSelect {
	return aq.Select().Aggregate(fns...)
}

func (aq *AccountQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range aq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, aq); err != nil {
				return err
			}
		}
	}
	for _, f := range aq.ctx.Fields {
		if !entaccount.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if aq.path != nil {
		prev, err := aq.path(ctx)
		if err != nil {
			return err
		}
		aq.sql = prev
	}
	return nil
}

func (aq *AccountQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*Account, error) {
	var (
		nodes       = []*Account{}
		_spec       = aq.querySpec()
		loadedTypes = [5]bool{
			aq.withRecoveryCodes != nil,
			aq.withAccessKeys != nil,
			aq.withIdentities != nil,
			aq.withLockers != nil,
			aq.withProperties != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*Account).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &Account{config: aq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, aq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := aq.withRecoveryCodes; query != nil {
		if err := aq.loadRecoveryCodes(ctx, query, nodes,
			func(n *Account) { n.Edges.RecoveryCodes = []*RecoveryCode{} },
			func(n *Account, e *RecoveryCode) { n.Edges.RecoveryCodes = append(n.Edges.RecoveryCodes, e) }); err != nil {
			return nil, err
		}
	}
	if query := aq.withAccessKeys; query != nil {
		if err := aq.loadAccessKeys(ctx, query, nodes,
			func(n *Account) { n.Edges.AccessKeys = []*AccessKey{} },
			func(n *Account, e *AccessKey) { n.Edges.AccessKeys = append(n.Edges.AccessKeys, e) }); err != nil {
			return nil, err
		}
	}
	if query := aq.withIdentities; query != nil {
		if err := aq.loadIdentities(ctx, query, nodes,
			func(n *Account) { n.Edges.Identities = []*Identity{} },
			func(n *Account, e *Identity) { n.Edges.Identities = append(n.Edges.Identities, e) }); err != nil {
			return nil, err
		}
	}
	if query := aq.withLockers; query != nil {
		if err := aq.loadLockers(ctx, query, nodes,
			func(n *Account) { n.Edges.Lockers = []*Locker{} },
			func(n *Account, e *Locker) { n.Edges.Lockers = append(n.Edges.Lockers, e) }); err != nil {
			return nil, err
		}
	}
	if query := aq.withProperties; query != nil {
		if err := aq.loadProperties(ctx, query, nodes,
			func(n *Account) { n.Edges.Properties = []*Property{} },
			func(n *Account, e *Property) { n.Edges.Properties = append(n.Edges.Properties, e) }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (aq *AccountQuery) loadRecoveryCodes(ctx context.Context, query *RecoveryCodeQuery, nodes []*Account, init func(*Account), assign func(*Account, *RecoveryCode)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[int]*Account)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	query.withFKs = true
	query.Where(predicate.RecoveryCode(func(s *sql.Selector) {
		s.Where(sql.InValues(entaccount.RecoveryCodesColumn, fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.account
		if fk == nil {
			return fmt.Errorf(`foreign-key "account" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "account" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}
func (aq *AccountQuery) loadAccessKeys(ctx context.Context, query *AccessKeyQuery, nodes []*Account, init func(*Account), assign func(*Account, *AccessKey)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[int]*Account)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	query.withFKs = true
	query.Where(predicate.AccessKey(func(s *sql.Selector) {
		s.Where(sql.InValues(entaccount.AccessKeysColumn, fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.account
		if fk == nil {
			return fmt.Errorf(`foreign-key "account" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "account" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}
func (aq *AccountQuery) loadIdentities(ctx context.Context, query *IdentityQuery, nodes []*Account, init func(*Account), assign func(*Account, *Identity)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[int]*Account)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	query.withFKs = true
	query.Where(predicate.Identity(func(s *sql.Selector) {
		s.Where(sql.InValues(entaccount.IdentitiesColumn, fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.account
		if fk == nil {
			return fmt.Errorf(`foreign-key "account" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "account" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}
func (aq *AccountQuery) loadLockers(ctx context.Context, query *LockerQuery, nodes []*Account, init func(*Account), assign func(*Account, *Locker)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[int]*Account)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	query.withFKs = true
	query.Where(predicate.Locker(func(s *sql.Selector) {
		s.Where(sql.InValues(entaccount.LockersColumn, fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.account
		if fk == nil {
			return fmt.Errorf(`foreign-key "account" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "account" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}
func (aq *AccountQuery) loadProperties(ctx context.Context, query *PropertyQuery, nodes []*Account, init func(*Account), assign func(*Account, *Property)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[int]*Account)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	query.withFKs = true
	query.Where(predicate.Property(func(s *sql.Selector) {
		s.Where(sql.InValues(entaccount.PropertiesColumn, fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.account
		if fk == nil {
			return fmt.Errorf(`foreign-key "account" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "account" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}

func (aq *AccountQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := aq.querySpec()
	_spec.Node.Columns = aq.ctx.Fields
	if len(aq.ctx.Fields) > 0 {
		_spec.Unique = aq.ctx.Unique != nil && *aq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, aq.driver, _spec)
}

func (aq *AccountQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := &sqlgraph.QuerySpec{
		Node: &sqlgraph.NodeSpec{
			Table:   entaccount.Table,
			Columns: entaccount.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: entaccount.FieldID,
			},
		},
		From:   aq.sql,
		Unique: true,
	}
	if unique := aq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	}
	if fields := aq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, entaccount.FieldID)
		for i := range fields {
			if fields[i] != entaccount.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := aq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := aq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := aq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := aq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (aq *AccountQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(aq.driver.Dialect())
	t1 := builder.Table(entaccount.Table)
	columns := aq.ctx.Fields
	if len(columns) == 0 {
		columns = entaccount.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if aq.sql != nil {
		selector = aq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if aq.ctx.Unique != nil && *aq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range aq.predicates {
		p(selector)
	}
	for _, p := range aq.order {
		p(selector)
	}
	if offset := aq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := aq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// AccountGroupBy is the group-by builder for Account entities.
type AccountGroupBy struct {
	selector
	build *AccountQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (agb *AccountGroupBy) Aggregate(fns ...AggregateFunc) *AccountGroupBy {
	agb.fns = append(agb.fns, fns...)
	return agb
}

// Scan applies the selector query and scans the result into the given value.
func (agb *AccountGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, agb.build.ctx, "GroupBy")
	if err := agb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*AccountQuery, *AccountGroupBy](ctx, agb.build, agb, agb.build.inters, v)
}

func (agb *AccountGroupBy) sqlScan(ctx context.Context, root *AccountQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(agb.fns))
	for _, fn := range agb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*agb.flds)+len(agb.fns))
		for _, f := range *agb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*agb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := agb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// AccountSelect is the builder for selecting fields of Account entities.
type AccountSelect struct {
	*AccountQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (as *AccountSelect) Aggregate(fns ...AggregateFunc) *AccountSelect {
	as.fns = append(as.fns, fns...)
	return as
}

// Scan applies the selector query and scans the result into the given value.
func (as *AccountSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, as.ctx, "Select")
	if err := as.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*AccountQuery, *AccountSelect](ctx, as.AccountQuery, as, as.inters, v)
}

func (as *AccountSelect) sqlScan(ctx context.Context, root *AccountQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(as.fns))
	for _, fn := range as.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*as.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := as.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
