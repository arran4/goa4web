package main

import (
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/configformat"
)

// configAsCmd implements "config as-*" commands.
type configAsCmd struct {
	*configCmd
	fs       *flag.FlagSet
	extended bool
}

func parseConfigAsCmd(parent *configCmd, name string, args []string) (*configAsCmd, error) {
	c := &configAsCmd{configCmd: parent}
	fs := newFlagSet(name)
	opts, err := configformat.ParseAsFlags(fs, args)
	if err != nil {
		return nil, err
	}
	c.fs = fs
	c.extended = opts.Extended

	return c, nil
}

func (c *configAsCmd) asEnvFile() error {
	out, err := configformat.FormatAsEnvFile(c.rootCmd.cfg, c.rootCmd.ConfigFile, c.rootCmd.dbReg, configformat.AsOptions{Extended: c.extended})
	if err != nil {
		return err
	}
	fmt.Print(out)
	return nil
}

func (c *configAsCmd) asEnv() error {
	out, err := configformat.FormatAsEnv(c.rootCmd.cfg, c.rootCmd.ConfigFile, c.rootCmd.dbReg, configformat.AsOptions{Extended: c.extended})
	if err != nil {
		return err
	}
	fmt.Print(out)
	return nil
}

func (c *configAsCmd) asJSON() error {
	out, err := configformat.FormatAsJSON(c.rootCmd.cfg, c.rootCmd.ConfigFile)
	if err != nil {
		return err
	}
	fmt.Print(out)
	return nil
}

func (c *configAsCmd) asCLI() error {
	out, err := configformat.FormatAsCLI(c.rootCmd.cfg, c.rootCmd.ConfigFile)
	if err != nil {
		return err
	}
	fmt.Print(out)
	return nil
}
