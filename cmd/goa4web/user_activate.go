package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"math"

	"github.com/arran4/goa4web/internal/db"
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
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	if c.ID == 0 {
		u, err := queries.SystemGetUserByUsername(ctx, sql.NullString{String: c.Username, Valid: true})
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}
		c.ID = int(u.Idusers)
	}
	c.rootCmd.Verbosef("restoring user %d", c.ID)
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	qtx := queries.WithTx(tx)
	if err := qtx.AdminRestoreUser(ctx, int32(c.ID)); err != nil {
		tx.Rollback()
		return fmt.Errorf("restore user: %w", err)
	}
	if err := qtx.AdminRestoreUserEmail(ctx, int32(c.ID)); err != nil {
		tx.Rollback()
		return fmt.Errorf("restore user email: %w", err)
	}
	if err := qtx.AdminRestoreUserPassword(ctx, int32(c.ID)); err != nil {
		tx.Rollback()
		return fmt.Errorf("restore user password: %w", err)
	}
	if err := qtx.AdminMarkUserRestored(ctx, int32(c.ID)); err != nil {
		tx.Rollback()
		return fmt.Errorf("mark user restored: %w", err)
	}
	rows, err := qtx.AdminListPendingDeactivatedComments(ctx, db.AdminListPendingDeactivatedCommentsParams{UsersIdusers: int32(c.ID), Limit: math.MaxInt32, Offset: 0})
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("select comments: %w", err)
	}
	for _, row := range rows {
		if err := qtx.AdminRestoreComment(ctx, db.AdminRestoreCommentParams{Text: row.Text, Idcomments: row.Idcomments}); err != nil {
			tx.Rollback()
			return fmt.Errorf("restore comment: %w", err)
		}
		if err := qtx.AdminMarkCommentRestored(ctx, row.Idcomments); err != nil {
			tx.Rollback()
			return fmt.Errorf("mark comment restored: %w", err)
		}
	}

	rowsW, err := qtx.AdminListPendingDeactivatedWritings(ctx, db.AdminListPendingDeactivatedWritingsParams{UsersIdusers: int32(c.ID), Limit: math.MaxInt32, Offset: 0})
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("select writings: %w", err)
	}
	for _, w := range rowsW {
		if err := qtx.AdminRestoreWriting(ctx, db.AdminRestoreWritingParams{
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

	rowsB, err := qtx.AdminListPendingDeactivatedBlogs(ctx, db.AdminListPendingDeactivatedBlogsParams{UsersIdusers: int32(c.ID), Limit: math.MaxInt32, Offset: 0})
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("select blogs: %w", err)
	}
	for _, b := range rowsB {
		if err := qtx.AdminRestoreBlog(ctx, db.AdminRestoreBlogParams{Blog: b.Blog, Idblogs: b.Idblogs}); err != nil {
			tx.Rollback()
			return fmt.Errorf("restore blog: %w", err)
		}
		if err := qtx.AdminMarkBlogRestored(ctx, b.Idblogs); err != nil {
			tx.Rollback()
			return fmt.Errorf("mark blog restored: %w", err)
		}
	}

	rowsI, err := qtx.AdminListPendingDeactivatedImageposts(ctx, db.AdminListPendingDeactivatedImagepostsParams{UsersIdusers: int32(c.ID), Limit: math.MaxInt32, Offset: 0})
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("select imageposts: %w", err)
	}
	for _, img := range rowsI {
		if err := qtx.AdminRestoreImagepost(ctx, db.AdminRestoreImagepostParams{Description: img.Description, Thumbnail: img.Thumbnail, Fullimage: img.Fullimage, Idimagepost: img.Idimagepost}); err != nil {
			tx.Rollback()
			return fmt.Errorf("restore imagepost: %w", err)
		}
		if err := qtx.AdminMarkImagepostRestored(ctx, img.Idimagepost); err != nil {
			tx.Rollback()
			return fmt.Errorf("mark imagepost restored: %w", err)
		}
	}

	rowsL, err := qtx.AdminListPendingDeactivatedLinks(ctx, db.AdminListPendingDeactivatedLinksParams{AuthorID: int32(c.ID), Limit: math.MaxInt32, Offset: 0})
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("select links: %w", err)
	}
	for _, l := range rowsL {
		if err := qtx.AdminRestoreLink(ctx, db.AdminRestoreLinkParams{Title: l.Title, Url: l.Url, Description: l.Description, ID: l.ID}); err != nil {
			tx.Rollback()
			return fmt.Errorf("restore link: %w", err)
		}
		if err := qtx.AdminMarkLinkRestored(ctx, l.ID); err != nil {
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
