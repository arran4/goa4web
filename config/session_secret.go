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

// LoadOrCreateSecret returns a secret using DefaultSessionSecretPath when no path
// is provided.
func LoadOrCreateSecret(fs core.FileSystem, cliSecret, path, envSecret, envSecretFile string) (string, error) {
	return secrets.LoadOrCreate(fs, cliSecret, path, envSecret, envSecretFile, DefaultSessionSecretPath)
}
