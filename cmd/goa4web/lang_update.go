package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// langUpdateCmd implements "lang update".
type langUpdateCmd struct {
	*langCmd
	fs   *flag.FlagSet
	ID   int
	Name string
}

func parseLangUpdateCmd(parent *langCmd, args []string) (*langUpdateCmd, error) {
	c := &langUpdateCmd{langCmd: parent}
	c.fs = newFlagSet("update")
	c.fs.IntVar(&c.ID, "id", 0, "language id")
	c.fs.StringVar(&c.Name, "name", "", "new name")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *langUpdateCmd) Run() error {
	if c.ID == 0 || c.Name == "" {
		return fmt.Errorf("id and name required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	err = queries.AdminRenameLanguage(ctx, db.AdminRenameLanguageParams{Nameof: sql.NullString{String: c.Name, Valid: true}, ID: int32(c.ID)})
	if err != nil {
		return fmt.Errorf("update language: %w", err)
	}
	return nil
}

func (c *langUpdateCmd) Usage() {
	executeUsage(c.fs.Output(), "lang_update_usage.txt", c)
}

func (c *langUpdateCmd) FlagGroups() []flagGroup {
	return append(c.langCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*langUpdateCmd)(nil)
