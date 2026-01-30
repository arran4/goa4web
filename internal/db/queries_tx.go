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
