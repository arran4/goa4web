package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/arran4/goa4web/internal/app/dbstart"
)

// testCmd implements the "test" command.
type testCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseTestCmd(parent *rootCmd, args []string) (*testCmd, error) {
	c := &testCmd{rootCmd: parent}
	c.fs = newFlagSet("test")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *testCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing test command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "migrations":
		cmd, err := parseTestMigrationsCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("migrations: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown test command %q", args[0])
	}
}

func (c *testCmd) Usage() {
	executeUsage(c.fs.Output(), "test_usage.txt", c)
}

func (c *testCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*testCmd)(nil)

// ----- migrations subcommands -----

type testMigrationsCmd struct {
	*testCmd
	fs *flag.FlagSet
}

func parseTestMigrationsCmd(parent *testCmd, args []string) (*testMigrationsCmd, error) {
	c := &testMigrationsCmd{testCmd: parent}
	c.fs = newFlagSet("migrations")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *testMigrationsCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing migrations command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "apply":
		cmd, err := parseTestMigrationsApplyCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("apply: %w", err)
		}
		return cmd.Run()
	case "clean":
		cmd, err := parseTestMigrationsCleanCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("clean: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown migrations command %q", args[0])
	}
}

func (c *testMigrationsCmd) Usage() {
	executeUsage(c.fs.Output(), "test_migrations_usage.txt", c)
}

func (c *testMigrationsCmd) FlagGroups() []flagGroup {
	return append(c.testCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*testMigrationsCmd)(nil)

// ----- apply all -----

type testMigrationsApplyCmd struct {
	*testMigrationsCmd
	fs         *flag.FlagSet
	DBType     string
	ConnString string
}

func parseTestMigrationsApplyCmd(parent *testMigrationsCmd, args []string) (*testMigrationsApplyCmd, error) {
	c := &testMigrationsApplyCmd{testMigrationsCmd: parent}
	c.fs = newFlagSet("apply")
	c.fs.StringVar(&c.DBType, "database-type", parent.rootCmd.cfg.DBDriver, "database driver")
	c.fs.StringVar(&c.ConnString, "connection-string", parent.rootCmd.cfg.DBConn, "database connection string")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *testMigrationsApplyCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 || args[0] != "all" {
		c.fs.Usage()
		return fmt.Errorf("missing 'all' subcommand")
	}
	connector, err := c.rootCmd.dbReg.Connector(c.DBType, c.ConnString)
	if err != nil {
		return err
	}
	sdb := sql.OpenDB(connector)
	defer func(sdb *sql.DB) {
		err := sdb.Close()
		if err != nil {
			log.Printf("failed to close DB connection: %v", err)
		}
	}(sdb)
	if err := sdb.Ping(); err != nil {
		return err
	}
	c.rootCmd.Verbosef("applying migrations using %s", c.DBType)
	if err := dbstart.Apply(context.Background(), sdb, os.DirFS("migrations"), true, c.DBType); err != nil {
		return err
	}
	c.rootCmd.Infof("migrations applied successfully")
	return nil
}

func (c *testMigrationsApplyCmd) Usage() {
	executeUsage(c.fs.Output(), "test_migrations_apply_usage.txt", c)
}

func (c *testMigrationsApplyCmd) FlagGroups() []flagGroup {
	return append(c.testMigrationsCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*testMigrationsApplyCmd)(nil)

// ----- clean -----

type testMigrationsCleanCmd struct {
	*testMigrationsCmd
	fs         *flag.FlagSet
	DBType     string
	ConnString string
}

func parseTestMigrationsCleanCmd(parent *testMigrationsCmd, args []string) (*testMigrationsCleanCmd, error) {
	c := &testMigrationsCleanCmd{testMigrationsCmd: parent}
	c.fs = newFlagSet("clean")
	c.fs.StringVar(&c.DBType, "database-type", parent.rootCmd.cfg.DBDriver, "database driver")
	c.fs.StringVar(&c.ConnString, "connection-string", parent.rootCmd.cfg.DBConn, "database connection string")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *testMigrationsCleanCmd) Run() error {
	connector, err := c.rootCmd.dbReg.Connector(c.DBType, c.ConnString)
	if err != nil {
		return err
	}
	sdb := sql.OpenDB(connector)
	defer func(sdb *sql.DB) {
		err := sdb.Close()
		if err != nil {
			log.Printf("failed to close DB connection: %v", err)
		}
	}(sdb)
	if err := sdb.Ping(); err != nil {
		return err
	}
	ctx := context.Background()
	var q string
	switch c.DBType {
	case "mysql":
		q = "SHOW TABLES"
	default:
		return fmt.Errorf("unsupported driver %s", c.DBType)
	}
	rows, err := sdb.QueryContext(ctx, q)
	if err != nil {
		return err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("failed to close DB rows: %v", err)
		}
	}(rows)
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return err
		}
		names = append(names, n)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, n := range names {
		if _, err := sdb.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", n)); err != nil {
			return err
		}
	}
	c.rootCmd.Infof("database cleaned")
	return nil
}

func (c *testMigrationsCleanCmd) Usage() {
	executeUsage(c.fs.Output(), "test_migrations_clean_usage.txt", c)
}

func (c *testMigrationsCleanCmd) FlagGroups() []flagGroup {
	return append(c.testMigrationsCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*testMigrationsCleanCmd)(nil)
