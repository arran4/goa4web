package main

import (
	"flag"
	"fmt"
)

// templatesCmd handles template-related subcommands.
type templatesCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseTemplatesCmd(parent *rootCmd, args []string) (*templatesCmd, error) {
	c := &templatesCmd{rootCmd: parent}
	c.fs = newFlagSet("templates")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *templatesCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing templates command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "extract":
		cmd, err := parseTemplatesExtractCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("extract: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown templates command %q", args[0])
	}
}

func (c *templatesCmd) Usage() {
	executeUsage(c.fs.Output(), "templates_usage.txt", c)
}

func (c *templatesCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*templatesCmd)(nil)
