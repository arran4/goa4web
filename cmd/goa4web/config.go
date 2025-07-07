package main

import (
	_ "embed"
	"flag"
	"fmt"
)

//go:embed templates/config_usage.txt
var configUsageTemplate string

// configCmd handles configuration utilities.
type configCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parseConfigCmd(parent *rootCmd, args []string) (*configCmd, error) {
	c := &configCmd{rootCmd: parent}
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *configCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing config command")
	}
	switch c.args[0] {
	case "reload":
		cmd, err := parseConfigReloadCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("reload: %w", err)
		}
		return cmd.Run()
	case "as-env":
		cmd, err := parseConfigAsCmd(c, "as-env", c.args[1:])
		if err != nil {
			return fmt.Errorf("as-env: %w", err)
		}
		return cmd.asEnv()
	case "as-env-file":
		cmd, err := parseConfigAsCmd(c, "as-env-file", c.args[1:])
		if err != nil {
			return fmt.Errorf("as-env-file: %w", err)
		}
		return cmd.asEnvFile()
	case "as-json":
		cmd, err := parseConfigAsCmd(c, "as-json", c.args[1:])
		if err != nil {
			return fmt.Errorf("as-json: %w", err)
		}
		return cmd.asJSON()
	case "as-cli":
		cmd, err := parseConfigAsCmd(c, "as-cli", c.args[1:])
		if err != nil {
			return fmt.Errorf("as-cli: %w", err)
		}
		return cmd.asCLI()
	case "add-json":
		cmd, err := parseConfigJSONAddCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("add-json: %w", err)
		}
		return cmd.Run()
	case "options":
		cmd, err := parseConfigOptionsCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("options: %w", err)
		}
		return cmd.Run()
	case "test":
		cmd, err := parseConfigTestCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("test: %w", err)
		}
		return cmd.Run()
	case "show":
		cmd, err := parseConfigShowCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("show: %w", err)
		}
		return cmd.Run()
	case "set":
		cmd, err := parseConfigSetCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("set: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown config command %q", c.args[0])
	}
}

// Usage prints command usage information with examples.
func (c *configCmd) Usage() {
	executeUsage(c.fs.Output(), configUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
