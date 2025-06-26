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
	fs   *flag.FlagSet
	ID   int
	args []string
}

func parseBoardDeleteCmd(parent *boardCmd, args []string) (*boardDeleteCmd, error) {
	c := &boardDeleteCmd{boardCmd: parent}
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	fs.IntVar(&c.ID, "id", 0, "board id")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if c.ID == 0 && fs.NArg() > 0 {
		_, err := fmt.Sscan(fs.Arg(0), &c.ID)
		if err == nil {
			c.args = fs.Args()[1:]
		} else {
			c.args = fs.Args()
		}
	} else {
		c.args = fs.Args()
	}
	c.fs = fs
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
