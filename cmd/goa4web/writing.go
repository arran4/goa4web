package main

import (
	"flag"
	"fmt"
)

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
	w := c.fs.Output()
	fmt.Fprintf(w, "Usage:\n  %s writing <command> [<args>]\n", c.rootCmd.fs.Name())
	fmt.Fprintln(w, "\nCommands:")
	fmt.Fprintln(w, "  tree\tshow writing categories")
	fmt.Fprintln(w, "  list\tlist writings")
	fmt.Fprintln(w, "  read\tread a writing")
	fmt.Fprintln(w, "  comments\tmanage comments for a writing")
	fmt.Fprintln(w, "\nExamples:")
	fmt.Fprintf(w, "  %s writing tree\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s writing list\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s writing read 1\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s writing comments list 1\n", c.rootCmd.fs.Name())
	c.fs.PrintDefaults()
}
