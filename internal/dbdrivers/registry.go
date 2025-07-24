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

// Registry lists registered database drivers.
type Registry struct {
	mu      sync.RWMutex
	drivers []DBDriver
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry { return &Registry{} }

// RegisterDriver adds d to the registry.
func (r *Registry) RegisterDriver(d DBDriver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, dr := range r.drivers {
		if dr.Name() == d.Name() {
			log.Printf("dbdrivers: driver %s already registered", d.Name())
			return
		}
	}
	r.drivers = append(r.drivers, d)
}

// Drivers returns a copy of registered drivers.
func (r *Registry) Drivers() []DBDriver {
	r.mu.RLock()
	ds := append([]DBDriver(nil), r.drivers...)
	r.mu.RUnlock()
	return ds
}

// Connector returns a connector for the named driver.
func (r *Registry) Connector(name, dsn string) (driver.Connector, error) {
	for _, d := range r.Drivers() {
		if d.Name() == name {
			return d.OpenConnector(dsn)
		}
	}
	return nil, fmt.Errorf("unsupported driver %s", name)
}

// Driver returns the driver by name.
func (r *Registry) Driver(name string) (DBDriver, error) {
	for _, d := range r.Drivers() {
		if d.Name() == name {
			return d, nil
		}
	}
	return nil, fmt.Errorf("unsupported driver %s", name)
}

// Backup invokes the driver's Backup method.
func (r *Registry) Backup(name, dsn, file string) error {
	d, err := r.Driver(name)
	if err != nil {
		return err
	}
	return d.Backup(dsn, file)
}

// Restore invokes the driver's Restore method.
func (r *Registry) Restore(name, dsn, file string) error {
	d, err := r.Driver(name)
	if err != nil {
		return err
	}
	return d.Restore(dsn, file)
}

// Names returns the names of registered drivers.
func (r *Registry) Names() []string {
	m := map[string]struct{}{}
	for _, d := range r.Drivers() {
		m[d.Name()] = struct{}{}
	}
	names := make([]string, 0, len(m))
	for n := range m {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// DefaultRegistry holds the package default drivers.
var DefaultRegistry = NewRegistry()

// RegisterDriver adds d to the DefaultRegistry.
func RegisterDriver(d DBDriver) { DefaultRegistry.RegisterDriver(d) }

// Connector uses DefaultRegistry to create a connector.
func Connector(name, dsn string) (driver.Connector, error) {
	return DefaultRegistry.Connector(name, dsn)
}

// Driver uses DefaultRegistry to find a driver.
func Driver(name string) (DBDriver, error) { return DefaultRegistry.Driver(name) }

// Backup uses DefaultRegistry to backup a database.
func Backup(name, dsn, file string) error { return DefaultRegistry.Backup(name, dsn, file) }

// Restore uses DefaultRegistry to restore a database.
func Restore(name, dsn, file string) error { return DefaultRegistry.Restore(name, dsn, file) }

// Names lists drivers in the DefaultRegistry.
func Names() []string { return DefaultRegistry.Names() }
