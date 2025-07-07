package main

import (
	_ "embed"
	"flag"
	"fmt"
)

//go:embed templates/lang_usage.txt
var langUsageTemplate string

type langCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parseLangCmd(parent *rootCmd, args []string) (*langCmd, error) {
	c := &langCmd{rootCmd: parent}
	fs := flag.NewFlagSet("lang", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *langCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing lang command")
	}
	switch c.args[0] {
	case "add":
		cmd, err := parseLangAddCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("add: %w", err)
		}
		return cmd.Run()
	case "list":
		cmd, err := parseLangListCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "update":
		cmd, err := parseLangUpdateCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown lang command %q", c.args[0])
	}
}

func (c *langCmd) Usage() {
	executeUsage(c.fs.Output(), langUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
