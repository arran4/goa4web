package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
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
	case "apply":
		cmd, err := parseScenarioApplyCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("apply: %w", err)
		}
		return cmd.Run()
	case "serve":
		cmd, err := parseScenarioServeCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("serve: %w", err)
		}
		return cmd.Run()
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

func loadScenarioPath(fsys fs.FS, rawPath string) (*scenario.Scenario, error) {
	if rawPath == "" {
		return nil, fmt.Errorf("scenario path required")
	}

	var targetPath string
	if fsys != nil {
		targetPath = path.Clean(strings.TrimPrefix(filepath.ToSlash(rawPath), "/"))
		info, err := fs.Stat(fsys, targetPath)
		if err != nil {
			return nil, fmt.Errorf("scenario path: %w", err)
		}
		if info.IsDir() {
			scenarioFile := path.Join(targetPath, "scenario.txtar")
			if _, err := fs.Stat(fsys, scenarioFile); err != nil {
				return nil, fmt.Errorf("scenario directory missing scenario.txtar: %w", err)
			}
			targetPath = scenarioFile
		}
	} else {
		// Production boundary: inspect OS filesystem
		info, err := os.Stat(rawPath)
		if err != nil {
			return nil, fmt.Errorf("scenario path: %w", err)
		}
		var dir string
		var filename string
		if info.IsDir() {
			dir = rawPath
			filename = "scenario.txtar"
			if _, err := os.Stat(filepath.Join(dir, filename)); err != nil {
				return nil, fmt.Errorf("scenario directory missing scenario.txtar: %w", err)
			}
		} else {
			dir = filepath.Dir(rawPath)
			filename = filepath.Base(rawPath)
		}
		fsys = os.DirFS(dir)
		targetPath = filename
	}

	return scenario.ParseFS(fsys, targetPath)
}

// scenarioValidateCmd implements "scenario validate <path>".
type scenarioValidateCmd struct {
	*scenarioCmd
	fs   *flag.FlagSet
	Path string
	fsys fs.FS
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

	sc, err := loadScenarioPath(c.fsys, c.Path)
	if err != nil {
		return err
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

// scenarioApplyCmd implements "scenario apply <path>".
type scenarioApplyCmd struct {
	*scenarioCmd
	fs      *flag.FlagSet
	Path    string
	fsys    fs.FS
	querier db.Querier
}

func parseScenarioApplyCmd(parent *scenarioCmd, args []string) (*scenarioApplyCmd, error) {
	c := &scenarioApplyCmd{scenarioCmd: parent}
	c.fs = newFlagSet("apply")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	if len(c.fs.Args()) > 0 {
		c.Path = c.fs.Arg(0)
	}
	return c, nil
}

func (c *scenarioApplyCmd) Run() error {
	if c.Path == "" {
		c.fs.Usage()
		return fmt.Errorf("scenario path required")
	}

	sc, err := loadScenarioPath(c.fsys, c.Path)
	if err != nil {
		return err
	}

	ctx := context.Background()
	var queries db.Querier
	if c.querier != nil {
		queries = c.querier
	} else {
		conn, err := c.DB()
		if err != nil {
			return fmt.Errorf("database: %w", err)
		}
		queries = db.New(conn)
	}

	cd := common.NewCoreData(ctx, queries, c.cfg)
	runner := scenario.NewRunner(cd)

	res, err := runner.Apply(ctx, sc)
	if err != nil {
		return fmt.Errorf("apply scenario: %w", err)
	}

	c.Infof("scenario %q applied successfully (%d events)", res.ScenarioName, res.EventsApplied)
	return nil
}

func (c *scenarioApplyCmd) Usage() {
	_ = executeUsage(c.fs.Output(), "scenario_apply_usage.txt", c)
}

func (c *scenarioApplyCmd) FlagGroups() []flagGroup {
	return append(c.scenarioCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*scenarioApplyCmd)(nil)
