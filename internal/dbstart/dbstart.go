package dbstart

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	common "github.com/arran4/goa4web/core/common"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	dbdrivers "github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/arran4/goa4web/internal/middleware"
	"github.com/arran4/goa4web/runtimeconfig"
)

var (
	dbPool         *sql.DB
	dbLogVerbosity int
)

// GetDBPool returns the active database connection pool.
func GetDBPool() *sql.DB { return dbPool }

// InitDB opens the database connection using the provided configuration
// and ensures the schema exists.
func InitDB(cfg runtimeconfig.RuntimeConfig) *common.UserError {
	dbLogVerbosity = cfg.DBLogVerbosity
	db.LogVerbosity = cfg.DBLogVerbosity
	conn := cfg.DBConn
	if conn == "" {
		return &common.UserError{Err: fmt.Errorf("connection string required"), ErrorMessage: "missing connection"}
	}
	c, err := dbdrivers.Connector(cfg.DBDriver, conn)
	if err != nil {
		return &common.UserError{Err: err, ErrorMessage: "failed to create connector"}
	}
	var connector driver.Connector = db.NewLoggingConnector(c)
	dbPool = sql.OpenDB(connector)
	if err := dbPool.Ping(); err != nil {
		return &common.UserError{Err: err, ErrorMessage: "failed to communicate with database"}
	}
	if err := EnsureSchema(context.Background(), dbPool); err != nil {
		return &common.UserError{Err: err, ErrorMessage: "failed to verify schema"}
	}
	middleware.SetDBPool(dbPool, dbLogVerbosity)
	if dbLogVerbosity > 0 {
		log.Printf("db pool stats after init: %+v", dbPool.Stats())
	}
	return nil
}

// PerformStartupChecks checks the database and upload directory configuration.
func PerformStartupChecks(cfg runtimeconfig.RuntimeConfig) error {
	if err := MaybeAutoMigrate(cfg); err != nil {
		return err
	}
	if ue := InitDB(cfg); ue != nil {
		return fmt.Errorf("%s: %w", ue.ErrorMessage, ue.Err)
	}
	if ue := CheckUploadDir(cfg); ue != nil {
		return fmt.Errorf("%s: %w", ue.ErrorMessage, ue.Err)
	}
	return nil
}

// CheckUploadDir verifies that the upload directory is accessible.
func CheckUploadDir(cfg runtimeconfig.RuntimeConfig) *common.UserError {
	if cfg.ImageUploadDir == "" {
		return &common.UserError{Err: fmt.Errorf("dir empty"), ErrorMessage: "image upload directory not set"}
	}
	if strings.HasPrefix(cfg.ImageUploadDir, "s3://") {
		// TODO: validate S3 upload targets
		return nil
	}
	info, err := os.Stat(cfg.ImageUploadDir)
	if (err != nil || !info.IsDir()) && cfg.CreateDirs {
		if err := os.MkdirAll(cfg.ImageUploadDir, 0o755); err != nil {
			return &common.UserError{Err: err, ErrorMessage: "image upload directory invalid"}
		}
		info, err = os.Stat(cfg.ImageUploadDir)
	}
	if err != nil || !info.IsDir() {
		return &common.UserError{Err: err, ErrorMessage: "image upload directory invalid"}
	}
	test := filepath.Join(cfg.ImageUploadDir, ".check")
	if err := os.WriteFile(test, []byte("ok"), 0644); err != nil {
		return &common.UserError{Err: err, ErrorMessage: "image upload directory not writable"}
	}
	os.Remove(test)

	if cfg.ImageCacheDir != "" {
		info, err := os.Stat(cfg.ImageCacheDir)
		if (err != nil || !info.IsDir()) && cfg.CreateDirs {
			if err := os.MkdirAll(cfg.ImageCacheDir, 0o755); err != nil {
				return &common.UserError{Err: err, ErrorMessage: "image cache directory invalid"}
			}
			info, err = os.Stat(cfg.ImageCacheDir)
		}
		if err != nil || !info.IsDir() {
			return &common.UserError{Err: err, ErrorMessage: "image cache directory invalid"}
		}
		test := filepath.Join(cfg.ImageCacheDir, ".check")
		if err := os.WriteFile(test, []byte("ok"), 0644); err != nil {
			return &common.UserError{Err: err, ErrorMessage: "image cache directory not writable"}
		}
		os.Remove(test)
	}
	return nil
}

// EnsureSchema creates core tables if they do not exist and inserts a version row.
func EnsureSchema(ctx context.Context, db *sql.DB) error {
	version, err := ensureVersionTable(ctx, db)
	if err != nil {
		return err
	}
	if version != hcommon.ExpectedSchemaVersion {
		msg := RenderSchemaMismatch(version, hcommon.ExpectedSchemaVersion)
		return fmt.Errorf(msg)
	}
	return nil
}
