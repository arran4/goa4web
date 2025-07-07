package main

import (
	"flag"
	"fmt"
)

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
	w := c.fs.Output()
	fmt.Fprintf(w, "Usage:\n  %s news comments <command> [<args>]\n", c.rootCmd.fs.Name())
	fmt.Fprintln(w, "\nCommands:")
	fmt.Fprintln(w, "  list\tlist comments for a news post")
	fmt.Fprintln(w, "  read\tread a comment or all comments")
	fmt.Fprintln(w, "\nExamples:")
	fmt.Fprintf(w, "  %s news comments list 3\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s news comments read 3 1\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s news comments read 3 all\n", c.rootCmd.fs.Name())
	c.fs.PrintDefaults()
}
