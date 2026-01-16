package main

import (
	"flag"
	"fmt"
)

// jmapCmd handles email-related subcommands.
type jmapCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseJmapCmd(parent *rootCmd, args []string) (*jmapCmd, error) {
	c := &jmapCmd{rootCmd: parent}
	c.fs = newFlagSet("jmap")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *jmapCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing jmap command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "test-config":
		return c.runTestConfig()
	case "test-send":
		return c.runTestSend()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown jmap command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *jmapCmd) Usage() {
	executeUsage(c.fs.Output(), "jmap_usage.txt", c)
}

func (c *jmapCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*jmapCmd)(nil)
