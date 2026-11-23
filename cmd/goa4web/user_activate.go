package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"math"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/searchindex"
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

func (c *userActivateCmd) Usage() {
	_ = executeUsage(c.fs.Output(), "user_activate_usage.txt", c)
}

func (c *userActivateCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userActivateCmd)(nil)

func (c *userActivateCmd) Run() error {
	if c.ID == 0 && c.Username == "" {
		return fmt.Errorf("id or username required")
	}
	conn, err := c.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := newCLIQueries(conn, c.cfg)
	if c.ID == 0 {
		u, err := queries.SystemGetUserByUsername(ctx, sql.NullString{String: c.Username, Valid: true})
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}
		c.ID = int(u.Idusers)
	}
	c.Verbosef("restoring user %d", c.ID)
	if err := withCLIJobTx(ctx, conn, c.cfg, func(qtx db.Querier, svc *jobs.Service) error {
		if err := qtx.AdminRestoreUser(ctx, int32(c.ID)); err != nil {
			return fmt.Errorf("restore user: %w", err)
		}
		if err := qtx.AdminRestoreUserEmail(ctx, int32(c.ID)); err != nil {
			return fmt.Errorf("restore user email: %w", err)
		}
		if err := qtx.AdminRestoreUserPassword(ctx, int32(c.ID)); err != nil {
			return fmt.Errorf("restore user password: %w", err)
		}
		if err := qtx.AdminMarkUserRestored(ctx, int32(c.ID)); err != nil {
			return fmt.Errorf("mark user restored: %w", err)
		}
		rows, err := qtx.AdminListPendingDeactivatedComments(ctx, db.AdminListPendingDeactivatedCommentsParams{UsersIdusers: int32(c.ID), Limit: math.MaxInt32, Offset: 0})
		if err != nil {
			return fmt.Errorf("select comments: %w", err)
		}
		for _, row := range rows {
			if err := qtx.AdminRestoreComment(ctx, db.AdminRestoreCommentParams{Text: row.Text, Idcomments: row.Idcomments}); err != nil {
				return fmt.Errorf("restore comment: %w", err)
			}
			if err := qtx.AdminMarkCommentRestored(ctx, row.Idcomments); err != nil {
				return fmt.Errorf("mark comment restored: %w", err)
			}
			if err := svc.RecordSearchDirty(ctx, qtx, searchindex.TypeComment, row.Idcomments); err != nil {
				return fmt.Errorf("record comment search dirty: %w", err)
			}
		}

		rowsW, err := qtx.AdminListPendingDeactivatedWritings(ctx, db.AdminListPendingDeactivatedWritingsParams{UsersIdusers: int32(c.ID), Limit: math.MaxInt32, Offset: 0})
		if err != nil {
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
				return fmt.Errorf("restore writing: %w", err)
			}
			if err := qtx.AdminMarkWritingRestored(ctx, w.Idwriting); err != nil {
				return fmt.Errorf("mark writing restored: %w", err)
			}
			if err := svc.RecordSearchDirty(ctx, qtx, searchindex.TypeWriting, w.Idwriting); err != nil {
				return fmt.Errorf("record writing search dirty: %w", err)
			}
		}

		rowsB, err := qtx.AdminListPendingDeactivatedBlogs(ctx, db.AdminListPendingDeactivatedBlogsParams{UsersIdusers: int32(c.ID), Limit: math.MaxInt32, Offset: 0})
		if err != nil {
			return fmt.Errorf("select blogs: %w", err)
		}
		for _, b := range rowsB {
			if err := qtx.AdminRestoreBlog(ctx, db.AdminRestoreBlogParams{Blog: b.Blog, Idblogs: b.Idblogs}); err != nil {
				return fmt.Errorf("restore blog: %w", err)
			}
			if err := qtx.AdminMarkBlogRestored(ctx, b.Idblogs); err != nil {
				return fmt.Errorf("mark blog restored: %w", err)
			}
		}

		rowsI, err := qtx.AdminListPendingDeactivatedImageposts(ctx, db.AdminListPendingDeactivatedImagepostsParams{UsersIdusers: int32(c.ID), Limit: math.MaxInt32, Offset: 0})
		if err != nil {
			return fmt.Errorf("select imageposts: %w", err)
		}
		for _, img := range rowsI {
			if err := qtx.AdminRestoreImagepost(ctx, db.AdminRestoreImagepostParams{Description: img.Description, Thumbnail: img.Thumbnail, Fullimage: img.Fullimage, Idimagepost: img.Idimagepost}); err != nil {
				return fmt.Errorf("restore imagepost: %w", err)
			}
			if err := qtx.AdminMarkImagepostRestored(ctx, img.Idimagepost); err != nil {
				return fmt.Errorf("mark imagepost restored: %w", err)
			}
			if err := svc.RecordSearchDirty(ctx, qtx, searchindex.TypeImage, img.Idimagepost); err != nil {
				return fmt.Errorf("record imagepost search dirty: %w", err)
			}
		}

		rowsL, err := qtx.AdminListPendingDeactivatedLinks(ctx, db.AdminListPendingDeactivatedLinksParams{AuthorID: int32(c.ID), Limit: math.MaxInt32, Offset: 0})
		if err != nil {
			return fmt.Errorf("select links: %w", err)
		}
		for _, l := range rowsL {
			if err := qtx.AdminRestoreLink(ctx, db.AdminRestoreLinkParams{Title: l.Title, Url: l.Url, Description: l.Description, ID: l.ID}); err != nil {
				return fmt.Errorf("restore link: %w", err)
			}
			if err := qtx.AdminMarkLinkRestored(ctx, l.ID); err != nil {
				return fmt.Errorf("mark link restored: %w", err)
			}
			if err := svc.RecordSearchDirty(ctx, qtx, searchindex.TypeLinker, l.ID); err != nil {
				return fmt.Errorf("record linker search dirty: %w", err)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	c.Infof("restored user %d", c.ID)
	return nil
}
