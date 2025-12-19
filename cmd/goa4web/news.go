package main

import (
	"flag"
	"fmt"
)

// newsCmd handles news management subcommands.
type newsCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseNewsCmd(parent *rootCmd, args []string) (*newsCmd, error) {
	c := &newsCmd{rootCmd: parent}
	c.fs = newFlagSet("news")

	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *newsCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing news command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "list":
		cmd, err := parseNewsListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "read":
		cmd, err := parseNewsReadCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		return cmd.Run()
	case "comments":
		cmd, err := parseNewsCommentsCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("comments: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown news command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *newsCmd) Usage() {
	executeUsage(c.fs.Output(), "news_usage.txt", c)
}

func (c *newsCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*newsCmd)(nil)
