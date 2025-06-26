package main

import (
	"flag"
	"fmt"
)

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
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown board command %q", c.args[0])
	}
}

// Usage prints command usage information with examples.
func (c *boardCmd) Usage() {
	w := c.fs.Output()
	fmt.Fprintf(w, "Usage:\n  %s board <command> [<args>]\n", c.rootCmd.fs.Name())
	fmt.Fprintln(w, "\nCommands:")
	fmt.Fprintln(w, "  list\tlist boards")
	fmt.Fprintln(w, "  create\tcreate a board")
	fmt.Fprintln(w, "  delete\tdelete a board")
	fmt.Fprintln(w, "\nExamples:")
	fmt.Fprintf(w, "  %s board list\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s board create -name foo -description 'bar'\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s board delete -id 1\n\n", c.rootCmd.fs.Name())
	c.fs.PrintDefaults()
}
