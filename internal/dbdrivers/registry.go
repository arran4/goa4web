package dbdrivers

import (
	"database/sql/driver"
	"fmt"
	"sort"

	"github.com/arran4/goa4web/internal/dbdrivers/mysql"
	"github.com/arran4/goa4web/internal/dbdrivers/postgres"
	"github.com/arran4/goa4web/internal/dbdrivers/sqlite"
)

// DBDriver describes a database driver and how to create connectors.
type DBDriver interface {
	Name() string
	Examples() []string
	OpenConnector(dsn string) (driver.Connector, error)
}

// Registry lists the built-in database drivers.
var Registry = []DBDriver{
	mysql.Driver{},
	postgres.Driver{},
	sqlite.Driver{},
}

// Connector returns a driver.Connector for the driver with the given name and
// DSN. It searches the Registry for a matching driver.
func Connector(name, dsn string) (driver.Connector, error) {
	for _, d := range Registry {
		if d.Name() == name {
			return d.OpenConnector(dsn)
		}
	}
	return nil, fmt.Errorf("unsupported driver %s", name)
}

// Names returns the names of all registered drivers.
func Names() []string {
	m := map[string]struct{}{}
	for _, d := range Registry {
		m[d.Name()] = struct{}{}
	}
	names := make([]string, 0, len(m))
	for n := range m {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
