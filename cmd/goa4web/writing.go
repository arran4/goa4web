package main

import (
	_ "embed"
	"flag"
	"fmt"
)

//go:embed templates/writing_usage.txt
var writingUsageTemplate string

// writingCmd handles writing management subcommands.
type writingCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parseWritingCmd(parent *rootCmd, args []string) (*writingCmd, error) {
	c := &writingCmd{rootCmd: parent}
	fs := flag.NewFlagSet("writing", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *writingCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing writing command")
	}
	switch c.args[0] {
	case "tree":
		cmd, err := parseWritingTreeCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("tree: %w", err)
		}
		return cmd.Run()
	case "list":
		cmd, err := parseWritingListCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "read":
		cmd, err := parseWritingReadCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		return cmd.Run()
	case "comments":
		cmd, err := parseWritingCommentsCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("comments: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown writing command %q", c.args[0])
	}
}

// Usage prints command usage information with examples.
func (c *writingCmd) Usage() {
	executeUsage(c.fs.Output(), writingUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
