package main

import (
	"context"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// boardListCmd implements "board list".
type boardListCmd struct {
	*boardCmd
	fs *flag.FlagSet
}

func parseBoardListCmd(parent *boardCmd, args []string) (*boardListCmd, error) {
	c := &boardListCmd{boardCmd: parent}
	c.fs = newFlagSet("list")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *boardListCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	rows, err := queries.GetAllImageBoards(ctx)
	if err != nil {
		return fmt.Errorf("list boards: %w", err)
	}
	for _, b := range rows {
		fmt.Printf("%d\t%s\t%s\n", b.Idimageboard, b.Title.String, b.Description.String)
	}
	return nil
}
