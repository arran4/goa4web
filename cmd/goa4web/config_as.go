package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/arran4/goa4web/config"
)

// configAsCmd implements "config as-*" commands.
type configAsCmd struct {
	*configCmd
	fs       *flag.FlagSet
	extended bool
}

func parseConfigAsCmd(parent *configCmd, name string, args []string) (*configAsCmd, error) {
	c := &configAsCmd{configCmd: parent}
	c.fs = newFlagSet(name)
	c.fs.BoolVar(&c.extended, "extended", false, "include extended usage")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func defaultMap() map[string]string {
	def := config.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" })
	m, _ := config.ToEnvMap(*def, "")
	return m
}

func (c *configAsCmd) asEnvFile() error {
	current, err := config.ToEnvMap(c.rootCmd.cfg, c.rootCmd.ConfigFile)
	if err != nil {
		return fmt.Errorf("env map: %w", err)
	}
	def := defaultMap()
	keys := make([]string, 0, len(current))
	for k := range current {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	usage := config.UsageMap()
	ext := config.ExtendedUsageMap(c.rootCmd.dbReg)
	ex := config.ExamplesMap()
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
	current, err := config.ToEnvMap(c.rootCmd.cfg, c.rootCmd.ConfigFile)
	if err != nil {
		return fmt.Errorf("env map: %w", err)
	}
	def := defaultMap()
	keys := make([]string, 0, len(current))
	for k := range current {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	usage := config.UsageMap()
	ext := config.ExtendedUsageMap(c.rootCmd.dbReg)
	ex := config.ExamplesMap()
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
	m, err := config.ToEnvMap(c.rootCmd.cfg, c.rootCmd.ConfigFile)
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
	current, err := config.ToEnvMap(c.rootCmd.cfg, c.rootCmd.ConfigFile)
	if err != nil {
		return fmt.Errorf("env map: %w", err)
	}
	def := defaultMap()
	var parts []string
	nameMap := config.NameMap()
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
