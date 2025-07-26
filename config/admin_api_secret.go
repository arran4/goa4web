package config

import (
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/secrets"
)

// adminAPISecretName is the filename used for storing the admin API secret.
const adminAPISecretName = "admin_api_secret"

// DefaultAdminAPISecretPath returns the default path for the admin API secret file based on the execution environment.
func DefaultAdminAPISecretPath() string {
	return secrets.DefaultPath(adminAPISecretName, EnvDocker)
}

// LoadOrCreateAdminAPISecret works like LoadOrCreateSecret but defaults to DefaultAdminAPISecretPath.
func LoadOrCreateAdminAPISecret(fs core.FileSystem, cliSecret, path string) (string, error) {
	return secrets.LoadOrCreate(fs, cliSecret, path, EnvAdminAPISecret, EnvAdminAPISecretFile, DefaultAdminAPISecretPath)
}
