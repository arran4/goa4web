package main

import (
	"flag"
	"fmt"
)

// newsCommentsCmd handles "news comments".
type newsCommentsCmd struct {
	*newsCmd
	fs *flag.FlagSet
}

func parseNewsCommentsCmd(parent *newsCmd, args []string) (*newsCommentsCmd, error) {
	c := &newsCommentsCmd{newsCmd: parent}
	c.fs = newFlagSet("comments")

	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *newsCommentsCmd) Run() error {
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
		cmd, err := parseNewsCommentsListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "read":
		cmd, err := parseNewsCommentsReadCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown comments command %q", args[0])
	}
}

func (c *newsCommentsCmd) Usage() {
	executeUsage(c.fs.Output(), "news_comments_usage.txt", c)
}

func (c *newsCommentsCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*newsCommentsCmd)(nil)
