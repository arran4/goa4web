package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/arran4/goa4web/internal/scenario"
)

// scenarioCmd handles scenario parsing, validation, and execution.
type scenarioCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseScenarioCmd(parent *rootCmd, args []string) (*scenarioCmd, error) {
	c := &scenarioCmd{rootCmd: parent}
	c.fs = newFlagSet("scenario")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *scenarioCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing scenario command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "validate":
		cmd, err := parseScenarioValidateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("validate: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown scenario command %q", args[0])
	}
}

func (c *scenarioCmd) Usage() {
	_ = executeUsage(c.fs.Output(), "scenario_usage.txt", c)
}

func (c *scenarioCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*scenarioCmd)(nil)

// scenarioValidateCmd implements "scenario validate <path>".
type scenarioValidateCmd struct {
	*scenarioCmd
	fs   *flag.FlagSet
	Path string
}

func parseScenarioValidateCmd(parent *scenarioCmd, args []string) (*scenarioValidateCmd, error) {
	c := &scenarioValidateCmd{scenarioCmd: parent}
	c.fs = newFlagSet("validate")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	if len(c.fs.Args()) > 0 {
		c.Path = c.fs.Arg(0)
	}
	return c, nil
}

func (c *scenarioValidateCmd) Run() error {
	if c.Path == "" {
		c.fs.Usage()
		return fmt.Errorf("scenario path required")
	}

	targetPath := c.Path
	info, err := os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("scenario path: %w", err)
	}

	var dir string
	var filename string

	if info.IsDir() {
		dir = targetPath
		filename = "scenario.txtar"
		if _, err := os.Stat(filepath.Join(dir, filename)); err != nil {
			return fmt.Errorf("scenario directory missing scenario.txtar: %w", err)
		}
	} else {
		dir = filepath.Dir(targetPath)
		filename = filepath.Base(targetPath)
	}

	dirFS := os.DirFS(dir)
	sc, err := scenario.ParseFS(dirFS, filename)
	if err != nil {
		return fmt.Errorf("parse scenario: %w", err)
	}

	if err := scenario.Validate(sc); err != nil {
		return fmt.Errorf("validate scenario: %w", err)
	}

	c.Infof("scenario %q valid (%d events)", sc.Meta.Name, len(sc.Events))
	return nil
}

func (c *scenarioValidateCmd) Usage() {
	_ = executeUsage(c.fs.Output(), "scenario_validate_usage.txt", c)
}

func (c *scenarioValidateCmd) FlagGroups() []flagGroup {
	return append(c.scenarioCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*scenarioValidateCmd)(nil)
