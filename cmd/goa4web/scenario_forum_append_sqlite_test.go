//go:build sqlite || sqlite3

package main

import (
	"context"
	"io"
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
		"username":		{username},
		"password":		{password},
		"task":			{"Login"},
		"gorilla.csrf.Token":	{token},
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

	var baseline int
	err = dbConn.QueryRowContext(ctx, "SELECT COUNT(*) FROM comments WHERE forumthread_id = ?", threadID).Scan(&baseline)
	if err != nil {
		t.Fatalf("failed to query baseline comments: %v", err)
	}

	threadURL := ts.URL + fmt.Sprintf("/private/topic/%d/thread/%d", topicID, threadID)

	noRedirectClient := *client
	noRedirectClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	// 1. First fresh reply
	replyHTML := scenarioHTTPGet(t, client, threadURL)
	token := scenarioCSRFToken(t, replyHTML)
	form := url.Values{
		"replytext":		{"Alice first fresh reply"},
		"task":			{"Reply"},
		"gorilla.csrf.Token":	{token},
	}

	req1, err := http.NewRequest(http.MethodPost, threadURL+"/reply", strings.NewReader(form.Encode()))
	if err != nil { t.Fatal(err) }
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req1.Header.Set("Referer", threadURL)
	resp1, err := noRedirectClient.Do(req1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp1.Body.Close()
	if resp1.StatusCode != http.StatusSeeOther {
		t.Fatalf("first reply expected status 303, got %d", resp1.StatusCode)
	}

	// Prove the mutation happened
	var rowCount int
	err = dbConn.QueryRowContext(ctx, "SELECT COUNT(*) FROM comments WHERE forumthread_id = ?", threadID).Scan(&rowCount)
	if err != nil {
		t.Fatal(err)
	}
	if rowCount != baseline+1 {
		t.Fatalf("expected %d comments, got %d", baseline+1, rowCount)
	}

	var aliceCommentID int32
	var aliceUID int32
	var text string
	err = dbConn.QueryRowContext(ctx, "SELECT idcomments, users_idusers, text FROM comments WHERE forumthread_id = ? ORDER BY idcomments DESC LIMIT 1", threadID).Scan(&aliceCommentID, &aliceUID, &text)
	if err != nil {
		t.Fatalf("failed to get last comment: %v", err)
	}
	if text != "Alice first fresh reply" {
		t.Fatalf("unexpected text: %s", text)
	}

	var expectedAliceUID int32
	err = dbConn.QueryRowContext(ctx, "SELECT idusers FROM users WHERE username = 'alice'").Scan(&expectedAliceUID)
	if err != nil { t.Fatal(err) }
	if aliceUID != expectedAliceUID {
		t.Fatalf("expected alice (%d) to own new comment, got %d", expectedAliceUID, aliceUID)
	}

	// Prove server config
	if got := srv.Config.PrivateForumPostAppendWindow; got != 60 {
		t.Fatalf("PrivateForumPostAppendWindow = %d, want 60", got)
	}

	// Prove explicit append grant
	var grantCount int
	err = dbConn.QueryRowContext(ctx, "SELECT COUNT(*) FROM grants WHERE section = 'privateforum_thread' AND item = 'thread' AND item_id = ? AND action = 'append' AND user_id = ? AND active = 1", threadID, aliceUID).Scan(&grantCount)
	if err != nil { t.Fatal(err) }
	if grantCount != 1 {
		t.Fatalf("expected 1 explicit append grant, got %d", grantCount)
	}

	// Create fresh diagnostic CoreData
	querier := db.NewForDriver(dbConn, "sqlite3")
	probeCD := common.NewCoreData(ctx, querier, srv.Config).ForUser(aliceUID)
	probeCD.SetCurrentThreadAndTopic(threadID, topicID)

	probeComments, err := probeCD.ThreadComments(threadID)
	if err != nil {
		t.Fatal(err)
	}
	if len(probeComments) == 0 {
		t.Fatal("fresh Alice CoreData sees no comments")
	}

	candidate := probeComments[len(probeComments)-1]

	if candidate.Idcomments != aliceCommentID {
		t.Fatalf("candidate ID %d != aliceCommentID %d", candidate.Idcomments, aliceCommentID)
	}
	if candidate.UsersIdusers != aliceUID {
		t.Fatalf("candidate UID %d != aliceUID %d", candidate.UsersIdusers, aliceUID)
	}
	if candidate.ForumthreadID != threadID {
		t.Fatalf("candidate ThreadID %d != %d", candidate.ForumthreadID, threadID)
	}
	if !candidate.IsOwner {
		t.Fatalf("candidate is not owner")
	}
	if !candidate.Written.Valid {
		t.Fatalf("candidate written timestamp is not valid")
	}

	if !probeCD.CanAppendToComment(candidate) {
		t.Fatalf("fresh probe CanAppendToComment returned false. Grant/window/read-marker issue.")
	}

	// Check read markers
	var newerMarkerCount int
	err = dbConn.QueryRowContext(ctx, "SELECT COUNT(*) FROM content_read_markers WHERE item = 'thread' AND item_id = ? AND last_comment_id >= ? AND user_id != ?", threadID, aliceCommentID, aliceUID).Scan(&newerMarkerCount)
	if err != nil { t.Fatal(err) }
	if newerMarkerCount > 0 {
		t.Fatalf("found %d read markers blocking append", newerMarkerCount)
	}

	// Finality check
	var newerComments int
	err = dbConn.QueryRowContext(ctx, "SELECT COUNT(*) FROM comments WHERE forumthread_id = ? AND idcomments > ?", threadID, aliceCommentID).Scan(&newerComments)
	if err != nil { t.Fatal(err) }
	if newerComments > 0 {
		t.Fatalf("expected 0 newer comments, got %d", newerComments)
	}

	// Re-fetch the page. If it still says Reply, the bug is in the request-scoped template projection
	replyHTML = scenarioHTTPGet(t, client, threadURL)
	if !strings.Contains(replyHTML, "Append:") {
		t.Fatalf("HTTP GET rendered 'Reply:', but CanAppendToComment probe returned true. Bug is in GET's viewer/selection/template projection.")
	}

	// 2. Second fresh reply
	token = scenarioCSRFToken(t, replyHTML)
	form = url.Values{
		"replytext":		{"Alice second fresh reply"},
		"task":			{"Reply"},
		"gorilla.csrf.Token":	{token},
	}

	req2, err := http.NewRequest(http.MethodPost, threadURL+"/reply", strings.NewReader(form.Encode()))
	if err != nil { t.Fatal(err) }
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("Referer", threadURL)
	resp2, err := noRedirectClient.Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusSeeOther {
		t.Fatalf("second reply expected status 303, got %d", resp2.StatusCode)
	}

	// 3. Third fresh reply
	replyHTML = scenarioHTTPGet(t, client, threadURL)
	token = scenarioCSRFToken(t, replyHTML)
	form = url.Values{
		"replytext":		{"Alice third fresh reply"},
		"task":			{"Reply"},
		"gorilla.csrf.Token":	{token},
	}

	req3, err := http.NewRequest(http.MethodPost, threadURL+"/reply", strings.NewReader(form.Encode()))
	if err != nil { t.Fatal(err) }
	req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req3.Header.Set("Referer", threadURL)
	resp3, err := noRedirectClient.Do(req3)
	if err != nil {
		t.Fatal(err)
	}
	defer resp3.Body.Close()

	// 4. Fourth fresh reply
	replyHTML = scenarioHTTPGet(t, client, threadURL)
	token = scenarioCSRFToken(t, replyHTML)
	form = url.Values{
		"replytext":		{"Alice fourth fresh reply"},
		"task":			{"Reply"},
		"gorilla.csrf.Token":	{token},
	}

	req4, err := http.NewRequest(http.MethodPost, threadURL+"/reply", strings.NewReader(form.Encode()))
	if err != nil { t.Fatal(err) }
	req4.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req4.Header.Set("Referer", threadURL)
	resp4, err := noRedirectClient.Do(req4)
	if err != nil {
		t.Fatal(err)
	}
	defer resp4.Body.Close()

	// Final verification
	var finalCount int
	err = dbConn.QueryRowContext(ctx, "SELECT COUNT(*) FROM comments WHERE forumthread_id = ?", threadID).Scan(&finalCount)
	if err != nil { t.Fatal(err) }
	if err != nil { t.Fatal(err) }
	if finalCount != baseline+1 {
		t.Fatalf("expected %d comments, got %d", baseline+1, finalCount)
	}

	var finalID int32
	var finalOwner int32
	var finalText string
	err = dbConn.QueryRowContext(ctx, "SELECT idcomments, users_idusers, text FROM comments WHERE forumthread_id = ? ORDER BY idcomments DESC LIMIT 1", threadID).Scan(&finalID, &finalOwner, &finalText)
	if err != nil { t.Fatal(err) }

	if finalID != aliceCommentID {
		t.Fatalf("expected final comment ID to remain %d, got %d", aliceCommentID, finalID)
	}
	if finalOwner != aliceUID {
		t.Fatalf("expected final comment to belong to alice (%d), got %d", aliceUID, finalOwner)
	}
	if !strings.Contains(finalText, "Alice first fresh reply") || !strings.Contains(finalText, "Alice fourth fresh reply") {
		t.Fatalf("last comment doesn't contain appended texts: %s", finalText)
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

	var baseline int
	err = dbConn.QueryRowContext(ctx, "SELECT COUNT(*) FROM comments WHERE forumthread_id = ?", threadID).Scan(&baseline)
	if err != nil {
		t.Fatalf("failed to query baseline comments: %v", err)
	}

	threadURL := ts.URL + fmt.Sprintf("/private/topic/%d/thread/%d", topicID, threadID)

	noRedirectClient := *client
	noRedirectClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	replyHTML := scenarioHTTPGet(t, client, threadURL)
	token := scenarioCSRFToken(t, replyHTML)
	form := url.Values{
		"replytext":		{"Alice first fresh reply (disabled)"},
		"task":			{"Reply"},
		"gorilla.csrf.Token":	{token},
	}

	req1, err := http.NewRequest(http.MethodPost, threadURL+"/reply", strings.NewReader(form.Encode()))
	if err != nil { t.Fatal(err) }
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req1.Header.Set("Referer", threadURL)
	resp1, err := noRedirectClient.Do(req1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp1.Body.Close()
	if resp1.StatusCode != http.StatusSeeOther {
		body, _ := io.ReadAll(resp1.Body)
		t.Fatalf("reply 1 status = %d, body=%s", resp1.StatusCode, body)
	}

	var count1 int
	err = dbConn.QueryRowContext(ctx, "SELECT COUNT(*) FROM comments WHERE forumthread_id = ?", threadID).Scan(&count1)
	if err != nil { t.Fatal(err) }
	if count1 != baseline+1 {
		t.Fatalf("expected %d comments after first reply, got %d", baseline+1, count1)
	}

	replyHTML = scenarioHTTPGet(t, client, threadURL)
	if strings.Contains(replyHTML, "Append:") {
		t.Fatalf("expected 'Reply:' in form, got 'Append:'")
	}

	token = scenarioCSRFToken(t, replyHTML)
	form = url.Values{
		"replytext":		{"Alice second fresh reply (disabled)"},
		"task":			{"Reply"},
		"gorilla.csrf.Token":	{token},
	}
	req2, err := http.NewRequest(http.MethodPost, threadURL+"/reply", strings.NewReader(form.Encode()))
	if err != nil { t.Fatal(err) }
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("Referer", threadURL)
	resp2, err := noRedirectClient.Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusSeeOther {
		body, _ := io.ReadAll(resp2.Body)
		t.Fatalf("reply 2 status = %d, body=%s", resp2.StatusCode, body)
	}

	var count2 int
	err = dbConn.QueryRowContext(ctx, "SELECT COUNT(*) FROM comments WHERE forumthread_id = ?", threadID).Scan(&count2)
	if err != nil { t.Fatal(err) }
	if count2 != baseline+2 {
		t.Fatalf("expected %d comments after second reply, got %d", baseline+2, count2)
	}

	rows, err := dbConn.QueryContext(ctx, "SELECT idcomments, users_idusers, text FROM comments WHERE forumthread_id = ? ORDER BY idcomments DESC LIMIT 2", threadID)
	if err != nil {
		t.Fatalf("failed to query new comments: %v", err)
	}
	defer rows.Close()

	type commentRow struct {
		id	int32
		uid	int32
		text	string
	}
	var newComments []commentRow
	for rows.Next() {
		var c commentRow
		rows.Scan(&c.id, &c.uid, &c.text)
		newComments = append(newComments, c)
	}

	if len(newComments) != 2 {
		t.Fatalf("expected to read 2 new comments, got %d", len(newComments))
	}

	c2 := newComments[0]
	c1 := newComments[1]

	var aliceID int32
	err = dbConn.QueryRowContext(ctx, "SELECT idusers FROM users WHERE username = 'alice'").Scan(&aliceID)
	if err != nil { t.Fatal(err) }

	if c1.uid != aliceID || c2.uid != aliceID {
		t.Fatalf("expected both comments to belong to alice (%d), got uid1=%d, uid2=%d", aliceID, c1.uid, c2.uid)
	}
	if c1.id == c2.id {
		t.Fatalf("expected two distinct comment IDs, but got same ID %d", c1.id)
	}
	if !strings.Contains(c1.text, "Alice first fresh reply (disabled)") {
		t.Fatalf("first comment text unexpected: %s", c1.text)
	}
	if !strings.Contains(c2.text, "Alice second fresh reply (disabled)") {
		t.Fatalf("second comment text unexpected: %s", c2.text)
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

	topicID, threadID := getTopicAndThreadIDForUser(ctx, t, dbConn, "alice")	// staff room thread

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

	var baseline int
	err = dbConn.QueryRowContext(ctx, "SELECT COUNT(*) FROM comments WHERE forumthread_id = ?", threadID).Scan(&baseline)
	if err != nil {
		t.Fatalf("failed to query baseline comments: %v", err)
	}

	threadURL := ts.URL + fmt.Sprintf("/private/topic/%d/thread/%d", topicID, threadID)

	noRedirectClient := *client
	noRedirectClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	replyHTML := scenarioHTTPGet(t, client, threadURL)
	token := scenarioCSRFToken(t, replyHTML)
	form := url.Values{
		"replytext":		{"Bob first fresh reply (denied)"},
		"task":			{"Reply"},
		"gorilla.csrf.Token":	{token},
	}

	req1, err := http.NewRequest(http.MethodPost, threadURL+"/reply", strings.NewReader(form.Encode()))
	if err != nil { t.Fatal(err) }
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req1.Header.Set("Referer", threadURL)
	resp1, err := noRedirectClient.Do(req1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp1.Body.Close()
	if resp1.StatusCode != http.StatusSeeOther {
		t.Fatalf("first reply expected status 303, got %d", resp1.StatusCode)
	}

	replyHTML = scenarioHTTPGet(t, client, threadURL)
	if strings.Contains(replyHTML, "Append:") {
		t.Fatalf("expected 'Reply:' in form, got 'Append:'")
	}

	token = scenarioCSRFToken(t, replyHTML)
	form = url.Values{
		"replytext":		{"Bob second fresh reply (denied)"},
		"task":			{"Reply"},
		"gorilla.csrf.Token":	{token},
	}

	req2, err := http.NewRequest(http.MethodPost, threadURL+"/reply", strings.NewReader(form.Encode()))
	if err != nil { t.Fatal(err) }
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("Referer", threadURL)
	resp2, err := noRedirectClient.Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusSeeOther {
		t.Fatalf("second reply expected status 303, got %d", resp2.StatusCode)
	}

	var finalCount int
	err = dbConn.QueryRowContext(ctx, "SELECT COUNT(*) FROM comments WHERE forumthread_id = ?", threadID).Scan(&finalCount)
	if err != nil { t.Fatal(err) }
	if err != nil { t.Fatal(err) }
	if finalCount != baseline+2 {
		t.Fatalf("expected %d comments, got %d", baseline+2, finalCount)
	}
}
