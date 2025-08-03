package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// boardCreateCmd implements "board create".
type boardCreateCmd struct {
	*boardCmd
	fs          *flag.FlagSet
	Parent      int
	Name        string
	Description string
}

func parseBoardCreateCmd(parent *boardCmd, args []string) (*boardCreateCmd, error) {
	c := &boardCreateCmd{boardCmd: parent}
	c.fs = newFlagSet("create")
	c.fs.IntVar(&c.Parent, "parent", 0, "parent board id")
	c.fs.StringVar(&c.Name, "name", "", "board name")
	c.fs.StringVar(&c.Description, "description", "", "board description")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *boardCreateCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	err = queries.AdminCreateImageBoard(ctx, db.AdminCreateImageBoardParams{
		ImageboardIdimageboard: int32(c.Parent),
		Title:                  sql.NullString{String: c.Name, Valid: c.Name != ""},
		Description:            sql.NullString{String: c.Description, Valid: c.Description != ""},
		ApprovalRequired:       false,
	})
	if err != nil {
		return fmt.Errorf("create board: %w", err)
	}
	return nil
}
