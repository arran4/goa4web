package config

import (
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/secrets"
)

// imageSignSecretName is the filename used for storing the image signing key.
const imageSignSecretName = "image_sign_secret"

// DefaultImageSignSecretPath returns the default path for the image signing key based on the execution environment.
func DefaultImageSignSecretPath() string {
	return secrets.DefaultPath(imageSignSecretName, EnvDocker)
}

// LoadOrCreateImageSignSecret works like LoadOrCreateSecret but defaults to DefaultImageSignSecretPath.
func LoadOrCreateImageSignSecret(fs core.FileSystem, cliSecret, path string) (string, error) {
	return secrets.LoadOrCreate(fs, cliSecret, path, EnvImageSignSecret, EnvImageSignSecretFile, DefaultImageSignSecretPath)
}
