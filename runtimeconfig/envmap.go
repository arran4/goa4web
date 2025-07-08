package runtimeconfig

import (
	"fmt"
	"os"
	"strconv"

	"github.com/arran4/goa4web/config"
)

// ToEnvMap converts cfg into a map keyed by environment variable name.
// The cfgPath argument sets the CONFIG_FILE entry and is used to
// resolve SESSION_SECRET_FILE when empty.
func ToEnvMap(cfg RuntimeConfig, cfgPath string) (map[string]string, error) {
	m := make(map[string]string)

	for _, o := range StringOptions {
		m[o.Env] = *o.Target(&cfg)
	}
	for _, o := range IntOptions {
		m[o.Env] = strconv.Itoa(*o.Target(&cfg))
	}
	for _, o := range BoolOptions {
		m[o.Env] = strconv.FormatBool(*o.Target(&cfg))
	}

	fileVals, err := config.LoadAppConfigFile(config.OSFS{}, cfgPath)
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
