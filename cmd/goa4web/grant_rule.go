package main

import (
	"flag"
	"fmt"
)

// grantCmd implements "grant" top-level command.
type grantCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseGrantCmd(parent *rootCmd, args []string) (*grantCmd, error) {
	c := &grantCmd{rootCmd: parent}
	c.fs = newFlagSet("grant")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *grantCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing grant command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "add":
		cmd, err := parseGrantAddCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("add: %w", err)
		}
		return cmd.Run()
	case "list":
		cmd, err := parseGrantListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "delete":
		cmd, err := parseGrantDeleteCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("delete: %w", err)
		}
		return cmd.Run()
	case "list-available":
		cmd := newGrantListAvailableCmd()
		if err := cmd.Init(args[1:]); err != nil {
			return fmt.Errorf("list-available: %w", err)
		}
		return cmd.Run()
	case "export":
		cmd := newGrantExportCmd()
		if err := cmd.Init(args[1:]); err != nil {
			return fmt.Errorf("export: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown grant command %q", args[0])
	}
}

func (c *grantCmd) Usage() {
	executeUsage(c.fs.Output(), "grant_usage.txt", c)
}

func (c *grantCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*grantCmd)(nil)
