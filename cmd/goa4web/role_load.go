package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// roleLoadCmd implements the "role load" subcommand.
type roleLoadCmd struct {
	*roleCmd
	fs   *flag.FlagSet
	role string
}

func parseRoleLoadCmd(parent *roleCmd, args []string) (*roleLoadCmd, error) {
	c := &roleLoadCmd{roleCmd: parent}
	fs := flag.NewFlagSet("load", flag.ContinueOnError)
	c.fs = fs
	fs.StringVar(&c.role, "role", "", "The name of the role to load.")
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if c.role == "" {
		return nil, fmt.Errorf("role name is required")
	}
	return c, nil
}

func (c *roleLoadCmd) Run() error {
	sdb, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(sdb)

	filename := fmt.Sprintf("%s.sql", c.role)
	filepath := filepath.Join("database", "roles", filename)
	log.Printf("Loading role from %s", filepath)

	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read role file: %w", err)
	}

	if err := runStatements(sdb, strings.NewReader(string(data))); err != nil {
		return fmt.Errorf("failed to apply role: %w", err)
	}

	log.Printf("Role %q loaded successfully.", c.role)
	return nil
}

func (c *roleLoadCmd) Usage() {
	executeUsage(c.fs.Output(), "role_load_usage.txt", c)
}

func (c *roleLoadCmd) FlagGroups() []flagGroup {
	return append(c.rootCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*roleLoadCmd)(nil)
