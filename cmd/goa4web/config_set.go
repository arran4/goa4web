package main

import (
	"flag"
	"fmt"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
)

// configSetCmd implements "config set".
type configSetCmd struct {
	*configCmd
	fs    *flag.FlagSet
	Key   string
	Value string
}

func parseConfigSetCmd(parent *configCmd, args []string) (*configSetCmd, error) {
	c := &configSetCmd{configCmd: parent}
	c.fs = newFlagSet("set")
	c.fs.StringVar(&c.Key, "key", "", "configuration key")
	c.fs.StringVar(&c.Value, "value", "", "configuration value")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *configSetCmd) Run() error {
	if c.Key == "" {
		return fmt.Errorf("key required")
	}
	path := c.rootCmd.ConfigFile
	if err := config.UpdateConfigKey(core.OSFS{}, path, c.Key, c.Value); err != nil {
		return fmt.Errorf("update config: %w", err)
	}
	if c.rootCmd.Verbosity > 0 {
		fmt.Printf("updated %s\n", c.Key)
	}
	return nil
}
