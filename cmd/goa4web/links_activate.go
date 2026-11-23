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

// linksActivateCmd implements "links activate".
type linksActivateCmd struct {
	*linksCmd
	fs *flag.FlagSet
	ID int
}

func parseLinksActivateCmd(parent *linksCmd, args []string) (*linksActivateCmd, error) {
	c := &linksActivateCmd{linksCmd: parent}
	c.fs = newFlagSet("activate")
	c.fs.IntVar(&c.ID, "id", 0, "link id")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *linksActivateCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	conn, err := c.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := newCLIQueries(conn, c.cfg)
	deactivated, err := queries.AdminIsLinkDeactivated(ctx, int32(c.ID))
	if err != nil {
		return fmt.Errorf("check deactivated: %w", err)
	}
	if !deactivated {
		return fmt.Errorf("link not deactivated")
	}
	rows, err := queries.AdminListDeactivatedLinks(ctx, db.AdminListDeactivatedLinksParams{Limit: math.MaxInt32, Offset: 0})
	if err != nil {
		return fmt.Errorf("list deactivated: %w", err)
	}
	var found *db.AdminListDeactivatedLinksRow
	for _, r := range rows {
		if int(r.ID) == c.ID {
			found = r
			break
		}
	}
	if found == nil {
		return fmt.Errorf("link %d not found", c.ID)
	}
	return withCLIJobTx(ctx, conn, c.cfg, func(qtx db.Querier, svc *jobs.Service) error {
		if err := qtx.AdminRestoreLink(ctx, db.AdminRestoreLinkParams{Title: found.Title, Url: found.Url, Description: found.Description, ID: found.ID}); err != nil {
			return fmt.Errorf("restore link: %w", err)
		}
		if err := qtx.AdminMarkLinkRestored(ctx, found.ID); err != nil {
			return fmt.Errorf("mark restored: %w", err)
		}
		if err := svc.RecordSearchDirty(ctx, qtx, searchindex.TypeLinker, found.ID); err != nil {
			return fmt.Errorf("record linker search dirty: %w", err)
		}
		return nil
	})
}
