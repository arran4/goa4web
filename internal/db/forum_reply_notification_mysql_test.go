//go:build !sqlite && !sqlite3

package db

import (
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/database"
	"github.com/stretchr/testify/require"
)

func TestMySQLForumReplyNotificationSource(t *testing.T) {
	ctx := t.Context()
	conn := openReplyThreadTestDatabase(t)
	_, err := conn.ExecContext(ctx, string(database.SchemaMySQL))
	require.NoError(t, err)
	execReplySQL(t, conn,
		"INSERT INTO users (idusers, username) VALUES (1, 'author'), (2, 'another author')",
		"INSERT INTO forumtopic (idforumtopic, handler, title) VALUES (5, 'private', 'Private topic')",
		"INSERT INTO forumthread (idforumthread, forumtopic_idforumtopic, firstpost) VALUES (6, 5, 7)",
		"INSERT INTO comments (idcomments, forumthread_id, users_idusers, text) VALUES (7, 6, 1, 'opener'), (8, 6, 1, 'current reply'), (9, 6, 2, 'another reply')",
		"INSERT INTO comment_appends (id, comment_id, users_idusers) VALUES (10, 8, 1), (11, 9, 2)",
	)
	q := NewForDriver(conn, "mysql")

	for _, appendID := range []int32{0, 10} {
		source, err := q.SystemGetForumReplyNotificationSource(ctx, SystemGetForumReplyNotificationSourceParams{CommentID: 8, ActorID: 1, AppendID: appendID})
		require.NoError(t, err)
		require.Equal(t, int32(6), source.ThreadID)
		require.Equal(t, int32(5), source.TopicID)
		require.Equal(t, "Private topic", source.TopicTitle.String)
		require.Equal(t, "private", source.Handler)
		require.Equal(t, "author", source.Username.String)
		require.Equal(t, "current reply", source.ReplyText.String)
	}

	for _, params := range []SystemGetForumReplyNotificationSourceParams{
		{CommentID: 8, ActorID: 2},
		{CommentID: 8, ActorID: 1, AppendID: 11},
	} {
		_, err := q.SystemGetForumReplyNotificationSource(ctx, params)
		require.ErrorIs(t, err, sql.ErrNoRows)
	}
	execReplySQL(t, conn, "UPDATE comments SET deleted_at = CURRENT_TIMESTAMP WHERE idcomments = 7")
	_, err = q.SystemGetForumReplyNotificationSource(ctx, SystemGetForumReplyNotificationSourceParams{CommentID: 8, ActorID: 1, AppendID: 10})
	require.ErrorIs(t, err, sql.ErrNoRows)
}
