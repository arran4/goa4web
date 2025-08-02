package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// boardUpdateCmd implements "board update".
type boardUpdateCmd struct {
	*boardCmd
	fs             *flag.FlagSet
	ID             int
	Parent         int
	Name           string
	Description    string
	ApprovalNeeded bool
}

func parseBoardUpdateCmd(parent *boardCmd, args []string) (*boardUpdateCmd, error) {
	c := &boardUpdateCmd{boardCmd: parent}
	fs, _, err := parseFlags("update", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.ID, "id", 0, "board id")
		fs.IntVar(&c.Parent, "parent", 0, "parent board id")
		fs.StringVar(&c.Name, "name", "", "board name")
		fs.StringVar(&c.Description, "description", "", "board description")
		fs.BoolVar(&c.ApprovalNeeded, "approval", false, "require approval")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func (c *boardUpdateCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(db)
	err = queries.UpdateImageBoard(ctx, db.UpdateImageBoardParams{
		Title:                  sql.NullString{String: c.Name, Valid: c.Name != ""},
		Description:            sql.NullString{String: c.Description, Valid: c.Description != ""},
		ImageboardIdimageboard: int32(c.Parent),
		ApprovalRequired:       c.ApprovalNeeded,
		Idimageboard:           int32(c.ID),
	})
	if err != nil {
		return fmt.Errorf("update board: %w", err)
	}
	return nil
}
