package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/arran4/goa4web/internal/configformat"
)

// configAsCmd implements "config as-*" commands.
type configAsCmd struct {
	*configCmd
	fs       *flag.FlagSet
	extended bool
	output   io.Writer
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
	out, err := configformat.FormatAsEnvFile(c.cfg, c.ConfigFile, c.dbReg, configformat.AsOptions{Extended: c.extended})
	if err != nil {
		return err
	}
	return c.writeOutput(out)
}

func (c *configAsCmd) asEnv() error {
	out, err := configformat.FormatAsEnv(c.cfg, c.ConfigFile, c.dbReg, configformat.AsOptions{Extended: c.extended})
	if err != nil {
		return err
	}
	return c.writeOutput(out)
}

func (c *configAsCmd) asJSON() error {
	out, err := configformat.FormatAsJSON(c.cfg, c.ConfigFile)
	if err != nil {
		return err
	}
	return c.writeOutput(out)
}

func (c *configAsCmd) asCLI() error {
	out, err := configformat.FormatAsCLI(c.cfg, c.ConfigFile)
	if err != nil {
		return err
	}
	return c.writeOutput(out)
}

func (c *configAsCmd) writeOutput(output string) error {
	w := c.output
	if w == nil {
		w = os.Stdout
	}
	_, err := fmt.Fprint(w, output)
	return err
}
