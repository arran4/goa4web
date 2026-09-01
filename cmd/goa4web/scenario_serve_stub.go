//go:build !sqlite && !sqlite3

package main

import (
	"errors"
	"flag"
)

// scenarioServeCmd is the non-SQLite stub.
type scenarioServeCmd struct {
	*scenarioCmd
	fs *flag.FlagSet
}

func parseScenarioServeCmd(parent *scenarioCmd, args []string) (*scenarioServeCmd, error) {
	c := &scenarioServeCmd{scenarioCmd: parent}
	c.fs = newFlagSet("serve")
	c.fs.Usage = c.Usage
	_ = c.fs.Parse(args)
	return c, nil
}

func (c *scenarioServeCmd) Run() error {
	return errors.New("scenario serve requires sqlite support: recompile goa4web with -tags sqlite")
}

func (c *scenarioServeCmd) Usage() {
	_ = executeUsage(c.fs.Output(), "scenario_serve_usage.txt", c)
}

func (c *scenarioServeCmd) FlagGroups() []flagGroup {
	return append(c.scenarioCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*scenarioServeCmd)(nil)
