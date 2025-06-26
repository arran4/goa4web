package main

import (
	"flag"
	"fmt"
)

type permCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parsePermCmd(parent *rootCmd, args []string) (*permCmd, error) {
	fs := flag.NewFlagSet("perm", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return &permCmd{rootCmd: parent, fs: fs, args: fs.Args()}, nil
}

func (c *permCmd) Run() error {
	if len(c.args) == 0 {
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
		return fmt.Errorf("unknown perm command %q", c.args[0])
	}
}
