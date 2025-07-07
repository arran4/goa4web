package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// userActivateCmd implements "user activate" to restore a deactivated user.
type userActivateCmd struct {
	*userCmd
	fs       *flag.FlagSet
	ID       int
	Username string
	args     []string
}

func parseUserActivateCmd(parent *userCmd, args []string) (*userActivateCmd, error) {
	c := &userActivateCmd{userCmd: parent}
	fs, rest, err := parseFlags("activate", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.ID, "id", 0, "user id")
		fs.StringVar(&c.Username, "username", "", "username")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	c.args = rest
	return c, nil
}

func (c *userActivateCmd) Run() error {
	if c.ID == 0 && c.Username == "" {
		return fmt.Errorf("id or username required")
	}
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	if c.ID == 0 {
		u, err := queries.GetUserByUsername(ctx, sql.NullString{String: c.Username, Valid: true})
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}
		c.ID = int(u.Idusers)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	qtx := queries.WithTx(tx)
	if err := qtx.RestoreUser(ctx, int32(c.ID)); err != nil {
		tx.Rollback()
		return fmt.Errorf("restore user: %w", err)
	}
	rows, err := qtx.PendingDeactivatedComments(ctx, int32(c.ID))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("select comments: %w", err)
	}
	for _, row := range rows {
		if err := qtx.RestoreComment(ctx, dbpkg.RestoreCommentParams{Text: row.Text, Idcomments: row.Idcomments}); err != nil {
			tx.Rollback()
			return fmt.Errorf("restore comment: %w", err)
		}
		if err := qtx.MarkCommentRestored(ctx, row.Idcomments); err != nil {
			tx.Rollback()
			return fmt.Errorf("mark comment restored: %w", err)
		}
	}

	rowsW, err := qtx.PendingDeactivatedWritings(ctx, int32(c.ID))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("select writings: %w", err)
	}
	for _, w := range rowsW {
		if err := qtx.RestoreWriting(ctx, dbpkg.RestoreWritingParams{
			Title:     w.Title,
			Writting:  w.Writting,
			Abstract:  w.Abstract,
			Private:   w.Private,
			Idwriting: w.Idwriting,
		}); err != nil {
			tx.Rollback()
			return fmt.Errorf("restore writing: %w", err)
		}
		if err := qtx.MarkWritingRestored(ctx, w.Idwriting); err != nil {
			tx.Rollback()
			return fmt.Errorf("mark writing restored: %w", err)
		}
	}

	rowsB, err := qtx.PendingDeactivatedBlogs(ctx, int32(c.ID))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("select blogs: %w", err)
	}
	for _, b := range rowsB {
		if err := qtx.RestoreBlog(ctx, dbpkg.RestoreBlogParams{Blog: b.Blog, Idblogs: b.Idblogs}); err != nil {
			tx.Rollback()
			return fmt.Errorf("restore blog: %w", err)
		}
		if err := qtx.MarkBlogRestored(ctx, b.Idblogs); err != nil {
			tx.Rollback()
			return fmt.Errorf("mark blog restored: %w", err)
		}
	}

	rowsI, err := qtx.PendingDeactivatedImageposts(ctx, int32(c.ID))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("select imageposts: %w", err)
	}
	for _, img := range rowsI {
		if err := qtx.RestoreImagepost(ctx, dbpkg.RestoreImagepostParams{Description: img.Description, Thumbnail: img.Thumbnail, Fullimage: img.Fullimage, Idimagepost: img.Idimagepost}); err != nil {
			tx.Rollback()
			return fmt.Errorf("restore imagepost: %w", err)
		}
		if err := qtx.MarkImagepostRestored(ctx, img.Idimagepost); err != nil {
			tx.Rollback()
			return fmt.Errorf("mark imagepost restored: %w", err)
		}
	}

	rowsL, err := qtx.PendingDeactivatedLinks(ctx, int32(c.ID))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("select links: %w", err)
	}
	for _, l := range rowsL {
		if err := qtx.RestoreLink(ctx, dbpkg.RestoreLinkParams{Title: l.Title, Url: l.Url, Description: l.Description, Idlinker: l.Idlinker}); err != nil {
			tx.Rollback()
			return fmt.Errorf("restore link: %w", err)
		}
		if err := qtx.MarkLinkRestored(ctx, l.Idlinker); err != nil {
			tx.Rollback()
			return fmt.Errorf("mark link restored: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	if c.rootCmd.Verbosity > 0 {
		fmt.Printf("restored user %d\n", c.ID)
	}
	return nil
}
