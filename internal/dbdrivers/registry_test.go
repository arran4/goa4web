package dbdrivers

import (
	"database/sql/driver"
	"reflect"
	"testing"
)

type mockDriver struct {
	name          string
	connector     driver.Connector
	openConnErr   error
	backupErr     error
	restoreErr    error
	calledBackup  bool
	calledRestore bool
	dsnArg        string
	fileArg       string
}

func (m *mockDriver) Name() string { return m.name }

func (m *mockDriver) Examples() []string { return []string{"example"} }

func (m *mockDriver) OpenConnector(dsn string) (driver.Connector, error) {
	m.dsnArg = dsn
	return m.connector, m.openConnErr
}

func (m *mockDriver) Backup(dsn, file string) error {
	m.calledBackup = true
	m.dsnArg = dsn
	m.fileArg = file
	return m.backupErr
}

func (m *mockDriver) Restore(dsn, file string) error {
	m.calledRestore = true
	m.dsnArg = dsn
	m.fileArg = file
	return m.restoreErr
}

// Ensure mockDriver implements DBDriver
var _ DBDriver = (*mockDriver)(nil)

func TestRegistry(t *testing.T) {
	t.Run("RegisterDriver", func(t *testing.T) {
		r := NewRegistry()
		d1 := &mockDriver{name: "driver1"}
		r.RegisterDriver(d1)

		drivers := r.Drivers()
		if len(drivers) != 1 {
			t.Errorf("expected 1 driver, got %d", len(drivers))
		}
		if drivers[0].Name() != "driver1" {
			t.Errorf("expected driver name 'driver1', got %s", drivers[0].Name())
		}

		// Test duplicate registration
		r.RegisterDriver(d1)
		drivers = r.Drivers()
		if len(drivers) != 1 {
			t.Errorf("expected 1 driver after duplicate registration, got %d", len(drivers))
		}
	})

	t.Run("Driver", func(t *testing.T) {
		r := NewRegistry()
		d1 := &mockDriver{name: "driver1"}
		r.RegisterDriver(d1)

		d, err := r.Driver("driver1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if d.Name() != "driver1" {
			t.Errorf("expected driver name 'driver1', got %s", d.Name())
		}

		_, err = r.Driver("unknown")
		if err == nil {
			t.Errorf("expected error for unknown driver, got nil")
		}
	})

	t.Run("Names", func(t *testing.T) {
		r := NewRegistry()
		r.RegisterDriver(&mockDriver{name: "b_driver"})
		r.RegisterDriver(&mockDriver{name: "a_driver"})

		names := r.Names()
		expected := []string{"a_driver", "b_driver"}
		if !reflect.DeepEqual(names, expected) {
			t.Errorf("expected names %v, got %v", expected, names)
		}
	})

	t.Run("Connector", func(t *testing.T) {
		r := NewRegistry()
		d1 := &mockDriver{name: "driver1"}
		r.RegisterDriver(d1)

		conn, err := r.Connector("driver1", "dsn")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if conn != nil { // mock returns nil connector by default
			t.Errorf("expected nil connector from mock, got %v", conn)
		}
		if d1.dsnArg != "dsn" {
			t.Errorf("expected dsn 'dsn', got %s", d1.dsnArg)
		}

		_, err = r.Connector("unknown", "dsn")
		if err == nil {
			t.Errorf("expected error for unknown driver, got nil")
		}
	})

	t.Run("Backup", func(t *testing.T) {
		r := NewRegistry()
		d1 := &mockDriver{name: "driver1"}
		r.RegisterDriver(d1)

		err := r.Backup("driver1", "dsn", "file")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !d1.calledBackup {
			t.Errorf("expected Backup to be called")
		}
		if d1.dsnArg != "dsn" || d1.fileArg != "file" {
			t.Errorf("expected args 'dsn', 'file', got %s, %s", d1.dsnArg, d1.fileArg)
		}

		err = r.Backup("unknown", "dsn", "file")
		if err == nil {
			t.Errorf("expected error for unknown driver, got nil")
		}
	})

	t.Run("Restore", func(t *testing.T) {
		r := NewRegistry()
		d1 := &mockDriver{name: "driver1"}
		r.RegisterDriver(d1)

		err := r.Restore("driver1", "dsn", "file")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !d1.calledRestore {
			t.Errorf("expected Restore to be called")
		}
		if d1.dsnArg != "dsn" || d1.fileArg != "file" {
			t.Errorf("expected args 'dsn', 'file', got %s, %s", d1.dsnArg, d1.fileArg)
		}

		err = r.Restore("unknown", "dsn", "file")
		if err == nil {
			t.Errorf("expected error for unknown driver, got nil")
		}
	})
}
