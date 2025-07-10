package dbdrivers

import (
	"database/sql/driver"
	"fmt"
	"log"
	"sort"
	"sync"
)

// DBDriver describes a database driver and how to create connectors.
type DBDriver interface {
	Name() string
	Examples() []string
	OpenConnector(dsn string) (driver.Connector, error)
	Backup(dsn, file string) error
	Restore(dsn, file string) error
}

// Registry lists the built-in database drivers.
var (
	regMu    sync.RWMutex
	Registry []DBDriver
)

// RegisterDriver adds d to the Registry.
func RegisterDriver(d DBDriver) {
	regMu.Lock()
	defer regMu.Unlock()
	for _, r := range Registry {
		if r.Name() == d.Name() {
			log.Printf("dbdrivers: driver %s already registered", d.Name())
			return
		}
	}
	Registry = append(Registry, d)
}

// Connector returns a driver.Connector for the driver with the given name and
// DSN. It searches the Registry for a matching driver.
func Connector(name, dsn string) (driver.Connector, error) {
	regMu.RLock()
	drivers := append([]DBDriver(nil), Registry...)
	regMu.RUnlock()
	for _, d := range drivers {
		if d.Name() == name {
			return d.OpenConnector(dsn)
		}
	}
	return nil, fmt.Errorf("unsupported driver %s", name)
}

// Driver looks up a registered driver by name.
func Driver(name string) (DBDriver, error) {
	regMu.RLock()
	drivers := append([]DBDriver(nil), Registry...)
	regMu.RUnlock()
	for _, d := range drivers {
		if d.Name() == name {
			return d, nil
		}
	}
	return nil, fmt.Errorf("unsupported driver %s", name)
}

// Backup invokes the driver's Backup method.
func Backup(name, dsn, file string) error {
	d, err := Driver(name)
	if err != nil {
		return err
	}
	return d.Backup(dsn, file)
}

// Restore invokes the driver's Restore method.
func Restore(name, dsn, file string) error {
	d, err := Driver(name)
	if err != nil {
		return err
	}
	return d.Restore(dsn, file)
}

// Names returns the names of all registered drivers.
func Names() []string {
	regMu.RLock()
	drivers := append([]DBDriver(nil), Registry...)
	regMu.RUnlock()
	m := map[string]struct{}{}
	for _, d := range drivers {
		m[d.Name()] = struct{}{}
	}
	names := make([]string, 0, len(m))
	for n := range m {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
