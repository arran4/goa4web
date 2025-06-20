package main

import (
	"context"
	"database/sql"
)

const createSubscription = `INSERT INTO subscriptions (users_idusers, forumthread_idforumthread) VALUES (?, ?) ON DUPLICATE KEY UPDATE users_idusers=users_idusers`

func (q *Queries) CreateSubscription(ctx context.Context, userID, threadID int32) error {
	_, err := q.db.ExecContext(ctx, createSubscription, userID, threadID)
	return err
}

const deleteSubscription = `DELETE FROM subscriptions WHERE users_idusers = ? AND forumthread_idforumthread = ?`

func (q *Queries) DeleteSubscription(ctx context.Context, userID, threadID int32) error {
	_, err := q.db.ExecContext(ctx, deleteSubscription, userID, threadID)
	return err
}

type ListSubscribersForThreadParams struct {
	ForumthreadIdforumthread int32
	Idusers                  int32
}

type ListSubscribersForThreadRow struct {
	Username sql.NullString
}

const listSubscribersForThread = `SELECT u.username FROM subscriptions s JOIN users u ON s.users_idusers = u.idusers JOIN preferences p ON u.idusers = p.users_idusers WHERE s.forumthread_idforumthread = ? AND p.emailforumupdates = 1 AND u.idusers != ?`

func (q *Queries) ListSubscribersForThread(ctx context.Context, arg ListSubscribersForThreadParams) ([]*ListSubscribersForThreadRow, error) {
	rows, err := q.db.QueryContext(ctx, listSubscribersForThread, arg.ForumthreadIdforumthread, arg.Idusers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*ListSubscribersForThreadRow
	for rows.Next() {
		var i ListSubscribersForThreadRow
		if err := rows.Scan(&i.Username); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
