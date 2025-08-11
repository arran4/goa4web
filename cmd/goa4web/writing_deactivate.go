package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// writingDeactivateCmd implements "writing deactivate".
type writingDeactivateCmd struct {
	*writingCmd
	fs *flag.FlagSet
	ID int
}

func parseWritingDeactivateCmd(parent *writingCmd, args []string) (*writingDeactivateCmd, error) {
	c := &writingDeactivateCmd{writingCmd: parent}
	c.fs = newFlagSet("deactivate")
	c.fs.IntVar(&c.ID, "id", 0, "writing id")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *writingDeactivateCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	w, err := queries.GetWritingForListerByID(ctx, db.GetWritingForListerByIDParams{ListerID: 0, Idwriting: int32(c.ID), ListerMatchID: sql.NullInt32{}})
	if err != nil {
		return fmt.Errorf("fetch writing: %w", err)
	}
	deactivated, err := queries.AdminIsWritingDeactivated(ctx, w.Idwriting)
	if err != nil {
		return fmt.Errorf("check deactivated: %w", err)
	}
	if deactivated {
		return fmt.Errorf("writing already deactivated")
	}
	if err := queries.AdminArchiveWriting(ctx, db.AdminArchiveWritingParams{
		Idwriting:         w.Idwriting,
		UsersIdusers:      w.UsersIdusers,
		ForumthreadID:     w.ForumthreadID,
		LanguageID:        w.LanguageID,
		WritingCategoryID: w.WritingCategoryID,
		Title:             w.Title,
		Published:         w.Published,
		Timezone:          w.Timezone,
		Writing:           w.Writing,
		Abstract:          w.Abstract,
		Private:           w.Private,
	}); err != nil {
		return fmt.Errorf("archive writing: %w", err)
	}
	if err := queries.AdminScrubWriting(ctx, db.AdminScrubWritingParams{Title: sql.NullString{String: "", Valid: true}, Writing: sql.NullString{String: "", Valid: true}, Abstract: sql.NullString{String: "", Valid: true}, Idwriting: w.Idwriting}); err != nil {
		return fmt.Errorf("scrub writing: %w", err)
	}
	return nil
}
