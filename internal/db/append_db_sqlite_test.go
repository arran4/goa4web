//go:build sqlite || sqlite3

package db

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"
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
	return db
}