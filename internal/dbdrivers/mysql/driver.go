package mysql

import (
	"database/sql/driver"

	"github.com/arran4/goa4web/internal/dbdrivers"
	sqlmysql "github.com/go-sql-driver/mysql"
)

// Driver implements the dbdrivers.DBDriver interface for MySQL.
type Driver struct{}

// Name returns the driver name.
func (Driver) Name() string { return "mysql" }

// Examples returns example DSN strings.
func (Driver) Examples() []string {
	return []string{
		"user:pass@tcp(127.0.0.1:3306)/dbname?parseTime=true",
		"user:pass@unix(/var/run/mysqld/mysqld.sock)/dbname?parseTime=true",
	}
}

// OpenConnector parses the DSN and returns a Connector.
func (Driver) OpenConnector(dsn string) (driver.Connector, error) {
	cfg, err := sqlmysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	return sqlmysql.NewConnector(cfg)
}

// Register registers the MySQL driver.
func Register() { dbdrivers.RegisterDriver(Driver{}) }
