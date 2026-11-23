package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// CommentAppender defines transactional forum comment append operations.
type CommentAppender interface {
	AppendCommentInSectionWithActivity(ctx context.Context, arg AppendCommentInSectionForCommenterParams) (rowsAffected int64, appendID int64, err error)
}

// CommentTransactor lets callers commit a comment mutation and its required effects together.
type CommentTransactor interface {
	WithCommentTransaction(ctx context.Context, write func(Querier) error) error
}

// WithCommentTransaction runs a comment mutation and its required effects atomically.
func (q *Queries) WithCommentTransaction(ctx context.Context, write func(Querier) error) error {
	tx, err := q.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin comment transaction: %w", err)
	}
	return writeCommentTransaction(tx, q.WithTx(tx), write)
}

func writeCommentTransaction(tx *sql.Tx, q Querier, write func(Querier) error) error {
	defer func() { _ = tx.Rollback() }()
	if err := write(q); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit comment transaction: %w", err)
	}
	return nil
}

// AppendCommentInSectionWithActivity appends to a comment and records its activity in one transaction.
func (q *Queries) AppendCommentInSectionWithActivity(ctx context.Context, arg AppendCommentInSectionForCommenterParams) (int64, int64, error) {
	return appendCommentWithActivity(ctx, q, arg)
}

func appendCommentWithActivity(ctx context.Context, transactor CommentTransactor, arg AppendCommentInSectionForCommenterParams) (int64, int64, error) {
	var rowsAffected, appendID int64
	err := transactor.WithCommentTransaction(ctx, func(q Querier) error {
		var err error
		rowsAffected, appendID, err = AppendCommentAndRecordActivity(ctx, q, arg)
		return err
	})
	if err != nil {
		return 0, 0, err
	}
	return rowsAffected, appendID, nil
}

// AppendCommentAndRecordActivity appends to a comment and records an activity when the conditional write succeeds.
// The caller supplies a transaction-bound querier so both writes roll back together.
func AppendCommentAndRecordActivity(ctx context.Context, q Querier, arg AppendCommentInSectionForCommenterParams) (int64, int64, error) {
	rowsAffected, err := q.AppendCommentInSectionForCommenter(ctx, arg)
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
	appendID, err := q.SystemCreateCommentAppend(ctx, SystemCreateCommentAppendParams{
		CommentID:    arg.CommentID,
		UsersIdusers: arg.CommenterID,
		CreatedAt:    createdAt,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("record comment append: %w", err)
	}
	return rowsAffected, appendID, nil
}
