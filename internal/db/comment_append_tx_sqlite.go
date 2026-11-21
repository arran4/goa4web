//go:build sqlite || sqlite3

package db

import (
	"context"
	"database/sql"
	"fmt"
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

// WithCommentTransaction runs a comment mutation and its required effects atomically.
func (s *sqliteQuerier) WithCommentTransaction(ctx context.Context, write func(Querier) error) error {
	tx, err := s.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin comment transaction: %w", err)
	}
	return writeCommentTransaction(tx, s.withTx(tx), write)
}

// AppendCommentInSectionWithActivity appends to a comment and records its activity in one SQLite transaction.
func (s *sqliteQuerier) AppendCommentInSectionWithActivity(ctx context.Context, arg AppendCommentInSectionForCommenterParams) (int64, int64, error) {
	return appendCommentWithActivity(ctx, s, arg)
}
