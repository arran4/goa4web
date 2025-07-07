package dbdrivers

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
)

type sqliteConnector struct {
	dsn string
}

func (c sqliteConnector) Connect(context.Context) (driver.Conn, error) {
	return (&sqlite3.SQLiteDriver{}).Open(c.dsn)
}

func (c sqliteConnector) Driver() driver.Driver { return &sqlite3.SQLiteDriver{} }

// Connector returns a driver.Connector for the provided driver name and DSN.
func Connector(driverName, dsn string) (driver.Connector, error) {
	switch driverName {
	case "mysql":
		cfg, err := mysql.ParseDSN(dsn)
		if err != nil {
			return nil, err
		}
		return mysql.NewConnector(cfg)
	case "postgres":
		return pq.NewConnector(dsn)
	case "sqlite3":
		return sqliteConnector{dsn: dsn}, nil
	default:
		return nil, fmt.Errorf("unsupported driver %s", driverName)
	}
}
