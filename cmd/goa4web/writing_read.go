package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"strconv"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// writingReadCmd implements "writing read".
type writingReadCmd struct {
	*writingCmd
	fs   *flag.FlagSet
	ID   int
	args []string
}

func parseWritingReadCmd(parent *writingCmd, args []string) (*writingReadCmd, error) {
	c := &writingReadCmd{writingCmd: parent}
	fs := flag.NewFlagSet("read", flag.ContinueOnError)
	fs.IntVar(&c.ID, "id", 0, "writing id")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	if c.ID == 0 && len(c.args) > 0 {
		if id, err := strconv.Atoi(c.args[0]); err == nil {
			c.ID = id
			c.args = c.args[1:]
		}
	}
	return c, nil
}

func (c *writingReadCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	row, err := queries.GetWritingByIdForUserDescendingByPublishedDate(ctx, dbpkg.GetWritingByIdForUserDescendingByPublishedDateParams{
		UserID:    0,
		ViewerID:  sql.NullInt32{Valid: false},
		Idwriting: int32(c.ID),
	})
	if err != nil {
		return fmt.Errorf("get writing: %w", err)
	}
	fmt.Printf("Title: %s\n", row.Title.String)
	fmt.Println(row.Writing.String)
	return nil
}
