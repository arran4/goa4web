package migrations

import (
	"os"
	"testing"

	"github.com/arran4/goa4web/database"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

func TestMySQLMigration0106PreservesPreferencesAndExistingMail(t *testing.T) {
	dsn := os.Getenv("GOA4WEB_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set GOA4WEB_TEST_MYSQL_DSN to run MySQL/MariaDB migration tests")
	}
	ctx := t.Context()
	conn := openTemporaryMySQLDatabase(t, dsn)
	_, err := conn.ExecContext(ctx, string(database.SchemaMySQL))
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, "CREATE TABLE schema_version (version INT NOT NULL)")
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, "INSERT INTO schema_version VALUES (106)")
	require.NoError(t, err)
	// Existing upgraded servers have the complete historic migration ledger.
	_, err = conn.ExecContext(ctx, "DELETE FROM goose_db_version")
	require.NoError(t, err)
	for version := 0; version <= 106; version++ {
		_, err = conn.ExecContext(ctx, "INSERT INTO goose_db_version (version_id, is_applied) VALUES (?, 1)", version)
		require.NoError(t, err)
	}
	provider, err := goose.NewProvider(goose.DialectMySQL, conn, ForDriver("mysql"))
	require.NoError(t, err)
	_, err = provider.DownTo(ctx, 105)
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, "INSERT INTO preferences (users_idusers, emailforumupdates, daily_digest_hour) VALUES (7, 1, 8)")
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, "INSERT INTO pending_emails (id, to_user_id, body) VALUES (34, 7, 'pre-existing queued MIME')")
	require.NoError(t, err)
	_, err = provider.UpTo(ctx, 106)
	require.NoError(t, err)
	var legacyVersion, generation, hour int
	var enabled bool
	var body string
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT version FROM schema_version").Scan(&legacyVersion))
	require.Equal(t, 106, legacyVersion)
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT optional_email_generation, emailforumupdates, daily_digest_hour FROM preferences WHERE users_idusers = 7").Scan(&generation, &enabled, &hour))
	require.Zero(t, generation)
	require.True(t, enabled)
	require.Equal(t, 8, hour)
	require.NoError(t, conn.QueryRowContext(ctx, "SELECT body FROM pending_emails WHERE id = 34").Scan(&body))
	require.Equal(t, "pre-existing queued MIME", body)
}
