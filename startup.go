package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"

	"github.com/go-sql-driver/mysql"
)

var (
	dbPool         *sql.DB
	dbLogVerbosity int
)

func InitDB() *UserError {
	cfg := loadDBConfig()
	dbLogVerbosity = cfg.LogVerbosity
	if cfg.User == "" {
		cfg.User = "a4web"
	}
	if cfg.Pass == "" {
		cfg.Pass = "a4web"
	}
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == "" {
		cfg.Port = "3306"
	}
	if cfg.Name == "" {
		cfg.Name = "a4web"
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.Name)
	mysqlCfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return &UserError{Err: err, ErrorMessage: "failed to parse DSN"}
	}
	baseConnector, err := mysql.NewConnector(mysqlCfg)
	if err != nil {
		return &UserError{Err: err, ErrorMessage: "failed to create connector"}
	}
	var connector driver.Connector = baseConnector
	if dbLogVerbosity > 0 {
		connector = loggingConnector{baseConnector}
	}
	dbPool = sql.OpenDB(connector)
	if err := dbPool.Ping(); err != nil {
		return &UserError{Err: err, ErrorMessage: "failed to communicate with database"}
	}
	if err := ensureSchema(context.Background(), dbPool); err != nil {
		return &UserError{Err: err, ErrorMessage: "failed to verify schema"}
	}
	if dbLogVerbosity > 0 {
		log.Printf("db pool stats after init: %+v", dbPool.Stats())
	}
	return nil
}

// checkDatabase attempts to connect and ping the configured database.
func checkDatabase() *UserError {
	return InitDB()
}

func performStartupChecks() error {
	if ue := checkDatabase(); ue != nil {
		return fmt.Errorf("%s: %w", ue.ErrorMessage, ue.Err)
	}
	return nil
}

// ensureSchema creates core tables if they do not exist and inserts a version row.
func ensureSchema(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_version (version INT NOT NULL)"); err != nil {
		return err
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_version").Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if _, err := db.ExecContext(ctx, "INSERT INTO schema_version (version) VALUES (1)"); err != nil {
			return err
		}
	}
	return nil
}
