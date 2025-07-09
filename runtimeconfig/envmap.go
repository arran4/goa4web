package runtimeconfig

import (
	"fmt"
	"os"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
)

// ToEnvMap converts cfg into a map keyed by environment variable name.
// The cfgPath argument sets the CONFIG_FILE entry and is used to
// resolve SESSION_SECRET_FILE when empty.
func ToEnvMap(cfg RuntimeConfig, cfgPath string) (map[string]string, error) {
	m := ValuesMap(cfg)

	fileVals, err := config.LoadAppConfigFile(core.OSFS{}, cfgPath)
	if err != nil {
		return nil, fmt.Errorf("load config file: %w", err)
	}

	m[config.EnvConfigFile] = cfgPath
	m[config.EnvSessionSecret] = os.Getenv(config.EnvSessionSecret)
	sessionFile := fileVals[config.EnvSessionSecretFile]
	if sessionFile == "" {
		sessionFile = os.Getenv(config.EnvSessionSecretFile)
	}
	if sessionFile == "" {
		sessionFile = DefaultSessionSecretPath()
	}
	m[config.EnvSessionSecretFile] = sessionFile

	return m, nil
}
