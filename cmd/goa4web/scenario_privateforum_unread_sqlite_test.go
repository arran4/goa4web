//go:build sqlite || sqlite3

package main

import (
	"context"
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

func TestE2EPrivateForumUnread(t *testing.T) {
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

	// Refresh CSRF token for subsequent tests if necessary
	noRedirectClient := *client
	noRedirectClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	// 1. Unscoped Unread
	body := scenarioHTTPGet(t, client, httpServer.URL+"/private/unread")

	// In the 100-private-forum scenario, Alice has:
	// - staff-welcome (opened by Alice, replied by Bob) -> Unread
	// - bob-staff-check-in (opened by Bob, replied by Alice) -> Read
	// - coordination-plan (opened by Alice, replied by Carol) -> Unread
	// So Alice has 2 unread private threads total.

	countMatches := strings.Count(body, "class=\"thread\"")
	require.Equal(t, 2, countMatches, "Expected 2 unread private threads for Alice in unscoped All Unread list. Body: %s", body)
	require.Contains(t, body, "Welcome to the staff room.", "Expected staff-welcome thread to be present")
	require.Contains(t, body, "Coordination plan for Alice and Carol", "Expected coordination-plan thread to be present")
	require.NotContains(t, body, "Bob opening a second Staff Room thread to exercise participant thread creation.", "Expected bob-staff-check-in to NOT be present")

	// Verify Private Custom Index
	body = scenarioHTTPGet(t, client, httpServer.URL+"/private")
	require.Contains(t, body, "All Unread (2)", "Expected CustomIndex to render All Unread (2)")

	var staffRoomTopicID string
	var coordinationTopicID string
	var projectRoomTopicID string

	// Since we know the titles, let's just get the DB IDs to be safe
	err = sqlDB.QueryRowContext(ctx, "SELECT idforumtopic FROM forumtopic WHERE title = 'Staff Room'").Scan(&staffRoomTopicID)
	require.NoError(t, err)
	err = sqlDB.QueryRowContext(ctx, "SELECT idforumtopic FROM forumtopic WHERE title = 'Coordination'").Scan(&coordinationTopicID)
	require.NoError(t, err)
	err = sqlDB.QueryRowContext(ctx, "SELECT idforumtopic FROM forumtopic WHERE title = 'Project Room'").Scan(&projectRoomTopicID)
	require.NoError(t, err)

	body = scenarioHTTPGet(t, client, httpServer.URL+"/private/topic/"+staffRoomTopicID+"/unread")
	countMatches = strings.Count(body, "class=\"thread\"")
	require.Equal(t, 1, countMatches, "Expected 1 unread private thread for Alice in Staff Room topic")
	require.Contains(t, body, "Welcome to the staff room.", "Expected staff-welcome thread to be present")

	body = scenarioHTTPGet(t, client, httpServer.URL+"/private/topic/"+coordinationTopicID+"/unread")
	countMatches = strings.Count(body, "class=\"thread\"")
	require.Equal(t, 1, countMatches, "Expected 1 unread private thread for Alice in Coordination topic")
	require.Contains(t, body, "Coordination plan for Alice and Carol", "Expected coordination-plan thread to be present")

	// Inaccessible Project Room content remains absent (Topic 3)
	resp, err := noRedirectClient.Get(httpServer.URL + "/private/topic/" + projectRoomTopicID + "/unread")
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode, "Expected 404 for inaccessible topic unread")

	// Verify Unread in Topic links
	body = scenarioHTTPGet(t, client, httpServer.URL+"/private/topic/"+staffRoomTopicID)
	require.Contains(t, body, "Unread in Topic (1)", "Expected CustomIndex to render Unread in Topic (1) for Staff Room")

	body = scenarioHTTPGet(t, client, httpServer.URL+"/private/topic/"+coordinationTopicID)
	require.Contains(t, body, "Unread in Topic (1)", "Expected CustomIndex to render Unread in Topic (1) for Coordination")
}
