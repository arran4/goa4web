package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// boardListCmd implements "board list".
type boardListCmd struct {
	*boardCmd
	fs     *flag.FlagSet
	limit  int
	offset int
}

func parseBoardListCmd(parent *boardCmd, args []string) (*boardListCmd, error) {
	c := &boardListCmd{boardCmd: parent}
	c.fs = newFlagSet("list")
	c.fs.IntVar(&c.limit, "limit", 100, "number of boards to list")
	c.fs.IntVar(&c.offset, "offset", 0, "offset into board list")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *boardListCmd) Run() error {
	sqldb, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(sqldb)
	rows, err := queries.AdminListBoards(ctx, db.AdminListBoardsParams{Limit: int32(c.limit), Offset: int32(c.offset)})
	if err != nil {
		return fmt.Errorf("list boards: %w", err)
	}
	for _, b := range rows {
		fmt.Printf("%d\t%s\t%s\n", b.Idimageboard, b.Title.String, b.Description.String)
	}
	return nil
}
