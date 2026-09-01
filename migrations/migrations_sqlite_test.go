//go:build sqlite || sqlite3

package migrations_test

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/app/dbstart"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/migrations"
	_ "modernc.org/sqlite"
)

func TestSQLiteMigrationsStepByStep(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_step.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open test sqlite db: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	for v := 1; v <= handlers.ExpectedSchemaVersion; v++ {
		filename := fmt.Sprintf("%04d_sqlite.sql", v)
		content, err := migrations.FS.ReadFile(filename)
		if err != nil {
			t.Fatalf("read migration %s: %v", filename, err)
		}

		rawSQL := string(content)
		// Only take the Up section
		if idx := strings.Index(rawSQL, "-- +goose Down"); idx != -1 {
			rawSQL = rawSQL[:idx]
		}
		lines := strings.Split(rawSQL, "\n")
		var cleanLines []string
		for _, l := range lines {
			trimmed := strings.TrimSpace(l)
			if strings.HasPrefix(trimmed, "-- +goose") {
				continue
			}
			cleanLines = append(cleanLines, l)
		}
		cleanSQL := strings.Join(cleanLines, "\n")

		// Execute migration
		if _, err := db.ExecContext(ctx, cleanSQL); err != nil {
			t.Fatalf("Migration %s failed: %v\nSQL was:\n%s", filename, err, cleanSQL)
		}
	}
}

func sqliteMigrationUpSQL(t *testing.T, version int) string {
	t.Helper()
	filename := fmt.Sprintf("%04d_sqlite.sql", version)
	content, err := migrations.FS.ReadFile(filename)
	if err != nil {
		t.Fatalf("read migration %s: %v", filename, err)
	}
	rawSQL := string(content)
	if idx := strings.Index(rawSQL, "-- +goose Down"); idx != -1 {
		rawSQL = rawSQL[:idx]
	}
	lines := strings.Split(rawSQL, "\n")
	cleanLines := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "-- +goose") {
			continue
		}
		cleanLines = append(cleanLines, line)
	}
	return strings.Join(cleanLines, "\n")
}

func TestSQLiteMigration0098RepairsNullableColumnsAndParticipantQuery(t *testing.T) {
	ctx := context.Background()
	dsn := fmt.Sprintf("file:migration_0098_%p?mode=memory&cache=shared", t)
	dbConn, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}
	t.Cleanup(func() { _ = dbConn.Close() })

	for version := 1; version <= 97; version++ {
		if _, err := dbConn.ExecContext(ctx, sqliteMigrationUpSQL(t, version)); err != nil {
			t.Fatalf("apply historical SQLite migration %04d: %v", version, err)
		}
	}

	if _, err := dbConn.ExecContext(ctx, `
		INSERT INTO users (idusers, username) VALUES (1, 'alice'), (2, 'bob');
		INSERT INTO blogs (idblogs, forumthread_id, users_idusers, language_id, blog)
		VALUES (30, 20, 1, 1, 'Preserve this blog');
		INSERT INTO forumtopic (idforumtopic, forumcategory_idforumcategory, language_id, title, description, handler)
		VALUES (10, 0, 1, 'Private chat with bob', 'Migration regression', 'private');
		INSERT INTO forumthread (idforumthread, forumtopic_idforumtopic) VALUES (20, 10);
		INSERT INTO grants (user_id, section, item, rule_type, item_id, action, active) VALUES
			(1, 'privateforum', 'topic', 'allow', 10, 'post', 1),
			(1, 'privateforum', 'topic', 'allow', 10, 'view', 1),
			(2, 'privateforum', 'topic', 'allow', 10, 'view', 1);
	`); err != nil {
		t.Fatalf("seed version 97 database: %v", err)
	}

	if _, err := dbConn.ExecContext(ctx, sqliteMigrationUpSQL(t, 98)); err != nil {
		t.Fatalf("apply SQLite migration 0098: %v", err)
	}

	affectedTables := []string{
		"blogs", "comments", "faq", "linker", "linker_queue", "preferences",
		"site_news", "writing", "deactivated_comments", "deactivated_writings",
		"deactivated_blogs", "deactivated_linker",
	}
	for _, table := range affectedTables {
		var notNull int
		err := dbConn.QueryRowContext(ctx,
			`SELECT "notnull" FROM pragma_table_info(?) WHERE name = 'language_id'`, table,
		).Scan(&notNull)
		if err != nil {
			t.Errorf("inspect %s.language_id: %v", table, err)
			continue
		}
		if notNull != 0 {
			t.Errorf("%s.language_id remains NOT NULL after migration 0098", table)
		}
	}

	var blogThreadID, blogLanguageID sql.NullInt64
	var blogText string
	if err := dbConn.QueryRowContext(ctx,
		`SELECT forumthread_id, language_id, blog FROM blogs WHERE idblogs = 30`,
	).Scan(&blogThreadID, &blogLanguageID, &blogText); err != nil {
		t.Fatalf("load preserved blog: %v", err)
	}
	if !blogThreadID.Valid || blogThreadID.Int64 != 20 || !blogLanguageID.Valid || blogLanguageID.Int64 != 1 || blogText != "Preserve this blog" {
		t.Fatalf("blog data changed during migration: thread=%v language=%v text=%q", blogThreadID, blogLanguageID, blogText)
	}
	var blogThreadNotNull int
	if err := dbConn.QueryRowContext(ctx,
		`SELECT "notnull" FROM pragma_table_info('blogs') WHERE name = 'forumthread_id'`,
	).Scan(&blogThreadNotNull); err != nil {
		t.Fatalf("inspect blogs.forumthread_id: %v", err)
	}
	if blogThreadNotNull != 0 {
		t.Error("blogs.forumthread_id remains NOT NULL after migration 0098")
	}
	if _, err := dbConn.ExecContext(ctx,
		`INSERT INTO blogs (users_idusers, forumthread_id, language_id, blog) VALUES (1, NULL, NULL, 'Nullable sentinels')`,
	); err != nil {
		t.Fatalf("insert blog with nullable migration-0060 fields: %v", err)
	}

	var faqPriorityIndex int
	if err := dbConn.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = 'faq_priority_idx'`,
	).Scan(&faqPriorityIndex); err != nil {
		t.Fatalf("inspect faq priority index: %v", err)
	}
	if faqPriorityIndex != 1 {
		t.Fatalf("faq_priority_idx count = %d, want 1", faqPriorityIndex)
	}

	queries := db.NewForDriver(dbConn, "sqlite3")
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserID(1))
	commentID, err := cd.CreatePrivateForumOpeningCommentForPoster(1, 20, 10, 0, "Unspecified language")
	if err != nil {
		t.Fatalf("CreatePrivateForumOpeningCommentForPoster language 0: %v", err)
	}
	var languageID sql.NullInt64
	if err := dbConn.QueryRowContext(ctx, `SELECT language_id FROM comments WHERE idcomments = ?`, commentID).Scan(&languageID); err != nil {
		t.Fatalf("load opening comment language: %v", err)
	}
	if languageID.Valid {
		t.Fatalf("opening comment language_id = %d, want NULL", languageID.Int64)
	}

	displayTitle, participants, total, err := cd.GetPrivateTopicDetails(10, "Private chat with bob")
	if err != nil {
		t.Fatalf("GetPrivateTopicDetails on migrated SQLite database: %v", err)
	}
	if displayTitle != "bob" || len(participants) != 1 || participants[0] != "bob" || total != 2 {
		t.Fatalf("private topic details = title %q, participants %v, total %d; want bob, [bob], 2", displayTitle, participants, total)
	}
}

func TestSQLiteMigrationsExecuteFromScratch(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_migration.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open test sqlite db: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	if err := dbstart.Apply(ctx, db, migrations.FS, testing.Verbose(), "sqlite3"); err != nil {
		t.Fatalf("failed to apply sqlite migrations: %v", err)
	}

	version, err := dbstart.SchemaVersionWithDriver(ctx, db, "sqlite3")
	if err != nil {
		t.Fatalf("failed to read sqlite schema version: %v", err)
	}

	if version != handlers.ExpectedSchemaVersion {
		t.Fatalf("expected schema version %d, got %d", handlers.ExpectedSchemaVersion, version)
	}

	// Verify key tables were created
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		t.Fatalf("failed to query users table: %v", err)
	}

	// Verify migration 0095 columns in forumthread
	var replyToCommentID, replyToThreadID sql.NullInt64
	err = db.QueryRowContext(ctx, "SELECT reply_to_comment_id, reply_to_thread_id FROM forumthread LIMIT 1").Scan(&replyToCommentID, &replyToThreadID)
	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("failed to query 0095 forumthread columns: %v", err)
	}

	// Verify long URLs > 255 chars with identical first 255 chars in SQLite
	prefix := "https://example.com/very/long/path/" + strings.Repeat("a", 230) + "?common=true&suffix="
	url1 := prefix + "one"
	url2 := prefix + "two"

	if _, err := db.ExecContext(ctx, `INSERT INTO external_links (url, clicks) VALUES (?, 10)`, url1); err != nil {
		t.Fatalf("insert long sqlite url 1: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO external_links (url, clicks) VALUES (?, 20)`, url2); err != nil {
		t.Fatalf("insert long sqlite url 2: %v", err)
	}

	var id1, id2, clicks1, clicks2 int
	if err := db.QueryRowContext(ctx, `SELECT id, clicks FROM external_links WHERE url = ?`, url1).Scan(&id1, &clicks1); err != nil {
		t.Fatalf("query long sqlite url 1: %v", err)
	}
	if err := db.QueryRowContext(ctx, `SELECT id, clicks FROM external_links WHERE url = ?`, url2).Scan(&id2, &clicks2); err != nil {
		t.Fatalf("query long sqlite url 2: %v", err)
	}
	if id1 == id2 || clicks1 != 10 || clicks2 != 20 {
		t.Fatalf("sqlite long urls collided: id1=%d, id2=%d, clicks1=%d, clicks2=%d", id1, id2, clicks1, clicks2)
	}
}
