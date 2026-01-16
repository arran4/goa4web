package main

import (
	"flag"
)

// jmapTestConfigCmd handles the 'jmap test-config' subcommand.
type jmapTestConfigCmd struct {
	*jmapCmd
	fs *flag.FlagSet
}

func parseJmapTestConfigCmd(parent *jmapCmd, args []string) (*jmapTestConfigCmd, error) {
	c := &jmapTestConfigCmd{jmapCmd: parent}
	c.fs = newFlagSet("test-config")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *jmapTestConfigCmd) Run() error {
	info, err := c.discoverJmapSession()
	if err != nil {
		return err
	}
	return c.printSessionInfo(info)
}

// Usage prints command usage information.
func (c *jmapTestConfigCmd) Usage() {
	executeUsage(c.fs.Output(), "jmap_test_config_usage.txt", c)
}

func (c *jmapTestConfigCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*jmapTestConfigCmd)(nil)
