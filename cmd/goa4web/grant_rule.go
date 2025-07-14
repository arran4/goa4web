package main

import (
	_ "embed"
	"flag"
	"fmt"
)

//go:embed templates/grant_usage.txt
var grantUsageTemplate string

// grantCmd implements "grant" top-level command.
type grantCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parseGrantCmd(parent *rootCmd, args []string) (*grantCmd, error) {
	c := &grantCmd{rootCmd: parent}
	fs := flag.NewFlagSet("grant", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *grantCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing grant command")
	}
	switch c.args[0] {
	case "add":
		cmd, err := parseGrantAddCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("add: %w", err)
		}
		return cmd.Run()
	case "list":
		cmd, err := parseGrantListCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "delete":
		cmd, err := parseGrantDeleteCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("delete: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown grant command %q", c.args[0])
	}
}

func (c *grantCmd) Usage() {
	executeUsage(c.fs.Output(), grantUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
