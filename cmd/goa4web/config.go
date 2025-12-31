package main

import (
	"flag"
	"fmt"
)

// configCmd handles configuration utilities.
type configCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseConfigCmd(parent *rootCmd, args []string) (*configCmd, error) {
	c := &configCmd{rootCmd: parent}
	c.fs = newFlagSet("config")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *configCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing config command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "reload":
		cmd, err := parseConfigReloadCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("reload: %w", err)
		}
		return cmd.Run()
	case "as-env":
		cmd, err := parseConfigAsCmd(c, "as-env", args[1:])
		if err != nil {
			return fmt.Errorf("as-env: %w", err)
		}
		return cmd.asEnv()
	case "as-env-file":
		cmd, err := parseConfigAsCmd(c, "as-env-file", args[1:])
		if err != nil {
			return fmt.Errorf("as-env-file: %w", err)
		}
		return cmd.asEnvFile()
	case "as-json":
		cmd, err := parseConfigAsCmd(c, "as-json", args[1:])
		if err != nil {
			return fmt.Errorf("as-json: %w", err)
		}
		return cmd.asJSON()
	case "as-cli":
		cmd, err := parseConfigAsCmd(c, "as-cli", args[1:])
		if err != nil {
			return fmt.Errorf("as-cli: %w", err)
		}
		return cmd.asCLI()
	case "add-json":
		cmd, err := parseConfigJSONAddCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("add-json: %w", err)
		}
		return cmd.Run()
	case "options":
		cmd, err := parseConfigOptionsCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("options: %w", err)
		}
		return cmd.Run()
	case "test":
		cmd, err := parseConfigTestCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("test: %w", err)
		}
		return cmd.Run()
	case "show":
		cmd, err := parseConfigShowCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("show: %w", err)
		}
		return cmd.Run()
	case "set":
		cmd, err := parseConfigSetCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("set: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown config command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *configCmd) Usage() {
	executeUsage(c.fs.Output(), "config_usage.txt", c)
}

func (c *configCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*configCmd)(nil)
