package config

import (
	"github.com/arran4/goa4web"
	"os"
	"path/filepath"

	"github.com/arran4/goa4web/core"
)

// TODO: build a small library for repeating secret helpers used across packages.

const defaultAPISecretName = ".admin_api_secret"

// DefaultAdminAPISecretPath returns the default path for the admin API secret file based on the execution environment.
func DefaultAdminAPISecretPath() string {
	if goa4web.Version == "dev" {
		return defaultAPISecretName
	}
	if os.Getenv(EnvDocker) != "" {
		return "/var/lib/goa4web/admin_api_secret"
	}
	if os.Getenv("HOME") == "" && os.Getenv("XDG_CONFIG_HOME") == "" {
		return "/var/lib/goa4web/admin_api_secret"
	}
	dir, err := os.UserConfigDir()
	if err == nil && dir != "" {
		return filepath.Join(dir, "goa4web", "admin_api_secret")
	}
	return defaultAPISecretName
}

// LoadOrCreateAdminAPISecret works like LoadOrCreateSecret but defaults to DefaultAdminAPISecretPath.
func LoadOrCreateAdminAPISecret(fs core.FileSystem, cliSecret, path string) (string, error) {
	return LoadOrCreateSecret(fs, cliSecret, path, EnvAdminAPISecret, EnvAdminAPISecretFile)
}
