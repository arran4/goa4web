package dbdrivers

import (
	"database/sql/driver"
	"fmt"
)

// Connector returns a driver.Connector for the provided driver name and DSN.
func Connector(driverName, dsn string) (driver.Connector, error) {
	fn, ok := connectors[driverName]
	if !ok {
		return nil, fmt.Errorf("unsupported driver %s", driverName)
	}
	return fn(dsn)
}
