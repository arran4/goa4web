//go:build sqlite || sqlite3

package migrations_test

import (
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/internal/db"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestSQLiteReplyEmailUpgradePreservesPreferencesAndExistingMail(t *testing.T) {
	ctx := t.Context()
	conn, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer conn.Close()
	conn.SetMaxOpenConns(1)
	for version := 1; version <= 105; version++ {
		_, err := conn.ExecContext(ctx, sqliteMigrationUpSQL(t, version))
		require.NoError(t, err, "apply version %d", version)
	}
	_, err = conn.ExecContext(ctx, `
		INSERT INTO users (idusers, username) VALUES (7, 'subscriber');
		INSERT INTO preferences (users_idusers, emailforumupdates, daily_digest_hour) VALUES (7, 1, 8);
		INSERT INTO pending_emails (id, to_user_id, body) VALUES (34, 7, 'pre-existing queued MIME');
	`)
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, sqliteMigrationUpSQL(t, 106))
	require.NoError(t, err)
	q := db.NewForDriver(conn, "sqlite")
	pref, err := q.GetPreferenceForLister(ctx, 7)
	require.NoError(t, err)
	require.True(t, pref.Emailforumupdates.Bool)
	require.Equal(t, int32(8), pref.DailyDigestHour.Int32)
	require.Zero(t, pref.OptionalEmailGeneration)
	require.NoError(t, q.UpdateEmailForumUpdatesForLister(ctx, db.UpdateEmailForumUpdatesForListerParams{
		EmailForumUpdates: sql.NullBool{Bool: false, Valid: true}, ListerID: 7,
	}))
	require.NoError(t, q.UpdateEmailForumUpdatesForLister(ctx, db.UpdateEmailForumUpdatesForListerParams{
		EmailForumUpdates: sql.NullBool{Bool: true, Valid: true}, ListerID: 7,
	}))
	pref, err = q.GetPreferenceForLister(ctx, 7)
	require.NoError(t, err)
	require.EqualValues(t, 1, pref.OptionalEmailGeneration)
	_, err = conn.ExecContext(ctx, sqliteMigrationDownSQL(t, 106))
	require.NoError(t, err)
	var body string
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT body FROM pending_emails WHERE id = 34").Scan(&body))
	require.Equal(t, "pre-existing queued MIME", body)
}
