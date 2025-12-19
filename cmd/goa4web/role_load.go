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
	file string
}

func parseRoleLoadCmd(parent *roleCmd, args []string) (*roleLoadCmd, error) {
	c := &roleLoadCmd{roleCmd: parent}
	fs := flag.NewFlagSet("load", flag.ContinueOnError)
	c.fs = fs
	fs.StringVar(&c.role, "role", "", "The name of the role to load.")
	fs.StringVar(&c.file, "file", "", "Optional path to a .sql file to load instead of the embedded role script.")
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

	var data []byte
	if c.file != "" {
		// Explicit filesystem file provided
		p := c.file
		if !strings.HasSuffix(strings.ToLower(p), ".sql") {
			p = p + ".sql"
		}
		abs, _ := filepath.Abs(p)
		log.Printf("Loading role %q from file %s", c.role, abs)
		b, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("failed to read role file: %w", err)
		}
		data = b
	} else {
		// Default to embedded role
		log.Printf("Loading role %q from embedded roles", c.role)
		b, err := readEmbeddedRole(c.role)
		if err != nil {
			return fmt.Errorf("failed to read embedded role %q: %w", c.role, err)
		}
		data = b
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
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*roleLoadCmd)(nil)
