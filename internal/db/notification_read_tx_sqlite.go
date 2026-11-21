//go:build sqlite || sqlite3

package db

import (
	"context"
	"database/sql"
)

// AcknowledgeThreadReadActivity atomically updates the user's thread read watermark and marks
// matching notifications read in a single SQLite transaction.
func (s *sqliteQuerier) AcknowledgeThreadReadActivity(ctx context.Context, arg AcknowledgeThreadReadActivityParams) error {
	return acknowledgeThreadReadActivityTx(ctx, s.BeginTx, func(tx *sql.Tx) Querier {
		return s.withTx(tx)
	}, arg)
}

// DeliverThreadNotificationForActivity atomically evaluates the recipient's thread read watermark
// under a SQLite write lock and inserts or updates the activity notification in a single transaction.
func (s *sqliteQuerier) DeliverThreadNotificationForActivity(ctx context.Context, arg SystemCreateNotificationForActivityParams) error {
	return deliverThreadNotificationForActivityTx(ctx, s, s.BeginTx, func(tx *sql.Tx) Querier {
		return s.withTx(tx)
	}, arg)
}
