package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	admin "github.com/arran4/goa4web/handlers/admin"
)

// configReloadCmd implements "config reload".
type configReloadCmd struct {
	*configCmd
	fs *flag.FlagSet
}

func parseConfigReloadCmd(parent *configCmd, args []string) (*configReloadCmd, error) {
	c := &configReloadCmd{configCmd: parent}
	c.fs = newFlagSet("reload")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *configReloadCmd) Run() error {
	cfgMap, err := config.LoadAppConfigFile(core.OSFS{}, c.rootCmd.ConfigFile)
	if err != nil && !errors.Is(err, config.ErrConfigFileNotFound) {
		return fmt.Errorf("load config file: %w", err)
	}
	c.rootCmd.Verbosef("reloading configuration")
	admin.Srv.Config = *config.GenerateRuntimeConfig(nil, cfgMap, os.Getenv)
	c.rootCmd.Infof("configuration reloaded")
	return nil
}
