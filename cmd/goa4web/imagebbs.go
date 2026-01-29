package main

import (
	"flag"
	"fmt"
)

// imagebbsCmd handles ImageBBS-related subcommands.
type imagebbsCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseImagebbsCmd(parent *rootCmd, args []string) (*imagebbsCmd, error) {
	c := &imagebbsCmd{rootCmd: parent}
	c.fs = newFlagSet("imagebbs")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *imagebbsCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing imagebbs command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "moderation":
		cmd, err := parseImagebbsModerationCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("moderation: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown imagebbs command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *imagebbsCmd) Usage() {
	executeUsage(c.fs.Output(), "imagebbs_usage.txt", c)
}

func (c *imagebbsCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*imagebbsCmd)(nil)
