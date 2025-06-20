package main

import (
	"context"
	"time"
)

type EmailQueueItem struct {
	IDEmailQueue int64
	Email        string
	Page         string
	Created      time.Time
}

func (q *Queries) EnqueueEmail(ctx context.Context, email, page string) error {
	const insert = "INSERT INTO emailQueue (email, page, created) VALUES (?, ?, NOW())"
	_, err := q.db.ExecContext(ctx, insert, email, page)
	return err
}

func (q *Queries) ListQueuedEmails(ctx context.Context, limit int32) ([]EmailQueueItem, error) {
	const selectQuery = "SELECT idemailQueue, email, page, created FROM emailQueue ORDER BY idemailQueue LIMIT ?"
	rows, err := q.db.QueryContext(ctx, selectQuery, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []EmailQueueItem
	for rows.Next() {
		var it EmailQueueItem
		if err := rows.Scan(&it.IDEmailQueue, &it.Email, &it.Page, &it.Created); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (q *Queries) DeleteQueuedEmail(ctx context.Context, id int64) error {
	const deleteQuery = "DELETE FROM emailQueue WHERE idemailQueue = ?"
	_, err := q.db.ExecContext(ctx, deleteQuery, id)
	return err
}
