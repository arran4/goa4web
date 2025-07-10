package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"sort"
	"strings"

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

func defaultMap() map[string]string {
	def := runtimeconfig.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" })
	m, _ := runtimeconfig.ToEnvMap(def, "")
	return m
}

func (c *configAsCmd) asEnvFile() error {
	current, err := runtimeconfig.ToEnvMap(c.rootCmd.cfg, c.rootCmd.ConfigFile)
	if err != nil {
		return fmt.Errorf("env map: %w", err)
	}
	def := defaultMap()
	keys := make([]string, 0, len(current))
	for k := range current {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	usage := runtimeconfig.UsageMap()
	ext := runtimeconfig.ExtendedUsageMap()
	ex := runtimeconfig.ExamplesMap()
	for _, k := range keys {
		u := usage[k]
		d := def[k]
		line := "# "
		if u != "" {
			line += fmt.Sprintf("%s (default: %s)", u, d)
		} else {
			line += fmt.Sprintf("default: %s", d)
		}
		if xs := ex[k]; len(xs) > 0 {
			line += fmt.Sprintf(" (examples: %s)", strings.Join(xs, ", "))
		}
		fmt.Println(line)
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
	current, err := runtimeconfig.ToEnvMap(c.rootCmd.cfg, c.rootCmd.ConfigFile)
	if err != nil {
		return fmt.Errorf("env map: %w", err)
	}
	def := defaultMap()
	keys := make([]string, 0, len(current))
	for k := range current {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	usage := runtimeconfig.UsageMap()
	ext := runtimeconfig.ExtendedUsageMap()
	ex := runtimeconfig.ExamplesMap()
	for _, k := range keys {
		u := usage[k]
		d := def[k]
		line := "# "
		if u != "" {
			line += fmt.Sprintf("%s (default: %s)", u, d)
		} else {
			line += fmt.Sprintf("default: %s", d)
		}
		if xs := ex[k]; len(xs) > 0 {
			line += fmt.Sprintf(" (examples: %s)", strings.Join(xs, ", "))
		}
		fmt.Println(line)
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
	m, err := runtimeconfig.ToEnvMap(c.rootCmd.cfg, c.rootCmd.ConfigFile)
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
	current, err := runtimeconfig.ToEnvMap(c.rootCmd.cfg, c.rootCmd.ConfigFile)
	if err != nil {
		return fmt.Errorf("env map: %w", err)
	}
	def := defaultMap()
	var parts []string
	nameMap := runtimeconfig.NameMap()
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
