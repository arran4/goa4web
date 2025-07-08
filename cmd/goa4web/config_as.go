package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
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
	fs       *flag.FlagSet
	extended bool
	args     []string
}

func parseConfigAsCmd(parent *configCmd, name string, args []string) (*configAsCmd, error) {
	c := &configAsCmd{configCmd: parent}
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.BoolVar(&c.extended, "extended", false, "include extended usage")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func envMapFromConfig(cfg runtimeconfig.RuntimeConfig, cfgPath string) (map[string]string, error) {
	m := make(map[string]string)
	for _, o := range runtimeconfig.StringOptions {
		m[o.Env] = *o.Target(&cfg)
	}
	for _, o := range runtimeconfig.IntOptions {
		m[o.Env] = strconv.Itoa(*o.Target(&cfg))
	}
	for _, o := range runtimeconfig.BoolOptions {
		m[o.Env] = strconv.FormatBool(*o.Target(&cfg))
	}

	fileVals, err := config.LoadAppConfigFile(core.OSFS{}, cfgPath)
	if err != nil {
		return nil, fmt.Errorf("load config file: %w", err)
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
	m[config.EnvSessionSecret] = first("", os.Getenv(config.EnvSessionSecret))
	sessionFile := first(fileVals[config.EnvSessionSecretFile], os.Getenv(config.EnvSessionSecretFile))
	if sessionFile == "" {
		sessionFile = runtimeconfig.DefaultSessionSecretPath()
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
	ext := extendedUsageMap()
	for _, k := range keys {
		u := usage[k]
		d := def[k]
		if u != "" {
			fmt.Printf("# %s (default: %s)\n", u, d)
		} else {
			fmt.Printf("# default: %s\n", d)
		}
		if c.extended {
			if e := ext[k]; e != "" {
				for _, line := range strings.Split(strings.TrimSuffix(e, "\n"), "\n") {
					fmt.Printf("# %s\n", line)
				}
			}
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
	ext := extendedUsageMap()
	for _, k := range keys {
		u := usage[k]
		d := def[k]
		if u != "" {
			fmt.Printf("# %s (default: %s)\n", u, d)
		} else {
			fmt.Printf("# default: %s\n", d)
		}
		if c.extended {
			if e := ext[k]; e != "" {
				for _, line := range strings.Split(strings.TrimSuffix(e, "\n"), "\n") {
					fmt.Printf("# %s\n", line)
				}
			}
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

func extendedUsageMap() map[string]string {
	m := make(map[string]string)
	for _, o := range runtimeconfig.StringOptions {
		if o.ExtendedUsage != "" {
			if txt, err := runtimeconfig.ExtendedUsage(o.ExtendedUsage); err == nil {
				m[o.Env] = txt
			}
		}
	}
	for _, o := range runtimeconfig.IntOptions {
		if o.ExtendedUsage != "" {
			if txt, err := runtimeconfig.ExtendedUsage(o.ExtendedUsage); err == nil {
				m[o.Env] = txt
			}
		}
	}
	for _, o := range runtimeconfig.BoolOptions {
		if o.ExtendedUsage != "" {
			if txt, err := runtimeconfig.ExtendedUsage(o.ExtendedUsage); err == nil {
				m[o.Env] = txt
			}
		}
	}
	return m
}

func usageMap() map[string]string {
	m := make(map[string]string)
	for _, o := range runtimeconfig.StringOptions {
		m[o.Env] = o.Usage
	}
	for _, o := range runtimeconfig.IntOptions {
		m[o.Env] = o.Usage
	}
	for _, o := range runtimeconfig.BoolOptions {
		m[o.Env] = o.Usage
	}
	m[config.EnvStatsStartYear] = "start year for usage stats"
	m[config.EnvConfigFile] = "path to config file"
	m[config.EnvSessionSecret] = "session secret key"
	m[config.EnvSessionSecretFile] = "path to session secret file"
	m[config.EnvAdminEmails] = "administrator email addresses"
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
	for _, o := range runtimeconfig.BoolOptions {
		if o.Name != "" {
			m[o.Env] = o.Name
		}
	}
	m[config.EnvConfigFile] = "config-file"
	m[config.EnvSessionSecret] = "session-secret"
	m[config.EnvSessionSecretFile] = "session-secret-file"
	return m
}
