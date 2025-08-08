package main

import (
	"flag"
	"fmt"
)

// linksRemapCmd groups commands for extracting and applying URL remappings.
type linksRemapCmd struct {
	*linksCmd
	fs *flag.FlagSet
}

func parseLinksRemapCmd(parent *linksCmd, args []string) (*linksRemapCmd, error) {
	c := &linksRemapCmd{linksCmd: parent}
	c.fs = newFlagSet("remap")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *linksRemapCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing remap command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "extract":
		cmd, err := parseLinksRemapExtractCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("extract: %w", err)
		}
		return cmd.Run()
	case "apply":
		cmd, err := parseLinksRemapApplyCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("apply: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown remap command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *linksRemapCmd) Usage() {
	executeUsage(c.fs.Output(), "links_remap_usage.txt", c)
}

func (c *linksRemapCmd) FlagGroups() []flagGroup {
	return append(c.linksCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*linksRemapCmd)(nil)
