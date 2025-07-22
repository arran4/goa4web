package main

import (
	"context"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// boardDeleteCmd implements "board delete".
type boardDeleteCmd struct {
	*boardCmd
	fs *flag.FlagSet
	ID int
}

func parseBoardDeleteCmd(parent *boardCmd, args []string) (*boardDeleteCmd, error) {
	c := &boardDeleteCmd{boardCmd: parent}
	c.fs = newFlagSet("delete")
	c.fs.IntVar(&c.ID, "id", 0, "board id")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	if c.ID == 0 && c.fs.NArg() > 0 {
		_, _ = fmt.Sscan(c.fs.Arg(0), &c.ID)
	}

	return c, nil
}

func (c *boardDeleteCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	if err := queries.DeleteImageBoard(ctx, int32(c.ID)); err != nil {
		return fmt.Errorf("delete board: %w", err)
	}
	return nil
}
