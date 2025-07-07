package main

import (
	_ "embed"
	"flag"
	"fmt"
)

//go:embed templates/faq_usage.txt
var faqUsageTemplate string

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
	executeUsage(c.fs.Output(), faqUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
