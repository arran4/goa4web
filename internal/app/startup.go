package app

import (
	"context"
	"database/sql"
	"fmt"
	dbstart2 "github.com/arran4/goa4web/internal/app/dbstart"
	"os"

	"github.com/arran4/goa4web/internal/dbdrivers"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/upload"
)

// PerformChecks checks DB connectivity and the upload provider.
func PerformChecks(cfg *config.RuntimeConfig, reg *dbdrivers.Registry) (*sql.DB, error) {
	if err := dbstart2.MaybeAutoMigrate(cfg, reg); err != nil {
		return nil, err
	}
	dbPool, ue := dbstart2.InitDB(cfg, reg)
	if ue != nil {
		return nil, fmt.Errorf("%s: %w", ue.ErrorMessage, ue.Err)
	}
	if ue := CheckUploadTarget(cfg); ue != nil {
		dbPool.Close()
		return nil, fmt.Errorf("%s: %w", ue.ErrorMessage, ue.Err)
	}
	return dbPool, nil
}

// CheckUploadTarget verifies that the configured upload backend is available.
func CheckUploadTarget(cfg *config.RuntimeConfig) *common.UserError {
	if cfg.ImageUploadDir == "" {
		return &common.UserError{Err: fmt.Errorf("dir empty"), ErrorMessage: "image upload directory not set"}
	}
	p := upload.ProviderFromConfig(cfg)
	if p == nil {
		return &common.UserError{Err: fmt.Errorf("no provider"), ErrorMessage: "image upload directory invalid"}
	}
	if err := p.Check(context.Background()); err != nil {
		return &common.UserError{Err: err, ErrorMessage: "image upload directory invalid"}
	}
	if cp := upload.CacheProviderFromConfig(cfg); cp != nil {
		if err := cp.Check(context.Background()); err != nil {
			return &common.UserError{Err: err, ErrorMessage: "image cache directory invalid"}
		}
	} else if cfg.ImageCacheDir != "" {
		if err := os.MkdirAll(cfg.ImageCacheDir, 0o755); err != nil {
			return &common.UserError{Err: err, ErrorMessage: "image cache directory invalid"}
		}
	}
	return nil
}
