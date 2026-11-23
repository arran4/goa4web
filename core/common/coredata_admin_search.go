package common

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/searchindex"
)

// AdminScrubComment scrubs a comment's text, marks it deleted, and records
// search durability work in the same transaction when durable jobs are configured.
func (cd *CoreData) AdminScrubComment(ctx context.Context, commentID int32, text string, recordLinks bool) error {
	return cd.adminUpdateComment(ctx, commentID, text, recordLinks, func(q db.Querier, callCtx context.Context) error {
		return q.AdminScrubComment(callCtx, db.AdminScrubCommentParams{
			Text:       sql.NullString{String: text, Valid: true},
			Idcomments: commentID,
		})
	})
}

// AdminEditComment updates a comment's text without altering its deleted state
// and records search and link durability work atomically when enabled.
func (cd *CoreData) AdminEditComment(ctx context.Context, commentID int32, text string) error {
	return cd.adminUpdateComment(ctx, commentID, text, true, func(q db.Querier, callCtx context.Context) error {
		return q.AdminUpdateCommentText(callCtx, db.AdminUpdateCommentTextParams{
			Text:       sql.NullString{String: text, Valid: true},
			Idcomments: commentID,
		})
	})
}

func (cd *CoreData) adminUpdateComment(ctx context.Context, commentID int32, text string, recordLinks bool, write func(db.Querier, context.Context) error) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	if ctx == nil {
		ctx = cd.ctx
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if cd.jobService == nil {
		return write(cd.queries, ctx)
	}
	tx, ok := cd.queries.(db.JobTransactor)
	if !ok {
		return fmt.Errorf("persist admin comment update: database does not support transactions")
	}
	var links []string
	if recordLinks && text != "" {
		links = externalLinksFromBody(text)
	}
	actor := sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID > 0}
	err := tx.WithJobTransaction(ctx, func(q db.Querier) error {
		if err := write(q, ctx); err != nil {
			return err
		}
		if err := cd.jobService.RecordSearchDirty(ctx, q, searchindex.TypeComment, commentID); err != nil {
			return err
		}
		if len(links) > 0 {
			if _, err := cd.jobService.RecordObservedExternalLinks(ctx, q, links, actor); err != nil {
				return err
			}
		}
		return nil
	})
	if err == nil {
		cd.jobService.SignalWakeup()
	}
	return err
}

// AdminDeactivateComment archives a comment, scrubs its text, and records
// search invalidation atomically when durable jobs are configured.
func (cd *CoreData) AdminDeactivateComment(ctx context.Context, archive db.AdminArchiveCommentParams, scrubText string) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	if ctx == nil {
		ctx = cd.ctx
	}
	if ctx == nil {
		ctx = context.Background()
	}
	write := func(q db.Querier) error {
		if err := q.AdminArchiveComment(ctx, archive); err != nil {
			return fmt.Errorf("archive comment: %w", err)
		}
		if err := q.AdminScrubComment(ctx, db.AdminScrubCommentParams{
			Text:       sql.NullString{String: scrubText, Valid: true},
			Idcomments: archive.Idcomments,
		}); err != nil {
			return fmt.Errorf("scrub comment: %w", err)
		}
		return nil
	}
	if cd.jobService == nil {
		return write(cd.queries)
	}
	tx, ok := cd.queries.(db.JobTransactor)
	if !ok {
		return fmt.Errorf("persist admin comment deactivation: database does not support transactions")
	}
	err := tx.WithJobTransaction(ctx, func(q db.Querier) error {
		if err := write(q); err != nil {
			return err
		}
		return cd.jobService.RecordSearchDirty(ctx, q, searchindex.TypeComment, archive.Idcomments)
	})
	if err == nil {
		cd.jobService.SignalWakeup()
	}
	return err
}

// AdminRestoreComment restores a deactivated comment's text, marks the archive
// restored, and records durable search and link work atomically when enabled.
func (cd *CoreData) AdminRestoreComment(ctx context.Context, commentID int32, text sql.NullString) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	if ctx == nil {
		ctx = cd.ctx
	}
	if ctx == nil {
		ctx = context.Background()
	}
	write := func(q db.Querier) error {
		if err := q.AdminRestoreComment(ctx, db.AdminRestoreCommentParams{
			Text:       text,
			Idcomments: commentID,
		}); err != nil {
			return fmt.Errorf("restore comment: %w", err)
		}
		if err := q.AdminMarkCommentRestored(ctx, commentID); err != nil {
			return fmt.Errorf("mark restored: %w", err)
		}
		return nil
	}
	if cd.jobService == nil {
		return write(cd.queries)
	}
	tx, ok := cd.queries.(db.JobTransactor)
	if !ok {
		return fmt.Errorf("persist admin comment restore: database does not support transactions")
	}
	var links []string
	if text.Valid && text.String != "" {
		links = externalLinksFromBody(text.String)
	}
	actor := sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID > 0}
	err := tx.WithJobTransaction(ctx, func(q db.Querier) error {
		if err := write(q); err != nil {
			return err
		}
		if err := cd.jobService.RecordSearchDirty(ctx, q, searchindex.TypeComment, commentID); err != nil {
			return err
		}
		if len(links) > 0 {
			if _, err := cd.jobService.RecordObservedExternalLinks(ctx, q, links, actor); err != nil {
				return err
			}
		}
		return nil
	})
	if err == nil {
		cd.jobService.SignalWakeup()
	}
	return err
}

// AdminUpdateLinkerItem updates a published linker item and records search
// dirty work atomically when durable jobs are configured.
func (cd *CoreData) AdminUpdateLinkerItem(ctx context.Context, arg db.AdminUpdateLinkerItemParams) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	if ctx == nil {
		ctx = cd.ctx
	}
	_, err := cd.withSearchWorkContext(ctx, searchindex.TypeLinker, func(q db.Querier) (int64, error) {
		if err := q.AdminUpdateLinkerItem(ctx, arg); err != nil {
			return 0, err
		}
		return int64(arg.ID), nil
	})
	return err
}

// AdminUpdateImagePost updates an image post and records search dirty work
// atomically when durable jobs are configured.
func (cd *CoreData) AdminUpdateImagePost(ctx context.Context, arg db.AdminUpdateImagePostParams) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	if ctx == nil {
		ctx = cd.ctx
	}
	_, err := cd.withSearchWorkContext(ctx, searchindex.TypeImage, func(q db.Querier) (int64, error) {
		if err := q.AdminUpdateImagePost(ctx, arg); err != nil {
			return 0, err
		}
		return int64(arg.Idimagepost), nil
	})
	return err
}

// AdminDeleteImagePost soft-deletes an image post and records search dirty
// clearance work atomically when durable jobs are configured.
func (cd *CoreData) AdminDeleteImagePost(ctx context.Context, postID int32) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	if ctx == nil {
		ctx = cd.ctx
	}
	_, err := cd.withSearchWorkContext(ctx, searchindex.TypeImage, func(q db.Querier) (int64, error) {
		if err := q.AdminDeleteImagePost(ctx, postID); err != nil {
			return 0, err
		}
		return int64(postID), nil
	})
	return err
}
