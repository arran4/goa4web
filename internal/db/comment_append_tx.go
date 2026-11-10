package db

import (
	"context"
	"fmt"
	"time"
)

// CommentAppender defines transactional forum comment append operations.
type CommentAppender interface {
	AppendCommentInSectionWithActivity(ctx context.Context, arg AppendCommentInSectionForCommenterParams) (rowsAffected int64, appendID int64, err error)
}

// AppendCommentInSectionWithActivity conditionally appends text to an existing comment within the append window
// and allocates an append activity ID in a single database transaction. If the conditional append affects 0 rows,
// no activity record is created and 0, 0, nil is returned. If either the append update or activity insert fails,
// the transaction is rolled back and an error is returned.
func (q *Queries) AppendCommentInSectionWithActivity(ctx context.Context, arg AppendCommentInSectionForCommenterParams) (int64, int64, error) {
	tx, err := q.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("begin append transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	qtx := q.WithTx(tx)
	rowsAffected, err := qtx.AppendCommentInSectionForCommenter(ctx, arg)
	if err != nil {
		return 0, 0, fmt.Errorf("append comment in section: %w", err)
	}
	if rowsAffected == 0 {
		return 0, 0, nil
	}

	createdAt := arg.Written.Time
	if !arg.Written.Valid {
		createdAt = time.Now().UTC()
	}

	appendID, err := qtx.SystemCreateCommentAppend(ctx, SystemCreateCommentAppendParams{
		CommentID:    arg.CommentID,
		UsersIdusers: arg.CommenterID,
		CreatedAt:    createdAt,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("record comment append: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("commit append transaction: %w", err)
	}

	return rowsAffected, appendID, nil
}
