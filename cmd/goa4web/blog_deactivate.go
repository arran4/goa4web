package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// blogDeactivateCmd implements "blog deactivate".
type blogDeactivateCmd struct {
	*blogCmd
	fs *flag.FlagSet
	ID int
}

func parseBlogDeactivateCmd(parent *blogCmd, args []string) (*blogDeactivateCmd, error) {
	c := &blogDeactivateCmd{blogCmd: parent}
	c.fs = newFlagSet("deactivate")
	c.fs.IntVar(&c.ID, "id", 0, "blog id")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *blogDeactivateCmd) Run() error {
	if c.ID == 0 {
		return fmt.Errorf("id required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(db)
	b, err := queries.GetBlogEntryForListerByID(ctx, db.GetBlogEntryForListerByIDParams{
		ListerID: 0,
		ID:       int32(c.ID),
		UserID:   sql.NullInt32{},
	})
	if err != nil {
		return fmt.Errorf("fetch blog: %w", err)
	}
	var threadID int32
	if b.ForumthreadID.Valid {
		threadID = b.ForumthreadID.Int32
	}
	if err := queries.ArchiveBlog(ctx, db.ArchiveBlogParams{
		Idblogs:            b.Idblogs,
		ForumthreadID:      threadID,
		UsersIdusers:       b.UsersIdusers,
		LanguageIdlanguage: b.LanguageIdlanguage,
		Blog:               b.Blog,
		Written:            sql.NullTime{Time: b.Written, Valid: true},
	}); err != nil {
		return fmt.Errorf("archive blog: %w", err)
	}
	if err := queries.ScrubBlog(ctx, db.ScrubBlogParams{Blog: sql.NullString{String: "", Valid: true}, Idblogs: b.Idblogs}); err != nil {
		return fmt.Errorf("scrub blog: %w", err)
	}
	return nil
}
