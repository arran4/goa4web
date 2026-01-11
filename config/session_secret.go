package config

import (
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/secrets"
)

// sessionSecretName is the filename used for storing the session secret.
const sessionSecretName = "session_secret"

// DefaultSessionSecretPath returns the default path for the session secret file
// based on the execution environment.
func DefaultSessionSecretPath() string {
	return secrets.DefaultPath(sessionSecretName, EnvDocker)
}

// LoadOrCreateSessionSecret returns a secret using DefaultSessionSecretPath when no path
// is provided.
func LoadOrCreateSessionSecret(fs core.FileSystem, cliSecret, path string) (string, error) {
	return secrets.LoadOrCreate(fs, cliSecret, path, EnvSessionSecret, EnvSessionSecretFile, DefaultSessionSecretPath)
}
