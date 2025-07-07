package dbdrivers

import (
	"database/sql/driver"
	"fmt"

	dbmysql "github.com/arran4/goa4web/internal/dbdrivers/mysql"
	dbpostgres "github.com/arran4/goa4web/internal/dbdrivers/postgres"
	dbsqlite "github.com/arran4/goa4web/internal/dbdrivers/sqlite"
)

// DBDriver provides connection creation and documentation details for a driver.
type DBDriver interface {
	// Name returns the identifier understood by database/sql.
	Name() string
	// Examples returns example DSNs for the driver.
	Examples() []string
	// OpenConnector returns a driver.Connector using the given DSN.
	OpenConnector(dsn string) (driver.Connector, error)
}

// Registry lists all supported database drivers.
var Registry = []DBDriver{
	dbmysql.Driver{},
	dbpostgres.Driver{},
	dbsqlite.Driver{},
}

// Connector returns a driver.Connector for the provided driver name.
func Connector(driverName, dsn string) (driver.Connector, error) {
	for _, drv := range Registry {
		if drv.Name() == driverName {
			return drv.OpenConnector(dsn)
		}
	}
	return nil, fmt.Errorf("unsupported driver %s", driverName)
}

// Names returns all registered driver identifiers.
func Names() []string {
	out := make([]string, len(Registry))
	for i, drv := range Registry {
		out[i] = drv.Name()
	}
	return out
}
