//go:build sqlite || sqlite3

package migrations_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/arran4/goa4web/internal/db"
	_ "modernc.org/sqlite"
)

func TestSQLitePendingEmailQueueUpgrade(t *testing.T) {
	ctx := t.Context()
	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetMaxOpenConns(1)
	for version := 1; version <= 104; version++ {
		if _, err := conn.ExecContext(ctx, sqliteMigrationUpSQL(t, version)); err != nil {
			t.Fatalf("apply SQLite migration %d: %v", version, err)
		}
	}
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO pending_emails (id, to_email, subject, body, html_body, to_user_id, error_count, digest_delivery_id)
		VALUES (17, 'old@example.test', 'Queued before upgrade', 'original body', 'original HTML', 4, 2, 9);
		INSERT INTO pending_emails (id, to_email, subject, body)
		VALUES (50, 'pruned@example.test', 'Pruned', 'pruned');
		DELETE FROM pending_emails WHERE id = 50;
	`); err != nil {
		t.Fatal(err)
	}

	q := db.NewForDriver(conn, "sqlite3")
	// Before 0105 this insert fails because omitted legacy columns have no default.
	if _, err := conn.ExecContext(ctx, sqliteMigrationUpSQL(t, 105)); err != nil {
		t.Fatal(err)
	}
	if err := q.SystemInsertDigestPendingEmail(ctx, db.SystemInsertDigestPendingEmailParams{
		ToUserID:         sql.NullInt32{Int32: 4, Valid: true},
		DigestDeliveryID: sql.NullInt32{Int32: 10, Valid: true},
		Body:             "grouped digest",
	}); err != nil {
		t.Fatalf("queue digest: %v", err)
	}
	if err := q.InsertPendingEmail(ctx, db.InsertPendingEmailParams{Body: "direct email", DirectEmail: true}); err != nil {
		t.Fatalf("queue direct email: %v", err)
	}
	assertLegacy := func() {
		t.Helper()
		var addr, subject, body, html string
		var uid, errors, digest int
		if err := conn.QueryRowContext(ctx, "SELECT to_email, subject, body, html_body, to_user_id, error_count, digest_delivery_id FROM pending_emails WHERE id = 17").Scan(&addr, &subject, &body, &html, &uid, &errors, &digest); err != nil {
			t.Fatal(err)
		}
		if addr != "old@example.test" || subject != "Queued before upgrade" || body != "original body" || html != "original HTML" || uid != 4 || errors != 2 || digest != 9 {
			t.Fatalf("legacy queued message changed: %q %q %q %q %d %d %d", addr, subject, body, html, uid, errors, digest)
		}
	}
	assertLegacy()
	var id int32
	if err := conn.QueryRowContext(ctx, "SELECT id FROM pending_emails WHERE digest_delivery_id = 10").Scan(&id); err != nil || id <= 50 {
		t.Fatalf("digest id = %d, err = %v, want id above deleted high-water mark 50", id, err)
	}
	if _, err := conn.ExecContext(ctx, sqliteMigrationDownSQL(t, 105)); err != nil {
		t.Fatal(err)
	}
	assertLegacy()
	var uid int
	if err := conn.QueryRowContext(ctx, "SELECT to_user_id FROM pending_emails WHERE body = 'direct email'").Scan(&uid); err != nil || uid != 0 {
		t.Fatalf("legacy direct recipient = %d, err = %v, want sentinel 0", uid, err)
	}
}

func TestCanonicalSQLitePendingEmailAcceptsCurrentQueueInsert(t *testing.T) {
	// Read the canonical schema file so this test verifies the exact SQLC
	// source independently of the versioned migration path.
	schema, err := os.ReadFile("../database/schema.sqlite.sql")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetMaxOpenConns(1)
	ctx := t.Context()
	if _, err := conn.ExecContext(ctx, string(schema)); err != nil {
		t.Fatal(err)
	}
	q := db.NewForDriver(conn, "sqlite3")
	if err := q.InsertPendingEmail(ctx, db.InsertPendingEmailParams{Body: "fresh direct email", DirectEmail: true}); err != nil {
		t.Fatalf("queue direct email on canonical schema: %v", err)
	}
}
