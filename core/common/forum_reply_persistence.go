package common

import (
	"context"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/jobs"
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

func (cd *CoreData) insertComment(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
	return insertCommentForCommenter(ctx, cd.queries, arg)
}

func (cd *CoreData) insertForumReply(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
	if cd.jobService == nil {
		return cd.insertComment(ctx, arg)
	}
	var commentID int64
	err := cd.withForumReplyTransaction(ctx, func(q db.Querier) (int32, int64, int32, error) {
		var err error
		commentID, err = insertCommentForCommenter(ctx, q, arg)
		return int32(commentID), 0, arg.CommenterID.Int32, err
	})
	if err != nil {
		return 0, err
	}
	return commentID, nil
}

func (cd *CoreData) appendForumReply(ctx context.Context, arg db.AppendCommentInSectionForCommenterParams) (int64, int64, error) {
	if cd.jobService == nil {
		return cd.AppendCommentInSectionWithActivity(ctx, arg)
	}
	var rowsAffected, appendID int64
	err := cd.withForumReplyTransaction(ctx, func(q db.Querier) (int32, int64, int32, error) {
		var err error
		rowsAffected, appendID, err = db.AppendCommentAndRecordActivity(ctx, q, arg)
		if err != nil || rowsAffected == 0 {
			return 0, 0, 0, err
		}
		return arg.CommentID, appendID, arg.CommenterID, nil
	})
	if err != nil {
		return 0, 0, err
	}
	return rowsAffected, appendID, nil
}

func (cd *CoreData) withForumReplyTransaction(ctx context.Context, write func(db.Querier) (commentID int32, appendID int64, actorID int32, err error)) error {
	transactor, ok := cd.queries.(db.CommentTransactor)
	if !ok {
		return fmt.Errorf("persist forum reply notification: database does not support comment transactions")
	}
	queued := false
	err := transactor.WithCommentTransaction(ctx, func(q db.Querier) error {
		commentID, appendID, actorID, err := write(q)
		if err != nil || commentID == 0 {
			return err
		}
		activity := buildForumReplyActivityID(commentID, appendID > 0, appendID)
		position := CalculateActivityPosition(commentID, appendID)
		if err := jobs.RecordForumReplyNotification(ctx, q, activity, position, commentID, actorID); err != nil {
			return err
		}
		queued = true
		return nil
	})
	if err == nil && queued {
		cd.jobService.SignalWakeup()
	}
	return err
}
