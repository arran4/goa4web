package main

import (
	"flag"
	"fmt"
)

// commentCmd handles comment management subcommands.
type commentCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseCommentCmd(parent *rootCmd, args []string) (*commentCmd, error) {
	c := &commentCmd{rootCmd: parent}
	c.fs = newFlagSet("comment")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *commentCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing comment command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "deactivate":
		cmd, err := parseCommentDeactivateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("deactivate: %w", err)
		}
		return cmd.Run()
	case "activate":
		cmd, err := parseCommentActivateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("activate: %w", err)
		}
		return cmd.Run()
	case "list-deactivated":
		cmd, err := parseCommentListDeactivatedCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list-deactivated: %w", err)
		}
		return cmd.Run()
	case "clean-bad":
		cmd, err := parseCommentCleanBadCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("clean-bad: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown comment command %q", args[0])
	}
}

func (c *commentCmd) Usage() {
	executeUsage(c.fs.Output(), "comment_usage.txt", c)
}

func (c *commentCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*commentCmd)(nil)
