package postgres

import (
	"database/sql/driver"

	"github.com/lib/pq"
)

// Driver implements the dbdrivers.DBDriver interface for PostgreSQL.
type Driver struct{}

// Name returns the driver name.
func (Driver) Name() string { return "postgres" }

// Examples returns example connection strings.
func (Driver) Examples() []string {
	return []string{
		"postgres://user:pass@localhost/dbname?sslmode=disable",
		"user=foo password=bar dbname=mydb sslmode=disable",
	}
}

// OpenConnector wraps pq.NewConnector.
func (Driver) OpenConnector(dsn string) (driver.Connector, error) {
	return pq.NewConnector(dsn)
}
