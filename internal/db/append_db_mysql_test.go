//go:build !sqlite && !sqlite3

package db

import (
	"strings"

	"context"
	"database/sql"
	"testing"
	"time"
)

func TestAppendCommentMySQL(t *testing.T) {
	database := openReplyThreadTestDatabase(t)
	createReplyThreadTestSchema(t, database)

	execReplySQL(t, database,
		`INSERT INTO users (idusers, username) VALUES (1, 'user1'), (2, 'user2')`,
		`INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (1000, 100, 1, 'initial text', DATE_SUB('2030-01-01 12:00:00', INTERVAL 30 MINUTE))`,
		`INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (1, 'forum', 'topic', 100, 'append', 1, 'allow')`,
	)

	queries := NewForDriver(database, "mysql")

	ctx := context.Background()

	// Perform append using actual queries.AppendCommentInSectionForCommenter
	now := time.Now()
	rowsAffected, err := queries.AppendCommentInSectionForCommenter(ctx, AppendCommentInSectionForCommenterParams{
		Text:             "new text",
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

	// Verify the text was actually appended correctly
	comment, err := queries.GetCommentById(ctx, 1000)
	if err != nil {
		t.Fatalf("get comment: %v", err)
	}

	expectedText := "initial text\n\n[hr]\n\nnew text"
	if comment.Text.String != expectedText {
		t.Errorf("expected text %q, got %q", expectedText, comment.Text.String)
	}
}

func TestAppendCommentConcurrencyMySQL(t *testing.T) {
	database := openReplyThreadTestDatabase(t)
	createReplyThreadTestSchema(t, database)

	execReplySQL(t, database,
		`INSERT INTO users (idusers, username) VALUES (1, 'user1'), (2, 'user2')`,
		`INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (1000, 100, 1, 'initial text', DATE_SUB('2030-01-01 12:00:00', INTERVAL 30 MINUTE))`,
		`INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (1, 'forum', 'topic', 100, 'append', 1, 'allow')`,
	)

	queries := NewForDriver(database, "mysql")

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

	// Because they ran concurrently and appended atomically with CONCAT,
	// both strings must be in the result, although order is not deterministic.
	finalText := comment.Text.String

	if !strings.Contains(finalText, "append 1") {
		t.Errorf("missing append 1 in: %q", finalText)
	}
	if !strings.Contains(finalText, "append 2") {
		t.Errorf("missing append 2 in: %q", finalText)
	}
}

func testAppendEligibilityMatrix(t *testing.T, database *sql.DB, queries Querier, ctx context.Context) {
	// Instead of using real queries (which might be missing in Querier stub or generated types),
	// we will manually setup the DB via exec.

	// Create user
	database.ExecContext(ctx, "INSERT INTO users (idusers, username) VALUES (?, ?)", 60, "poster")
	database.ExecContext(ctx, "INSERT INTO users (idusers, username) VALUES (?, ?)", 61, "reader")

	// Create thread
	database.ExecContext(ctx, "INSERT INTO forumthread (idforumthread, poster, language_id) VALUES (?, ?, ?)", 1000, 60, 1)

	// Create comment
	database.ExecContext(ctx, "INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (?, ?, ?, ?, '2030-01-01 12:00:00')", 2000, 1000, 60, "Initial post")
	database.ExecContext(ctx, "UPDATE forumthread SET firstpost = 2000, lastpost = 2000 WHERE idforumthread = 1000")

	// Create grant for poster
	database.ExecContext(ctx, "INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (?, ?, ?, ?, ?, ?, ?)", 60, "forum", "topic", 30, "append", 1, "allow")

	tests := []struct {
		name     string
		setup    func()
		wantRows bool
	}{
		{
			name:     "no marker -> 1 row",
			setup:    func() {},
			wantRows: true,
		},
		{
			name: "author's own marker -> 1 row",
			setup: func() {
				database.ExecContext(ctx, "INSERT INTO content_read_markers (item, item_id, user_id, last_comment_id) VALUES (?, ?, ?, ?)", "thread", 1000, 60, 2000)
			},
			wantRows: true,
		},
		{
			name: "other user's older marker -> 1 row",
			setup: func() {
				database.ExecContext(ctx, "INSERT INTO content_read_markers (item, item_id, user_id, last_comment_id) VALUES (?, ?, ?, ?)", "thread", 1000, 61, 1999)
			},
			wantRows: true,
		},
		{
			name: "other user's marker at comment -> 0 rows",
			setup: func() {
				database.ExecContext(ctx, "INSERT INTO content_read_markers (item, item_id, user_id, last_comment_id) VALUES (?, ?, ?, ?)", "thread", 1000, 61, 2000)
			},
			wantRows: false,
		},
		{
			name: "other user's marker beyond -> 0 rows",
			setup: func() {
				database.ExecContext(ctx, "INSERT INTO content_read_markers (item, item_id, user_id, last_comment_id) VALUES (?, ?, ?, ?)", "thread", 1000, 61, 2001)
			},
			wantRows: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Clear all read markers
			database.ExecContext(ctx, "DELETE FROM content_read_markers WHERE item_id = ? AND item = ?", 1000, "thread")
			// Reset comment text and written time
			database.ExecContext(ctx, "UPDATE comments SET text = 'Initial post', written = '2030-01-01 12:00:00' WHERE idcomments = 2000")

			tc.setup()

			rowsAffected, err := queries.AppendCommentInSectionForCommenter(ctx, AppendCommentInSectionForCommenterParams{
				Text:             "\n\n[hr]\n\nnew text",
				AppendWindowMins: 60, // normal window
				Section:          "forum",
				ItemType:         sql.NullString{String: "topic", Valid: true},
				ItemID:           sql.NullInt32{Int32: 30, Valid: true},
				CommentID:        2000,
				CommenterID:      60,
				ForumthreadID:    1000,
				GrantUserID:      sql.NullInt32{Int32: 60, Valid: true},
				Written:          sql.NullTime{Time: time.Now(), Valid: true},
			})
			hasAppended := rowsAffected > 0
			_ = err

			if hasAppended != tc.wantRows {
				t.Fatalf("Expected wantRows=%v, got=%v", tc.wantRows, hasAppended)
			}
		})
	}
}

func TestMySQLAppendEligibilityMatrix(t *testing.T) {
	database := openReplyThreadTestDatabase(t)
	createReplyThreadTestSchema(t, database)
	queries := NewForDriver(database, "mysql")
	ctx := context.Background()

	// Clear out any previous test data
	database.Exec("DELETE FROM grants")
	database.Exec("DELETE FROM content_read_markers")
	database.Exec("DELETE FROM comments")
	database.Exec("DELETE FROM forumthread")
	database.Exec("DELETE FROM users")

	testAppendEligibilityMatrix(t, database, queries, ctx)
}
