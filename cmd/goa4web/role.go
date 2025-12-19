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
	case "load":
		cmd, err := parseRoleLoadCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("load: %w", err)
		}
		return cmd.Run()
	case "reset":
		cmd, err := parseRoleResetCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("reset: %w", err)
		}
		return cmd.Run()
	case "apply":
		cmd, err := parseRoleApplyCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("apply: %w", err)
		}
		return cmd.Run()
	case "remove":
		cmd, err := parseRoleRemoveCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("remove: %w", err)
		}
		return cmd.Run()
	case "users":
		cmd, err := parseRoleUsersCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("users: %w", err)
		}
		return cmd.Run()
	case "list":
		cmd, err := parseRoleListCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "inspect":
		cmd, err := parseRoleInspectCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("inspect: %w", err)
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
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*roleCmd)(nil)
