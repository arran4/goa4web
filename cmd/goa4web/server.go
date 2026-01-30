package main

import (
	"flag"
	"fmt"
)

// serverCmd handles server management commands.
type serverCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseServerCmd(parent *rootCmd, args []string) (*serverCmd, error) {
	c := &serverCmd{rootCmd: parent}
	c.fs = newFlagSet("server")

	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *serverCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing server command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "shutdown":
		cmd, err := parseServerShutdownCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		return cmd.Run()
	case "stats":
		cmd, err := parseServerStatsCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("stats: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown server command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *serverCmd) Usage() {
	executeUsage(c.fs.Output(), "server_usage.txt", c)
}

func (c *serverCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*serverCmd)(nil)
