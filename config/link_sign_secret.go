package config

import (
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/secrets"
)

// linkSignSecretName is the filename used for storing the external link signing key.
const linkSignSecretName = "link_sign_secret"

// DefaultLinkSignSecretPath returns the default path for the external link signing key based on the execution environment.
func DefaultLinkSignSecretPath() string {
	return secrets.DefaultPath(linkSignSecretName, EnvDocker)
}

// LoadOrCreateLinkSignSecret works like LoadOrCreateSecret but defaults to DefaultLinkSignSecretPath.
func LoadOrCreateLinkSignSecret(fs core.FileSystem, cliSecret, path string) (string, error) {
	return secrets.LoadOrCreate(fs, cliSecret, path, EnvLinkSignSecret, EnvLinkSignSecretFile, DefaultLinkSignSecretPath)
}
