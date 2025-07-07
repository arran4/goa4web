package main

import (
	_ "embed"
	"flag"
	"fmt"
)

//go:embed templates/server_usage.txt
var serverUsageTemplate string

// serverCmd handles server management commands.
type serverCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parseServerCmd(parent *rootCmd, args []string) (*serverCmd, error) {
	c := &serverCmd{rootCmd: parent}
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *serverCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing server command")
	}
	switch c.args[0] {
	case "shutdown":
		cmd, err := parseServerShutdownCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown server command %q", c.args[0])
	}
}

// Usage prints command usage information with examples.
func (c *serverCmd) Usage() {
	executeUsage(c.fs.Output(), serverUsageTemplate, c.fs, c.rootCmd.fs.Name())
}
