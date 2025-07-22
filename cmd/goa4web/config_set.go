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
	args  []string
}

func parseConfigSetCmd(parent *configCmd, args []string) (*configSetCmd, error) {
	c := &configSetCmd{configCmd: parent}
	fs := flag.NewFlagSet("set", flag.ContinueOnError)
	fs.StringVar(&c.Key, "key", "", "configuration key")
	fs.StringVar(&c.Value, "value", "", "configuration value")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *configSetCmd) Run() error {
	if c.Key == "" {
		return fmt.Errorf("key required")
	}
	path := c.rootCmd.ConfigFile
	c.rootCmd.Verbosef("updating %s in %s", c.Key, path)
	if err := config.UpdateConfigKey(core.OSFS{}, path, c.Key, c.Value); err != nil {
		return fmt.Errorf("update config: %w", err)
	}
	c.rootCmd.Infof("updated %s", c.Key)
	return nil
}
