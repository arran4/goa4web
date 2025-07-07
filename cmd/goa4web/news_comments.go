package main

import (
	_ "embed"
	"flag"
	"fmt"
)

//go:embed templates/news_comments_usage.txt
var newsCommentsUsageTemplate string

// newsCommentsCmd handles "news comments".
type newsCommentsCmd struct {
	*newsCmd
	fs   *flag.FlagSet
	args []string
}

func parseNewsCommentsCmd(parent *newsCmd, args []string) (*newsCommentsCmd, error) {
	c := &newsCommentsCmd{newsCmd: parent}
	fs := flag.NewFlagSet("comments", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *newsCommentsCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing comments command")
	}
	switch c.args[0] {
	case "list":
		cmd, err := parseNewsCommentsListCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "read":
		cmd, err := parseNewsCommentsReadCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown comments command %q", c.args[0])
	}
}

func (c *newsCommentsCmd) Usage() {
	executeUsage(c.fs.Output(), newsCommentsUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
