package main

import (
	"flag"
	"fmt"
)

// blogCommentsCmd handles "blog comments".
type blogCommentsCmd struct {
	*blogCmd
	fs *flag.FlagSet
}

func parseBlogCommentsCmd(parent *blogCmd, args []string) (*blogCommentsCmd, error) {
	c := &blogCommentsCmd{blogCmd: parent}
	c.fs = newFlagSet("comments")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *blogCommentsCmd) Run() error {
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
		cmd, err := parseBlogCommentsListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "read":
		cmd, err := parseBlogCommentsReadCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown comments command %q", args[0])
	}
}

func (c *blogCommentsCmd) Usage() {
	executeUsage(c.fs.Output(), "blog_comments_usage.txt", c)
}

func (c *blogCommentsCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*blogCommentsCmd)(nil)
