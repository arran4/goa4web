package main

import (
	"encoding/json"
	"flag"
	"fmt"
)

// configShowCmd implements "config show".
type configShowCmd struct {
	*configCmd
	fs   *flag.FlagSet
	args []string
}

func parseConfigShowCmd(parent *configCmd, args []string) (*configShowCmd, error) {
	c := &configShowCmd{configCmd: parent}
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *configShowCmd) Run() error {
	b, err := json.MarshalIndent(c.rootCmd.cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	fmt.Println(string(b))
	return nil
}
