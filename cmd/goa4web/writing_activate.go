package main

import (
	"context"
	"flag"
	"fmt"
	"math"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/searchindex"
)

// writingActivateCmd implements "writing activate".
type writingActivateCmd struct {
	*writingCmd
	fs *flag.FlagSet
	ID int
}

func parseWritingActivateCmd(parent *writingCmd, args []string) (*writingActivateCmd, error) {
	c := &writingActivateCmd{writingCmd: parent}
	c.fs = newFlagSet("activate")
	c.fs.IntVar(&c.ID, "id", 0, "writing id")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *writingActivateCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	conn, err := c.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := newCLIQueries(conn, c.cfg)
	deactivated, err := queries.AdminIsWritingDeactivated(ctx, int32(c.ID))
	if err != nil {
		return fmt.Errorf("check deactivated: %w", err)
	}
	if !deactivated {
		return fmt.Errorf("writing not deactivated")
	}
	rows, err := queries.AdminListDeactivatedWritings(ctx, db.AdminListDeactivatedWritingsParams{Limit: math.MaxInt32, Offset: 0})
	if err != nil {
		return fmt.Errorf("list deactivated: %w", err)
	}
	var found *db.AdminListDeactivatedWritingsRow
	for _, r := range rows {
		if int(r.Idwriting) == c.ID {
			found = r
			break
		}
	}
	if found == nil {
		return fmt.Errorf("writing %d not found", c.ID)
	}
	return withCLIJobTx(ctx, conn, c.cfg, func(qtx db.Querier, svc *jobs.Service) error {
		if err := qtx.AdminRestoreWriting(ctx, db.AdminRestoreWritingParams{Title: found.Title, Writing: found.Writing, Abstract: found.Abstract, Private: found.Private, Idwriting: found.Idwriting}); err != nil {
			return fmt.Errorf("restore writing: %w", err)
		}
		if err := qtx.AdminMarkWritingRestored(ctx, found.Idwriting); err != nil {
			return fmt.Errorf("mark restored: %w", err)
		}
		if err := svc.RecordSearchDirty(ctx, qtx, searchindex.TypeWriting, found.Idwriting); err != nil {
			return fmt.Errorf("record writing search dirty: %w", err)
		}
		return nil
	})
}
