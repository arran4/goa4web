package main

import (
	"flag"
	"fmt"
)

// blogCommentsCmd handles "blog comments".
type blogCommentsCmd struct {
	*blogCmd
	fs   *flag.FlagSet
	args []string
}

func parseBlogCommentsCmd(parent *blogCmd, args []string) (*blogCommentsCmd, error) {
	c := &blogCommentsCmd{blogCmd: parent}
	fs := flag.NewFlagSet("comments", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *blogCommentsCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing comments command")
	}
	switch c.args[0] {
	case "list":
		cmd, err := parseBlogCommentsListCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "read":
		cmd, err := parseBlogCommentsReadCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown comments command %q", c.args[0])
	}
}

func (c *blogCommentsCmd) Usage() {
	w := c.fs.Output()
	fmt.Fprintf(w, "Usage:\n  %s blog comments <command> [<args>]\n", c.rootCmd.fs.Name())
	fmt.Fprintln(w, "\nCommands:")
	fmt.Fprintln(w, "  list\tlist comments for a blog")
	fmt.Fprintln(w, "  read\tread a comment or all comments")
	fmt.Fprintln(w, "\nExamples:")
	fmt.Fprintf(w, "  %s blog comments list 3\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s blog comments read 3 1\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s blog comments read 3 all\n", c.rootCmd.fs.Name())
	c.fs.PrintDefaults()
}
