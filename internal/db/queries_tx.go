package db

import (
	"context"
	"database/sql"
	"fmt"
)

// BeginTx starts a transaction using the underlying database connection.
func (q *Queries) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if txer, ok := q.db.(interface {
		BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
	}); ok {
		return txer.BeginTx(ctx, opts)
	}
	return nil, fmt.Errorf("database does not support transactions")
}

// JobTransactor runs related queue and source changes in one transaction.
type JobTransactor interface {
	WithJobTransaction(context.Context, func(Querier) error) error
}

// WithJobTransaction commits job and source updates together.
func (q *Queries) WithJobTransaction(ctx context.Context, update func(Querier) error) error {
	tx, err := q.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	return updateJobTransaction(tx, q.WithTx(tx), update)
}

func updateJobTransaction(tx *sql.Tx, q Querier, update func(Querier) error) error {
	defer func() { _ = tx.Rollback() }()
	if err := update(q); err != nil {
		return err
	}
	return tx.Commit()
}
