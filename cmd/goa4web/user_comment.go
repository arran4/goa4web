package main

import (
	"flag"
	"fmt"
)

// userCommentCmd handles "user comment".
type userCommentCmd struct {
	*userCmd
	fs *flag.FlagSet
}

func parseUserCommentCmd(parent *userCmd, args []string) (*userCommentCmd, error) {
	c := &userCommentCmd{userCmd: parent}
	c.fs = newFlagSet("comment")

	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *userCommentCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing comment command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "add":
		cmd, err := parseUserCommentAddCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("add: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown comment command %q", args[0])
	}
}

func (c *userCommentCmd) Usage() {
	executeUsage(c.fs.Output(), "user_comment_usage.txt", c)
}

func (c *userCommentCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userCommentCmd)(nil)
