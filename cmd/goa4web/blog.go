package main

import (
	"flag"
	"fmt"
)

// blogCmd handles blog management subcommands.
type blogCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseBlogCmd(parent *rootCmd, args []string) (*blogCmd, error) {
	c := &blogCmd{rootCmd: parent}
	c.fs = newFlagSet("blog")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *blogCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing blog command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "create":
		cmd, err := parseBlogCreateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		return cmd.Run()
	case "list":
		cmd, err := parseBlogListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "read":
		cmd, err := parseBlogReadCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		return cmd.Run()
	case "comments":
		cmd, err := parseBlogCommentsCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("comments: %w", err)
		}
		return cmd.Run()
	case "update":
		cmd, err := parseBlogUpdateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}
		return cmd.Run()
	case "deactivate":
		cmd, err := parseBlogDeactivateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("deactivate: %w", err)
		}
		return cmd.Run()
	case "activate":
		cmd, err := parseBlogActivateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("activate: %w", err)
		}
		return cmd.Run()
	case "list-deactivated":
		cmd, err := parseBlogListDeactivatedCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list-deactivated: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown blog command %q", args[0])
	}
}

func (c *blogCmd) Usage() {
	executeUsage(c.fs.Output(), "blog_usage.txt", c)
}

func (c *blogCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*blogCmd)(nil)
