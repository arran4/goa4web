package sqlite

import (
	"context"
	"database/sql/driver"

	"github.com/mattn/go-sqlite3"
)

// Driver implements the dbdrivers.DBDriver interface for SQLite.
type Driver struct{}

// Name returns the driver name used by database/sql.
func (Driver) Name() string { return "sqlite3" }

// Examples returns example DSN strings.
func (Driver) Examples() []string {
	return []string{
		"file:./db.sqlite?_fk=1",
		":memory:",
	}
}

// OpenConnector creates a connector for the SQLite driver.
type connector struct{ dsn string }

func (c connector) Connect(ctx context.Context) (driver.Conn, error) {
	return (&sqlite3.SQLiteDriver{}).Open(c.dsn)
}

func (c connector) Driver() driver.Driver { return &sqlite3.SQLiteDriver{} }

func (Driver) OpenConnector(dsn string) (driver.Connector, error) {
	return connector{dsn: dsn}, nil
}
