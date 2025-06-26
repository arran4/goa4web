package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// boardCreateCmd implements "board create".
type boardCreateCmd struct {
	*boardCmd
	fs          *flag.FlagSet
	Parent      int
	Name        string
	Description string
	args        []string
}

func parseBoardCreateCmd(parent *boardCmd, args []string) (*boardCreateCmd, error) {
	c := &boardCreateCmd{boardCmd: parent}
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	fs.IntVar(&c.Parent, "parent", 0, "parent board id")
	fs.StringVar(&c.Name, "name", "", "board name")
	fs.StringVar(&c.Description, "description", "", "board description")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = fs.Args()
	return c, nil
}

func (c *boardCreateCmd) Run() error {
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	err = queries.CreateImageBoard(ctx, dbpkg.CreateImageBoardParams{
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
