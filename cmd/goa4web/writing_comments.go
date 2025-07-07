package main

import (
	"flag"
	"fmt"
)

// writingCommentsCmd handles "writing comments".
type writingCommentsCmd struct {
	*writingCmd
	fs   *flag.FlagSet
	args []string
}

func parseWritingCommentsCmd(parent *writingCmd, args []string) (*writingCommentsCmd, error) {
	c := &writingCommentsCmd{writingCmd: parent}
	fs := flag.NewFlagSet("comments", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *writingCommentsCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing comments command")
	}
	switch c.args[0] {
	case "list":
		cmd, err := parseWritingCommentsListCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "read":
		cmd, err := parseWritingCommentsReadCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown comments command %q", c.args[0])
	}
}

func (c *writingCommentsCmd) Usage() {
	w := c.fs.Output()
	fmt.Fprintf(w, "Usage:\n  %s writing comments <command> [<args>]\n", c.rootCmd.fs.Name())
	fmt.Fprintln(w, "\nCommands:")
	fmt.Fprintln(w, "  list\tlist comments for a writing")
	fmt.Fprintln(w, "  read\tread a comment or all comments")
	fmt.Fprintln(w, "\nExamples:")
	fmt.Fprintf(w, "  %s writing comments list 3\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s writing comments read 3 1\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s writing comments read 3 all\n", c.rootCmd.fs.Name())
	c.fs.PrintDefaults()
}
