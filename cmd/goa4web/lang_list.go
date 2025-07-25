package main

import (
	"context"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// langListCmd implements "lang list".
type langListCmd struct {
	*langCmd
	fs *flag.FlagSet
}

func parseLangListCmd(parent *langCmd, args []string) (*langListCmd, error) {
	c := &langListCmd{langCmd: parent}
	c.fs = newFlagSet("list")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *langListCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	langs, err := queries.AllLanguages(ctx)
	if err != nil {
		return fmt.Errorf("list languages: %w", err)
	}
	for _, l := range langs {
		fmt.Printf("%d\t%s\n", l.Idlanguage, l.Nameof.String)
	}
	return nil
}
