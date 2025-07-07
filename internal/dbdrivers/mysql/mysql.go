package mysql

import (
	"database/sql/driver"

	"github.com/arran4/goa4web/internal/dbdrivers"
	sqlmysql "github.com/go-sql-driver/mysql"
)

func connector(dsn string) (driver.Connector, error) {
	cfg, err := sqlmysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	return sqlmysql.NewConnector(cfg)
}

// Register registers the mysql connector with the dbdrivers registry.
func Register() {
	dbdrivers.RegisterConnector("mysql", connector)
}
