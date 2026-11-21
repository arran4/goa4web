//go:build sqlite || sqlite3

package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// BeginTx starts a transaction using the underlying SQLite database connection.
func (s *sqliteQuerier) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if txer, ok := s.db.(interface {
		BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
	}); ok {
		return txer.BeginTx(ctx, opts)
	}
	return nil, fmt.Errorf("database does not support transactions")
}

func (s *sqliteQuerier) withTx(tx *sql.Tx) *sqliteQuerier {
	return &sqliteQuerier{
		db: tx,
		q:  s.q.WithTx(tx),
	}
}

// AppendCommentInSectionWithActivity conditionally appends text to an existing comment within the append window
// and allocates an append activity ID in a single SQLite database transaction. If the conditional append affects 0 rows,
// no activity record is created and 0, 0, nil is returned. If either the append update or activity insert fails,
// the transaction is rolled back and an error is returned.
func (s *sqliteQuerier) AppendCommentInSectionWithActivity(ctx context.Context, arg AppendCommentInSectionForCommenterParams) (int64, int64, error) {
	tx, err := s.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("begin append transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stx := s.withTx(tx)
	rowsAffected, err := stx.AppendCommentInSectionForCommenter(ctx, arg)
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

	appendID, err := stx.SystemCreateCommentAppend(ctx, SystemCreateCommentAppendParams{
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
