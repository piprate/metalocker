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
	"log"

	"github.com/piprate/metalocker/storage/rdb/ent/migrate"

	"github.com/piprate/metalocker/storage/rdb/ent/accesskey"
	entaccount "github.com/piprate/metalocker/storage/rdb/ent/account"
	"github.com/piprate/metalocker/storage/rdb/ent/did"
	"github.com/piprate/metalocker/storage/rdb/ent/identity"
	"github.com/piprate/metalocker/storage/rdb/ent/locker"
	"github.com/piprate/metalocker/storage/rdb/ent/property"
	"github.com/piprate/metalocker/storage/rdb/ent/recoverycode"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

// Client is the client that holds all ent builders.
type Client struct {
	config
	// Schema is the client for creating, migrating and dropping schema.
	Schema *migrate.Schema
	// AccessKey is the client for interacting with the AccessKey builders.
	AccessKey *AccessKeyClient
	// Account is the client for interacting with the Account builders.
	Account *AccountClient
	// DID is the client for interacting with the DID builders.
	DID *DIDClient
	// Identity is the client for interacting with the Identity builders.
	Identity *IdentityClient
	// Locker is the client for interacting with the Locker builders.
	Locker *LockerClient
	// Property is the client for interacting with the Property builders.
	Property *PropertyClient
	// RecoveryCode is the client for interacting with the RecoveryCode builders.
	RecoveryCode *RecoveryCodeClient
}

// NewClient creates a new client configured with the given options.
func NewClient(opts ...Option) *Client {
	cfg := config{log: log.Println, hooks: &hooks{}, inters: &inters{}}
	cfg.options(opts...)
	client := &Client{config: cfg}
	client.init()
	return client
}

func (c *Client) init() {
	c.Schema = migrate.NewSchema(c.driver)
	c.AccessKey = NewAccessKeyClient(c.config)
	c.Account = NewAccountClient(c.config)
	c.DID = NewDIDClient(c.config)
	c.Identity = NewIdentityClient(c.config)
	c.Locker = NewLockerClient(c.config)
	c.Property = NewPropertyClient(c.config)
	c.RecoveryCode = NewRecoveryCodeClient(c.config)
}

// Open opens a database/sql.DB specified by the driver name and
// the data source name, and returns a new client attached to it.
// Optional parameters can be added for configuring the client.
func Open(driverName, dataSourceName string, options ...Option) (*Client, error) {
	switch driverName {
	case dialect.MySQL, dialect.Postgres, dialect.SQLite:
		drv, err := sql.Open(driverName, dataSourceName)
		if err != nil {
			return nil, err
		}
		return NewClient(append(options, Driver(drv))...), nil
	default:
		return nil, fmt.Errorf("unsupported driver: %q", driverName)
	}
}

// Tx returns a new transactional client. The provided context
// is used until the transaction is committed or rolled back.
func (c *Client) Tx(ctx context.Context) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, errors.New("ent: cannot start a transaction within a transaction")
	}
	tx, err := newTx(ctx, c.driver)
	if err != nil {
		return nil, fmt.Errorf("ent: starting a transaction: %w", err)
	}
	cfg := c.config
	cfg.driver = tx
	return &Tx{
		ctx:          ctx,
		config:       cfg,
		AccessKey:    NewAccessKeyClient(cfg),
		Account:      NewAccountClient(cfg),
		DID:          NewDIDClient(cfg),
		Identity:     NewIdentityClient(cfg),
		Locker:       NewLockerClient(cfg),
		Property:     NewPropertyClient(cfg),
		RecoveryCode: NewRecoveryCodeClient(cfg),
	}, nil
}

// BeginTx returns a transactional client with specified options.
func (c *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, errors.New("ent: cannot start a transaction within a transaction")
	}
	tx, err := c.driver.(interface {
		BeginTx(context.Context, *sql.TxOptions) (dialect.Tx, error)
	}).BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("ent: starting a transaction: %w", err)
	}
	cfg := c.config
	cfg.driver = &txDriver{tx: tx, drv: c.driver}
	return &Tx{
		ctx:          ctx,
		config:       cfg,
		AccessKey:    NewAccessKeyClient(cfg),
		Account:      NewAccountClient(cfg),
		DID:          NewDIDClient(cfg),
		Identity:     NewIdentityClient(cfg),
		Locker:       NewLockerClient(cfg),
		Property:     NewPropertyClient(cfg),
		RecoveryCode: NewRecoveryCodeClient(cfg),
	}, nil
}

// Debug returns a new debug-client. It's used to get verbose logging on specific operations.
//
//	client.Debug().
//		AccessKey.
//		Query().
//		Count(ctx)
func (c *Client) Debug() *Client {
	if c.debug {
		return c
	}
	cfg := c.config
	cfg.driver = dialect.Debug(c.driver, c.log)
	client := &Client{config: cfg}
	client.init()
	return client
}

// Close closes the database connection and prevents new queries from starting.
func (c *Client) Close() error {
	return c.driver.Close()
}

// Use adds the mutation hooks to all the entity clients.
// In order to add hooks to a specific client, call: `client.Node.Use(...)`.
func (c *Client) Use(hooks ...Hook) {
	c.AccessKey.Use(hooks...)
	c.Account.Use(hooks...)
	c.DID.Use(hooks...)
	c.Identity.Use(hooks...)
	c.Locker.Use(hooks...)
	c.Property.Use(hooks...)
	c.RecoveryCode.Use(hooks...)
}

// Intercept adds the query interceptors to all the entity clients.
// In order to add interceptors to a specific client, call: `client.Node.Intercept(...)`.
func (c *Client) Intercept(interceptors ...Interceptor) {
	c.AccessKey.Intercept(interceptors...)
	c.Account.Intercept(interceptors...)
	c.DID.Intercept(interceptors...)
	c.Identity.Intercept(interceptors...)
	c.Locker.Intercept(interceptors...)
	c.Property.Intercept(interceptors...)
	c.RecoveryCode.Intercept(interceptors...)
}

// Mutate implements the ent.Mutator interface.
func (c *Client) Mutate(ctx context.Context, m Mutation) (Value, error) {
	switch m := m.(type) {
	case *AccessKeyMutation:
		return c.AccessKey.mutate(ctx, m)
	case *AccountMutation:
		return c.Account.mutate(ctx, m)
	case *DIDMutation:
		return c.DID.mutate(ctx, m)
	case *IdentityMutation:
		return c.Identity.mutate(ctx, m)
	case *LockerMutation:
		return c.Locker.mutate(ctx, m)
	case *PropertyMutation:
		return c.Property.mutate(ctx, m)
	case *RecoveryCodeMutation:
		return c.RecoveryCode.mutate(ctx, m)
	default:
		return nil, fmt.Errorf("ent: unknown mutation type %T", m)
	}
}

// AccessKeyClient is a client for the AccessKey schema.
type AccessKeyClient struct {
	config
}

// NewAccessKeyClient returns a client for the AccessKey from the given config.
func NewAccessKeyClient(c config) *AccessKeyClient {
	return &AccessKeyClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `accesskey.Hooks(f(g(h())))`.
func (c *AccessKeyClient) Use(hooks ...Hook) {
	c.hooks.AccessKey = append(c.hooks.AccessKey, hooks...)
}

// Use adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `accesskey.Intercept(f(g(h())))`.
func (c *AccessKeyClient) Intercept(interceptors ...Interceptor) {
	c.inters.AccessKey = append(c.inters.AccessKey, interceptors...)
}

// Create returns a builder for creating a AccessKey entity.
func (c *AccessKeyClient) Create() *AccessKeyCreate {
	mutation := newAccessKeyMutation(c.config, OpCreate)
	return &AccessKeyCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of AccessKey entities.
func (c *AccessKeyClient) CreateBulk(builders ...*AccessKeyCreate) *AccessKeyCreateBulk {
	return &AccessKeyCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for AccessKey.
func (c *AccessKeyClient) Update() *AccessKeyUpdate {
	mutation := newAccessKeyMutation(c.config, OpUpdate)
	return &AccessKeyUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *AccessKeyClient) UpdateOne(ak *AccessKey) *AccessKeyUpdateOne {
	mutation := newAccessKeyMutation(c.config, OpUpdateOne, withAccessKey(ak))
	return &AccessKeyUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *AccessKeyClient) UpdateOneID(id int) *AccessKeyUpdateOne {
	mutation := newAccessKeyMutation(c.config, OpUpdateOne, withAccessKeyID(id))
	return &AccessKeyUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for AccessKey.
func (c *AccessKeyClient) Delete() *AccessKeyDelete {
	mutation := newAccessKeyMutation(c.config, OpDelete)
	return &AccessKeyDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *AccessKeyClient) DeleteOne(ak *AccessKey) *AccessKeyDeleteOne {
	return c.DeleteOneID(ak.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *AccessKeyClient) DeleteOneID(id int) *AccessKeyDeleteOne {
	builder := c.Delete().Where(accesskey.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &AccessKeyDeleteOne{builder}
}

// Query returns a query builder for AccessKey.
func (c *AccessKeyClient) Query() *AccessKeyQuery {
	return &AccessKeyQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeAccessKey},
		inters: c.Interceptors(),
	}
}

// Get returns a AccessKey entity by its id.
func (c *AccessKeyClient) Get(ctx context.Context, id int) (*AccessKey, error) {
	return c.Query().Where(accesskey.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *AccessKeyClient) GetX(ctx context.Context, id int) *AccessKey {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryAccount queries the account edge of a AccessKey.
func (c *AccessKeyClient) QueryAccount(ak *AccessKey) *AccountQuery {
	query := (&AccountClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := ak.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(accesskey.Table, accesskey.FieldID, id),
			sqlgraph.To(entaccount.Table, entaccount.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, accesskey.AccountTable, accesskey.AccountColumn),
		)
		fromV = sqlgraph.Neighbors(ak.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *AccessKeyClient) Hooks() []Hook {
	return c.hooks.AccessKey
}

// Interceptors returns the client interceptors.
func (c *AccessKeyClient) Interceptors() []Interceptor {
	return c.inters.AccessKey
}

func (c *AccessKeyClient) mutate(ctx context.Context, m *AccessKeyMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&AccessKeyCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&AccessKeyUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&AccessKeyUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&AccessKeyDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown AccessKey mutation op: %q", m.Op())
	}
}

// AccountClient is a client for the Account schema.
type AccountClient struct {
	config
}

// NewAccountClient returns a client for the Account from the given config.
func NewAccountClient(c config) *AccountClient {
	return &AccountClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `entaccount.Hooks(f(g(h())))`.
func (c *AccountClient) Use(hooks ...Hook) {
	c.hooks.Account = append(c.hooks.Account, hooks...)
}

// Use adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `entaccount.Intercept(f(g(h())))`.
func (c *AccountClient) Intercept(interceptors ...Interceptor) {
	c.inters.Account = append(c.inters.Account, interceptors...)
}

// Create returns a builder for creating a Account entity.
func (c *AccountClient) Create() *AccountCreate {
	mutation := newAccountMutation(c.config, OpCreate)
	return &AccountCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Account entities.
func (c *AccountClient) CreateBulk(builders ...*AccountCreate) *AccountCreateBulk {
	return &AccountCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Account.
func (c *AccountClient) Update() *AccountUpdate {
	mutation := newAccountMutation(c.config, OpUpdate)
	return &AccountUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *AccountClient) UpdateOne(a *Account) *AccountUpdateOne {
	mutation := newAccountMutation(c.config, OpUpdateOne, withAccount(a))
	return &AccountUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *AccountClient) UpdateOneID(id int) *AccountUpdateOne {
	mutation := newAccountMutation(c.config, OpUpdateOne, withAccountID(id))
	return &AccountUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Account.
func (c *AccountClient) Delete() *AccountDelete {
	mutation := newAccountMutation(c.config, OpDelete)
	return &AccountDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *AccountClient) DeleteOne(a *Account) *AccountDeleteOne {
	return c.DeleteOneID(a.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *AccountClient) DeleteOneID(id int) *AccountDeleteOne {
	builder := c.Delete().Where(entaccount.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &AccountDeleteOne{builder}
}

// Query returns a query builder for Account.
func (c *AccountClient) Query() *AccountQuery {
	return &AccountQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeAccount},
		inters: c.Interceptors(),
	}
}

// Get returns a Account entity by its id.
func (c *AccountClient) Get(ctx context.Context, id int) (*Account, error) {
	return c.Query().Where(entaccount.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *AccountClient) GetX(ctx context.Context, id int) *Account {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryRecoveryCodes queries the recovery_codes edge of a Account.
func (c *AccountClient) QueryRecoveryCodes(a *Account) *RecoveryCodeQuery {
	query := (&RecoveryCodeClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := a.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(entaccount.Table, entaccount.FieldID, id),
			sqlgraph.To(recoverycode.Table, recoverycode.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, entaccount.RecoveryCodesTable, entaccount.RecoveryCodesColumn),
		)
		fromV = sqlgraph.Neighbors(a.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// QueryAccessKeys queries the access_keys edge of a Account.
func (c *AccountClient) QueryAccessKeys(a *Account) *AccessKeyQuery {
	query := (&AccessKeyClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := a.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(entaccount.Table, entaccount.FieldID, id),
			sqlgraph.To(accesskey.Table, accesskey.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, entaccount.AccessKeysTable, entaccount.AccessKeysColumn),
		)
		fromV = sqlgraph.Neighbors(a.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// QueryIdentities queries the identities edge of a Account.
func (c *AccountClient) QueryIdentities(a *Account) *IdentityQuery {
	query := (&IdentityClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := a.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(entaccount.Table, entaccount.FieldID, id),
			sqlgraph.To(identity.Table, identity.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, entaccount.IdentitiesTable, entaccount.IdentitiesColumn),
		)
		fromV = sqlgraph.Neighbors(a.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// QueryLockers queries the lockers edge of a Account.
func (c *AccountClient) QueryLockers(a *Account) *LockerQuery {
	query := (&LockerClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := a.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(entaccount.Table, entaccount.FieldID, id),
			sqlgraph.To(locker.Table, locker.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, entaccount.LockersTable, entaccount.LockersColumn),
		)
		fromV = sqlgraph.Neighbors(a.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// QueryProperties queries the properties edge of a Account.
func (c *AccountClient) QueryProperties(a *Account) *PropertyQuery {
	query := (&PropertyClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := a.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(entaccount.Table, entaccount.FieldID, id),
			sqlgraph.To(property.Table, property.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, entaccount.PropertiesTable, entaccount.PropertiesColumn),
		)
		fromV = sqlgraph.Neighbors(a.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *AccountClient) Hooks() []Hook {
	return c.hooks.Account
}

// Interceptors returns the client interceptors.
func (c *AccountClient) Interceptors() []Interceptor {
	return c.inters.Account
}

func (c *AccountClient) mutate(ctx context.Context, m *AccountMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&AccountCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&AccountUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&AccountUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&AccountDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown Account mutation op: %q", m.Op())
	}
}

// DIDClient is a client for the DID schema.
type DIDClient struct {
	config
}

// NewDIDClient returns a client for the DID from the given config.
func NewDIDClient(c config) *DIDClient {
	return &DIDClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `did.Hooks(f(g(h())))`.
func (c *DIDClient) Use(hooks ...Hook) {
	c.hooks.DID = append(c.hooks.DID, hooks...)
}

// Use adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `did.Intercept(f(g(h())))`.
func (c *DIDClient) Intercept(interceptors ...Interceptor) {
	c.inters.DID = append(c.inters.DID, interceptors...)
}

// Create returns a builder for creating a DID entity.
func (c *DIDClient) Create() *DIDCreate {
	mutation := newDIDMutation(c.config, OpCreate)
	return &DIDCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of DID entities.
func (c *DIDClient) CreateBulk(builders ...*DIDCreate) *DIDCreateBulk {
	return &DIDCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for DID.
func (c *DIDClient) Update() *DIDUpdate {
	mutation := newDIDMutation(c.config, OpUpdate)
	return &DIDUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *DIDClient) UpdateOne(d *DID) *DIDUpdateOne {
	mutation := newDIDMutation(c.config, OpUpdateOne, withDID(d))
	return &DIDUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *DIDClient) UpdateOneID(id int) *DIDUpdateOne {
	mutation := newDIDMutation(c.config, OpUpdateOne, withDIDID(id))
	return &DIDUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for DID.
func (c *DIDClient) Delete() *DIDDelete {
	mutation := newDIDMutation(c.config, OpDelete)
	return &DIDDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *DIDClient) DeleteOne(d *DID) *DIDDeleteOne {
	return c.DeleteOneID(d.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *DIDClient) DeleteOneID(id int) *DIDDeleteOne {
	builder := c.Delete().Where(did.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &DIDDeleteOne{builder}
}

// Query returns a query builder for DID.
func (c *DIDClient) Query() *DIDQuery {
	return &DIDQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeDID},
		inters: c.Interceptors(),
	}
}

// Get returns a DID entity by its id.
func (c *DIDClient) Get(ctx context.Context, id int) (*DID, error) {
	return c.Query().Where(did.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *DIDClient) GetX(ctx context.Context, id int) *DID {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// Hooks returns the client hooks.
func (c *DIDClient) Hooks() []Hook {
	return c.hooks.DID
}

// Interceptors returns the client interceptors.
func (c *DIDClient) Interceptors() []Interceptor {
	return c.inters.DID
}

func (c *DIDClient) mutate(ctx context.Context, m *DIDMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&DIDCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&DIDUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&DIDUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&DIDDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown DID mutation op: %q", m.Op())
	}
}

// IdentityClient is a client for the Identity schema.
type IdentityClient struct {
	config
}

// NewIdentityClient returns a client for the Identity from the given config.
func NewIdentityClient(c config) *IdentityClient {
	return &IdentityClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `identity.Hooks(f(g(h())))`.
func (c *IdentityClient) Use(hooks ...Hook) {
	c.hooks.Identity = append(c.hooks.Identity, hooks...)
}

// Use adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `identity.Intercept(f(g(h())))`.
func (c *IdentityClient) Intercept(interceptors ...Interceptor) {
	c.inters.Identity = append(c.inters.Identity, interceptors...)
}

// Create returns a builder for creating a Identity entity.
func (c *IdentityClient) Create() *IdentityCreate {
	mutation := newIdentityMutation(c.config, OpCreate)
	return &IdentityCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Identity entities.
func (c *IdentityClient) CreateBulk(builders ...*IdentityCreate) *IdentityCreateBulk {
	return &IdentityCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Identity.
func (c *IdentityClient) Update() *IdentityUpdate {
	mutation := newIdentityMutation(c.config, OpUpdate)
	return &IdentityUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *IdentityClient) UpdateOne(i *Identity) *IdentityUpdateOne {
	mutation := newIdentityMutation(c.config, OpUpdateOne, withIdentity(i))
	return &IdentityUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *IdentityClient) UpdateOneID(id int) *IdentityUpdateOne {
	mutation := newIdentityMutation(c.config, OpUpdateOne, withIdentityID(id))
	return &IdentityUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Identity.
func (c *IdentityClient) Delete() *IdentityDelete {
	mutation := newIdentityMutation(c.config, OpDelete)
	return &IdentityDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *IdentityClient) DeleteOne(i *Identity) *IdentityDeleteOne {
	return c.DeleteOneID(i.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *IdentityClient) DeleteOneID(id int) *IdentityDeleteOne {
	builder := c.Delete().Where(identity.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &IdentityDeleteOne{builder}
}

// Query returns a query builder for Identity.
func (c *IdentityClient) Query() *IdentityQuery {
	return &IdentityQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeIdentity},
		inters: c.Interceptors(),
	}
}

// Get returns a Identity entity by its id.
func (c *IdentityClient) Get(ctx context.Context, id int) (*Identity, error) {
	return c.Query().Where(identity.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *IdentityClient) GetX(ctx context.Context, id int) *Identity {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryAccount queries the account edge of a Identity.
func (c *IdentityClient) QueryAccount(i *Identity) *AccountQuery {
	query := (&AccountClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := i.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(identity.Table, identity.FieldID, id),
			sqlgraph.To(entaccount.Table, entaccount.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, identity.AccountTable, identity.AccountColumn),
		)
		fromV = sqlgraph.Neighbors(i.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *IdentityClient) Hooks() []Hook {
	return c.hooks.Identity
}

// Interceptors returns the client interceptors.
func (c *IdentityClient) Interceptors() []Interceptor {
	return c.inters.Identity
}

func (c *IdentityClient) mutate(ctx context.Context, m *IdentityMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&IdentityCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&IdentityUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&IdentityUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&IdentityDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown Identity mutation op: %q", m.Op())
	}
}

// LockerClient is a client for the Locker schema.
type LockerClient struct {
	config
}

// NewLockerClient returns a client for the Locker from the given config.
func NewLockerClient(c config) *LockerClient {
	return &LockerClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `locker.Hooks(f(g(h())))`.
func (c *LockerClient) Use(hooks ...Hook) {
	c.hooks.Locker = append(c.hooks.Locker, hooks...)
}

// Use adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `locker.Intercept(f(g(h())))`.
func (c *LockerClient) Intercept(interceptors ...Interceptor) {
	c.inters.Locker = append(c.inters.Locker, interceptors...)
}

// Create returns a builder for creating a Locker entity.
func (c *LockerClient) Create() *LockerCreate {
	mutation := newLockerMutation(c.config, OpCreate)
	return &LockerCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Locker entities.
func (c *LockerClient) CreateBulk(builders ...*LockerCreate) *LockerCreateBulk {
	return &LockerCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Locker.
func (c *LockerClient) Update() *LockerUpdate {
	mutation := newLockerMutation(c.config, OpUpdate)
	return &LockerUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *LockerClient) UpdateOne(l *Locker) *LockerUpdateOne {
	mutation := newLockerMutation(c.config, OpUpdateOne, withLocker(l))
	return &LockerUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *LockerClient) UpdateOneID(id int) *LockerUpdateOne {
	mutation := newLockerMutation(c.config, OpUpdateOne, withLockerID(id))
	return &LockerUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Locker.
func (c *LockerClient) Delete() *LockerDelete {
	mutation := newLockerMutation(c.config, OpDelete)
	return &LockerDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *LockerClient) DeleteOne(l *Locker) *LockerDeleteOne {
	return c.DeleteOneID(l.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *LockerClient) DeleteOneID(id int) *LockerDeleteOne {
	builder := c.Delete().Where(locker.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &LockerDeleteOne{builder}
}

// Query returns a query builder for Locker.
func (c *LockerClient) Query() *LockerQuery {
	return &LockerQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeLocker},
		inters: c.Interceptors(),
	}
}

// Get returns a Locker entity by its id.
func (c *LockerClient) Get(ctx context.Context, id int) (*Locker, error) {
	return c.Query().Where(locker.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *LockerClient) GetX(ctx context.Context, id int) *Locker {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryAccount queries the account edge of a Locker.
func (c *LockerClient) QueryAccount(l *Locker) *AccountQuery {
	query := (&AccountClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := l.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(locker.Table, locker.FieldID, id),
			sqlgraph.To(entaccount.Table, entaccount.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, locker.AccountTable, locker.AccountColumn),
		)
		fromV = sqlgraph.Neighbors(l.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *LockerClient) Hooks() []Hook {
	return c.hooks.Locker
}

// Interceptors returns the client interceptors.
func (c *LockerClient) Interceptors() []Interceptor {
	return c.inters.Locker
}

func (c *LockerClient) mutate(ctx context.Context, m *LockerMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&LockerCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&LockerUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&LockerUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&LockerDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown Locker mutation op: %q", m.Op())
	}
}

// PropertyClient is a client for the Property schema.
type PropertyClient struct {
	config
}

// NewPropertyClient returns a client for the Property from the given config.
func NewPropertyClient(c config) *PropertyClient {
	return &PropertyClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `property.Hooks(f(g(h())))`.
func (c *PropertyClient) Use(hooks ...Hook) {
	c.hooks.Property = append(c.hooks.Property, hooks...)
}

// Use adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `property.Intercept(f(g(h())))`.
func (c *PropertyClient) Intercept(interceptors ...Interceptor) {
	c.inters.Property = append(c.inters.Property, interceptors...)
}

// Create returns a builder for creating a Property entity.
func (c *PropertyClient) Create() *PropertyCreate {
	mutation := newPropertyMutation(c.config, OpCreate)
	return &PropertyCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Property entities.
func (c *PropertyClient) CreateBulk(builders ...*PropertyCreate) *PropertyCreateBulk {
	return &PropertyCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Property.
func (c *PropertyClient) Update() *PropertyUpdate {
	mutation := newPropertyMutation(c.config, OpUpdate)
	return &PropertyUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *PropertyClient) UpdateOne(pr *Property) *PropertyUpdateOne {
	mutation := newPropertyMutation(c.config, OpUpdateOne, withProperty(pr))
	return &PropertyUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *PropertyClient) UpdateOneID(id int) *PropertyUpdateOne {
	mutation := newPropertyMutation(c.config, OpUpdateOne, withPropertyID(id))
	return &PropertyUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Property.
func (c *PropertyClient) Delete() *PropertyDelete {
	mutation := newPropertyMutation(c.config, OpDelete)
	return &PropertyDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *PropertyClient) DeleteOne(pr *Property) *PropertyDeleteOne {
	return c.DeleteOneID(pr.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *PropertyClient) DeleteOneID(id int) *PropertyDeleteOne {
	builder := c.Delete().Where(property.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &PropertyDeleteOne{builder}
}

// Query returns a query builder for Property.
func (c *PropertyClient) Query() *PropertyQuery {
	return &PropertyQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeProperty},
		inters: c.Interceptors(),
	}
}

// Get returns a Property entity by its id.
func (c *PropertyClient) Get(ctx context.Context, id int) (*Property, error) {
	return c.Query().Where(property.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *PropertyClient) GetX(ctx context.Context, id int) *Property {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryAccount queries the account edge of a Property.
func (c *PropertyClient) QueryAccount(pr *Property) *AccountQuery {
	query := (&AccountClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := pr.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(property.Table, property.FieldID, id),
			sqlgraph.To(entaccount.Table, entaccount.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, property.AccountTable, property.AccountColumn),
		)
		fromV = sqlgraph.Neighbors(pr.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *PropertyClient) Hooks() []Hook {
	return c.hooks.Property
}

// Interceptors returns the client interceptors.
func (c *PropertyClient) Interceptors() []Interceptor {
	return c.inters.Property
}

func (c *PropertyClient) mutate(ctx context.Context, m *PropertyMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&PropertyCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&PropertyUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&PropertyUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&PropertyDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown Property mutation op: %q", m.Op())
	}
}

// RecoveryCodeClient is a client for the RecoveryCode schema.
type RecoveryCodeClient struct {
	config
}

// NewRecoveryCodeClient returns a client for the RecoveryCode from the given config.
func NewRecoveryCodeClient(c config) *RecoveryCodeClient {
	return &RecoveryCodeClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `recoverycode.Hooks(f(g(h())))`.
func (c *RecoveryCodeClient) Use(hooks ...Hook) {
	c.hooks.RecoveryCode = append(c.hooks.RecoveryCode, hooks...)
}

// Use adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `recoverycode.Intercept(f(g(h())))`.
func (c *RecoveryCodeClient) Intercept(interceptors ...Interceptor) {
	c.inters.RecoveryCode = append(c.inters.RecoveryCode, interceptors...)
}

// Create returns a builder for creating a RecoveryCode entity.
func (c *RecoveryCodeClient) Create() *RecoveryCodeCreate {
	mutation := newRecoveryCodeMutation(c.config, OpCreate)
	return &RecoveryCodeCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of RecoveryCode entities.
func (c *RecoveryCodeClient) CreateBulk(builders ...*RecoveryCodeCreate) *RecoveryCodeCreateBulk {
	return &RecoveryCodeCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for RecoveryCode.
func (c *RecoveryCodeClient) Update() *RecoveryCodeUpdate {
	mutation := newRecoveryCodeMutation(c.config, OpUpdate)
	return &RecoveryCodeUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *RecoveryCodeClient) UpdateOne(rc *RecoveryCode) *RecoveryCodeUpdateOne {
	mutation := newRecoveryCodeMutation(c.config, OpUpdateOne, withRecoveryCode(rc))
	return &RecoveryCodeUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *RecoveryCodeClient) UpdateOneID(id int) *RecoveryCodeUpdateOne {
	mutation := newRecoveryCodeMutation(c.config, OpUpdateOne, withRecoveryCodeID(id))
	return &RecoveryCodeUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for RecoveryCode.
func (c *RecoveryCodeClient) Delete() *RecoveryCodeDelete {
	mutation := newRecoveryCodeMutation(c.config, OpDelete)
	return &RecoveryCodeDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *RecoveryCodeClient) DeleteOne(rc *RecoveryCode) *RecoveryCodeDeleteOne {
	return c.DeleteOneID(rc.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *RecoveryCodeClient) DeleteOneID(id int) *RecoveryCodeDeleteOne {
	builder := c.Delete().Where(recoverycode.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &RecoveryCodeDeleteOne{builder}
}

// Query returns a query builder for RecoveryCode.
func (c *RecoveryCodeClient) Query() *RecoveryCodeQuery {
	return &RecoveryCodeQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeRecoveryCode},
		inters: c.Interceptors(),
	}
}

// Get returns a RecoveryCode entity by its id.
func (c *RecoveryCodeClient) Get(ctx context.Context, id int) (*RecoveryCode, error) {
	return c.Query().Where(recoverycode.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *RecoveryCodeClient) GetX(ctx context.Context, id int) *RecoveryCode {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryAccount queries the account edge of a RecoveryCode.
func (c *RecoveryCodeClient) QueryAccount(rc *RecoveryCode) *AccountQuery {
	query := (&AccountClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := rc.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(recoverycode.Table, recoverycode.FieldID, id),
			sqlgraph.To(entaccount.Table, entaccount.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, recoverycode.AccountTable, recoverycode.AccountColumn),
		)
		fromV = sqlgraph.Neighbors(rc.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *RecoveryCodeClient) Hooks() []Hook {
	return c.hooks.RecoveryCode
}

// Interceptors returns the client interceptors.
func (c *RecoveryCodeClient) Interceptors() []Interceptor {
	return c.inters.RecoveryCode
}

func (c *RecoveryCodeClient) mutate(ctx context.Context, m *RecoveryCodeMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&RecoveryCodeCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&RecoveryCodeUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&RecoveryCodeUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&RecoveryCodeDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown RecoveryCode mutation op: %q", m.Op())
	}
}
