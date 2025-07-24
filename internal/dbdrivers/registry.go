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

// Registry maintains registered database drivers.
type Registry struct {
	mu      sync.RWMutex
	drivers []DBDriver
}

// NewRegistry returns an empty driver registry.
func NewRegistry() *Registry { return &Registry{} }

// RegisterDriver adds d to the Registry.
func (r *Registry) RegisterDriver(d DBDriver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, exist := range r.drivers {
		if exist.Name() == d.Name() {
			log.Printf("dbdrivers: driver %s already registered", d.Name())
			return
		}
	}
	r.drivers = append(r.drivers, d)
}

// Connector returns a driver.Connector for the driver with the given name and
// DSN. It searches the Registry for a matching driver.
func (r *Registry) Connector(name, dsn string) (driver.Connector, error) {
	r.mu.RLock()
	drivers := append([]DBDriver(nil), r.drivers...)
	r.mu.RUnlock()
	for _, d := range drivers {
		if d.Name() == name {
			return d.OpenConnector(dsn)
		}
	}
	return nil, fmt.Errorf("unsupported driver %s", name)
}

// Driver looks up a registered driver by name.
func (r *Registry) Driver(name string) (DBDriver, error) {
	r.mu.RLock()
	drivers := append([]DBDriver(nil), r.drivers...)
	r.mu.RUnlock()
	for _, d := range drivers {
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

// Names returns the names of all registered drivers.
func (r *Registry) Names() []string {
	r.mu.RLock()
	drivers := append([]DBDriver(nil), r.drivers...)
	r.mu.RUnlock()
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
