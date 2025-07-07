package mysql

import (
	"database/sql/driver"

	mysqldriver "github.com/go-sql-driver/mysql"
)

// Driver implements the dbdrivers.DBDriver interface for MySQL.
type Driver struct{}

func (Driver) Name() string { return "mysql" }

func (Driver) Examples() []string {
	return []string{
		"user:pass@tcp(localhost:3306)/a4web",
		"user:pass@unix(/var/run/mysqld/mysqld.sock)/a4web",
	}
}

func (Driver) OpenConnector(dsn string) (driver.Connector, error) {
	cfg, err := mysqldriver.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	return mysqldriver.NewConnector(cfg)
}
