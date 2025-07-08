package db

import (
	"context"
	"database/sql"
)

type InsertNotificationParams struct {
	UsersIdusers int32
	Link         sql.NullString
	Message      sql.NullString
}

func (q *Queries) InsertNotification(ctx context.Context, arg InsertNotificationParams) error {
	_, err := q.db.ExecContext(ctx,
		"INSERT INTO notifications (users_idusers, link, message) VALUES (?, ?, ?)",
		arg.UsersIdusers, arg.Link, arg.Message)
	return err
}

func (q *Queries) CountUnreadNotifications(ctx context.Context, userID int32) (int32, error) {
	var c int32
	err := q.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM notifications WHERE users_idusers = ? AND read_at IS NULL",
		userID).Scan(&c)
	return c, err
}

func (q *Queries) GetUnreadNotifications(ctx context.Context, userID int32) ([]*Notification, error) {
	rows, err := q.db.QueryContext(ctx,
		"SELECT id, users_idusers, link, message, created_at, read_at FROM notifications WHERE users_idusers = ? AND read_at IS NULL ORDER BY id DESC",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UsersIdusers, &n.Link, &n.Message, &n.CreatedAt, &n.ReadAt); err != nil {
			return nil, err
		}
		items = append(items, &n)
	}
	return items, rows.Err()
}

func (q *Queries) MarkNotificationRead(ctx context.Context, id int32) error {
	_, err := q.db.ExecContext(ctx, "UPDATE notifications SET read_at = NOW() WHERE id = ?", id)
	return err
}

func (q *Queries) PurgeReadNotifications(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx,
		"DELETE FROM notifications WHERE read_at IS NOT NULL AND read_at < (NOW() - INTERVAL 24 HOUR)")
	return err
}

type LastNotificationByMessageParams struct {
	UsersIdusers int32
	Message      string
}

func (q *Queries) LastNotificationByMessage(ctx context.Context, arg LastNotificationByMessageParams) (*Notification, error) {
	row := q.db.QueryRowContext(ctx,
		"SELECT id, users_idusers, link, message, created_at, read_at FROM notifications WHERE users_idusers = ? AND message = ? ORDER BY id DESC LIMIT 1",
		arg.UsersIdusers, arg.Message)
	var n Notification
	err := row.Scan(&n.ID, &n.UsersIdusers, &n.Link, &n.Message, &n.CreatedAt, &n.ReadAt)
	if err != nil {
		return nil, err
	}
	return &n, nil
}
