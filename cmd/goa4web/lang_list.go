package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
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
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	langs, err := queries.SystemListLanguages(ctx)
	if err != nil {
		return fmt.Errorf("list languages: %w", err)
	}
	for _, l := range langs {
		fmt.Printf("%d\t%s\n", l.Idlanguage, l.Nameof.String)
	}
	return nil
}
