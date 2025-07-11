package main

import (
	_ "embed"
	"flag"
	"fmt"
)

//go:embed templates/ipban_usage.txt
var ipBanUsageTemplate string

// ipBanCmd implements IP ban management commands.
type ipBanCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parseIpBanCmd(parent *rootCmd, args []string) (*ipBanCmd, error) {
	c := &ipBanCmd{rootCmd: parent}
	fs := flag.NewFlagSet("ipban", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *ipBanCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing ipban command")
	}
	switch c.args[0] {
	case "add":
		cmd, err := parseIpBanAddCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("add: %w", err)
		}
		return cmd.Run()
	case "list":
		cmd, err := parseIpBanListCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "delete":
		cmd, err := parseIpBanDeleteCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("delete: %w", err)
		}
		return cmd.Run()
	case "update":
		cmd, err := parseIpBanUpdateCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown ipban command %q", c.args[0])
	}
}

// Usage prints command usage information with examples.
func (c *ipBanCmd) Usage() {
	executeUsage(c.fs.Output(), ipBanUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
