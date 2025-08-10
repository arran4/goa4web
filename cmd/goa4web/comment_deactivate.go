package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// commentDeactivateCmd implements "comment deactivate".
type commentDeactivateCmd struct {
	*commentCmd
	fs *flag.FlagSet
	ID int
}

func parseCommentDeactivateCmd(parent *commentCmd, args []string) (*commentDeactivateCmd, error) {
	c := &commentDeactivateCmd{commentCmd: parent}
	c.fs = newFlagSet("deactivate")
	c.fs.IntVar(&c.ID, "id", 0, "comment id")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *commentDeactivateCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	cm, err := queries.GetCommentById(ctx, int32(c.ID))
	if err != nil {
		return fmt.Errorf("fetch comment: %w", err)
	}
	deactivated, err := queries.AdminIsCommentDeactivated(ctx, cm.Idcomments)
	if err != nil {
		return fmt.Errorf("check deactivated: %w", err)
	}
	if deactivated {
		return fmt.Errorf("comment already deactivated")
	}
	var langID int32
	if cm.LanguageIdlanguage.Valid {
		langID = cm.LanguageIdlanguage.Int32
	}
	if err := queries.AdminArchiveComment(ctx, db.AdminArchiveCommentParams{
		Idcomments:         cm.Idcomments,
		ForumthreadID:      cm.ForumthreadID,
		UsersIdusers:       cm.UsersIdusers,
		LanguageIdlanguage: langID,
		Written:            cm.Written,
		Text:               cm.Text,
	}); err != nil {
		return fmt.Errorf("archive comment: %w", err)
	}
	if err := queries.AdminScrubComment(ctx, db.AdminScrubCommentParams{Text: sql.NullString{String: "", Valid: true}, Idcomments: cm.Idcomments}); err != nil {
		return fmt.Errorf("scrub comment: %w", err)
	}
	return nil
}
