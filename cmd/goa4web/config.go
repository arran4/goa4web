package main

import (
	"flag"
	"fmt"
)

// configCmd handles configuration utilities.
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
	case "reload":
		cmd, err := parseConfigReloadCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("reload: %w", err)
		}
		return cmd.Run()
	case "as-env":
		cmd, err := parseConfigAsCmd(c, "as-env", c.args[1:])
		if err != nil {
			return fmt.Errorf("as-env: %w", err)
		}
		return cmd.asEnv()
	case "as-env-file":
		cmd, err := parseConfigAsCmd(c, "as-env-file", c.args[1:])
		if err != nil {
			return fmt.Errorf("as-env-file: %w", err)
		}
		return cmd.asEnvFile()
	case "as-json":
		cmd, err := parseConfigAsCmd(c, "as-json", c.args[1:])
		if err != nil {
			return fmt.Errorf("as-json: %w", err)
		}
		return cmd.asJSON()
	case "as-cli":
		cmd, err := parseConfigAsCmd(c, "as-cli", c.args[1:])
		if err != nil {
			return fmt.Errorf("as-cli: %w", err)
		}
		return cmd.asCLI()
	case "add-json":
		cmd, err := parseConfigJSONAddCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("add-json: %w", err)
		}
		return cmd.Run()
	case "options":
		cmd, err := parseConfigOptionsCmd(c, c.args[1:])
		if err != nil {
			return fmt.Errorf("options: %w", err)
		}
		return cmd.Run()
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
	fmt.Fprintln(w, "  reload\treload configuration from file")
	fmt.Fprintln(w, "  as-env\toutput configuration as export statements")
	fmt.Fprintln(w, "  as-env-file\toutput configuration as env file")
	fmt.Fprintln(w, "  as-json\toutput configuration as JSON")
	fmt.Fprintln(w, "  as-cli\toutput configuration as CLI flags")
	fmt.Fprintln(w, "  add-json\tadd missing options to JSON file")
	fmt.Fprintln(w, "  options\tlist available configuration options")
	fmt.Fprintln(w, "\nExamples:")
	fmt.Fprintf(w, "  %s config reload\n\n", c.rootCmd.fs.Name())
	fmt.Fprintln(w, "  show\tdisplay runtime configuration")
	fmt.Fprintln(w, "  set\tupdate configuration file")
	fmt.Fprintln(w, "\nExamples:")
	fmt.Fprintf(w, "  %s config show\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s config set -key DB_HOST -value localhost\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s config as-env-file > config.env\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s config as-cli\n\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s config options\n", c.rootCmd.fs.Name())
	fmt.Fprintf(w, "  %s config add-json -file cfg.json\n", c.rootCmd.fs.Name())
	c.fs.PrintDefaults()
}
