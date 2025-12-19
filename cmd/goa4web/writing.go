package main

import (
	"flag"
	"fmt"
)

// writingCmd handles writing management subcommands.
type writingCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseWritingCmd(parent *rootCmd, args []string) (*writingCmd, error) {
	c := &writingCmd{rootCmd: parent}
	c.fs = newFlagSet("writing")

	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *writingCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing writing command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "tree":
		cmd, err := parseWritingTreeCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("tree: %w", err)
		}
		return cmd.Run()
	case "list":
		cmd, err := parseWritingListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "read":
		cmd, err := parseWritingReadCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		return cmd.Run()
	case "comments":
		cmd, err := parseWritingCommentsCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("comments: %w", err)
		}
		return cmd.Run()
	case "deactivate":
		cmd, err := parseWritingDeactivateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("deactivate: %w", err)
		}
		return cmd.Run()
	case "activate":
		cmd, err := parseWritingActivateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("activate: %w", err)
		}
		return cmd.Run()
	case "list-deactivated":
		cmd, err := parseWritingListDeactivatedCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list-deactivated: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown writing command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *writingCmd) Usage() {
	executeUsage(c.fs.Output(), "writing_usage.txt", c)
}

func (c *writingCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*writingCmd)(nil)
