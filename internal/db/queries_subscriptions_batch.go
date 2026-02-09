package db

import (
	"context"
	"strings"
)

// BatchInsertSubscriptions performs a bulk insert for subscriptions.
func (q *Queries) BatchInsertSubscriptions(ctx context.Context, params []InsertSubscriptionParams) error {
	if len(params) == 0 {
		return nil
	}

	query := "INSERT INTO subscriptions (users_idusers, pattern, method) VALUES "
	vals := make([]interface{}, 0, len(params)*3)
	placeholders := make([]string, 0, len(params))

	for _, p := range params {
		placeholders = append(placeholders, "(?, ?, ?)")
		vals = append(vals, p.UsersIdusers, p.Pattern, p.Method)
	}

	query += strings.Join(placeholders, ",")

	_, err := q.db.ExecContext(ctx, query, vals...)
	return err
}

// BatchDeleteSubscriptions performs a bulk delete for subscriptions based on subscriber, pattern, and method.
func (q *Queries) BatchDeleteSubscriptions(ctx context.Context, params []DeleteSubscriptionForSubscriberParams) error {
	if len(params) == 0 {
		return nil
	}

	// Use tuple syntax for efficient deletion: DELETE FROM subscriptions WHERE (users_idusers, pattern, method) IN ((?, ?, ?), ...)
	query := "DELETE FROM subscriptions WHERE (users_idusers, pattern, method) IN ("
	vals := make([]interface{}, 0, len(params)*3)
	placeholders := make([]string, 0, len(params))

	for _, p := range params {
		placeholders = append(placeholders, "(?, ?, ?)")
		vals = append(vals, p.SubscriberID, p.Pattern, p.Method)
	}

	query += strings.Join(placeholders, ",") + ")"

	_, err := q.db.ExecContext(ctx, query, vals...)
	return err
}
