package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/arran4/goa4web/core"
	admin "github.com/arran4/goa4web/handlers/admin"
	"github.com/arran4/goa4web/runtimeconfig"
)

// configReloadCmd implements "config reload".
type configReloadCmd struct {
	*configCmd
	fs   *flag.FlagSet
	args []string
}

func parseConfigReloadCmd(parent *configCmd, args []string) (*configReloadCmd, error) {
	c := &configReloadCmd{configCmd: parent}
	fs := flag.NewFlagSet("reload", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *configReloadCmd) Run() error {
	cfgMap := admin.LoadAppConfigFile(core.OSFS{}, c.rootCmd.ConfigFile)
	admin.Srv.Config = runtimeconfig.GenerateRuntimeConfig(nil, cfgMap, os.Getenv)
	if c.rootCmd.Verbosity > 0 {
		fmt.Println("configuration reloaded")
	}
	return nil
}
