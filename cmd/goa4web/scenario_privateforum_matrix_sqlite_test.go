//go:build sqlite || sqlite3

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/testdata/scenarios"
)

func TestScenarioServePrivateForumPermissionMatrix(t *testing.T) {
	ctx := context.Background()

	root, err := parseRoot([]string{"goa4web", "scenario", "serve", "100-private-forum"})
	if err != nil {
		t.Fatalf("parseRoot: %v", err)
	}
	defer root.Close()
	parent, err := parseScenarioCmd(root, []string{"serve", "100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}
	serveCmd, err := parseScenarioServeCmd(parent, []string{"100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}
	serveCmd.fsys = scenarios.FS

	srv, dbConn, cleanup, err := serveCmd.Bootstrap(ctx)
	if err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	defer cleanup()

	querier := db.NewForDriver(dbConn, "sqlite3")
	baseCD := common.NewCoreData(ctx, querier, srv.Config)
	userIDs := loadScenarioUserIDs(t, ctx, dbConn)
	topicIDs := loadScenarioTopicIDs(t, ctx, dbConn)
	threadIDs := loadScenarioThreadIDs(t, ctx, dbConn)

	wantTopics := map[string][]string{
		"alice": {"Coordination", "Staff Room"},
		"bob":   {"Staff Room"},
		"carol": {"Coordination", "Project Room"},
		"dave":  {"Project Room"},
	}
	for username, expected := range wantTopics {
		t.Run("topic-list/"+username, func(t *testing.T) {
			userCD := baseCD.ForUser(userIDs[username])
			topics, err := userCD.PrivateForumTopics()
			if err != nil {
				t.Fatalf("PrivateForumTopics: %v", err)
			}
			got := make([]string, 0, len(topics))
			for _, topic := range topics {
				got = append(got, topic.DisplayTitle)
				if topic.TotalParticipants != 2 {
					t.Errorf("%q participant count = %d, want 2", topic.DisplayTitle, topic.TotalParticipants)
				}
			}
			sort.Strings(got)
			if fmt.Sprint(got) != fmt.Sprint(expected) {
				t.Fatalf("visible topics = %v, want %v", got, expected)
			}
		})
	}

	members := map[string]map[string]bool{
		"alice": {"Staff Room": true, "Coordination": true},
		"bob":   {"Staff Room": true},
		"carol": {"Project Room": true, "Coordination": true},
		"dave":  {"Project Room": true},
	}
	threadTopics := map[string]string{
		"staff-welcome":     "Staff Room",
		"staff-check-in":    "Staff Room",
		"project-kickoff":   "Project Room",
		"coordination-plan": "Coordination",
	}
	for username, userID := range userIDs {
		for topicName, topicID := range topicIDs {
			t.Run("topic-open/"+username+"/"+topicName, func(t *testing.T) {
				_, err := baseCD.ForUser(userID).ForumTopicByID(topicID)
				assertScenarioVisibility(t, err, members[username][topicName])
			})
		}
		for threadName, threadID := range threadIDs {
			t.Run("thread-open/"+username+"/"+threadName, func(t *testing.T) {
				_, err := baseCD.ForUser(userID).ForumThreadByID(threadID)
				assertScenarioVisibility(t, err, members[username][threadTopics[threadName]])
			})
		}
	}

	for username, userID := range userIDs {
		for threadName, threadID := range threadIDs {
			topicName := threadTopics[threadName]
			if members[username][topicName] {
				continue
			}
			t.Run("reply-denied/"+username+"/"+threadName, func(t *testing.T) {
				var before int
				if err := dbConn.QueryRowContext(ctx, "SELECT count(*) FROM comments WHERE forumthread_id = ?", threadID).Scan(&before); err != nil {
					t.Fatalf("count comments before reply: %v", err)
				}
				_, err := baseCD.ReplyForumThread(ctx, common.ReplyForumThreadParams{
					ActorID:        userID,
					ThreadID:       threadID,
					Text:           "This unauthorized reply must not be stored.",
					Private:        true,
					EnforceHandler: true,
				})
				if err == nil {
					t.Fatal("unauthorized reply unexpectedly succeeded")
				}
				var notFound common.ForumResourceNotFoundError
				var forbidden common.ForumOperationForbiddenError
				if !errors.As(err, &notFound) && !errors.As(err, &forbidden) {
					t.Fatalf("unauthorized reply error = %T %v, want not-found or forbidden", err, err)
				}
				var after int
				if err := dbConn.QueryRowContext(ctx, "SELECT count(*) FROM comments WHERE forumthread_id = ?", threadID).Scan(&after); err != nil {
					t.Fatalf("count comments after reply: %v", err)
				}
				if after != before {
					t.Fatalf("comment count changed from %d to %d after denied reply", before, after)
				}
			})
		}
	}

	wantThreadUsers := map[string][]string{
		"staff-welcome":     {"alice", "bob"},
		"staff-check-in":    {"alice", "bob"},
		"project-kickoff":   {"carol", "dave"},
		"coordination-plan": {"alice", "carol"},
	}
	for threadName, expected := range wantThreadUsers {
		t.Run("thread-grants/"+threadName, func(t *testing.T) {
			grants, err := querier.AdminListGrantsByThreadID(ctx, sql.NullInt32{Int32: threadIDs[threadName], Valid: true})
			if err != nil {
				t.Fatalf("AdminListGrantsByThreadID: %v", err)
			}
			seen := map[string]bool{}
			for _, grant := range grants {
				if grant.Username.Valid {
					seen[grant.Username.String] = true
				}
			}
			got := make([]string, 0, len(seen))
			for username := range seen {
				got = append(got, username)
			}
			sort.Strings(got)
			if fmt.Sprint(got) != fmt.Sprint(expected) {
				t.Fatalf("thread grant users = %v, want %v", got, expected)
			}
		})
	}
}

func loadScenarioUserIDs(t *testing.T, ctx context.Context, dbConn *sql.DB) map[string]int32 {
	t.Helper()
	ids := make(map[string]int32)
	for _, username := range []string{"alice", "bob", "carol", "dave"} {
		var id int32
		if err := dbConn.QueryRowContext(ctx, "SELECT idusers FROM users WHERE username = ?", username).Scan(&id); err != nil {
			t.Fatalf("query user %q: %v", username, err)
		}
		ids[username] = id
	}
	return ids
}

func loadScenarioTopicIDs(t *testing.T, ctx context.Context, dbConn *sql.DB) map[string]int32 {
	t.Helper()
	ids := make(map[string]int32)
	for _, title := range []string{"Staff Room", "Project Room", "Coordination"} {
		var id int32
		if err := dbConn.QueryRowContext(ctx, "SELECT idforumtopic FROM forumtopic WHERE title = ?", title).Scan(&id); err != nil {
			t.Fatalf("query topic %q: %v", title, err)
		}
		ids[title] = id
	}
	return ids
}

func loadScenarioThreadIDs(t *testing.T, ctx context.Context, dbConn *sql.DB) map[string]int32 {
	t.Helper()
	texts := map[string]string{
		"staff-welcome":     "Welcome to the staff room. This conversation is for Alice and Bob.",
		"staff-check-in":    "Bob opening a second Staff Room thread to exercise participant thread creation.",
		"project-kickoff":   "Project Room kickoff for Carol and Dave.",
		"coordination-plan": "Coordination plan for Alice and Carol across the two teams.",
	}
	ids := make(map[string]int32)
	for name, body := range texts {
		var id int32
		if err := dbConn.QueryRowContext(ctx, "SELECT forumthread_id FROM comments WHERE text = ?", body).Scan(&id); err != nil {
			t.Fatalf("query thread %q: %v", name, err)
		}
		ids[name] = id
	}
	return ids
}

func assertScenarioVisibility(t *testing.T, err error, visible bool) {
	t.Helper()
	if visible && err != nil {
		t.Fatalf("authorized resource lookup failed: %v", err)
	}
	if !visible && !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unauthorized resource lookup error = %v, want sql.ErrNoRows", err)
	}
}
