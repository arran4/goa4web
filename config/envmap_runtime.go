package config

import (
	"fmt"
	"os"

	"github.com/arran4/goa4web/core"
)

// ToEnvMap converts cfg into a map keyed by environment variable name.
// The cfgPath argument sets the CONFIG_FILE entry and is used to
// resolve SESSION_SECRET_FILE when empty.
func ToEnvMap(cfg *RuntimeConfig, cfgPath string) (map[string]string, error) {
	m := ValuesMap(*cfg)

	fileVals, err := LoadAppConfigFile(core.OSFS{}, cfgPath)
	if err != nil {
		return nil, fmt.Errorf("load config file: %w", err)
	}

	m[EnvConfigFile] = cfgPath
	m[EnvSessionSecret] = os.Getenv(EnvSessionSecret)
	sessionFile := fileVals[EnvSessionSecretFile]
	if sessionFile == "" {
		sessionFile = os.Getenv(EnvSessionSecretFile)
	}
	if sessionFile == "" {
		sessionFile = DefaultSessionSecretPath()
	}
	m[EnvSessionSecretFile] = sessionFile

	return m, nil
}
