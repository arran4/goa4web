package main

import (
	"flag"
	"fmt"
)

// shareCmd implements "share".
type shareCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseShareCmd(parent *rootCmd, args []string) (*shareCmd, error) {
	c := &shareCmd{rootCmd: parent}
	c.fs = newFlagSet("share")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *shareCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing share command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "sign":
		cmd, err := parseShareSignCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("sign: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown share command %q", args[0])
	}
}

func (c *shareCmd) Usage() {
	executeUsage(c.fs.Output(), "share_usage.txt", c)
}

func (c *shareCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*shareCmd)(nil)
