package main

import (
	"flag"
	"fmt"
)

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
	w := c.fs.Output()
	fmt.Fprintf(w, "Usage:\n  %s blog <command> [<args>]\n", c.rootCmd.fs.Name())
	fmt.Fprintln(w, "\nCommands:")
	fmt.Fprintln(w, "  create\tcreate a blog entry")
	fmt.Fprintln(w, "  list\tlist blog entries")
	fmt.Fprintln(w, "  read\tread a blog entry")
	fmt.Fprintln(w, "  comments\tmanage blog comments")
	fmt.Fprintln(w, "  update\tupdate a blog entry")
	fmt.Fprintln(w, "  deactivate\tdeactivate a blog entry")
	fmt.Fprintln(w, "\nExamples:")
	fmt.Fprintf(w, "  %s blog create -user 1 -lang 1 -text 'hi'\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s blog list -user 1\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s blog read 1\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s blog comments list 1\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s blog update -id 1 -text 'changed'\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s blog deactivate -id 1\n\n", c.rootCmd.fs.Name())
	c.fs.PrintDefaults()
}
