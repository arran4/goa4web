package main

import (
	_ "embed"
	"flag"
	"fmt"
)

// roleCmd implements "role" top-level command.
type roleCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parseRoleCmd(parent *rootCmd, args []string) (*roleCmd, error) {
	c := &roleCmd{rootCmd: parent}
	fs := flag.NewFlagSet("role", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *roleCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing role command")
	}
	if err := usageIfHelp(c.fs, c.args); err != nil {
		return err
	}
	switch c.args[0] {
	case "users":
		cmd, err := parseRoleUsersCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("users: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown role command %q", c.args[0])
	}
}

func (c *roleCmd) Usage() {
	executeUsage(c.fs.Output(), "role_usage.txt", c)
}

func (c *roleCmd) FlagGroups() []flagGroup {
	return append(c.rootCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*roleCmd)(nil)
