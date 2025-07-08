package core

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"

	common "github.com/arran4/goa4web/core/common"
)

// LoadSessionSecret returns the session secret using the following priority:
//  1. cliSecret if non-empty
//  2. the environment variable named envSecret
//  3. contents of the file at path. If path is empty it uses envSecretFile
//     or a default path determined at runtime.
//
// If the file does not exist, a new random secret is generated and saved.
func defaultSessionSecretPath() string {
	const envDocker = "GOA4WEB_DOCKER" // environment variable indicating docker mode
	if common.Version == "dev" {
		return ".session_secret"
	}
	if os.Getenv(envDocker) != "" {
		return "/var/lib/goa4web/session_secret"
	}
	if os.Getenv("HOME") == "" && os.Getenv("XDG_CONFIG_HOME") == "" {
		return "/var/lib/goa4web/session_secret"
	}
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, "goa4web", "session_secret")
	}
	return ".session_secret"
}

func LoadSessionSecret(fs FileSystem, cliSecret, path, envSecret, envSecretFile string) (string, error) {
	if cliSecret != "" {
		return cliSecret, nil
	}

	if env := os.Getenv(envSecret); env != "" {
		return env, nil
	}

	if path == "" {
		path = os.Getenv(envSecretFile)
		if path == "" {
			path = defaultSessionSecretPath()
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

	// Generate a new secret and store it
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
