package common

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

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

func (cd *CoreData) withForumLinkObservations(ctx context.Context, links []string, actorID sql.NullInt32, write func(db.Querier) (bool, error)) error {
	if len(links) == 0 || cd.jobService == nil {
		_, err := write(cd.queries)
		return err
	}
	tx, ok := cd.queries.(db.CommentTransactor)
	if !ok {
		return fmt.Errorf("persist forum link observations: database does not support comment transactions")
	}
	queued := false
	err := tx.WithCommentTransaction(ctx, func(q db.Querier) error {
		written, err := write(q)
		if err != nil || !written {
			return err
		}
		queued, err = cd.jobService.RecordObservedExternalLinks(ctx, q, links, actorID)
		return err
	})
	if err == nil && queued {
		cd.jobService.SignalWakeup()
	}
	return err
}

func (cd *CoreData) appendForumCommentWithObservedLinks(ctx context.Context, body string, arg db.AppendCommentInSectionForCommenterParams) (int64, int64, error) {
	if cd.jobService == nil {
		return cd.AppendCommentInSectionWithActivity(ctx, arg)
	}
	links := externalLinksFromBody(body)
	if len(links) == 0 {
		return cd.AppendCommentInSectionWithActivity(ctx, arg)
	}
	var rows, appendID int64
	actor := sql.NullInt32{Int32: arg.CommenterID, Valid: arg.CommenterID != 0}
	err := cd.withForumLinkObservations(ctx, links, actor, func(q db.Querier) (bool, error) {
		var err error
		rows, appendID, err = db.AppendCommentAndRecordActivity(ctx, q, arg)
		return rows != 0, err
	})
	if err != nil {
		return 0, 0, err
	}
	return rows, appendID, nil
}
