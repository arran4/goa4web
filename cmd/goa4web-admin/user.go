package main

import (
	"flag"
	"fmt"
)

type userCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parseUserCmd(parent *rootCmd, args []string) (*userCmd, error) {
	fs := flag.NewFlagSet("user", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return &userCmd{rootCmd: parent, fs: fs, args: fs.Args()}, nil
}

func (c *userCmd) Run() error {
	if len(c.args) == 0 {
		return fmt.Errorf("missing user command")
	}
	switch c.args[0] {
	case "add":
		cmd, err := parseUserAddCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("add: %w", err)
		}
		return cmd.Run()
	case "add-admin":
		cmd, err := parseUserAddAdminCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("add-admin: %w", err)
		}
		return cmd.Run()
	case "make-admin":
		cmd, err := parseUserMakeAdminCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("make-admin: %w", err)
		}
		return cmd.Run()
	case "list":
		cmd, err := parseUserListCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	default:
		return fmt.Errorf("unknown user command %q", c.args[0])
	}
}
