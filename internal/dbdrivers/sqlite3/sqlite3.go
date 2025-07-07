package sqlite3

import (
	"context"
	"database/sql/driver"

	"github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/mattn/go-sqlite3"
)

type connector struct{ dsn string }

func (c connector) Connect(ctx context.Context) (driver.Conn, error) {
	return (&sqlite3.SQLiteDriver{}).Open(c.dsn)
}

func (c connector) Driver() driver.Driver { return &sqlite3.SQLiteDriver{} }

func connectorFunc(dsn string) (driver.Connector, error) {
	return connector{dsn: dsn}, nil
}

// Register registers the sqlite3 connector with the dbdrivers registry.
func Register() {
	dbdrivers.RegisterConnector("sqlite3", connectorFunc)
}
