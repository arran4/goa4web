package main

import (
	"flag"
	"fmt"
)

// ipBanCmd implements IP ban management commands.
type ipBanCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseIpBanCmd(parent *rootCmd, args []string) (*ipBanCmd, error) {
	c := &ipBanCmd{rootCmd: parent}
	c.fs = newFlagSet("ipban")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *ipBanCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing ipban command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "add":
		cmd, err := parseIpBanAddCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("add: %w", err)
		}
		return cmd.Run()
	case "list":
		cmd, err := parseIpBanListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "delete":
		cmd, err := parseIpBanDeleteCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("delete: %w", err)
		}
		return cmd.Run()
	case "update":
		cmd, err := parseIpBanUpdateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown ipban command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *ipBanCmd) Usage() {
	executeUsage(c.fs.Output(), "ipban_usage.txt", c)
}

func (c *ipBanCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*ipBanCmd)(nil)
