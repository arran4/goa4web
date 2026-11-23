//go:build sqlite || sqlite3

package db

import "context"

// WithJobTransaction commits job and source updates together.
func (q *sqliteQuerier) WithJobTransaction(ctx context.Context, update func(Querier) error) error {
	tx, err := q.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	return updateJobTransaction(tx, q.withTx(tx), update)
}
