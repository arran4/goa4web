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
