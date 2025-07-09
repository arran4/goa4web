package runtimeconfig

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	common "github.com/arran4/goa4web/core/common"
)

// defaultSecretName is used for local development when no other path is found.
const defaultSecretName = ".session_secret"

// FileSystem abstracts file operations for loading and storing the secret.
type FileSystem = core.FileSystem

// DefaultSessionSecretPath returns the default path for the session secret file
// based on the execution environment.
func DefaultSessionSecretPath() string {
	if common.Version == "dev" {
		return defaultSecretName
	}
	if os.Getenv(config.EnvDocker) != "" {
		return "/var/lib/goa4web/session_secret"
	}
	if os.Getenv("HOME") == "" && os.Getenv("XDG_CONFIG_HOME") == "" {
		return "/var/lib/goa4web/session_secret"
	}
	dir, err := os.UserConfigDir()
	if err == nil && dir != "" {
		return filepath.Join(dir, "goa4web", "session_secret")
	}
	return defaultSecretName
}

// LoadOrCreateSecret returns a secret using the following priority:
//  1. cliSecret if non-empty
//  2. the environment variable named envSecret
//  3. contents of the file at path. If path is empty it uses envSecretFile
//     or DefaultSessionSecretPath().
//
// If the file does not exist, a new random secret is generated and saved.
func LoadOrCreateSecret(fs FileSystem, cliSecret, path, envSecret, envSecretFile string) (string, error) {
	if cliSecret != "" {
		return cliSecret, nil
	}

	if env := os.Getenv(envSecret); env != "" {
		return env, nil
	}

	if path == "" {
		path = os.Getenv(envSecretFile)
		if path == "" {
			path = DefaultSessionSecretPath()
		}
	}

	b, err := fs.ReadFile(path)
	if err == nil {
		secret := strings.TrimSpace(string(b))
		if secret != "" {
			return secret, nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	secret := hex.EncodeToString(buf)
	if err := fs.WriteFile(path, []byte(secret), 0600); err != nil {
		return "", err
	}
	return secret, nil
}
