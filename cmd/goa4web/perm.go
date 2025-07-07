package main

import (
	_ "embed"
	"flag"
	"fmt"
)

//go:embed templates/perm_usage.txt
var permUsageTemplate string

type permCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parsePermCmd(parent *rootCmd, args []string) (*permCmd, error) {
	c := &permCmd{rootCmd: parent}
	fs := flag.NewFlagSet("perm", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *permCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing perm command")
	}
	switch c.args[0] {
	case "grant":
		cmd, err := parsePermGrantCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("grant: %w", err)
		}
		return cmd.Run()
	case "revoke":
		cmd, err := parsePermRevokeCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("revoke: %w", err)
		}
		return cmd.Run()
	case "update":
		cmd, err := parsePermUpdateCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}
		return cmd.Run()
	case "list":
		cmd, err := parsePermListCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown perm command %q", c.args[0])
	}
}

// Usage prints command usage information with examples.
func (c *permCmd) Usage() {
	executeUsage(c.fs.Output(), permUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
