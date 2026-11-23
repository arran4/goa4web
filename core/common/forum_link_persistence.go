package common

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/searchindex"
)

func (cd *CoreData) recordsCommentWork(section consts.PermissionSection) bool {
	if cd.jobService == nil {
		return false
	}
	switch section {
	case consts.PermissionSectionForum, consts.PermissionSectionPrivateForum, consts.PermissionSectionPrivateForumThread,
		consts.PermissionSectionBlogs, consts.PermissionSectionImageBBS, consts.PermissionSectionLinker,
		consts.PermissionSectionNews, consts.PermissionSectionWriting:
		return true
	default:
		return false
	}
}

func insertCommentForCommenter(ctx context.Context, q db.Querier, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
	result, err := q.CreateCommentInSectionForCommenter(ctx, arg)
	if err != nil {
		return 0, err
	}
	// SQLite preserves its previous insert ID when the grant-checked INSERT inserts no row.
	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return 0, err
	}
	return result.LastInsertId()
}

func (cd *CoreData) withCommentWork(ctx context.Context, recordWork bool, body string, actorID sql.NullInt32, write func(db.Querier) (int64, error)) error {
	if !recordWork {
		_, err := write(cd.queries)
		return err
	}
	tx, ok := cd.queries.(db.CommentTransactor)
	if !ok {
		return fmt.Errorf("persist comment work: database does not support comment transactions")
	}
	links := externalLinksFromBody(body)
	written := false
	err := tx.WithCommentTransaction(ctx, func(q db.Querier) error {
		commentID, err := write(q)
		if err != nil || commentID == 0 {
			return err
		}
		if err := cd.jobService.RecordSearchDirty(ctx, q, searchindex.TypeComment, int32(commentID)); err != nil {
			return err
		}
		if _, err := cd.jobService.RecordObservedExternalLinks(ctx, q, links, actorID); err != nil {
			return err
		}
		written = true
		return nil
	})
	if err == nil && written {
		cd.jobService.SignalWakeup()
	}
	return err
}

func (cd *CoreData) appendForumCommentWithWork(ctx context.Context, body string, arg db.AppendCommentInSectionForCommenterParams) (int64, int64, error) {
	if cd.jobService == nil {
		return cd.AppendCommentInSectionWithActivity(ctx, arg)
	}
	var rows, appendID int64
	actor := sql.NullInt32{Int32: arg.CommenterID, Valid: arg.CommenterID != 0}
	err := cd.withCommentWork(ctx, true, body, actor, func(q db.Querier) (int64, error) {
		var err error
		rows, appendID, err = db.AppendCommentAndRecordActivity(ctx, q, arg)
		if rows == 0 || err != nil {
			return 0, err
		}
		return int64(arg.CommentID), nil
	})
	if err != nil {
		return 0, 0, err
	}
	return rows, appendID, nil
}

// UpdateCommentForEditorWithWork updates a comment and records its durable
// search revision and observed external links in one transaction when enabled.
func (cd *CoreData) UpdateCommentForEditorWithWork(ctx context.Context, arg db.UpdateCommentForEditorParams) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var rows int64
	err := cd.withCommentWork(ctx, cd.jobService != nil, arg.Text.String, arg.EditorID, func(q db.Querier) (int64, error) {
		var err error
		rows, err = q.UpdateCommentForEditor(ctx, arg)
		if err != nil || rows == 0 {
			return 0, err
		}
		return int64(arg.CommentID), nil
	})
	return rows, err
}
