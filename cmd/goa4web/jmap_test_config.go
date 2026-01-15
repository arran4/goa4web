package main

import (
	"flag"
)

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
	ep, acc, id, _, err := discoverJmapLogic(c.jmapCmd)
	if err != nil {
		return err
	}
	c.Infof("JMAP Provider Configured:\nEndpoint: %s\nUser: %s\nAccountID: %s\nIdentityID: %s\n", ep, c.cfg.EmailJMAPUser, acc, id)
	return nil
}

func (c *jmapTestConfigCmd) Usage() {
	executeUsage(c.fs.Output(), "jmap_test_config_usage.txt", c)
}
