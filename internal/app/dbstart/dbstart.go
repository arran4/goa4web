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

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dbdrivers"
	dbmysql "github.com/arran4/goa4web/internal/dbdrivers/mysql"
)

// parseS3Dir validates S3 paths in the form s3://bucket/prefix. The bucket
// component is required while the prefix may be empty.
func parseS3Dir(raw string) (bucket, prefix string, err error) {
	p := strings.TrimPrefix(raw, "s3://")
	if p == "" {
		return "", "", fmt.Errorf("missing bucket")
	}
	parts := strings.SplitN(p, "/", 2)
	bucket = parts[0]
	if bucket == "" {
		return "", "", fmt.Errorf("missing bucket")
	}
	if len(parts) == 2 {
		prefix = strings.TrimPrefix(parts[1], "/")
	}
	return bucket, prefix, nil
}

// InitDB opens the database connection using the provided configuration
// and ensures the schema exists.
func InitDB(cfg *config.RuntimeConfig, reg *dbdrivers.Registry) (*sql.DB, *common.UserError) {
	conn := cfg.DBConn
	if conn == "" {
		return nil, &common.UserError{Err: fmt.Errorf("connection string required"), ErrorMessage: "missing connection"}
	}
	if cfg.DBDriver == "mysql" {
		dbmysql.SetTimezone(cfg.DBTimezone)
	}
	c, err := reg.Connector(cfg.DBDriver, conn)
	if err != nil {
		return nil, &common.UserError{Err: err, ErrorMessage: "failed to create connector"}
	}
	var connector driver.Connector = db.NewLoggingConnector(c, cfg.DBLogVerbosity)
	dbPool := sql.OpenDB(connector)
	if err := dbPool.Ping(); err != nil {
		dbPool.Close()
		return nil, &common.UserError{Err: err, ErrorMessage: "failed to communicate with database"}
	}
	if err := EnsureSchema(context.Background(), dbPool); err != nil {
		dbPool.Close()
		return nil, &common.UserError{Err: err, ErrorMessage: "failed to verify schema"}
	}
	if cfg.DBLogVerbosity > 0 {
		log.Printf("db pool stats after init: %+v", dbPool.Stats())
	}
	return dbPool, nil
}

// PerformStartupChecks checks the database and upload directory configuration.
func PerformStartupChecks(cfg *config.RuntimeConfig, reg *dbdrivers.Registry) (*sql.DB, error) {
	if err := MaybeAutoMigrate(cfg, reg); err != nil {
		return nil, err
	}
	dbPool, ue := InitDB(cfg, reg)
	if ue != nil {
		return nil, fmt.Errorf("%s: %w", ue.ErrorMessage, ue.Err)
	}
	if ue := CheckUploadDir(cfg); ue != nil {
		dbPool.Close()
		return nil, fmt.Errorf("%s: %w", ue.ErrorMessage, ue.Err)
	}
	return dbPool, nil
}

// CheckUploadDir verifies that the upload directory is accessible.
func CheckUploadDir(cfg *config.RuntimeConfig) *common.UserError {
	if cfg.ImageUploadDir == "" {
		return &common.UserError{Err: fmt.Errorf("dir empty"), ErrorMessage: "image upload directory not set"}
	}
	if strings.HasPrefix(cfg.ImageUploadDir, "s3://") {
		if _, _, err := parseS3Dir(cfg.ImageUploadDir); err != nil {
			return &common.UserError{Err: err, ErrorMessage: "image upload directory invalid"}
		}
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
	return ensureSchemaWithQuerier(ctx, sqlDBQuerier{db: db})
}

func ensureSchemaWithQuerier(ctx context.Context, q execQuerier) error {
	version, err := ensureVersionTable(ctx, q)
	if err != nil {
		return err
	}
	if version != handlers.ExpectedSchemaVersion {
		msg := RenderSchemaMismatch(version, handlers.ExpectedSchemaVersion)
		return fmt.Errorf("%s", msg)
	}
	return nil
}
