//go:build sqlite || sqlite3

package db

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestAppendCommentSQLite(t *testing.T) {
	database := openReplyThreadTestDatabaseSQLite(t)
	createReplyThreadTestSchema(t, database)

	execReplySQL(t, database,
		`INSERT INTO users (idusers, username) VALUES (1, 'user1'), (2, 'user2')`,
		`INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (1000, 100, 1, 'initial text', datetime('now', '-30 minutes'))`,
		`INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (1, 'forum', 'topic', 100, 'append', 1, 'allow')`,
	)

	queries := NewForDriver(database, "sqlite")
	ctx := context.Background()
	now := time.Now()

	rowsAffected, err := queries.AppendCommentInSectionForCommenter(ctx, AppendCommentInSectionForCommenterParams{
		Text:             "new text sqlite",
		AppendWindowMins: 60,
		Section:          "forum",
		ItemType:         sql.NullString{String: "topic", Valid: true},
		GrantUserID:      sql.NullInt32{Int32: 1, Valid: true},
		ItemID:           sql.NullInt32{Int32: 100, Valid: true},
		Written:          sql.NullTime{Time: now, Valid: true},
		CommentID:        1000,
		CommenterID:      1,
		ForumthreadID:    100,
	})

	if err != nil {
		t.Fatalf("append comment: %v", err)
	}

	if rowsAffected != 1 {
		t.Fatalf("expected 1 row affected, got %d", rowsAffected)
	}

	comment, err := queries.GetCommentById(ctx, 1000)
	if err != nil {
		t.Fatalf("get comment: %v", err)
	}

	expectedText := "initial text\n\n[hr]\n\nnew text sqlite"
	if comment.Text.String != expectedText {
		t.Errorf("expected text %q, got %q", expectedText, comment.Text.String)
	}
	if !comment.Written.Valid || comment.Written.Time.Sub(now).Abs() > time.Second {
		t.Errorf("written = %v, want %v", comment.Written, now)
	}
	var commentCount int
	if err := database.QueryRowContext(ctx, "SELECT COUNT(*) FROM comments WHERE forumthread_id = ?", 100).Scan(&commentCount); err != nil {
		t.Fatalf("count comments: %v", err)
	}
	if commentCount != 1 || comment.Idcomments != 1000 {
		t.Fatalf("append changed comment identity/count: id=%d count=%d", comment.Idcomments, commentCount)
	}
}

func TestAppendCommentConcurrencySQLite(t *testing.T) {
	database := openReplyThreadTestDatabaseSQLite(t)
	createReplyThreadTestSchema(t, database)

	execReplySQL(t, database,
		`INSERT INTO users (idusers, username) VALUES (1, 'user1'), (2, 'user2')`,
		`INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (1000, 100, 1, 'initial text', datetime('now', '-30 minutes'))`,
		`INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (1, 'forum', 'topic', 100, 'append', 1, 'allow')`,
	)

	queries := NewForDriver(database, "sqlite")
	ctx := context.Background()

	// Simulate concurrent append attempts
	ch1 := make(chan error)
	ch2 := make(chan error)

	now := time.Now()

	go func() {
		_, err := queries.AppendCommentInSectionForCommenter(ctx, AppendCommentInSectionForCommenterParams{
			Text:             "append 1",
			AppendWindowMins: 60,
			Section:          "forum",
			ItemType:         sql.NullString{String: "topic", Valid: true},
			GrantUserID:      sql.NullInt32{Int32: 1, Valid: true},
			ItemID:           sql.NullInt32{Int32: 100, Valid: true},
			Written:          sql.NullTime{Time: now, Valid: true},
			CommentID:        1000,
			CommenterID:      1,
			ForumthreadID:    100,
		})
		ch1 <- err
	}()

	go func() {
		_, err := queries.AppendCommentInSectionForCommenter(ctx, AppendCommentInSectionForCommenterParams{
			Text:             "append 2",
			AppendWindowMins: 60,
			Section:          "forum",
			ItemType:         sql.NullString{String: "topic", Valid: true},
			GrantUserID:      sql.NullInt32{Int32: 1, Valid: true},
			ItemID:           sql.NullInt32{Int32: 100, Valid: true},
			Written:          sql.NullTime{Time: now, Valid: true},
			CommentID:        1000,
			CommenterID:      1,
			ForumthreadID:    100,
		})
		ch2 <- err
	}()

	err1 := <-ch1
	err2 := <-ch2

	if err1 != nil {
		t.Fatalf("append 1: %v", err1)
	}
	if err2 != nil {
		t.Fatalf("append 2: %v", err2)
	}

	comment, err := queries.GetCommentById(ctx, 1000)
	if err != nil {
		t.Fatalf("get comment: %v", err)
	}

	finalText := comment.Text.String

	if !strings.Contains(finalText, "append 1") {
		t.Errorf("missing append 1 in: %q", finalText)
	}
	if !strings.Contains(finalText, "append 2") {
		t.Errorf("missing append 2 in: %q", finalText)
	}
}

func openReplyThreadTestDatabaseSQLite(t *testing.T) *sql.DB {
	t.Helper()
	// Using file::memory:?cache=shared instead of just :memory: for concurrent accesses
	db, err := sql.Open("sqlite", "file:"+t.Name()+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func mustExec(t *testing.T, ctx context.Context, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		t.Fatalf("setup exec failed: %v", err)
	}
}

func testAppendEligibilityMatrix(t *testing.T, database *sql.DB, queries Querier, ctx context.Context) {
	// Instead of using real queries (which might be missing in Querier stub or generated types),
	// we will manually setup the DB via exec.

	// Create user
	mustExec(t, ctx, database, "INSERT INTO users (idusers, username) VALUES (?, ?)", 60, "poster")
	mustExec(t, ctx, database, "INSERT INTO users (idusers, username) VALUES (?, ?)", 61, "reader")

	// Create thread
	mustExec(t, ctx, database, "INSERT INTO forumthread (idforumthread, firstpost, lastposter, forumtopic_idforumtopic) VALUES (?, ?, ?, ?)", 1000, 2000, 60, 30)

	// Create comment
	mustExec(t, ctx, database, "INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (?, ?, ?, ?, '2030-01-01 12:00:00')", 2000, 1000, 60, "Initial post")
	mustExec(t, ctx, database, "UPDATE forumthread SET firstpost = 2000, lastposter = 60 WHERE idforumthread = 1000")

	// Create grant for poster
	mustExec(t, ctx, database, "INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (?, ?, ?, ?, ?, ?, ?)", 60, "forum", "topic", 30, "append", 1, "allow")

	tests := []struct {
		name     string
		setup    func()
		section  string
		itemType string
		itemID   int32
		wantRows bool
	}{
		{
			name:    "no marker -> 1 row",
			section: "forum", itemType: "topic", itemID: 30,
			setup:    func() {},
			wantRows: true,
		},
		{
			name:    "author's own marker -> 1 row",
			section: "forum", itemType: "topic", itemID: 30,
			setup: func() {
				mustExec(t, ctx, database, "INSERT INTO content_read_markers (item, item_id, user_id, last_comment_id) VALUES (?, ?, ?, ?)", "thread", 1000, 60, 2000)
			},
			wantRows: true,
		},
		{
			name:    "other user's older marker -> 1 row",
			section: "forum", itemType: "topic", itemID: 30,
			setup: func() {
				mustExec(t, ctx, database, "INSERT INTO content_read_markers (item, item_id, user_id, last_comment_id) VALUES (?, ?, ?, ?)", "thread", 1000, 61, 1999)
			},
			wantRows: true,
		},
		{
			name:    "other user's marker at comment -> 0 rows",
			section: "forum", itemType: "topic", itemID: 30,
			setup: func() {
				mustExec(t, ctx, database, "INSERT INTO content_read_markers (item, item_id, user_id, last_comment_id) VALUES (?, ?, ?, ?)", "thread", 1000, 61, 2000)
			},
			wantRows: false,
		},
		{
			name: "other user's marker beyond -> 0 rows",
			setup: func() {
				mustExec(t, ctx, database, "INSERT INTO content_read_markers (item, item_id, user_id, last_comment_id) VALUES (?, ?, ?, ?)", "thread", 1000, 61, 2001)
			},
			wantRows: false,
		},
		{
			name: "newer comment exists -> 0 rows",
			setup: func() {
				mustExec(t, ctx, database, "INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (?, ?, ?, ?, '2030-01-01 12:01:00')", 2001, 1000, 60, "newer")
			},
			wantRows: false,
		},
		{
			name: "wrong owner -> 0 rows",
			setup: func() {
				mustExec(t, ctx, database, "UPDATE comments SET users_idusers = 61 WHERE idcomments = 2000")
			},
			wantRows: false,
		},
		{
			name: "append grant missing -> 0 rows",
			setup: func() {
				mustExec(t, ctx, database, "DELETE FROM grants")
			},
			wantRows: false,
		},
		{
			name: "outside window -> 0 rows",
			setup: func() {
				mustExec(t, ctx, database, "UPDATE comments SET written = '2000-01-01 12:00:00' WHERE idcomments = 2000")
			},
			wantRows: false,
		},
		{
			name: "privateforum_thread/thread grant -> 1 row",
			setup: func() {

				mustExec(t, ctx, database, "INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (?, ?, ?, ?, ?, ?, ?)", 60, "privateforum_thread", "thread", 1000, "append", 1, "allow")
			},
			section:  "privateforum_thread",
			itemType: "thread",
			itemID:   1000,
			wantRows: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mustExec(t, ctx, database, "DELETE FROM comments WHERE idcomments > 2000")
			mustExec(t, ctx, database, "DELETE FROM content_read_markers")
			mustExec(t, ctx, database, "DELETE FROM grants")
			mustExec(t, ctx, database, "UPDATE comments SET users_idusers = 60, text = 'Initial post', written = '2030-01-01 12:00:00' WHERE idcomments = 2000")

			section := "forum"
			if tc.section == "privateforum_thread" {
				section = tc.section
			}
			if section == "forum" {
				mustExec(t, ctx, database, "INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (?, ?, ?, ?, ?, ?, ?)", 60, "forum", "topic", 30, "append", 1, "allow")
			}

			tc.setup()

			rowsAffected, err := queries.AppendCommentInSectionForCommenter(ctx, AppendCommentInSectionForCommenterParams{
				Text:             "new text",
				AppendWindowMins: 60, // normal window
				Section: func() string {
					if tc.section == "" {
						return "forum"
					} else {
						return tc.section
					}
				}(),
				ItemType: sql.NullString{String: func() string {
					if tc.itemType == "" {
						return "topic"
					} else {
						return tc.itemType
					}
				}(), Valid: true},
				ItemID: sql.NullInt32{Int32: func() int32 {
					if tc.itemID == 0 {
						return 30
					} else {
						return tc.itemID
					}
				}(), Valid: true},
				CommentID:     2000,
				CommenterID:   60,
				ForumthreadID: 1000,
				GrantUserID:   sql.NullInt32{Int32: 60, Valid: true},
				Written:       sql.NullTime{Time: time.Now(), Valid: true},
			})
			hasAppended := rowsAffected > 0
			if err != nil {
				t.Fatalf("err: %v", err)
			}

			if hasAppended != tc.wantRows {
				t.Fatalf("Expected wantRows=%v, got=%v", tc.wantRows, hasAppended)
			}
			if tc.wantRows {
				cRow, _ := queries.GetCommentById(ctx, 2000)
				if cRow.Text.String != "Initial post\n\n[hr]\n\nnew text" {
					t.Fatalf("Expected exact text, got: %q", cRow.Text.String)
				}
			}
		})
	}
}

func TestSQLiteAppendEligibilityMatrix(t *testing.T) {
	database := openReplyThreadTestDatabaseSQLite(t)
	createReplyThreadTestSchema(t, database)
	queries := NewForDriver(database, "sqlite")
	ctx := context.Background()
	testAppendEligibilityMatrix(t, database, queries, ctx)
}
