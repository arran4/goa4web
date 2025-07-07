package main

import (
	_ "embed"
	"flag"
	"fmt"
)

//go:embed templates/board_usage.txt
var boardUsageTemplate string

// boardCmd handles board management subcommands.
type boardCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parseBoardCmd(parent *rootCmd, args []string) (*boardCmd, error) {
	c := &boardCmd{rootCmd: parent}
	fs := flag.NewFlagSet("board", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *boardCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing board command")
	}
	switch c.args[0] {
	case "list":
		cmd, err := parseBoardListCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "create":
		cmd, err := parseBoardCreateCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		return cmd.Run()
	case "delete":
		cmd, err := parseBoardDeleteCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("delete: %w", err)
		}
		return cmd.Run()
	case "update":
		cmd, err := parseBoardUpdateCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown board command %q", c.args[0])
	}
}

// Usage prints command usage information with examples.
func (c *boardCmd) Usage() {
	executeUsage(c.fs.Output(), boardUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
