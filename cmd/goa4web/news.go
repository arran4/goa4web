package main

import (
	"flag"
	"fmt"
)

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
	w := c.fs.Output()
	fmt.Fprintf(w, "Usage:\n  %s news <command> [<args>]\n", c.rootCmd.fs.Name())
	fmt.Fprintln(w, "\nCommands:")
	fmt.Fprintln(w, "  list\tlist news posts")
	fmt.Fprintln(w, "  read\tread a news post")
	fmt.Fprintln(w, "  comments\tmanage news comments")
	fmt.Fprintln(w, "\nExamples:")
	fmt.Fprintf(w, "  %s news list\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s news read 1\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s news comments list 1\n", c.rootCmd.fs.Name())
	c.fs.PrintDefaults()
}
