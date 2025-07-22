package main

import (
	"flag"
	"fmt"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
)

// configJSONAddCmd implements "config add-json".
type configJSONAddCmd struct {
	*configCmd
	fs   *flag.FlagSet
	File string
}

func parseConfigJSONAddCmd(parent *configCmd, args []string) (*configJSONAddCmd, error) {
	c := &configJSONAddCmd{configCmd: parent}
	c.fs = newFlagSet("add-json")
	c.fs.StringVar(&c.File, "file", "", "json file to update")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *configJSONAddCmd) Run() error {
	if c.File == "" {
		return fmt.Errorf("file required")
	}
	values, err := config.ToEnvMap(c.rootCmd.cfg, c.rootCmd.ConfigFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if err := config.AddMissingJSONOptions(core.OSFS{}, c.File, values); err != nil {
		return fmt.Errorf("update json: %w", err)
	}
	if c.rootCmd.Verbosity > 0 {
		fmt.Printf("updated %s\n", c.File)
	}
	return nil
}
