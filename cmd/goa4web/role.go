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
	if c.fs.NArg() == 0 {
		c.Usage()
		return fmt.Errorf("missing subcommand")
	}

	switch c.fs.Arg(0) {
	case "list":
		cmd, err := parseRoleListCmd(c, c.fs.Args()[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	case "apply":
		cmd, err := parseRoleApplyCmd(c, c.fs.Args()[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	case "load":
		cmd, err := parseRoleLoadCmd(c, c.fs.Args()[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	case "reset":
		cmd, err := parseRoleResetCmd(c, c.fs.Args()[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	case "template":
		cmd, err := parseRoleTemplateCmd(c, c.fs.Args()[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	case "remove":
		cmd, err := parseRoleRemoveCmd(c, c.fs.Args()[1:])
		if err != nil {
			return err
		}
		return cmd.Run()
	case "inspect":
		cmd, err := parseRoleInspectCmd(c, c.fs.Args()[1:])
		if err != nil {
			return fmt.Errorf("inspect: %w", err)
		}
		return cmd.Run()
	default:
		c.Usage()
		return fmt.Errorf("unknown subcommand: %s", c.fs.Arg(0))
	}
}

func (c *roleCmd) Usage() {
	executeUsage(os.Stdout, "role_usage.txt", c)
}

func (c *roleCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*roleCmd)(nil)
