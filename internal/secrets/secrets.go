package secrets

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/arran4/goa4web"
	"github.com/arran4/goa4web/core"
)

// DefaultPath returns the default path for a secret file named name.
// dockerEnv specifies the environment variable used to detect Docker builds.
func DefaultPath(name, dockerEnv string) string {
	devName := "." + name
	if goa4web.Version == "dev" {
		return devName
	}
	if os.Getenv(dockerEnv) != "" {
		return filepath.Join("/var/lib/goa4web", name)
	}
	if os.Getenv("HOME") == "" && os.Getenv("XDG_CONFIG_HOME") == "" {
		return filepath.Join("/var/lib/goa4web", name)
	}
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, "goa4web", name)
	}
	return devName
}

// LoadOrCreate returns a secret using the following priority:
//  1. cliSecret if non-empty
//  2. the environment variable named envSecret
//  3. contents of the file at path. If path is empty it uses envSecretFile
//     or defaultPath.
//
// If the file does not exist, a new random secret is generated and saved.
func LoadOrCreate(fs core.FileSystem, cliSecret, path, envSecret, envSecretFile string, defaultPath func() string) (string, error) {
	if cliSecret != "" {
		return cliSecret, nil
	}
	if env := os.Getenv(envSecret); env != "" {
		return env, nil
	}
	if path == "" {
		path = os.Getenv(envSecretFile)
		if path == "" {
			path = defaultPath()
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
