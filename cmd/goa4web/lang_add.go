package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// langAddCmd implements "lang add".
type langAddCmd struct {
	*langCmd
	fs   *flag.FlagSet
	Code string
	Name string
}

func parseLangAddCmd(parent *langCmd, args []string) (*langAddCmd, error) {
	c := &langAddCmd{langCmd: parent}
	c.fs = newFlagSet("add")
	c.fs.StringVar(&c.Code, "code", "", "language code")
	c.fs.StringVar(&c.Name, "name", "", "language name")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *langAddCmd) Run() error {
	if c.Code == "" || c.Name == "" {
		return fmt.Errorf("code and name required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	c.rootCmd.Verbosef("adding language %s (%s)", c.Name, c.Code)
	if _, err := queries.AdminInsertLanguage(ctx, sql.NullString{String: c.Name, Valid: true}); err != nil {
		return fmt.Errorf("insert language: %w", err)
	}
	c.rootCmd.Infof("added language %s (%s)", c.Name, c.Code)
	return nil
}

func (c *langAddCmd) Usage() {
	executeUsage(c.fs.Output(), "lang_add_usage.txt", c)
}

func (c *langAddCmd) FlagGroups() []flagGroup {
	return append(c.langCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*langAddCmd)(nil)
