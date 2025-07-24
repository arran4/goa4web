//go:build sqlite

package sqlite

import (
	"context"
	"database/sql/driver"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/mattn/go-sqlite3"
)

// Driver implements the dbdrivers.DBDriver interface for SQLite.
type Driver struct{}

// Name returns the driver name used by database/sql.
func (Driver) Name() string { return "sqlite3" }

// Examples returns example DSN strings.
func (Driver) Examples() []string {
	return []string{
		"file:./db.sqlite?_fk=1",
		":memory:",
	}
}

// OpenConnector creates a connector for the SQLite driver.
type connector struct{ dsn string }

func (c connector) Connect(ctx context.Context) (driver.Conn, error) {
	return (&sqlite3.SQLiteDriver{}).Open(c.dsn)
}

func (c connector) Driver() driver.Driver { return &sqlite3.SQLiteDriver{} }

func (Driver) OpenConnector(dsn string) (driver.Connector, error) {
	return connector{dsn: dsn}, nil
}

func pathFromConn(conn string) string {
	path := conn
	if strings.HasPrefix(path, "file:") {
		if u, err := url.Parse(path); err == nil {
			path = u.Path
		}
	}
	return path
}

// Backup dumps the SQLite database to file using the sqlite3 command.
func (Driver) Backup(dsn, file string) error {
	path := pathFromConn(dsn)
	cmd := exec.Command("sqlite3", path, ".dump")
	outFile, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer outFile.Close()
	cmd.Stdout = outFile
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("backup: %w", err)
	}
	return nil
}

// Restore loads the SQLite database from file using the sqlite3 command.
func (Driver) Restore(dsn, file string) error {
	path := pathFromConn(dsn)
	inFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer inFile.Close()
	cmd := exec.Command("sqlite3", path)
	cmd.Stdin = inFile
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sqlite restore: %w", err)
	}
	return nil
}

// Register registers the SQLite driver.
func Register(r *dbdrivers.Registry) { r.RegisterDriver(Driver{}) }
