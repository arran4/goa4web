package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"math"

	dbpkg "github.com/arran4/goa4web/internal/db"
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
	db, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := dbpkg.New(db)
	u, err := queries.GetUserByUsername(ctx, sql.NullString{String: c.Username, Valid: true})
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	c.rootCmd.Verbosef("deactivating user %s", c.Username)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	qtx := queries.WithTx(tx)
	if err := qtx.AdminArchiveUser(ctx, u.Idusers); err != nil {
		tx.Rollback()
		return fmt.Errorf("archive user: %w", err)
	}
	newName := randomString(16)
	if err := qtx.AdminScrubUser(ctx, dbpkg.AdminScrubUserParams{Username: sql.NullString{String: newName, Valid: true}, Idusers: u.Idusers}); err != nil {
		tx.Rollback()
		return fmt.Errorf("scrub user: %w", err)
	}
	comments, err := qtx.GetAllCommentsByUser(ctx, u.Idusers)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("list comments: %w", err)
	}
	for _, cm := range comments {
		if err := qtx.AdminArchiveComment(ctx, dbpkg.AdminArchiveCommentParams{
			Idcomments:         cm.Idcomments,
			ForumthreadID:      cm.ForumthreadID,
			UsersIdusers:       cm.UsersIdusers,
			LanguageIdlanguage: cm.LanguageIdlanguage,
			Written:            cm.Written,
			Text:               cm.Text,
		}); err != nil {
			tx.Rollback()
			return fmt.Errorf("archive comment: %w", err)
		}
		scrub := scrubText(cm.Text.String)
		if err := qtx.AdminScrubComment(ctx, dbpkg.AdminScrubCommentParams{Text: sql.NullString{String: scrub, Valid: true}, Idcomments: cm.Idcomments}); err != nil {
			tx.Rollback()
			return fmt.Errorf("scrub comment: %w", err)
		}
	}
	writings, err := qtx.GetAllWritingsByUser(ctx, dbpkg.GetAllWritingsByUserParams{
		ViewerID:      u.Idusers,
		AuthorID:      u.Idusers,
		ViewerMatchID: sql.NullInt32{Int32: u.Idusers, Valid: true},
	})
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("list writings: %w", err)
	}
	for _, w := range writings {
		if err := qtx.AdminArchiveWriting(ctx, dbpkg.AdminArchiveWritingParams{
			Idwriting:          w.Idwriting,
			UsersIdusers:       w.UsersIdusers,
			ForumthreadID:      w.ForumthreadID,
			LanguageIdlanguage: w.LanguageIdlanguage,
			WritingCategoryID:  w.WritingCategoryID,
			Title:              w.Title,
			Published:          w.Published,
			Writing:            w.Writing,
			Abstract:           w.Abstract,
			Private:            w.Private,
		}); err != nil {
			tx.Rollback()
			return fmt.Errorf("archive writing: %w", err)
		}
		if err := qtx.AdminScrubWriting(ctx, dbpkg.AdminScrubWritingParams{
			Title:     sql.NullString{String: scrubText(w.Title.String), Valid: w.Title.Valid},
			Writing:   sql.NullString{String: scrubText(w.Writing.String), Valid: w.Writing.Valid},
			Abstract:  sql.NullString{String: scrubText(w.Abstract.String), Valid: w.Abstract.Valid},
			Idwriting: w.Idwriting,
		}); err != nil {
			tx.Rollback()
			return fmt.Errorf("scrub writing: %w", err)
		}
	}
	blogs, err := qtx.GetAllBlogEntriesByUser(ctx, u.Idusers)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("list blogs: %w", err)
	}
	for _, b := range blogs {
		var threadID int32
		if b.ForumthreadID.Valid {
			threadID = b.ForumthreadID.Int32
		}
		if err := qtx.AdminArchiveBlog(ctx, dbpkg.AdminArchiveBlogParams{
			Idblogs:            b.Idblogs,
			ForumthreadID:      threadID,
			UsersIdusers:       b.UsersIdusers,
			LanguageIdlanguage: b.LanguageIdlanguage,
			Blog:               b.Blog,
			Written:            sql.NullTime{Time: b.Written, Valid: true},
		}); err != nil {
			tx.Rollback()
			return fmt.Errorf("archive blog: %w", err)
		}
		if err := qtx.AdminScrubBlog(ctx, dbpkg.AdminScrubBlogParams{Blog: sql.NullString{String: scrubText(b.Blog.String), Valid: b.Blog.Valid}, Idblogs: b.Idblogs}); err != nil {
			tx.Rollback()
			return fmt.Errorf("scrub blog: %w", err)
		}
	}
	imgs, err := qtx.GetImagePostsByUserDescending(ctx, dbpkg.GetImagePostsByUserDescendingParams{UsersIdusers: u.Idusers, Limit: math.MaxInt32, Offset: 0})
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("list images: %w", err)
	}
	for _, img := range imgs {
		if err := qtx.AdminArchiveImagepost(ctx, dbpkg.AdminArchiveImagepostParams{
			Idimagepost:            img.Idimagepost,
			ForumthreadID:          img.ForumthreadID,
			UsersIdusers:           img.UsersIdusers,
			ImageboardIdimageboard: img.ImageboardIdimageboard,
			Posted:                 img.Posted,
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
	links, err := qtx.GetLinkerItemsByUserDescending(ctx, dbpkg.GetLinkerItemsByUserDescendingParams{UsersIdusers: u.Idusers, Limit: math.MaxInt32, Offset: 0})
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("list links: %w", err)
	}
	for _, l := range links {
		if err := qtx.AdminArchiveLink(ctx, dbpkg.AdminArchiveLinkParams{
			Idlinker:           l.Idlinker,
			LanguageIdlanguage: l.LanguageIdlanguage,
			UsersIdusers:       l.UsersIdusers,
			LinkerCategoryID:   l.LinkerCategoryID,
			ForumthreadID:      l.ForumthreadID,
			Title:              l.Title,
			Url:                l.Url,
			Description:        l.Description,
			Listed:             l.Listed,
		}); err != nil {
			tx.Rollback()
			return fmt.Errorf("archive link: %w", err)
		}
		if err := qtx.AdminScrubLink(ctx, dbpkg.AdminScrubLinkParams{Title: sql.NullString{String: scrubText(l.Title.String), Valid: l.Title.Valid}, Idlinker: l.Idlinker}); err != nil {
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
