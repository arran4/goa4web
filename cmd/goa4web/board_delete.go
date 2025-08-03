package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
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
		if _, err := fmt.Sscan(c.fs.Arg(0), &c.ID); err != nil {
			return nil, fmt.Errorf("parse id: %w", err)
		}
	}

	return c, nil
}

func (c *boardDeleteCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	if err := queries.DeleteImageBoard(ctx, int32(c.ID)); err != nil {
		return fmt.Errorf("delete board: %w", err)
	}
	return nil
}
