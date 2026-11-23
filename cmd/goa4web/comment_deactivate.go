package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/searchindex"
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
	conn, err := c.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := newCLIQueries(conn, c.cfg)
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
	return withCLIJobTx(ctx, conn, c.cfg, func(qtx db.Querier, svc *jobs.Service) error {
		if err := qtx.AdminArchiveComment(ctx, db.AdminArchiveCommentParams{
			Idcomments:    cm.Idcomments,
			ForumthreadID: cm.ForumthreadID,
			UsersIdusers:  cm.UsersIdusers,
			LanguageID:    cm.LanguageID,
			Written:       cm.Written,
			Text:          cm.Text,
			Timezone:      cm.Timezone,
		}); err != nil {
			return fmt.Errorf("archive comment: %w", err)
		}
		if err := qtx.AdminScrubComment(ctx, db.AdminScrubCommentParams{Text: sql.NullString{String: "", Valid: true}, Idcomments: cm.Idcomments}); err != nil {
			return fmt.Errorf("scrub comment: %w", err)
		}
		if err := svc.RecordSearchDirty(ctx, qtx, searchindex.TypeComment, cm.Idcomments); err != nil {
			return fmt.Errorf("record comment search invalidation: %w", err)
		}
		return nil
	})
}
