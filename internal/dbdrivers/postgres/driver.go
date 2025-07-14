package postgres

import (
	"database/sql/driver"
	"fmt"
	"os"
	"os/exec"

	"github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/lib/pq"
)

// Driver implements the dbdrivers.DBDriver interface for PostgreSQL.
type Driver struct{}

// Name returns the driver name.
func (Driver) Name() string { return "postgres" }

// Examples returns example connection strings.
func (Driver) Examples() []string {
	return []string{
		"postgres://user:pass@localhost/dbname?sslmode=disable",
		"user=foo password=bar dbname=mydb sslmode=disable",
	}
}

// OpenConnector wraps pq.NewConnector.
func (Driver) OpenConnector(dsn string) (driver.Connector, error) {
	return pq.NewConnector(dsn)
}

// Backup uses pg_dump to create a database backup.
func (Driver) Backup(dsn, file string) error {
	if dsn == "" {
		return fmt.Errorf("connection string required")
	}
	cmd := exec.Command("pg_dump", "--dbname="+dsn)
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

// Restore loads a database backup using psql.
func (Driver) Restore(dsn, file string) error {
	if dsn == "" {
		return fmt.Errorf("connection string required")
	}
	inFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer inFile.Close()
	cmd := exec.Command("psql", dsn)
	cmd.Stdin = inFile
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restore: %w", err)
	}
	return nil
}

// Register registers the PostgreSQL driver.
func Register() { dbdrivers.RegisterDriver(Driver{}) }
