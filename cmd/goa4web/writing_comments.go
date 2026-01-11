package main

import (
	"flag"
	"fmt"
)

// writingCommentsCmd handles "writing comments".
type writingCommentsCmd struct {
	*writingCmd
	fs *flag.FlagSet
}

func parseWritingCommentsCmd(parent *writingCmd, args []string) (*writingCommentsCmd, error) {
	c := &writingCommentsCmd{writingCmd: parent}
	c.fs = newFlagSet("comments")

	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *writingCommentsCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing comments command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "list":
		cmd, err := parseWritingCommentsListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "read":
		cmd, err := parseWritingCommentsReadCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown comments command %q", args[0])
	}
}

func (c *writingCommentsCmd) Usage() {
	executeUsage(c.fs.Output(), "writing_comments_usage.txt", c)
}

func (c *writingCommentsCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*writingCommentsCmd)(nil)
