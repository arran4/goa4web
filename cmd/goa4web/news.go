package main

import (
	_ "embed"
	"flag"
	"fmt"
)

//go:embed templates/news_usage.txt
var newsUsageTemplate string

// newsCmd handles news management subcommands.
type newsCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parseNewsCmd(parent *rootCmd, args []string) (*newsCmd, error) {
	c := &newsCmd{rootCmd: parent}
	fs := flag.NewFlagSet("news", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *newsCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing news command")
	}
	switch c.args[0] {
	case "list":
		cmd, err := parseNewsListCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "read":
		cmd, err := parseNewsReadCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		return cmd.Run()
	case "comments":
		cmd, err := parseNewsCommentsCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("comments: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown news command %q", c.args[0])
	}
}

// Usage prints command usage information with examples.
func (c *newsCmd) Usage() {
	executeUsage(c.fs.Output(), newsUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
