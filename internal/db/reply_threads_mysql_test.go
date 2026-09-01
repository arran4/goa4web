package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

func TestGetReplyThreadsForListerFiltersPrivateChildrenAndOrdersUnread(t *testing.T) {
	database := openReplyThreadTestDatabase(t)
	createReplyThreadTestSchema(t, database)
	execReplySQL(t, database,
		`INSERT INTO roles (id, name) VALUES (1, 'anyone'), (2, 'member')`,
		`INSERT INTO users (idusers, username) VALUES (10, 'viewer'), (11, 'excluded'), (12, 'author')`,
		`INSERT INTO user_roles (users_idusers, role_id) VALUES (10, 2), (11, 2)`,
		`INSERT INTO forumtopic (idforumtopic, title, handler) VALUES (5, 'Private', 'private')`,
		`INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (10, 'privateforum', 'topic', 5, 'view', 1, 'allow')`,
		`INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (1000, 100, 12, 'visible', '2026-08-16 10:00:00'), (1001, 101, 12, 'hidden', '2026-08-16 11:00:00'), (1002, 102, 12, 'anyone', '2026-08-16 09:00:00')`,
		`INSERT INTO forumthread (idforumthread, firstpost, lastposter, forumtopic_idforumtopic, comments, lastaddition, reply_to_comment_id, reply_to_thread_id) VALUES (100, 1000, 12, 5, 0, '2026-08-16 12:00:00', 500, 50), (101, 1001, 12, 5, 0, '2026-08-16 13:00:00', 500, 50), (102, 1002, 12, 5, 0, '2026-08-16 11:00:00', 500, 50)`,
		`INSERT INTO grants (user_id, role_id, section, item, item_id, action, active, rule_type) VALUES (10, NULL, 'privateforum_thread', 'thread', 100, 'view', 1, 'allow'), (11, NULL, 'privateforum_thread', 'thread', 101, 'view', 1, 'allow'), (NULL, 1, 'privateforum_thread', 'thread', 102, 'view', 1, 'allow')`,
		`INSERT INTO content_private_labels (item, item_id, user_id, label, invert) VALUES ('thread', 100, 10, 'unread', true)`,
	)
	queries := New(database)
	rows, err := queries.GetReplyThreadsForLister(context.Background(), GetReplyThreadsForListerParams{
		ViewerID:        10,
		ReplyToThreadID: sql.NullInt32{Int32: 50, Valid: true},
		ViewerMatchID:   sql.NullInt32{Int32: 10, Valid: true},
	})
	if err != nil {
		t.Fatalf("list visible forks: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("visible forks = %d, want 2: %#v", len(rows), rows)
	}
	if rows[0].Idforumthread != 102 || rows[0].IsUnread != 1 {
		t.Errorf("first fork = thread %d unread=%d, want unread anyone-role thread 102", rows[0].Idforumthread, rows[0].IsUnread)
	}
	if rows[1].Idforumthread != 100 || rows[1].IsUnread != 0 {
		t.Errorf("second fork = thread %d unread=%d, want read user-granted thread 100", rows[1].Idforumthread, rows[1].IsUnread)
	}
	for _, row := range rows {
		if row.Idforumthread == 101 {
			t.Fatal("private child without a viewer thread grant was enumerated")
		}
	}
}

func TestGetReplyThreadsForListerPreservesPublicForumVisibility(t *testing.T) {
	database := openReplyThreadTestDatabase(t)
	createReplyThreadTestSchema(t, database)
	execReplySQL(t, database,
		`INSERT INTO roles (id, name) VALUES (1, 'anyone')`,
		`INSERT INTO users (idusers, username) VALUES (10, 'viewer'), (12, 'author')`,
		`INSERT INTO forumtopic (idforumtopic, title, handler) VALUES (6, 'Public', 'forum')`,
		`INSERT INTO grants (role_id, section, item, item_id, action, active, rule_type) VALUES (1, 'forum', 'topic', 6, 'view', 1, 'allow')`,
		`INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (2000, 200, 12, 'public fork', '2026-08-16 10:00:00')`,
		`INSERT INTO forumthread (idforumthread, firstpost, lastposter, forumtopic_idforumtopic, comments, lastaddition, reply_to_comment_id, reply_to_thread_id) VALUES (200, 2000, 12, 6, 0, '2026-08-16 10:00:00', 600, 60)`,
	)
	rows, err := New(database).GetReplyThreadsForLister(context.Background(), GetReplyThreadsForListerParams{
		ViewerID:        10,
		ReplyToThreadID: sql.NullInt32{Int32: 60, Valid: true},
		ViewerMatchID:   sql.NullInt32{Int32: 10, Valid: true},
	})
	if err != nil {
		t.Fatalf("list public forks: %v", err)
	}
	if len(rows) != 1 || rows[0].Idforumthread != 200 {
		t.Fatalf("public forks = %#v, want thread 200", rows)
	}
}

func TestSystemCopyPrivateThreadGrantsDoesNotBroadenToTopicParticipants(t *testing.T) {
	database := openReplyThreadTestDatabase(t)
	createReplyThreadTestSchema(t, database)
	execReplySQL(t, database,
		`INSERT INTO users (idusers, username) VALUES (10, 'source viewer'), (11, 'topic only')`,
		`INSERT INTO grants (user_id, role_id, section, item, item_id, action, active, rule_type) VALUES (10, NULL, 'privateforum_thread', 'thread', 300, 'view', 1, 'allow'), (10, NULL, 'privateforum_thread', 'thread', 300, 'reply', 1, 'allow'), (NULL, 1, 'privateforum_thread', 'thread', 300, 'view', 1, 'allow'), (11, NULL, 'privateforum', 'topic', 5, 'view', 1, 'allow')`,
	)
	if err := New(database).SystemCopyPrivateThreadGrantsToThread(context.Background(), SystemCopyPrivateThreadGrantsToThreadParams{
		SrcThreadID: sql.NullInt32{Int32: 300, Valid: true},
		DstThreadID: sql.NullInt32{Int32: 301, Valid: true},
	}); err != nil {
		t.Fatalf("copy source thread grants: %v", err)
	}
	rows, err := database.Query(`SELECT user_id, role_id, action FROM grants WHERE section = 'privateforum_thread' AND item_id = 301 ORDER BY action, user_id, role_id`)
	if err != nil {
		t.Fatalf("list copied grants: %v", err)
	}
	defer func() { _ = rows.Close() }()
	var count int
	var copiedAnyoneRole bool
	for rows.Next() {
		var userID, roleID sql.NullInt32
		var action string
		if err := rows.Scan(&userID, &roleID, &action); err != nil {
			t.Fatalf("scan copied grant: %v", err)
		}
		if userID.Valid && userID.Int32 != 10 {
			t.Fatalf("unexpected copied grant user: %#v", userID)
		}
		if roleID.Valid {
			if roleID.Int32 != 1 {
				t.Fatalf("unexpected copied grant role: %#v", roleID)
			}
			copiedAnyoneRole = true
		}
		count++
	}
	if count != 3 || !copiedAnyoneRole {
		t.Fatalf("copied grants = %d, copied anyone role = %t; want source user and role grants only", count, copiedAnyoneRole)
	}
}

func TestSystemDeleteUninitializedThreadRemovesForkAndCopiedGrants(t *testing.T) {
	database := openReplyThreadTestDatabase(t)
	createReplyThreadTestSchema(t, database)
	execReplySQL(t, database,
		`INSERT INTO forumtopic (idforumtopic, title, handler) VALUES (5, 'Private', 'private')`,
		`INSERT INTO forumthread (idforumthread, forumtopic_idforumtopic, reply_to_comment_id, reply_to_thread_id) VALUES (400, 5, 40, 30)`,
		`INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (10, 'privateforum_thread', 'thread', 400, 'view', 1, 'allow')`,
	)
	if err := New(database).SystemDeleteUninitializedThread(context.Background(), 400); err != nil {
		t.Fatalf("delete uninitialized fork: %v", err)
	}
	var threadCount, grantCount int
	if err := database.QueryRow(`SELECT COUNT(*) FROM forumthread WHERE idforumthread = 400`).Scan(&threadCount); err != nil {
		t.Fatalf("count fork rows: %v", err)
	}
	if err := database.QueryRow(`SELECT COUNT(*) FROM grants WHERE section = 'privateforum_thread' AND item_id = 400`).Scan(&grantCount); err != nil {
		t.Fatalf("count fork grants: %v", err)
	}
	if threadCount != 0 || grantCount != 0 {
		t.Fatalf("cleanup left thread=%d grants=%d", threadCount, grantCount)
	}
}

func createReplyThreadTestSchema(t *testing.T, database *sql.DB) {
	t.Helper()
	execReplySQL(t, database,
		`CREATE TABLE roles (id INT PRIMARY KEY, name VARCHAR(255) NOT NULL, is_admin BOOLEAN NOT NULL DEFAULT 0)`,
		`CREATE TABLE users (idusers INT PRIMARY KEY, username VARCHAR(255))`,
		`CREATE TABLE user_roles (users_idusers INT NOT NULL, role_id INT NOT NULL)`,
		`CREATE TABLE user_language (users_idusers INT NOT NULL, language_id INT NOT NULL)`,
		`CREATE TABLE forumtopic (idforumtopic INT PRIMARY KEY, language_id INT NULL, title VARCHAR(255), handler VARCHAR(255) NOT NULL)`,
		`CREATE TABLE forumthread (idforumthread INT PRIMARY KEY, firstpost INT NOT NULL DEFAULT 0, lastposter INT NOT NULL DEFAULT 0, forumtopic_idforumtopic INT NOT NULL, comments INT NULL, lastaddition DATETIME NULL, locked BOOLEAN NULL, reply_to_comment_id INT NULL, reply_to_thread_id INT NULL, deleted_at DATETIME NULL)`,
		`CREATE TABLE comments (idcomments INT PRIMARY KEY, forumthread_id INT NOT NULL, users_idusers INT NOT NULL, language_id INT NULL, text TEXT NULL, written DATETIME NULL, timezone VARCHAR(255) NULL, deleted_at DATETIME NULL, last_index DATETIME NULL)`,
		`CREATE TABLE grants (id INT AUTO_INCREMENT PRIMARY KEY, created_at DATETIME NULL, user_id INT NULL, role_id INT NULL, section VARCHAR(255) NOT NULL, item VARCHAR(255) NULL, rule_type VARCHAR(255) NOT NULL, item_id INT NULL, item_rule VARCHAR(255) NULL, action VARCHAR(255) NOT NULL, extra VARCHAR(255) NULL, active BOOLEAN NOT NULL)`,
		`CREATE TABLE content_read_markers (item VARCHAR(255), item_id INT, user_id INT, last_comment_id INT, unread INT)`,
		`CREATE TABLE content_private_labels (item VARCHAR(255) NOT NULL, item_id INT NOT NULL, user_id INT NOT NULL, label VARCHAR(255) NOT NULL, invert BOOLEAN NOT NULL)`,
		`CREATE TABLE content_public_labels (item VARCHAR(255) NOT NULL, item_id INT NOT NULL, label VARCHAR(255) NOT NULL)`,
		`CREATE TABLE content_label_status (item VARCHAR(255) NOT NULL, item_id INT NOT NULL, label VARCHAR(255) NOT NULL)`,
	)
}

func execReplySQL(t *testing.T, database *sql.DB, statements ...string) {
	t.Helper()
	for _, statement := range statements {
		if _, err := database.Exec(statement); err != nil {
			t.Fatalf("execute %q: %v", statement, err)
		}
	}
}

func openReplyThreadTestDatabase(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("GOA4WEB_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set GOA4WEB_TEST_MYSQL_DSN to run MySQL fork query tests")
	}
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("parse MySQL DSN: %v", err)
	}
	databaseName := fmt.Sprintf("goa4web_forks_%d", time.Now().UnixNano())
	adminConfig := *cfg
	adminConfig.DBName = ""
	adminDB, err := sql.Open("mysql", adminConfig.FormatDSN())
	if err != nil {
		t.Fatalf("open MySQL admin connection: %v", err)
	}
	if _, err := adminDB.Exec("CREATE DATABASE `" + databaseName + "`"); err != nil {
		_ = adminDB.Close()
		t.Fatalf("create temporary database: %v", err)
	}
	t.Cleanup(func() {
		_, _ = adminDB.Exec("DROP DATABASE `" + databaseName + "`")
		_ = adminDB.Close()
	})
	testConfig := *cfg
	testConfig.DBName = databaseName
	database, err := sql.Open("mysql", testConfig.FormatDSN())
	if err != nil {
		t.Fatalf("open temporary database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return database
}
