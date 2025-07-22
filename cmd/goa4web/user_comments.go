package main

import (
	"flag"
	"fmt"
)

// userCommentsCmd handles "user comments".
type userCommentsCmd struct {
	*userCmd
	fs *flag.FlagSet
}

func parseUserCommentsCmd(parent *userCmd, args []string) (*userCommentsCmd, error) {
	c := &userCommentsCmd{userCmd: parent}
	c.fs = newFlagSet("comments")

	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *userCommentsCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing comments command")
	}
	switch args[0] {
	case "list":
		cmd, err := parseUserCommentsListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "add":
		cmd, err := parseUserCommentsAddCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("add: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown comments command %q", args[0])
	}
}

func (c *userCommentsCmd) Usage() {
	executeUsage(c.fs.Output(), templateString("user_comments_usage.txt"), c.fs, c.rootCmd.fs.Name())
}
