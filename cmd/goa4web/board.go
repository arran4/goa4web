package main

import (
	"flag"
	"fmt"
)

// boardCmd handles board management subcommands.
type boardCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseBoardCmd(parent *rootCmd, args []string) (*boardCmd, error) {
	c := &boardCmd{rootCmd: parent}
	c.fs = newFlagSet("board")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *boardCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing board command")
	}
	switch args[0] {
	case "list":
		cmd, err := parseBoardListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "create":
		cmd, err := parseBoardCreateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		return cmd.Run()
	case "delete":
		cmd, err := parseBoardDeleteCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("delete: %w", err)
		}
		return cmd.Run()
	case "update":
		cmd, err := parseBoardUpdateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown board command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *boardCmd) Usage() {
	executeUsage(c.fs.Output(), templateString("board_usage.txt"), c.fs, c.rootCmd.fs.Name())
}
