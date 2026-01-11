package config

import (
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/secrets"
)

// shareSignSecretName is the filename used for storing the share signing key.
const shareSignSecretName = "share_sign_secret"

// DefaultShareSignSecretPath returns the default path for the share signing key based on the execution environment.
func DefaultShareSignSecretPath() string {
	return secrets.DefaultPath(shareSignSecretName, EnvDocker)
}

// LoadOrCreateShareSignSecret retrieves the share signing secret from the
// environment or a file, generating a new secret if needed.
func LoadOrCreateShareSignSecret(fs core.FileSystem, val, path string) (string, error) {
	return secrets.LoadOrCreate(fs, val, path, EnvShareSignSecret, EnvShareSignSecretFile, DefaultShareSignSecretPath)
}
