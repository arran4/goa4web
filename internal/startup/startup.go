package startup

import (
	"context"
	"fmt"
	"os"

	"github.com/arran4/goa4web/config"
	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/dbstart"
	"github.com/arran4/goa4web/internal/upload"
)

// PerformChecks checks DB connectivity and the upload provider.
func PerformChecks(cfg config.RuntimeConfig) error {
	if err := dbstart.MaybeAutoMigrate(cfg); err != nil {
		return err
	}
	if ue := dbstart.InitDB(cfg); ue != nil {
		return fmt.Errorf("%s: %w", ue.ErrorMessage, ue.Err)
	}
	if ue := CheckUploadTarget(cfg); ue != nil {
		return fmt.Errorf("%s: %w", ue.ErrorMessage, ue.Err)
	}
	return nil
}

// CheckUploadTarget verifies that the configured upload backend is available.
func CheckUploadTarget(cfg config.RuntimeConfig) *common.UserError {
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
