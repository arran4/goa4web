package postgres

import (
	"database/sql/driver"

	"github.com/lib/pq"
)

// Driver implements the dbdrivers.DBDriver interface for PostgreSQL.
type Driver struct{}

func (Driver) Name() string { return "postgres" }

func (Driver) Examples() []string {
	return []string{
		"postgres://user:pass@localhost/a4web?sslmode=disable",
	}
}

func (Driver) OpenConnector(dsn string) (driver.Connector, error) {
	return pq.NewConnector(dsn)
}
