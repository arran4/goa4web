package main

import (
	"flag"
	"fmt"
)

// announcementCmd handles announcement-related subcommands.
type announcementCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseAnnouncementCmd(parent *rootCmd, args []string) (*announcementCmd, error) {
	c := &announcementCmd{rootCmd: parent}
	c.fs = newFlagSet("announcement")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *announcementCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing announcement command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "list":
		cmd, err := parseAnnouncementListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "add":
		cmd, err := parseAnnouncementAddCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("add: %w", err)
		}
		return cmd.Run()
	case "delete":
		cmd, err := parseAnnouncementDeleteCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("delete: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown announcement command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *announcementCmd) Usage() {
	executeUsage(c.fs.Output(), "announcement_usage.txt", c)
}

func (c *announcementCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*announcementCmd)(nil)
