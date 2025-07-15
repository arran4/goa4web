package main

import (
	_ "embed"
	"flag"
	"fmt"
)

//go:embed templates/user_comments_usage.txt
var userCommentsUsageTemplate string

// userCommentsCmd handles "user comments".
type userCommentsCmd struct {
	*userCmd
	fs   *flag.FlagSet
	args []string
}

func parseUserCommentsCmd(parent *userCmd, args []string) (*userCommentsCmd, error) {
	c := &userCommentsCmd{userCmd: parent}
	fs := flag.NewFlagSet("comments", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *userCommentsCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing comments command")
	}
	switch c.args[0] {
	case "list":
		cmd, err := parseUserCommentsListCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "add":
		cmd, err := parseUserCommentsAddCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("add: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown comments command %q", c.args[0])
	}
}

func (c *userCommentsCmd) Usage() {
	executeUsage(c.fs.Output(), userCommentsUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
