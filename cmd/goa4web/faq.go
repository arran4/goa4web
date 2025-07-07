package main

import (
	"flag"
	"fmt"
)

// faqCmd handles FAQ management subcommands.
type faqCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parseFaqCmd(parent *rootCmd, args []string) (*faqCmd, error) {
	c := &faqCmd{rootCmd: parent}
	fs := flag.NewFlagSet("faq", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *faqCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing faq command")
	}
	switch c.args[0] {
	case "tree":
		cmd, err := parseFaqTreeCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("tree: %w", err)
		}
		return cmd.Run()
	case "read":
		cmd, err := parseFaqReadCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown faq command %q", c.args[0])
	}
}

func (c *faqCmd) Usage() {
	w := c.fs.Output()
	fmt.Fprintf(w, "Usage:\n  %s faq <command> [<args>]\n", c.rootCmd.fs.Name())
	fmt.Fprintln(w, "\nCommands:")
	fmt.Fprintln(w, "  tree\tshow FAQ categories and questions")
	fmt.Fprintln(w, "  read\tread a FAQ question")
	fmt.Fprintln(w, "\nExamples:")
	fmt.Fprintf(w, "  %s faq tree\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s faq read 1\n", c.rootCmd.fs.Name())
	c.fs.PrintDefaults()
}
