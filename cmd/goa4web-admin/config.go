package main

import (
	"flag"
	"fmt"
)

// configCmd handles configuration management commands.
type configCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parseConfigCmd(parent *rootCmd, args []string) (*configCmd, error) {
	c := &configCmd{rootCmd: parent}
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *configCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing config command")
	}
	switch c.args[0] {
	case "show":
		cmd, err := parseConfigShowCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("show: %w", err)
		}
		return cmd.Run()
	case "set":
		cmd, err := parseConfigSetCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("set: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown config command %q", c.args[0])
	}
}

// Usage prints command usage information with examples.
func (c *configCmd) Usage() {
	w := c.fs.Output()
	fmt.Fprintf(w, "Usage:\n  %s config <command> [<args>]\n", c.rootCmd.fs.Name())
	fmt.Fprintln(w, "\nCommands:")
	fmt.Fprintln(w, "  show\tdisplay runtime configuration")
	fmt.Fprintln(w, "  set\tupdate configuration file")
	fmt.Fprintln(w, "\nExamples:")
	fmt.Fprintf(w, "  %s config show\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s config set -key DB_HOST -value localhost\n\n", c.rootCmd.fs.Name())
	c.fs.PrintDefaults()
}
