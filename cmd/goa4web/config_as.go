package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/runtimeconfig"
)

// configAsCmd implements "config as-*" commands.
type configAsCmd struct {
	*configCmd
	fs   *flag.FlagSet
	args []string
}

func parseConfigAsCmd(parent *configCmd, name string, args []string) (*configAsCmd, error) {
	c := &configAsCmd{configCmd: parent}
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func parseBool(v string) (bool, bool) {
	if v == "" {
		return false, false
	}
	switch strings.ToLower(v) {
	case "0", "false", "off", "no":
		return false, true
	case "1", "true", "on", "yes":
		return true, true
	default:
		return false, false
	}
}

func envMapFromConfig(cfg runtimeconfig.RuntimeConfig, cfgPath string) (map[string]string, error) {
	m := make(map[string]string)
	v := reflect.ValueOf(cfg)
	for _, o := range runtimeconfig.StringOptions {
		m[o.Env] = v.FieldByName(o.Field).String()
	}
	for _, o := range runtimeconfig.IntOptions {
		m[o.Env] = strconv.Itoa(int(v.FieldByName(o.Field).Int()))
	}
	m[config.EnvFeedsEnabled] = strconv.FormatBool(cfg.FeedsEnabled)
	m[config.EnvStatsStartYear] = strconv.Itoa(cfg.StatsStartYear)

	fileVals, err := config.LoadAppConfigFile(core.OSFS{}, cfgPath)
	if err != nil {
		return nil, fmt.Errorf("load config file: %w", err)
	}

	boolVal := func(file, env string, def bool) string {
		if b, ok := parseBool(file); ok {
			return strconv.FormatBool(b)
		}
		if b, ok := parseBool(env); ok {
			return strconv.FormatBool(b)
		}
		return strconv.FormatBool(def)
	}

	first := func(vals ...string) string {
		for _, v := range vals {
			if v != "" {
				return v
			}
		}
		return ""
	}

	m[config.EnvConfigFile] = cfgPath
	m[config.EnvEmailEnabled] = boolVal(fileVals[config.EnvEmailEnabled], os.Getenv(config.EnvEmailEnabled), true)
	m[config.EnvNotificationsEnabled] = boolVal(fileVals[config.EnvNotificationsEnabled], os.Getenv(config.EnvNotificationsEnabled), true)
	m[config.EnvCSRFEnabled] = boolVal(fileVals[config.EnvCSRFEnabled], os.Getenv(config.EnvCSRFEnabled), true)
	m[config.EnvAdminNotify] = boolVal(fileVals[config.EnvAdminNotify], os.Getenv(config.EnvAdminNotify), true)
	m[config.EnvAdminEmails] = first(fileVals[config.EnvAdminEmails], os.Getenv(config.EnvAdminEmails))
	m[config.EnvSessionSecret] = first("", os.Getenv(config.EnvSessionSecret))
	sessionFile := first(fileVals[config.EnvSessionSecretFile], os.Getenv(config.EnvSessionSecretFile))
	if sessionFile == "" {
		sessionFile = ".session_secret"
	}
	m[config.EnvSessionSecretFile] = sessionFile

	return m, nil
}

func defaultMap() map[string]string {
	def := runtimeconfig.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" })
	m, _ := envMapFromConfig(def, "")
	return m
}

func (c *configAsCmd) asEnvFile() error {
	current, err := envMapFromConfig(c.rootCmd.cfg, c.rootCmd.ConfigFile)
	if err != nil {
		return fmt.Errorf("env map: %w", err)
	}
	def := defaultMap()
	keys := make([]string, 0, len(current))
	for k := range current {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	usage := usageMap()
	for _, k := range keys {
		u := usage[k]
		d := def[k]
		if u != "" {
			fmt.Printf("# %s (default: %s)\n", u, d)
		} else {
			fmt.Printf("# default: %s\n", d)
		}
		fmt.Printf("%s=%s\n", k, current[k])
	}
	return nil
}

func (c *configAsCmd) asEnv() error {
	current, err := envMapFromConfig(c.rootCmd.cfg, c.rootCmd.ConfigFile)
	if err != nil {
		return fmt.Errorf("env map: %w", err)
	}
	def := defaultMap()
	keys := make([]string, 0, len(current))
	for k := range current {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	usage := usageMap()
	for _, k := range keys {
		u := usage[k]
		d := def[k]
		if u != "" {
			fmt.Printf("# %s (default: %s)\n", u, d)
		} else {
			fmt.Printf("# default: %s\n", d)
		}
		fmt.Printf("export %s=%s\n", k, current[k])
	}
	return nil
}

func (c *configAsCmd) asJSON() error {
	m, err := envMapFromConfig(c.rootCmd.cfg, c.rootCmd.ConfigFile)
	if err != nil {
		return fmt.Errorf("env map: %w", err)
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	fmt.Println(string(b))
	return nil
}

func (c *configAsCmd) asCLI() error {
	current, err := envMapFromConfig(c.rootCmd.cfg, c.rootCmd.ConfigFile)
	if err != nil {
		return fmt.Errorf("env map: %w", err)
	}
	def := defaultMap()
	var parts []string
	nameMap := nameMap()
	for env, val := range current {
		if def[env] == val {
			continue
		}
		n := nameMap[env]
		if n == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("--%s=%s", n, val))
	}
	sort.Strings(parts)
	fmt.Println(strings.Join(parts, " "))
	return nil
}

func usageMap() map[string]string {
	m := make(map[string]string)
	for _, o := range runtimeconfig.StringOptions {
		m[o.Env] = o.Usage
	}
	for _, o := range runtimeconfig.IntOptions {
		m[o.Env] = o.Usage
	}
	m[config.EnvFeedsEnabled] = "enable or disable feeds"
	m[config.EnvStatsStartYear] = "start year for usage stats"
	m[config.EnvConfigFile] = "path to config file"
	m[config.EnvEmailEnabled] = "enable sending queued emails"
	m[config.EnvNotificationsEnabled] = "enable internal notifications"
	m[config.EnvCSRFEnabled] = "enable or disable CSRF protection"
	m[config.EnvSessionSecret] = "session secret key"
	m[config.EnvSessionSecretFile] = "path to session secret file"
	m[config.EnvAdminEmails] = "administrator email addresses"
	m[config.EnvAdminNotify] = "enable admin notification emails"
	return m
}

func nameMap() map[string]string {
	m := make(map[string]string)
	for _, o := range runtimeconfig.StringOptions {
		m[o.Env] = o.Name
	}
	for _, o := range runtimeconfig.IntOptions {
		m[o.Env] = o.Name
	}
	// feeds-enabled and stats-start-year only used for CLI
	m[config.EnvFeedsEnabled] = "feeds-enabled"
	m[config.EnvStatsStartYear] = "stats-start-year"
	m[config.EnvConfigFile] = "config-file"
	m[config.EnvSessionSecret] = "session-secret"
	m[config.EnvSessionSecretFile] = "session-secret-file"
	return m
}
