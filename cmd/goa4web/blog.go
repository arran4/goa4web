package main

import (
	_ "embed"
	"flag"
	"fmt"
)

//go:embed templates/blog_usage.txt
var blogUsageTemplate string

// blogCmd handles blog management subcommands.
type blogCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parseBlogCmd(parent *rootCmd, args []string) (*blogCmd, error) {
	c := &blogCmd{rootCmd: parent}
	fs := flag.NewFlagSet("blog", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *blogCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing blog command")
	}
	switch c.args[0] {
	case "create":
		cmd, err := parseBlogCreateCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		return cmd.Run()
	case "list":
		cmd, err := parseBlogListCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "read":
		cmd, err := parseBlogReadCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		return cmd.Run()
	case "comments":
		cmd, err := parseBlogCommentsCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("comments: %w", err)
		}
		return cmd.Run()
	case "update":
		cmd, err := parseBlogUpdateCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}
		return cmd.Run()
	case "deactivate":
		cmd, err := parseBlogDeactivateCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("deactivate: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown blog command %q", c.args[0])
	}
}

func (c *blogCmd) Usage() {
	executeUsage(c.fs.Output(), blogUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
