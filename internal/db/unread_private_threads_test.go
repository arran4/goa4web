//go:build sqlite || sqlite3
package db_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/arran4/goa4web/internal/db"
	_ "modernc.org/sqlite"
	"github.com/stretchr/testify/require"
)

func TestUnreadPrivateThreadsQueries(t *testing.T) {
	sqlDB, err := sql.Open("sqlite", "file:"+t.Name()+"?mode=memory&cache=shared")
	require.NoError(t, err)
	defer sqlDB.Close()

	ctx := context.Background()
	_, err = sqlDB.ExecContext(ctx, `
		CREATE TABLE forumtopic (
			idforumtopic INTEGER PRIMARY KEY,
			handler TEXT,
			language_id INTEGER,
			title TEXT,
			description TEXT,
			threads INTEGER,
			comments INTEGER,
			lastaddition DATETIME,
			lastposter INTEGER,
			forumcategory_idforumcategory INTEGER
		);
		CREATE TABLE forumthread (
			idforumthread INTEGER PRIMARY KEY,
			forumtopic_idforumtopic INTEGER,
			lastaddition DATETIME,
			lastposter INTEGER,
			comments INTEGER,
			firstpost INTEGER
		);
		CREATE TABLE comments (
			idcomments INTEGER PRIMARY KEY,
			users_idusers INTEGER,
			written DATETIME,
			text TEXT,
			language_id INTEGER
		);
		CREATE TABLE users (
			idusers INTEGER PRIMARY KEY,
			username TEXT
		);
		CREATE TABLE user_roles (
			users_idusers INTEGER,
			role_id INTEGER
		);
		CREATE TABLE roles (
			id INTEGER PRIMARY KEY,
			name TEXT
		);
		CREATE TABLE grants (
			id INTEGER PRIMARY KEY,
			user_id INTEGER,
			role_id INTEGER,
			section TEXT,
			item TEXT,
			action TEXT,
			active INTEGER,
			item_id INTEGER,
			created_at DATETIME,
			rule_type TEXT,
			item_rule TEXT,
			extra TEXT
		);
		CREATE TABLE content_private_labels (
			item TEXT,
			item_id INTEGER,
			user_id INTEGER,
			label TEXT,
			invert BOOLEAN
		);
	`)
	require.NoError(t, err)

	q := db.NewForDriver(sqlDB, "sqlite")

	now := time.Now()

	// Insert test data
	// User
	_, err = sqlDB.ExecContext(ctx, "INSERT INTO users (idusers, username) VALUES (1, 'user1'), (2, 'user2')")
	require.NoError(t, err)

	// Topics
	_, err = sqlDB.ExecContext(ctx, "INSERT INTO forumtopic (idforumtopic, handler, title, forumcategory_idforumcategory) VALUES (1, 'private', 'Topic 1', 1), (2, 'private', 'Topic 2', 1)")
	require.NoError(t, err)

	// Grants (user1 can see both topics and their threads)
	_, err = sqlDB.ExecContext(ctx, "INSERT INTO grants (user_id, section, item, action, active, item_id) VALUES (1, 'privateforum', 'topic', 'see', 1, 1), (1, 'privateforum', 'topic', 'see', 1, 2)")
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(ctx, "INSERT INTO grants (user_id, section, item, action, active, item_id) VALUES (1, 'privateforum_thread', 'thread', 'view', 1, 101), (1, 'privateforum_thread', 'thread', 'view', 1, 102), (1, 'privateforum_thread', 'thread', 'view', 1, 103)")
	require.NoError(t, err)

	// Comments (firstposts)
	// Thread 101 (in Topic 1): by user2 (unread for user1)
	_, err = sqlDB.ExecContext(ctx, "INSERT INTO comments (idcomments, users_idusers, written, text) VALUES (201, 2, ?, 'post 1'), (202, 2, ?, 'post 2'), (203, 2, ?, 'post 3')", now, now, now)
	require.NoError(t, err)

	// Threads
	_, err = sqlDB.ExecContext(ctx, "INSERT INTO forumthread (idforumthread, forumtopic_idforumtopic, lastaddition, lastposter, comments, firstpost) VALUES (101, 1, ?, 2, 1, 201)", now)
	require.NoError(t, err)
	// Thread 102 (in Topic 2): by user2 (unread for user1)
	_, err = sqlDB.ExecContext(ctx, "INSERT INTO forumthread (idforumthread, forumtopic_idforumtopic, lastaddition, lastposter, comments, firstpost) VALUES (102, 2, ?, 2, 1, 202)", now)
	require.NoError(t, err)
	// Thread 103 (in Topic 2): by user2, explicitly marked READ for user1
	_, err = sqlDB.ExecContext(ctx, "INSERT INTO forumthread (idforumthread, forumtopic_idforumtopic, lastaddition, lastposter, comments, firstpost) VALUES (103, 2, ?, 2, 1, 203)", now)
	require.NoError(t, err)

	// Read labels
	_, err = sqlDB.ExecContext(ctx, "INSERT INTO content_private_labels (item, item_id, user_id, label, invert) VALUES ('thread', 103, 1, 'unread', 1)")
	require.NoError(t, err)

	// 1. Scoped "Unread in Topic 1"
	rows, err := q.ListUnreadPrivateThreadsForUser(ctx, db.ListUnreadPrivateThreadsForUserParams{
		GranteeID:   1,
		GrantUserID: sql.NullInt32{Int32: 1, Valid: true},
		TopicID:     sql.NullInt32{Int32: 1, Valid: true},
		Limit:       50,
		Offset:      0,
	})
	require.NoError(t, err)
	require.Len(t, rows, 1, "Expected 1 unread thread in Topic 1")
	require.Equal(t, int32(101), rows[0].Idforumthread)

	// Count scoped
	count, err := q.CountUnreadPrivateThreadsForUser(ctx, db.CountUnreadPrivateThreadsForUserParams{
		GranteeID:   1,
		GrantUserID: sql.NullInt32{Int32: 1, Valid: true},
		TopicID:     sql.NullInt32{Int32: 1, Valid: true},
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), count, "Expected count 1 for Topic 1")

	// 2. Unscoped "All Unread"
	rows, err = q.ListUnreadPrivateThreadsForUser(ctx, db.ListUnreadPrivateThreadsForUserParams{
		GranteeID:   1,
		GrantUserID: sql.NullInt32{Int32: 1, Valid: true},
		TopicID:     sql.NullInt32{Valid: false},
		Limit:       50,
		Offset:      0,
	})
	require.NoError(t, err)
	require.Len(t, rows, 2, "Expected 2 unread threads total (Topic 1 and Topic 2, skipping explicitly read)")

	var unscopedIDs []int32
	for _, r := range rows {
		unscopedIDs = append(unscopedIDs, r.Idforumthread)
	}
	require.ElementsMatch(t, []int32{101, 102}, unscopedIDs)

	// Count unscoped
	count, err = q.CountUnreadPrivateThreadsForUser(ctx, db.CountUnreadPrivateThreadsForUserParams{
		GranteeID:   1,
		GrantUserID: sql.NullInt32{Int32: 1, Valid: true},
		TopicID:     sql.NullInt32{Valid: false},
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), count, "Expected count 2 for unscoped All Unread")
}
