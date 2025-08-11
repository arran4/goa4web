package main

import (
	"context"
	"flag"
	"fmt"
	"math"

	"github.com/arran4/goa4web/internal/db"
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
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
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
	if err := queries.AdminRestoreLink(ctx, db.AdminRestoreLinkParams{Title: found.Title, Url: found.Url, Description: found.Description, ID: found.ID}); err != nil {
		return fmt.Errorf("restore link: %w", err)
	}
	if err := queries.AdminMarkLinkRestored(ctx, found.ID); err != nil {
		return fmt.Errorf("mark restored: %w", err)
	}
	return nil
}
