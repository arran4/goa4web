package main

import (
	"flag"
	"fmt"
)

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
	w := c.fs.Output()
	fmt.Fprintf(w, "Usage:\n  %s ipban <command> [<args>]\n", c.rootCmd.fs.Name())
	fmt.Fprintln(w, "\nCommands:")
	fmt.Fprintln(w, "  add\tadd an IP ban")
	fmt.Fprintln(w, "  list\tlist banned IPs")
	fmt.Fprintln(w, "  delete\tremove an IP ban")
	fmt.Fprintln(w, "  update\tupdate an IP ban")
	fmt.Fprintln(w, "\nExamples:")
	fmt.Fprintf(w, "  %s ipban add -ip 192.168.1.1 -reason spam\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s ipban list\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s ipban update -id 1 -reason updated\n\n", c.rootCmd.fs.Name())
	c.fs.PrintDefaults()
}
