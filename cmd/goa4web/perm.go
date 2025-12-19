package main

import (
	"flag"
	"fmt"
)

type permCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parsePermCmd(parent *rootCmd, args []string) (*permCmd, error) {
	c := &permCmd{rootCmd: parent}
	c.fs = newFlagSet("perm")

	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *permCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing perm command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "grant":
		cmd, err := parsePermGrantCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("grant: %w", err)
		}
		return cmd.Run()
	case "revoke":
		cmd, err := parsePermRevokeCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("revoke: %w", err)
		}
		return cmd.Run()
	case "update":
		cmd, err := parsePermUpdateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}
		return cmd.Run()
	case "list":
		cmd, err := parsePermListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown perm command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *permCmd) Usage() {
	executeUsage(c.fs.Output(), "perm_usage.txt", c)
}

func (c *permCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*permCmd)(nil)
