package main

import (
	"flag"
	"fmt"
)

// testVerificationCmd implements the "verification" subcommand under "test".
type testVerificationCmd struct {
	*testCmd
	fs *flag.FlagSet
}

func parseTestVerificationCmd(parent *testCmd, args []string) (*testVerificationCmd, error) {
	c := &testVerificationCmd{testCmd: parent}
	c.fs = newFlagSet("verification")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *testVerificationCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing verification command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "template":
		cmd, err := parseTestVerificationTemplateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("template: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown verification command %q", args[0])
	}
}

func (c *testVerificationCmd) Usage() {
	executeUsage(c.fs.Output(), "test_verification_usage.txt", c)
}

func (c *testVerificationCmd) FlagGroups() []flagGroup {
	return append(c.testCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*testVerificationCmd)(nil)
