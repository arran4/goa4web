package config

import (
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/secrets"
)

// LoadOrCreateShareSignSecret retrieves the share signing secret from the
// environment or a file, generating a new secret if needed.
func LoadOrCreateShareSignSecret(fs core.FileSystem, val, path string) (string, error) {
	return secrets.LoadOrCreate(fs, val, path, EnvShareSignSecret, EnvShareSignSecretFile, func() string {
		return secrets.DefaultPath("share_sign_secret", EnvDocker)
	})
}
