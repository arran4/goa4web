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
}

func parseUserActivateCmd(parent *userCmd, args []string) (*userActivateCmd, error) {
	c := &userActivateCmd{userCmd: parent}
	fs, _, err := parseFlags("activate", args, func(fs *flag.FlagSet) {
		fs.IntVar(&c.ID, "id", 0, "user id")
		fs.StringVar(&c.Username, "username", "", "username")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
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
	c.rootCmd.Verbosef("restoring user %d", c.ID)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	qtx := queries.WithTx(tx)
	if err := qtx.AdminRestoreUser(ctx, int32(c.ID)); err != nil {
		tx.Rollback()
		return fmt.Errorf("restore user: %w", err)
	}
	rows, err := qtx.AdminPendingDeactivatedComments(ctx, int32(c.ID))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("select comments: %w", err)
	}
	for _, row := range rows {
		if err := qtx.AdminRestoreComment(ctx, dbpkg.AdminRestoreCommentParams{Text: row.Text, Idcomments: row.Idcomments}); err != nil {
			tx.Rollback()
			return fmt.Errorf("restore comment: %w", err)
		}
		if err := qtx.AdminMarkCommentRestored(ctx, row.Idcomments); err != nil {
			tx.Rollback()
			return fmt.Errorf("mark comment restored: %w", err)
		}
	}

	rowsW, err := qtx.AdminPendingDeactivatedWritings(ctx, int32(c.ID))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("select writings: %w", err)
	}
	for _, w := range rowsW {
		if err := qtx.AdminRestoreWriting(ctx, dbpkg.AdminRestoreWritingParams{
			Title:     w.Title,
			Writing:   w.Writing,
			Abstract:  w.Abstract,
			Private:   w.Private,
			Idwriting: w.Idwriting,
		}); err != nil {
			tx.Rollback()
			return fmt.Errorf("restore writing: %w", err)
		}
		if err := qtx.AdminMarkWritingRestored(ctx, w.Idwriting); err != nil {
			tx.Rollback()
			return fmt.Errorf("mark writing restored: %w", err)
		}
	}

	rowsB, err := qtx.AdminPendingDeactivatedBlogs(ctx, int32(c.ID))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("select blogs: %w", err)
	}
	for _, b := range rowsB {
		if err := qtx.AdminRestoreBlog(ctx, dbpkg.AdminRestoreBlogParams{Blog: b.Blog, Idblogs: b.Idblogs}); err != nil {
			tx.Rollback()
			return fmt.Errorf("restore blog: %w", err)
		}
		if err := qtx.AdminMarkBlogRestored(ctx, b.Idblogs); err != nil {
			tx.Rollback()
			return fmt.Errorf("mark blog restored: %w", err)
		}
	}

	rowsI, err := qtx.AdminPendingDeactivatedImageposts(ctx, int32(c.ID))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("select imageposts: %w", err)
	}
	for _, img := range rowsI {
		if err := qtx.AdminRestoreImagepost(ctx, dbpkg.AdminRestoreImagepostParams{Description: img.Description, Thumbnail: img.Thumbnail, Fullimage: img.Fullimage, Idimagepost: img.Idimagepost}); err != nil {
			tx.Rollback()
			return fmt.Errorf("restore imagepost: %w", err)
		}
		if err := qtx.AdminMarkImagepostRestored(ctx, img.Idimagepost); err != nil {
			tx.Rollback()
			return fmt.Errorf("mark imagepost restored: %w", err)
		}
	}

	rowsL, err := qtx.AdminPendingDeactivatedLinks(ctx, int32(c.ID))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("select links: %w", err)
	}
	for _, l := range rowsL {
		if err := qtx.AdminRestoreLink(ctx, dbpkg.AdminRestoreLinkParams{Title: l.Title, Url: l.Url, Description: l.Description, Idlinker: l.Idlinker}); err != nil {
			tx.Rollback()
			return fmt.Errorf("restore link: %w", err)
		}
		if err := qtx.AdminMarkLinkRestored(ctx, l.Idlinker); err != nil {
			tx.Rollback()
			return fmt.Errorf("mark link restored: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	c.rootCmd.Infof("restored user %d", c.ID)
	return nil
}
