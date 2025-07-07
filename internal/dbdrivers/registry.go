package dbdrivers

import (
	"database/sql/driver"
)

// connectors holds factories for creating driver connectors by name.
var connectors = map[string]func(string) (driver.Connector, error){}

// RegisterConnector registers a function to construct a driver.Connector for the
// given name.
func RegisterConnector(name string, fn func(string) (driver.Connector, error)) {
	connectors[name] = fn
}
