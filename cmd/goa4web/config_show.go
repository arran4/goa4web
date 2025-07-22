package main

import (
	"encoding/json"
	"flag"
	"fmt"
)

// configShowCmd implements "config show".
type configShowCmd struct {
	*configCmd
	fs *flag.FlagSet
}

func parseConfigShowCmd(parent *configCmd, args []string) (*configShowCmd, error) {
	c := &configShowCmd{configCmd: parent}
	c.fs = newFlagSet("show")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

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
