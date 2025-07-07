package sqlite

import (
	"context"
	"database/sql/driver"

	"github.com/mattn/go-sqlite3"
)

// Driver implements the dbdrivers.DBDriver interface for SQLite.
type Driver struct{}

func (Driver) Name() string { return "sqlite3" }

func (Driver) Examples() []string { return []string{"file:./a4web.db"} }

func (Driver) OpenConnector(dsn string) (driver.Connector, error) {
	return sqliteConnector{dsn: dsn}, nil
}

// sqliteConnector wraps the sqlite3 driver to satisfy driver.Connector.
type sqliteConnector struct{ dsn string }

func (c sqliteConnector) Connect(context.Context) (driver.Conn, error) {
	return (&sqlite3.SQLiteDriver{}).Open(c.dsn)
}

func (c sqliteConnector) Driver() driver.Driver { return &sqlite3.SQLiteDriver{} }
