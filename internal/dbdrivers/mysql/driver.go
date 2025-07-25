package mysql

import (
	"database/sql/driver"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/arran4/goa4web/internal/dbdrivers"
	sqlmysql "github.com/go-sql-driver/mysql"
)

// Driver implements the dbdrivers.DBDriver interface for MySQL.
type Driver struct{}

// Name returns the driver name.
func (Driver) Name() string { return "mysql" }

// Examples returns example DSN strings.
func (Driver) Examples() []string {
	return []string{
		"user:pass@tcp(127.0.0.1:3306)/dbname?parseTime=true",
		"user:pass@unix(/var/run/mysqld/mysqld.sock)/dbname?parseTime=true",
	}
}

// OpenConnector parses the DSN and returns a Connector.
func (Driver) OpenConnector(dsn string) (driver.Connector, error) {
	cfg, err := sqlmysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	return sqlmysql.NewConnector(cfg)
}

// Backup dumps the database to file using mysqldump.
func (Driver) Backup(dsn, file string) error {
	if dsn == "" {
		return fmt.Errorf("connection string required")
	}
	mcfg, err := sqlmysql.ParseDSN(dsn)
	if err != nil {
		return fmt.Errorf("parse DSN: %w", err)
	}
	host, port, _ := strings.Cut(mcfg.Addr, ":")
	args := []string{"-h", host, "-P", port, "-u", mcfg.User, fmt.Sprintf("-p%s", mcfg.Passwd), mcfg.DBName}
	cmd := exec.Command("mysqldump", args...)
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

// Restore loads the database from the provided file using mysql.
func (Driver) Restore(dsn, file string) error {
	if dsn == "" {
		return fmt.Errorf("connection string required")
	}
	inFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer inFile.Close()
	mcfg, err := sqlmysql.ParseDSN(dsn)
	if err != nil {
		return fmt.Errorf("parse DSN: %w", err)
	}
	host, port, _ := strings.Cut(mcfg.Addr, ":")
	args := []string{"-h", host, "-P", port, "-u", mcfg.User, fmt.Sprintf("-p%s", mcfg.Passwd), mcfg.DBName}
	cmd := exec.Command("mysql", args...)
	cmd.Stdin = inFile
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restore: %w", err)
	}
	return nil
}

// Register registers the MySQL driver.
func Register(r *dbdrivers.Registry) { r.RegisterDriver(Driver{}) }
