package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// linksDeactivateCmd implements "links deactivate".
type linksDeactivateCmd struct {
	*linksCmd
	fs *flag.FlagSet
	ID int
}

func parseLinksDeactivateCmd(parent *linksCmd, args []string) (*linksDeactivateCmd, error) {
	c := &linksDeactivateCmd{linksCmd: parent}
	c.fs = newFlagSet("deactivate")
	c.fs.IntVar(&c.ID, "id", 0, "link id")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *linksDeactivateCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	l, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending(ctx, int32(c.ID))
	if err != nil {
		return fmt.Errorf("fetch link: %w", err)
	}
	deactivated, err := queries.AdminIsLinkDeactivated(ctx, l.ID)
	if err != nil {
		return fmt.Errorf("check deactivated: %w", err)
	}
	if deactivated {
		return fmt.Errorf("link already deactivated")
	}
	if err := queries.AdminArchiveLink(ctx, db.AdminArchiveLinkParams{
		Idlinker:         l.ID,
		LanguageID:       l.LanguageID,
		UsersIdusers:     l.AuthorID,
		LinkerCategoryID: l.CategoryID,
		ForumthreadID:    l.ThreadID,
		Title:            l.Title,
		Url:              l.Url,
		Description:      l.Description,
		Listed:           l.Listed,
	}); err != nil {
		return fmt.Errorf("archive link: %w", err)
	}
	if err := queries.AdminScrubLink(ctx, db.AdminScrubLinkParams{Title: sql.NullString{String: "", Valid: true}, ID: l.ID}); err != nil {
		return fmt.Errorf("scrub link: %w", err)
	}
	return nil
}
