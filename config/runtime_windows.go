//go:build windows

package config

import (
	"os"
	"path/filepath"

	"github.com/arran4/goa4web"
)

func defaultDataDir() string {
	if os.Getenv(EnvDocker) != "" {
		return "/var/lib/goa4web"
	}
	if goa4web.Version == "dev" {
		return ".data"
	}
	// On Windows, use %PROGRAMDATA% for system-wide or %LOCALAPPDATA% for user
	if pd := os.Getenv("ProgramData"); pd != "" {
		return filepath.Join(pd, "goa4web")
	}
	if lad := os.Getenv("LOCALAPPDATA"); lad != "" {
		return filepath.Join(lad, "goa4web", "data")
	}
	return ".data"
}

func defaultCacheDir() string {
	if os.Getenv(EnvDocker) != "" {
		return "/var/cache/goa4web/thumbnails"
	}
	if goa4web.Version == "dev" {
		return ".data/cache"
	}
	if lad := os.Getenv("LOCALAPPDATA"); lad != "" {
		return filepath.Join(lad, "goa4web", "cache")
	}
	return ".data/cache"
}
