//go:build sqlite || sqlite3

package migrations_test

import (
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestSQLiteAdminCanCreateLanguageNeutralTopicAfterUpgrade(t *testing.T) {
	ctx := t.Context()
	conn, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer conn.Close()
	conn.SetMaxOpenConns(1)
	for version := 1; version <= handlers.ExpectedSchemaVersion; version++ {
		_, err := conn.ExecContext(ctx, sqliteMigrationUpSQL(t, version))
		require.NoError(t, err, "apply historical SQLite schema version %d", version)
	}
	q := db.NewForDriver(conn, "sqlite")
	args := db.AdminCreateForumTopicParams{
		Title:       sql.NullString{String: "Language-neutral announcements", Valid: true},
		Description: sql.NullString{String: "Visible across languages", Valid: true},
	}
	id, err := q.AdminCreateForumTopic(ctx, args)
	require.NoError(t, err)
	var language int32
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT language_id FROM forumtopic WHERE idforumtopic = ?", id).Scan(&language))
	require.Zero(t, language)

	args.LanguageID = sql.NullInt32{Int32: 1, Valid: true}
	id, err = q.AdminCreateForumTopic(ctx, args)
	require.NoError(t, err)
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT language_id FROM forumtopic WHERE idforumtopic = ?", id).Scan(&language))
	require.EqualValues(t, 1, language)
}
