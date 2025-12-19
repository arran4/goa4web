package main

import (
	"flag"
	"fmt"
	"os"
)

type roleCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseRoleCmd(parent *rootCmd, args []string) (*roleCmd, error) {
	c := &roleCmd{rootCmd: parent}
	fs := flag.NewFlagSet("role", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *roleCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing role command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "load":
		cmd, err := parseRoleLoadCmd(c, args[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	case "reset":
		cmd, err := parseRoleResetCmd(c, args[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	case "apply":
		cmd, err := parseRoleApplyCmd(c, args[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	case "remove":
		cmd, err := parseRoleRemoveCmd(c, args[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	case "users":
		cmd, err := parseRoleUsersCmd(c, args[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	case "list":
		cmd, err := parseRoleListCmd(c, args[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	case "inspect":
		cmd, err := parseRoleInspectCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("inspect: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown role command %q", args[0])
	}
}

func (c *roleCmd) Usage() {
	executeUsage(os.Stdout, "role_usage.txt", c)
}

func (c *roleCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*roleCmd)(nil)
