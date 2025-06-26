package goa4web

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	common "github.com/arran4/goa4web/core/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/emailutil"
	"github.com/arran4/goa4web/internal/middleware"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/runtimeconfig"
	"github.com/go-sql-driver/mysql"
)

var (
	dbPool         *sql.DB
	dbLogVerbosity int
)

// InitDB opens the database connection using the provided configuration
// and ensures the schema exists.
func InitDB(cfg runtimeconfig.RuntimeConfig) *common.UserError {
	dbLogVerbosity = cfg.DBLogVerbosity
	db.LogVerbosity = cfg.DBLogVerbosity
	if cfg.DBUser == "" {
		cfg.DBUser = "a4web"
	}
	if cfg.DBPass == "" {
		cfg.DBPass = "a4web"
	}
	if cfg.DBHost == "" {
		cfg.DBHost = "localhost"
	}
	if cfg.DBPort == "" {
		cfg.DBPort = "3306"
	}
	if cfg.DBName == "" {
		cfg.DBName = "a4web"
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
	mysqlCfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return &common.UserError{Err: err, ErrorMessage: "failed to parse DSN"}
	}
	baseConnector, err := mysql.NewConnector(mysqlCfg)
	if err != nil {
		return &common.UserError{Err: err, ErrorMessage: "failed to create connector"}
	}
	var connector driver.Connector = db.NewLoggingConnector(baseConnector)
	dbPool = sql.OpenDB(connector)
	if err := dbPool.Ping(); err != nil {
		return &common.UserError{Err: err, ErrorMessage: "failed to communicate with database"}
	}
	if err := ensureSchema(context.Background(), dbPool); err != nil {
		return &common.UserError{Err: err, ErrorMessage: "failed to verify schema"}
	}
	middleware.SetDBPool(dbPool, dbLogVerbosity)
	if dbLogVerbosity > 0 {
		log.Printf("db pool stats after init: %+v", dbPool.Stats())
	}
	return nil
}

// checkDatabase attempts to connect and ping the configured database.
func checkDatabase(cfg runtimeconfig.RuntimeConfig) *common.UserError {
	return InitDB(cfg)
}

func performStartupChecks(cfg runtimeconfig.RuntimeConfig) error {
	if ue := checkDatabase(cfg); ue != nil {
		return fmt.Errorf("%s: %w", ue.ErrorMessage, ue.Err)
	}
	if ue := checkUploadDir(cfg); ue != nil {
		return fmt.Errorf("%s: %w", ue.ErrorMessage, ue.Err)
	}
	return nil
}

func checkUploadDir(cfg runtimeconfig.RuntimeConfig) *common.UserError {
	if cfg.ImageUploadDir == "" {
		return &common.UserError{Err: fmt.Errorf("dir empty"), ErrorMessage: "image upload directory not set"}
	}
	if strings.HasPrefix(cfg.ImageUploadDir, "s3://") {
		// TODO: validate S3 upload targets
		return nil
	}
	info, err := os.Stat(cfg.ImageUploadDir)
	if err != nil || !info.IsDir() {
		return &common.UserError{Err: err, ErrorMessage: "image upload directory invalid"}
	}
	test := filepath.Join(cfg.ImageUploadDir, ".check")
	if err := os.WriteFile(test, []byte("ok"), 0644); err != nil {
		return &common.UserError{Err: err, ErrorMessage: "image upload directory not writable"}
	}
	os.Remove(test)
	return nil
}

// ensureSchema creates core tables if they do not exist and inserts a version row.
func ensureSchema(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_version (version INT NOT NULL)"); err != nil {
		return fmt.Errorf("create schema_version: %w", err)
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_version").Scan(&count); err != nil {
		return fmt.Errorf("count schema_version: %w", err)
	}
	if count == 0 {
		if _, err := db.ExecContext(ctx, "INSERT INTO schema_version (version) VALUES (1)"); err != nil {
			return fmt.Errorf("insert schema_version: %w", err)
		}
	}
	var version int
	if err := db.QueryRowContext(ctx, "SELECT version FROM schema_version").Scan(&version); err != nil {
		return fmt.Errorf("select schema_version: %w", err)
	}
	if version != ExpectedSchemaVersion {
		return fmt.Errorf("database schema version %d does not match expected %d", version, ExpectedSchemaVersion)
	}
	return nil
}

// startWorkers launches goroutines for email processing and notification cleanup.
func startWorkers(ctx context.Context, db *sql.DB, provider email.Provider) {
	log.Printf("Starting email worker")
	safeGo(func() { emailutil.EmailQueueWorker(ctx, New(db), provider, time.Minute) })
	log.Printf("Starting notification purger worker")
	safeGo(func() { notifications.NotificationPurgeWorker(ctx, New(db), time.Hour) })
}
