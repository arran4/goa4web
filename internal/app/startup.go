package app

import (
	"context"
	"database/sql"
	"fmt"
	dbstart2 "github.com/arran4/goa4web/internal/app/dbstart"
	"os"
	"path/filepath"
	"strings"

	"github.com/arran4/goa4web/internal/db"
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
	if ue := CheckMediaFiles(cfg, dbPool); ue != nil {
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

// CheckMediaFiles verifies that the most recent media files exist on disk.
func CheckMediaFiles(cfg *config.RuntimeConfig, dbPool *sql.DB) *common.UserError {
	q := db.New(dbPool)
	ctx := context.Background()

	var missing []string

	// Check uploaded images
	if cfg.ImageUploadDir != "" {
		imgs, err := q.AdminListUploadedImages(ctx, db.AdminListUploadedImagesParams{
			Limit:  5,
			Offset: 0,
		})
		if err == nil {
			for _, img := range imgs {
				if !img.Path.Valid {
					continue
				}
				cleanPath := strings.TrimPrefix(img.Path.String, "/")
				cleanPath = strings.TrimPrefix(cleanPath, "uploads")
				cleanPath = strings.TrimPrefix(cleanPath, "/")
				p := filepath.Join(cfg.ImageUploadDir, cleanPath)
				if _, err := os.Stat(p); os.IsNotExist(err) {
					missing = append(missing, fmt.Sprintf("Uploaded image missing: %s", p))
				}
			}
		}
	}

	// Check cached images
	if cfg.ImageCacheDir != "" {
		links, err := q.AdminListExternalLinks(ctx, db.AdminListExternalLinksParams{
			Limit:  5,
			Offset: 0,
		})
		if err == nil {
			for _, link := range links {
				if link.CardImageCache.Valid {
					id := strings.TrimPrefix(link.CardImageCache.String, "cache:")
					if len(id) >= 4 {
						sub1, sub2 := id[:2], id[2:4]
						p := filepath.Join(cfg.ImageCacheDir, sub1, sub2, id)
						if _, err := os.Stat(p); os.IsNotExist(err) {
							missing = append(missing, fmt.Sprintf("Cached card image missing: %s", p))
						}
					}
				}
				if link.FaviconCache.Valid {
					id := strings.TrimPrefix(link.FaviconCache.String, "cache:")
					if len(id) >= 4 {
						sub1, sub2 := id[:2], id[2:4]
						p := filepath.Join(cfg.ImageCacheDir, sub1, sub2, id)
						if _, err := os.Stat(p); os.IsNotExist(err) {
							missing = append(missing, fmt.Sprintf("Cached favicon missing: %s", p))
						}
					}
				}
			}
		}
	}

	if len(missing) > 0 {
		msg := fmt.Sprintf("Found %d missing media files (checking recent 5 uploaded and cached):\n%s\n\n", len(missing), strings.Join(missing, "\n"))
		msg += fmt.Sprintf("Configured Image Upload Dir: %s\n", cfg.ImageUploadDir)
		msg += fmt.Sprintf("Configured Image Cache Dir: %s\n", cfg.ImageCacheDir)
		msg += fmt.Sprintf("Default Image Upload Dir: %s\n", filepath.Join(config.DefaultDataDir(), "images"))
		msg += fmt.Sprintf("Default Image Cache Dir: %s\n", config.DefaultCacheDir())
		msg += "Please check if your configuration points to the correct directory."

		return &common.UserError{
			Err:          fmt.Errorf("missing media files"),
			ErrorMessage: msg,
		}
	}

	return nil
}
