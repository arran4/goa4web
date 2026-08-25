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
		`INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (1000, 100, 1, 'initial text', DATE_SUB(NOW(), INTERVAL 30 MINUTE))`,
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
		`INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (1000, 100, 1, 'initial text', DATE_SUB(NOW(), INTERVAL 30 MINUTE))`,
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
