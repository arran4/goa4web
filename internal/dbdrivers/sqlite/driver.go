//go:build sqlite || sqlite3

package sqlite

import (
	"context"
	"database/sql/driver"
	"fmt"
	"os"
	"os/exec"

	"github.com/arran4/goa4web/internal/dbdrivers"
	sqlitesql "modernc.org/sqlite"
)

// Driver implements the dbdrivers.DBDriver interface for SQLite.
type Driver struct {
	name string
}

// NewDriver creates a new SQLite driver with the specified name.
func NewDriver(name string) Driver {
	return Driver{name: name}
}

// Name returns the driver name.
func (d Driver) Name() string {
	if d.name != "" {
		return d.name
	}
	return "sqlite3"
}

// Examples returns example DSN strings.
func (Driver) Examples() []string {
	return []string{
		"data.db",
		"file:data.db?cache=shared&mode=rwc",
		":memory:",
	}
}

type sqliteConnector struct {
	dsn string
	drv driver.Driver
}

func (c sqliteConnector) Connect(context.Context) (driver.Conn, error) {
	return c.drv.Open(c.dsn)
}

func (c sqliteConnector) Driver() driver.Driver {
	return c.drv
}

// OpenConnector parses the DSN and returns a Connector.
func (d Driver) OpenConnector(dsn string) (driver.Connector, error) {
	return sqliteConnector{
		dsn: dsn,
		drv: &sqlitesql.Driver{},
	}, nil
}

// Backup dumps the database to file using sqlite3 .dump.
func (Driver) Backup(dsn, file string) error {
	if dsn == "" {
		return fmt.Errorf("connection string required")
	}
	outFile, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	cmd := exec.Command("sqlite3", dsn, ".dump")
	cmd.Stdout = outFile
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("backup: %w", err)
	}
	return nil
}

// Restore loads the database from the provided file using sqlite3 CLI.
func (Driver) Restore(dsn, file string) error {
	if dsn == "" {
		return fmt.Errorf("connection string required")
	}
	inFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = inFile.Close() }()

	cmd := exec.Command("sqlite3", dsn)
	cmd.Stdin = inFile
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restore: %w", err)
	}
	return nil
}

// Register registers SQLite drivers ("sqlite3" and "sqlite") in the registry.
func Register(r *dbdrivers.Registry) {
	r.RegisterDriver(Driver{name: "sqlite3"})
	r.RegisterDriver(Driver{name: "sqlite"})
}
