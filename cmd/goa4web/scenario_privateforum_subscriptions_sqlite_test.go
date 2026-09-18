//go:build sqlite || sqlite3

package main

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/testdata/scenarios"
	"github.com/stretchr/testify/require"
)

func TestE2EPrivateForumSubscriptions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	root, err := parseRoot([]string{"goa4web", "scenario", "serve", "100-private-forum"})
	require.NoError(t, err)
	defer root.Close()

	parent, err := parseScenarioCmd(root, []string{"serve", "100-private-forum"})
	require.NoError(t, err)
	serveCmd, err := parseScenarioServeCmd(parent, []string{"100-private-forum"})
	require.NoError(t, err)
	serveCmd.fsys = scenarios.FS

	srv, sqlDB, cleanup, err := serveCmd.Bootstrap(ctx)
	require.NoError(t, err)
	defer cleanup()

	httpServer := httptest.NewServer(srv.Router)
	defer httpServer.Close()

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	client := &http.Client{Jar: jar}

	loginPage := scenarioHTTPGet(t, client, httpServer.URL+"/login")
	loginToken := scenarioCSRFToken(t, loginPage)
	loginForm := url.Values{
		"username":           {"alice"},
		"password":           {"alice-test"},
		"task":               {"Login"},
		"gorilla.csrf.Token": {loginToken},
	}
	loginResponse := scenarioHTTPPostForm(t, client, httpServer.URL+"/login", loginForm)
	if strings.Contains(loginResponse, "Invalid username or password") {
		t.Fatal("Alice's scenario credentials were rejected")
	}

	// Fetch read markers for Staff Welcome thread
	var staffWelcomeThreadID int32
	err = sqlDB.QueryRowContext(ctx, "SELECT forumthread_id FROM comments WHERE text LIKE '%Welcome to the staff room%' LIMIT 1").Scan(&staffWelcomeThreadID)
	require.NoError(t, err)

	// Fetch alice marker
	var aliceMarker sql.NullInt32
	err = sqlDB.QueryRowContext(ctx, "SELECT last_comment_id FROM content_read_markers WHERE user_id = (SELECT idusers FROM users WHERE username = 'alice') AND item_id = ?", staffWelcomeThreadID).Scan(&aliceMarker)
	require.NoError(t, err)
	// Fetch bob marker
	var bobMarker sql.NullInt32
	err = sqlDB.QueryRowContext(ctx, "SELECT last_comment_id FROM content_read_markers WHERE user_id = (SELECT idusers FROM users WHERE username = 'bob') AND item_id = ?", staffWelcomeThreadID).Scan(&bobMarker)
	require.NoError(t, err)

	// Fetch comment IDs
	var firstPostID int32
	err = sqlDB.QueryRowContext(ctx, "SELECT idcomments FROM comments WHERE forumthread_id = ? ORDER BY written ASC LIMIT 1", staffWelcomeThreadID).Scan(&firstPostID)
	require.NoError(t, err)
	var latestPostID int32
	err = sqlDB.QueryRowContext(ctx, "SELECT idcomments FROM comments WHERE forumthread_id = ? ORDER BY written DESC LIMIT 1", staffWelcomeThreadID).Scan(&latestPostID)
	require.NoError(t, err)

	// After Alice read staff-welcome in the scenario, her marker advanced to Bob's *first* reply. Then Bob replied *again*.
	// Therefore Bob's marker should be latestPostID. Alice's should be the previous one.
	require.True(t, bobMarker.Valid)
	require.Equal(t, latestPostID, bobMarker.Int32, "Bob's marker should be at his latest reply")

	require.True(t, aliceMarker.Valid)
	require.NotEqual(t, latestPostID, aliceMarker.Int32, "Alice's marker should NOT advance to Bob's new reply")

	// 1. Unread state after reading and then receiving a reply
	// In the scenario:
	// - alice-read.event: marks staff-welcome as read.
	// - bob-staff-reply.event: Bob replies to staff-welcome, making it unread again.
	body := scenarioHTTPGet(t, client, httpServer.URL+"/private/unread")

	// Alice has 2 unread threads total now:
	// staff-welcome, coordination-plan
	countMatches := strings.Count(body, "class=\"thread\"")
	require.Equal(t, 2, countMatches, "Expected 2 unread private threads for Alice in unscoped All Unread list. Body: %s", body)
	require.Contains(t, body, "Welcome to the staff room.", "Expected staff-welcome thread to be present")
	require.Contains(t, body, "Coordination plan for Alice and Carol", "Expected coordination-plan thread to be present")
	require.NotContains(t, body, "Bob opening a second Staff Room thread to exercise participant thread creation.", "Expected bob-staff-check-in to NOT be present")

	var staffRoomTopicID string
	err = sqlDB.QueryRowContext(ctx, "SELECT idforumtopic FROM forumtopic WHERE title = 'Staff Room'").Scan(&staffRoomTopicID)
	require.NoError(t, err)

	rows, err := sqlDB.QueryContext(ctx, "SELECT pattern FROM subscriptions WHERE users_idusers = (SELECT idusers FROM users WHERE username = 'alice')")
	require.NoError(t, err)
	defer rows.Close()
	var patterns []string
	for rows.Next() {
		var p string
		require.NoError(t, rows.Scan(&p))
		patterns = append(patterns, p)
	}

	hasStaffRoomTopicSub := false
	for _, p := range patterns {
		if strings.Contains(p, "topic/"+staffRoomTopicID+"/*") && strings.Contains(p, "create thread") {
			hasStaffRoomTopicSub = true
		}
	}
	require.False(t, hasStaffRoomTopicSub, "Expected Alice to have no topic subscriptions to staff-room due to unsubscribe")
}
