package main

import (
	"flag"
	"fmt"
)

// filesCmd implements file management utilities.
type filesCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseFilesCmd(parent *rootCmd, args []string) (*filesCmd, error) {
	c := &filesCmd{rootCmd: parent}
	c.fs = newFlagSet("files")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *filesCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing files command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "list":
		cmd, err := parseFilesListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "purge":
		cmd, err := parseFilesPurgeCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("purge: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown files command %q", args[0])
	}
}

func (c *filesCmd) Usage() {
	executeUsage(c.fs.Output(), "files_usage.txt", c)
}

func (c *filesCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*filesCmd)(nil)
