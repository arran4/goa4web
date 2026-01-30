//go:build !windows

package config

import (
	"os"
	"path/filepath"

	"github.com/arran4/goa4web"
)

func defaultDataDir() string {
	// Check for Docker environment or running as root (likely a system service)
	if os.Getenv(EnvDocker) != "" || os.Geteuid() == 0 {
		return "/var/lib/goa4web"
	}
	if goa4web.Version == "dev" {
		return ".data"
	}
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "goa4web")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "goa4web")
}

func defaultCacheDir() string {
	// Check for Docker environment or running as root (likely a system service)
	if os.Getenv(EnvDocker) != "" || os.Geteuid() == 0 {
		return "/var/cache/goa4web/thumbnails"
	}
	if goa4web.Version == "dev" {
		return ".data/cache"
	}
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "goa4web")
	}
	if dir, err := os.UserCacheDir(); err == nil {
		return filepath.Join(dir, "goa4web")
	}
	return ".data/cache"
}
