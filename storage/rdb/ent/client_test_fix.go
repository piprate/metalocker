package ent

import (
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
)

// SetOneOpenConnection sets maxOpenConns to 1. We use it in SQLite based unit tests to avoid 'database locked'
// concurrency issues.
func (c *Client) SetOneOpenConnection() {
	drvSource := c.config.driver
	if debugDrv, wrapped := drvSource.(*dialect.DebugDriver); wrapped {
		drvSource = debugDrv.Driver
	}
	drv, ok := drvSource.(*sql.Driver)
	if !ok {
		panic("can't set SetMaxOpenConns: not a SQL driver")
	}
	drv.DB().SetMaxOpenConns(1)
}
