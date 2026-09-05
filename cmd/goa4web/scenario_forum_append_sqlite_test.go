//go:build sqlite || sqlite3

package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func newTestScenarioServeCmd(t *testing.T, args []string) (*scenarioServeCmd, func()) {
	t.Helper()
	root, err := parseRoot(args)
	if err != nil {
		t.Fatalf("parseRoot: %v", err)
	}
	parent, err := parseScenarioCmd(root, []string{"serve", args[len(args)-1]})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}
	serveCmd, err := parseScenarioServeCmd(parent, []string{args[len(args)-1]})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}
	serveCmd.fsys = os.DirFS("../../")
	return serveCmd, func() { root.Close() }
}

func doScenarioLogin(t *testing.T, ts *httptest.Server, client *http.Client, username, password string) {
	loginHTML := scenarioHTTPGet(t, client, ts.URL+"/login")
	token := scenarioCSRFToken(t, loginHTML)
	form := url.Values{
		"username":           {username},
		"password":           {password},
		"task":               {"Login"},
		"gorilla.csrf.Token": {token},
	}
	scenarioHTTPPostForm(t, client, ts.URL+"/login", form)
}

func getTopicAndThreadIDForUser(ctx context.Context, t *testing.T, dbConn db.DBTX, username string) (int32, int32) {
	var uid int32
	err := dbConn.QueryRowContext(ctx, "SELECT idusers FROM users WHERE username = ?;", username).Scan(&uid)
	if err != nil {
		t.Fatalf("query user: %v", err)
	}
	var topicID, threadID int32
	err = dbConn.QueryRowContext(ctx, "SELECT idforumtopic FROM forumtopic WHERE title = 'Staff Room';").Scan(&topicID)
	if err != nil {
		t.Fatalf("query Staff Room topic: %v", err)
	}
	err = dbConn.QueryRowContext(ctx, "SELECT forumthread_id FROM comments WHERE text = 'Welcome to the staff room. This conversation is for Alice and Bob.';").Scan(&threadID)
	if err != nil {
		t.Fatalf("query staff-welcome thread via comment: %v", err)
	}
	return topicID, threadID
}

func TestScenarioForumAppend_RapidPosts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serveCmd, cleanupRoot := newTestScenarioServeCmd(t, []string{"goa4web", "--private-forum-post-append-window=60", "scenario", "serve", "testdata/scenarios/100-private-forum"})
	defer cleanupRoot()

	srv, dbConn, cleanupServe, err := serveCmd.Bootstrap(ctx)
	if err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	defer cleanupServe()

	topicID, threadID := getTopicAndThreadIDForUser(ctx, t, dbConn, "alice")

	ts := httptest.NewServer(srv.Router)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	doScenarioLogin(t, ts, client, "alice", "alice-test")

	querier := db.NewForDriver(dbConn, "sqlite3")
	cd := common.NewCoreData(ctx, querier, srv.Config)
	comments, err := cd.ThreadComments(threadID)
	if err != nil {
		t.Fatalf("ThreadComments baseline: %v", err)
	}
	baseline := len(comments)

	threadURL := ts.URL + fmt.Sprintf("/private/topic/%d/thread/%d", topicID, threadID)

	// 1. First fresh reply
	replyHTML := scenarioHTTPGet(t, client, threadURL)
	token := scenarioCSRFToken(t, replyHTML)
	form := url.Values{
		"replytext":          {"Alice first fresh reply"},
		"task":               {"Reply"},
		"gorilla.csrf.Token": {token},
	}
	scenarioHTTPPostForm(t, client, threadURL+"/reply", form)

	// Check UI for Append:
	replyHTML = scenarioHTTPGet(t, client, threadURL)
	if !strings.Contains(replyHTML, "Append:") {
		t.Fatalf("expected 'Append:' in reply form, got: %s", replyHTML)
	}

	// 2. Second fresh reply
	token = scenarioCSRFToken(t, replyHTML)
	form = url.Values{
		"replytext":          {"Alice second fresh reply"},
		"task":               {"Reply"},
		"gorilla.csrf.Token": {token},
	}
	scenarioHTTPPostForm(t, client, threadURL+"/reply", form)

	// 3. Third fresh reply
	replyHTML = scenarioHTTPGet(t, client, threadURL)
	token = scenarioCSRFToken(t, replyHTML)
	form = url.Values{
		"replytext":          {"Alice third fresh reply"},
		"task":               {"Reply"},
		"gorilla.csrf.Token": {token},
	}
	scenarioHTTPPostForm(t, client, threadURL+"/reply", form)

	// 4. Fourth fresh reply
	replyHTML = scenarioHTTPGet(t, client, threadURL)
	token = scenarioCSRFToken(t, replyHTML)
	form = url.Values{
		"replytext":          {"Alice fourth fresh reply"},
		"task":               {"Reply"},
		"gorilla.csrf.Token": {token},
	}
	scenarioHTTPPostForm(t, client, threadURL+"/reply", form)

	comments, err = cd.ThreadComments(threadID)
	if err != nil {
		t.Fatalf("ThreadComments end: %v", err)
	}

	if len(comments) != baseline+1 {
		t.Fatalf("expected %d comments, got %d", baseline+1, len(comments))
	}

	lastComment := comments[len(comments)-1]
	if !strings.Contains(lastComment.Text.String, "Alice first fresh reply") || !strings.Contains(lastComment.Text.String, "Alice fourth fresh reply") {
		t.Fatalf("last comment doesn't contain appended texts: %s", lastComment.Text.String)
	}
}

func TestScenarioForumAppend_DisabledByConfig(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serveCmd, cleanupRoot := newTestScenarioServeCmd(t, []string{"goa4web", "--private-forum-post-append-window=0", "scenario", "serve", "testdata/scenarios/100-private-forum"})
	defer cleanupRoot()

	srv, dbConn, cleanupServe, err := serveCmd.Bootstrap(ctx)
	if err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	defer cleanupServe()

	topicID, threadID := getTopicAndThreadIDForUser(ctx, t, dbConn, "alice")

	ts := httptest.NewServer(srv.Router)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	doScenarioLogin(t, ts, client, "alice", "alice-test")

	querier := db.NewForDriver(dbConn, "sqlite3")
	cd := common.NewCoreData(ctx, querier, srv.Config)
	comments, _ := cd.ThreadComments(threadID)
	baseline := len(comments)

	threadURL := ts.URL + fmt.Sprintf("/private/topic/%d/thread/%d", topicID, threadID)

	replyHTML := scenarioHTTPGet(t, client, threadURL)
	token := scenarioCSRFToken(t, replyHTML)
	form := url.Values{
		"replytext":          {"Alice first fresh reply (disabled)"},
		"task":               {"Reply"},
		"gorilla.csrf.Token": {token},
	}
	scenarioHTTPPostForm(t, client, threadURL+"/reply", form)

	replyHTML = scenarioHTTPGet(t, client, threadURL)
	if strings.Contains(replyHTML, "Append:") {
		t.Fatalf("expected 'Reply:' in form, got 'Append:'")
	}

	token = scenarioCSRFToken(t, replyHTML)
	form = url.Values{
		"replytext":          {"Alice second fresh reply (disabled)"},
		"task":               {"Reply"},
		"gorilla.csrf.Token": {token},
	}
	scenarioHTTPPostForm(t, client, threadURL+"/reply", form)

	comments, _ = cd.ThreadComments(threadID)

	if len(comments) != baseline+2 {
		t.Fatalf("expected %d comments, got %d", baseline+2, len(comments))
	}
	if comments[len(comments)-1].Idcomments == comments[len(comments)-2].Idcomments {
		t.Fatalf("expected two distinct comment IDs, but they were the same")
	}
}

func TestScenarioForumAppend_PermissionDenial(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serveCmd, cleanupRoot := newTestScenarioServeCmd(t, []string{"goa4web", "--private-forum-post-append-window=60", "scenario", "serve", "testdata/scenarios/100-private-forum"})
	defer cleanupRoot()

	srv, dbConn, cleanupServe, err := serveCmd.Bootstrap(ctx)
	if err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	defer cleanupServe()

	var bobID int32
	err = dbConn.QueryRowContext(ctx, "SELECT idusers FROM users WHERE username = 'bob';").Scan(&bobID)
	if err != nil {
		t.Fatalf("query bob idusers: %v", err)
	}

	topicID, threadID := getTopicAndThreadIDForUser(ctx, t, dbConn, "alice") // staff room thread

	// Remove Bob's append grant so he only has reply.
	_, err = dbConn.Exec("DELETE FROM grants WHERE item_id = ? AND action = 'append' AND user_id = ?", threadID, bobID)
	if err != nil {
		t.Fatalf("failed to revoke grant: %v", err)
	}

	ts := httptest.NewServer(srv.Router)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	doScenarioLogin(t, ts, client, "bob", "bob-test")

	querier := db.NewForDriver(dbConn, "sqlite3")
	cd := common.NewCoreData(ctx, querier, srv.Config)
	comments, _ := cd.ThreadComments(threadID)
	baseline := len(comments)

	threadURL := ts.URL + fmt.Sprintf("/private/topic/%d/thread/%d", topicID, threadID)

	replyHTML := scenarioHTTPGet(t, client, threadURL)
	token := scenarioCSRFToken(t, replyHTML)
	form := url.Values{
		"replytext": {"Bob first fresh reply (denied)"},
		"task": {"Reply"},
		"gorilla.csrf.Token": {token},
	}
	scenarioHTTPPostForm(t, client, threadURL+"/reply", form)

	replyHTML = scenarioHTTPGet(t, client, threadURL)
	if strings.Contains(replyHTML, "Append:") {
		t.Fatalf("expected 'Reply:' in form, got 'Append:'")
	}

	token = scenarioCSRFToken(t, replyHTML)
	form = url.Values{
		"replytext": {"Bob second fresh reply (denied)"},
		"task": {"Reply"},
		"gorilla.csrf.Token": {token},
	}
	scenarioHTTPPostForm(t, client, threadURL+"/reply", form)

	comments, _ = cd.ThreadComments(threadID)

	if len(comments) != baseline+2 {
		t.Fatalf("expected %d comments, got %d", baseline+2, len(comments))
	}
	if comments[len(comments)-1].Idcomments == comments[len(comments)-2].Idcomments {
		t.Fatalf("expected two distinct comment IDs, but they were the same")
	}
}
