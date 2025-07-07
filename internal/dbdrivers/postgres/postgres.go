package postgres

import (
	"database/sql/driver"

	"github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/lib/pq"
)

func connector(dsn string) (driver.Connector, error) {
	return pq.NewConnector(dsn)
}

// Register registers the postgres connector with the dbdrivers registry.
func Register() {
	dbdrivers.RegisterConnector("postgres", connector)
}
