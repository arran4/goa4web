package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// langAddCmd implements "lang add".
type langAddCmd struct {
	*langCmd
	fs   *flag.FlagSet
	Code string
	Name string
	args []string
}

func parseLangAddCmd(parent *langCmd, args []string) (*langAddCmd, error) {
	c := &langAddCmd{langCmd: parent}
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.StringVar(&c.Code, "code", "", "language code")
	fs.StringVar(&c.Name, "name", "", "language name")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *langAddCmd) Run() error {
	if c.Code == "" || c.Name == "" {
		return fmt.Errorf("code and name required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	c.rootCmd.Verbosef("adding language %s (%s)", c.Name, c.Code)
	if _, err := queries.InsertLanguage(ctx, sql.NullString{String: c.Name, Valid: true}); err != nil {
		return fmt.Errorf("insert language: %w", err)
	}
	c.rootCmd.Infof("added language %s (%s)", c.Name, c.Code)
	return nil
}
