package dbstart

import (
	"fmt"

	sqlmysql "github.com/go-sql-driver/mysql"
)

// MigrationConnectionString enables multi-statement SQL only on the short-lived
// migration connection. Legacy Goose StatementBegin blocks contain multiple SQL
// statements, but regular application connections must retain their DSN policy.
func MigrationConnectionString(driverName, dsn string) (string, error) {
	if driverName != "mysql" {
		return dsn, nil
	}
	cfg, err := sqlmysql.ParseDSN(dsn)
	if err != nil {
		return "", fmt.Errorf("parse MySQL migration DSN: %w", err)
	}
	cfg.MultiStatements = true
	return cfg.FormatDSN(), nil
}
