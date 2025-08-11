package main

import (
	"flag"
	"fmt"

	coretemplates "github.com/arran4/goa4web/core/templates"
)

// templatesExtractCmd implements "templates extract".
type templatesExtractCmd struct {
	*templatesCmd
	fs  *flag.FlagSet
	dir string
}

func parseTemplatesExtractCmd(parent *templatesCmd, args []string) (*templatesExtractCmd, error) {
	c := &templatesExtractCmd{templatesCmd: parent}
	c.fs = newFlagSet("extract")
	c.fs.StringVar(&c.dir, "dir", "", "destination directory")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *templatesExtractCmd) Run() error {
	if c.dir == "" {
		return fmt.Errorf("dir flag is required")
	}
	if err := coretemplates.WriteToDir(c.dir); err != nil {
		return fmt.Errorf("write templates: %w", err)
	}
	return nil
}

func (c *templatesExtractCmd) Usage() {
	executeUsage(c.fs.Output(), "templates_extract_usage.txt", c)
}

func (c *templatesExtractCmd) FlagGroups() []flagGroup {
	return append(c.templatesCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*templatesExtractCmd)(nil)
