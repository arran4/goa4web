package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/config"
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

func envMapFromConfig(cfg runtimeconfig.RuntimeConfig) map[string]string {
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
	return m
}

func defaultMap() map[string]string {
	def := runtimeconfig.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" })
	return envMapFromConfig(def)
}

func (c *configAsCmd) asEnv() error {
	current := envMapFromConfig(c.rootCmd.cfg)
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

func (c *configAsCmd) asJSON() error {
	m := envMapFromConfig(c.rootCmd.cfg)
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	fmt.Println(string(b))
	return nil
}

func (c *configAsCmd) asCLI() error {
	current := envMapFromConfig(c.rootCmd.cfg)
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
	return m
}
