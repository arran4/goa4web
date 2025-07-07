package runtimeconfig

import (
	"os"
	"path/filepath"

	"github.com/arran4/goa4web/config"
	common "github.com/arran4/goa4web/core/common"
)

const defaultSecretName = ".session_secret"

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
