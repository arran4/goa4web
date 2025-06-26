package main

import (
	"flag"
	"fmt"
)

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
	w := c.fs.Output()
	fmt.Fprintf(w, "Usage:\n  %s lang <command> [<args>]\n", c.rootCmd.fs.Name())
	fmt.Fprintln(w, "\nCommands:")
	fmt.Fprintln(w, "  list\tlist languages")
	fmt.Fprintln(w, "  add\tadd a language")
	fmt.Fprintln(w, "  update\tupdate a language")
	fmt.Fprintln(w, "\nExamples:")
	fmt.Fprintf(w, "  %s lang list\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s lang add --code en --name English\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s lang update -id 1 -name New\n\n", c.rootCmd.fs.Name())
	c.fs.PrintDefaults()
}
