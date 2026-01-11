package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"math"

	"github.com/arran4/goa4web/internal/db"
)

// userDeactivateCmd implements "user deactivate".
type userDeactivateCmd struct {
	*userCmd
	fs       *flag.FlagSet
	Username string
}

func parseUserDeactivateCmd(parent *userCmd, args []string) (*userDeactivateCmd, error) {
	c := &userDeactivateCmd{userCmd: parent}
	fs, _, err := parseFlags("deactivate", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.Username, "username", "", "username")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

func scrubText(s string) string {
	if s == "" {
		return s
	}
	return randomString(len(s))
}

func (c *userDeactivateCmd) Run() error {
	if c.Username == "" {
		return fmt.Errorf("username required")
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	u, err := queries.SystemGetUserByUsername(ctx, sql.NullString{String: c.Username, Valid: true})
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	deactivated, err := queries.AdminIsUserDeactivated(ctx, u.Idusers)
	if err != nil {
		return fmt.Errorf("check deactivated: %w", err)
	}
	if deactivated {
		return fmt.Errorf("user already deactivated")
	}
	c.rootCmd.Verbosef("deactivating user %s", c.Username)
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	qtx := queries.WithTx(tx)
	if err := qtx.AdminArchiveUser(ctx, u.Idusers); err != nil {
		tx.Rollback()
		return fmt.Errorf("archive user: %w", err)
	}
	newName := randomString(16)
	if err := qtx.AdminScrubUser(ctx, db.AdminScrubUserParams{Username: sql.NullString{String: newName, Valid: true}, Idusers: u.Idusers}); err != nil {
		tx.Rollback()
		return fmt.Errorf("scrub user: %w", err)
	}
	if err := qtx.AdminScrubUserEmails(ctx, u.Idusers); err != nil {
		tx.Rollback()
		return fmt.Errorf("scrub user emails: %w", err)
	}
	if err := qtx.AdminScrubUserPasswords(ctx, u.Idusers); err != nil {
		tx.Rollback()
		return fmt.Errorf("scrub user passwords: %w", err)
	}
	comments, err := qtx.AdminGetAllCommentsByUser(ctx, u.Idusers)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("list comments: %w", err)
	}
	for _, cm := range comments {
		deactivated, err := qtx.AdminIsCommentDeactivated(ctx, cm.Idcomments)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("check comment deactivated: %w", err)
		}
		if deactivated {
			tx.Rollback()
			return fmt.Errorf("comment %d already deactivated", cm.Idcomments)
		}
		if err := qtx.AdminArchiveComment(ctx, db.AdminArchiveCommentParams{
			Idcomments:    cm.Idcomments,
			ForumthreadID: cm.ForumthreadID,
			UsersIdusers:  cm.UsersIdusers,
			LanguageID:    cm.LanguageID,
			Written:       cm.Written,
			Text:          cm.Text,
			Timezone:      cm.Timezone,
		}); err != nil {
			tx.Rollback()
			return fmt.Errorf("archive comment: %w", err)
		}
		scrub := scrubText(cm.Text.String)
		if err := qtx.AdminScrubComment(ctx, db.AdminScrubCommentParams{Text: sql.NullString{String: scrub, Valid: true}, Idcomments: cm.Idcomments}); err != nil {
			tx.Rollback()
			return fmt.Errorf("scrub comment: %w", err)
		}
	}
	writings, err := qtx.AdminGetAllWritingsByAuthor(ctx, u.Idusers)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("list writings: %w", err)
	}
	for _, w := range writings {
		deactivated, err := qtx.AdminIsWritingDeactivated(ctx, w.Idwriting)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("check writing deactivated: %w", err)
		}
		if deactivated {
			tx.Rollback()
			return fmt.Errorf("writing %d already deactivated", w.Idwriting)
		}
		if err := qtx.AdminArchiveWriting(ctx, db.AdminArchiveWritingParams{
			Idwriting:         w.Idwriting,
			UsersIdusers:      w.UsersIdusers,
			ForumthreadID:     w.ForumthreadID,
			LanguageID:        w.LanguageID,
			WritingCategoryID: w.WritingCategoryID,
			Title:             w.Title,
			Published:         w.Published,
			Timezone:          w.Timezone,
			Writing:           w.Writing,
			Abstract:          w.Abstract,
			Private:           w.Private,
		}); err != nil {
			tx.Rollback()
			return fmt.Errorf("archive writing: %w", err)
		}
		if err := qtx.AdminScrubWriting(ctx, db.AdminScrubWritingParams{
			Title:     sql.NullString{String: scrubText(w.Title.String), Valid: w.Title.Valid},
			Writing:   sql.NullString{String: scrubText(w.Writing.String), Valid: w.Writing.Valid},
			Abstract:  sql.NullString{String: scrubText(w.Abstract.String), Valid: w.Abstract.Valid},
			Idwriting: w.Idwriting,
		}); err != nil {
			tx.Rollback()
			return fmt.Errorf("scrub writing: %w", err)
		}
	}
	blogs, err := qtx.AdminGetAllBlogEntriesByUser(ctx, u.Idusers)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("list blogs: %w", err)
	}
	for _, b := range blogs {
		deactivated, err := qtx.AdminIsBlogDeactivated(ctx, b.Idblogs)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("check blog deactivated: %w", err)
		}
		if deactivated {
			tx.Rollback()
			return fmt.Errorf("blog %d already deactivated", b.Idblogs)
		}
		var threadID int32
		if b.ForumthreadID.Valid {
			threadID = b.ForumthreadID.Int32
		}
		if err := qtx.AdminArchiveBlog(ctx, db.AdminArchiveBlogParams{
			Idblogs:       b.Idblogs,
			ForumthreadID: threadID,
			UsersIdusers:  b.UsersIdusers,
			LanguageID:    b.LanguageID,
			Blog:          b.Blog,
			Written:       sql.NullTime{Time: b.Written, Valid: true},
			Timezone:      b.Timezone,
		}); err != nil {
			tx.Rollback()
			return fmt.Errorf("archive blog: %w", err)
		}
		if err := qtx.AdminScrubBlog(ctx, db.AdminScrubBlogParams{Blog: sql.NullString{String: scrubText(b.Blog.String), Valid: b.Blog.Valid}, Idblogs: b.Idblogs}); err != nil {
			tx.Rollback()
			return fmt.Errorf("scrub blog: %w", err)
		}
	}
	imgs, err := qtx.GetImagePostsByUserDescending(ctx, db.GetImagePostsByUserDescendingParams{UsersIdusers: u.Idusers, Limit: math.MaxInt32, Offset: 0})
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("list images: %w", err)
	}
	for _, img := range imgs {
		deactivated, err := qtx.AdminIsImagepostDeactivated(ctx, img.Idimagepost)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("check imagepost deactivated: %w", err)
		}
		if deactivated {
			tx.Rollback()
			return fmt.Errorf("imagepost %d already deactivated", img.Idimagepost)
		}
		if err := qtx.AdminArchiveImagepost(ctx, db.AdminArchiveImagepostParams{
			Idimagepost:            img.Idimagepost,
			ForumthreadID:          img.ForumthreadID,
			UsersIdusers:           img.UsersIdusers,
			ImageboardIdimageboard: img.ImageboardIdimageboard,
			Posted:                 img.Posted,
			Timezone:               img.Timezone,
			Description:            img.Description,
			Thumbnail:              img.Thumbnail,
			Fullimage:              img.Fullimage,
			FileSize:               img.FileSize,
			Approved:               sql.NullBool{Bool: img.Approved, Valid: true},
		}); err != nil {
			tx.Rollback()
			return fmt.Errorf("archive imagepost: %w", err)
		}
		if err := qtx.AdminScrubImagepost(ctx, img.Idimagepost); err != nil {
			tx.Rollback()
			return fmt.Errorf("scrub imagepost: %w", err)
		}
	}
	links, err := qtx.GetLinkerItemsByUserDescending(ctx, db.GetLinkerItemsByUserDescendingParams{AuthorID: u.Idusers, Limit: math.MaxInt32, Offset: 0})
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("list links: %w", err)
	}
	for _, l := range links {
		deactivated, err := qtx.AdminIsLinkDeactivated(ctx, l.ID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("check link deactivated: %w", err)
		}
		if deactivated {
			tx.Rollback()
			return fmt.Errorf("link %d already deactivated", l.ID)
		}
		if err := qtx.AdminArchiveLink(ctx, db.AdminArchiveLinkParams{
			ID:          l.ID,
			LanguageID:  l.LanguageID,
			AuthorID:    l.AuthorID,
			CategoryID:  l.CategoryID,
			ThreadID:    l.ThreadID,
			Title:       l.Title,
			Url:         l.Url,
			Description: l.Description,
			Listed:      l.Listed,
			Timezone:    l.Timezone,
		}); err != nil {
			tx.Rollback()
			return fmt.Errorf("archive link: %w", err)
		}
		if err := qtx.AdminScrubLink(ctx, db.AdminScrubLinkParams{Title: sql.NullString{String: scrubText(l.Title.String), Valid: l.Title.Valid}, ID: l.ID}); err != nil {
			tx.Rollback()
			return fmt.Errorf("scrub link: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	c.rootCmd.Infof("deactivated user %s", c.Username)
	return nil
}
