package main

import (
	"context"
	"flag"
	"fmt"
	"math"

	"github.com/arran4/goa4web/internal/db"
)

// commentActivateCmd implements "comment activate".
type commentActivateCmd struct {
	*commentCmd
	fs *flag.FlagSet
	ID int
}

func parseCommentActivateCmd(parent *commentCmd, args []string) (*commentActivateCmd, error) {
	c := &commentActivateCmd{commentCmd: parent}
	c.fs = newFlagSet("activate")
	c.fs.IntVar(&c.ID, "id", 0, "comment id")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *commentActivateCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	deactivated, err := queries.AdminIsCommentDeactivated(ctx, int32(c.ID))
	if err != nil {
		return fmt.Errorf("check deactivated: %w", err)
	}
	if !deactivated {
		return fmt.Errorf("comment not deactivated")
	}
	rows, err := queries.AdminListDeactivatedComments(ctx, db.AdminListDeactivatedCommentsParams{Limit: math.MaxInt32, Offset: 0})
	if err != nil {
		return fmt.Errorf("list deactivated: %w", err)
	}
	var found *db.AdminListDeactivatedCommentsRow
	for _, r := range rows {
		if int(r.Idcomments) == c.ID {
			found = r
			break
		}
	}
	if found == nil {
		return fmt.Errorf("comment %d not found", c.ID)
	}
	if err := queries.AdminRestoreComment(ctx, db.AdminRestoreCommentParams{Text: found.Text, Idcomments: found.Idcomments}); err != nil {
		return fmt.Errorf("restore comment: %w", err)
	}
	if err := queries.AdminMarkCommentRestored(ctx, found.Idcomments); err != nil {
		return fmt.Errorf("mark restored: %w", err)
	}
	return nil
}
