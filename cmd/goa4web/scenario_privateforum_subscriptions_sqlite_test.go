//go:build sqlite || sqlite3

package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/scenario"
	"github.com/arran4/goa4web/testdata/scenarios"
	"github.com/stretchr/testify/require"
)

func buildScenarioTxtar(sc *scenario.Scenario) string {
	var sb strings.Builder
	sb.WriteString("-- scenario.meta --\n")
	sb.WriteString("Format: " + sc.Meta.Format + "\n")
	sb.WriteString("Name: " + sc.Meta.Name + "\n")
	sb.WriteString("Description: " + sc.Meta.Description + "\n\n")

	for _, evt := range sc.Events {
		sb.WriteString("-- " + evt.File + " --\n")
		sb.WriteString("Op: " + evt.Op + "\n")
		for _, k := range evt.Headers.Keys() {
			for _, v := range evt.Headers.Values(k) {
				if k != "Op" {
					sb.WriteString(k + ": " + v + "\n")
				}
			}
		}
		if evt.Body != "" {
			sb.WriteString("\n" + evt.Body + "\n")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func runScenarioAndAssert(t *testing.T, eventsToRun int, assertFunc func(*testing.T, *http.Client, string, *sql.DB, *common.CoreData)) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	data, err := fs.ReadFile(scenarios.FS, "100-private-forum/scenario.txtar")
	require.NoError(t, err)

	sc, err := scenario.Parse(data, scenarios.FS)
	require.NoError(t, err)

	sc.Events = sc.Events[:eventsToRun]
	txtarContent := buildScenarioTxtar(sc)

	tmpDir := t.TempDir()
	scenarioFile := filepath.Join(tmpDir, "scenario.txtar")
	err = os.WriteFile(scenarioFile, []byte(txtarContent), 0644)
	require.NoError(t, err)

	root, err := parseRoot([]string{"goa4web", "scenario", "serve", scenarioFile})
	require.NoError(t, err)
	defer root.Close()

	parent, err := parseScenarioCmd(root, []string{"serve", scenarioFile})
	require.NoError(t, err)
	serveCmd, err := parseScenarioServeCmd(parent, []string{scenarioFile})
	require.NoError(t, err)

	// Since we are writing to real FS, we can use nil fsys
	serveCmd.fsys = nil

	srv, sqlDB, cleanup, err := serveCmd.Bootstrap(ctx)
	require.NoError(t, err)
	defer cleanup()

	cd := common.NewCoreData(ctx, db.NewForDriver(sqlDB, "sqlite3"), srv.Config)

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
		t.Fatalf("Alice's scenario credentials were rejected because test state wasn't persistent? Oh, it's starting fresh each time! %d events run", eventsToRun)
	}

	assertFunc(t, client, httpServer.URL, sqlDB, cd)
}

func coreDataForUser(ctx context.Context, cd *common.CoreData, userID int32) *common.CoreData {
	return cd.ForUser(userID)
}

func TestE2EPrivateForumSubscriptionsIncremental(t *testing.T) {
	data, err := fs.ReadFile(scenarios.FS, "100-private-forum/scenario.txtar")
	require.NoError(t, err)

	sc, err := scenario.Parse(data, scenarios.FS)
	require.NoError(t, err)

	var baseEventsCount, readEventCount, replyEventCount, unsubEventCount, subEventCount int
	for i, e := range sc.Events {
		if strings.Contains(e.File, "400-alice-read.event") {
			baseEventsCount = i
			readEventCount = i + 1
		} else if strings.Contains(e.File, "410-bob-staff-reply.event") {
			replyEventCount = i + 1
		} else if strings.Contains(e.File, "500-alice-unsubscribe.event") {
			unsubEventCount = i + 1
		} else if strings.Contains(e.File, "510-alice-subscribe.event") {
			subEventCount = i + 1
		}
	}

	t.Run("Base", func(t *testing.T) {
		runScenarioAndAssert(t, baseEventsCount, func(t *testing.T, client *http.Client, url string, sqlDB *sql.DB, cd *common.CoreData) {
			body := scenarioHTTPGet(t, client, url+"/private/unread")
			countMatches := strings.Count(body, "class=\"thread\"")
			require.Equal(t, 2, countMatches, "Expected 2 unread private threads for Alice initially")

			var aliceID int32
			errAlice := sqlDB.QueryRow("SELECT idusers FROM users WHERE username = 'alice'").Scan(&aliceID)
			require.NoError(t, errAlice)

			aliceCD := coreDataForUser(context.Background(), cd, aliceID)

			var staffRoomTopicID int32
			errStaff := sqlDB.QueryRow("SELECT idforumtopic FROM forumtopic WHERE title = 'Staff Room'").Scan(&staffRoomTopicID)
			require.NoError(t, errStaff)
			var coordinationTopicID int32
			errCoord := sqlDB.QueryRow("SELECT idforumtopic FROM forumtopic WHERE title = 'Coordination'").Scan(&coordinationTopicID)
			require.NoError(t, errCoord)
			var projectRoomTopicID int32
			err = sqlDB.QueryRow("SELECT idforumtopic FROM forumtopic WHERE title = 'Project Room'").Scan(&projectRoomTopicID)
			require.NoError(t, err)

			pattern1 := fmt.Sprintf("create thread:/private/topic/%d/*", staffRoomTopicID)
			hasSub := aliceCD.HasSubscription(pattern1, "internal")
			require.True(t, hasSub, "Alice should be automatically subscribed to Staff Room")

			pattern2 := fmt.Sprintf("create thread:/private/topic/%d/*", coordinationTopicID)
			hasSub = aliceCD.HasSubscription(pattern2, "internal")
			require.True(t, hasSub, "Alice should be automatically subscribed to Coordination")

			pattern3 := fmt.Sprintf("create thread:/private/topic/%d/*", projectRoomTopicID)
			hasSub = aliceCD.HasSubscription(pattern3, "internal")
			require.False(t, hasSub, "Alice should have no Project Room subscription")

			var staffWelcomeThreadID int32
			err = sqlDB.QueryRow("SELECT forumthread_id FROM comments WHERE text LIKE '%Welcome to the staff room%' LIMIT 1").Scan(&staffWelcomeThreadID)
			require.NoError(t, err)

			aliceThreadPattern := fmt.Sprintf("reply:/private/topic/%d/thread/%d/*", staffRoomTopicID, staffWelcomeThreadID)
			hasSub = aliceCD.HasSubscription(aliceThreadPattern, "internal")

			require.True(t, hasSub, "Alice should be automatically subscribed to staff-welcome thread")
		})
	})

	t.Run("Read", func(t *testing.T) {
		runScenarioAndAssert(t, readEventCount, func(t *testing.T, client *http.Client, url string, sqlDB *sql.DB, cd *common.CoreData) {
			body := scenarioHTTPGet(t, client, url+"/private/unread")
			countMatches := strings.Count(body, "class=\"thread\"")
			require.Equal(t, 1, countMatches, "Expected 1 unread private thread for Alice after reading staff-welcome")
			require.NotContains(t, body, "Welcome to the staff room.", "Staff Room unread should be cleared")

			var aliceID int32
			errAlice := sqlDB.QueryRow("SELECT idusers FROM users WHERE username = 'alice'").Scan(&aliceID)
			require.NoError(t, errAlice)

			var staffWelcomeThreadID int32
			err = sqlDB.QueryRow("SELECT forumthread_id FROM comments WHERE text LIKE '%Welcome to the staff room%' LIMIT 1").Scan(&staffWelcomeThreadID)
			require.NoError(t, err)

			aliceCD := coreDataForUser(context.Background(), cd, aliceID)

			marker, err := aliceCD.ThreadReadMarker(staffWelcomeThreadID)
			require.NoError(t, err)

			var bobFirstReplyID int32
			err = sqlDB.QueryRow("SELECT idcomments FROM comments WHERE forumthread_id = ? ORDER BY written DESC LIMIT 1", staffWelcomeThreadID).Scan(&bobFirstReplyID)
			require.NoError(t, err)

			require.Equal(t, bobFirstReplyID, marker, "Alice's marker should equal Bob's first reply exactly")
		})
	})

	t.Run("Reply", func(t *testing.T) {
		runScenarioAndAssert(t, replyEventCount, func(t *testing.T, client *http.Client, url string, sqlDB *sql.DB, cd *common.CoreData) {
			body := scenarioHTTPGet(t, client, url+"/private/unread")
			countMatches := strings.Count(body, "class=\"thread\"")
			require.Equal(t, 2, countMatches, "Expected 2 unread private threads for Alice after bob replied")

			var aliceID, bobID int32
			errAlice := sqlDB.QueryRow("SELECT idusers FROM users WHERE username = 'alice'").Scan(&aliceID)
			require.NoError(t, errAlice)
			errBob := sqlDB.QueryRow("SELECT idusers FROM users WHERE username = 'bob'").Scan(&bobID)
			require.NoError(t, errBob)

			var staffWelcomeThreadID int32
			err = sqlDB.QueryRow("SELECT forumthread_id FROM comments WHERE text LIKE '%Welcome to the staff room%' LIMIT 1").Scan(&staffWelcomeThreadID)
			require.NoError(t, err)

			aliceCD := coreDataForUser(context.Background(), cd, aliceID)
			bobCD := coreDataForUser(context.Background(), cd, bobID)

			aliceMarker, err := aliceCD.ThreadReadMarker(staffWelcomeThreadID)
			require.NoError(t, err)
			bobMarker, err := bobCD.ThreadReadMarker(staffWelcomeThreadID)
			require.NoError(t, err)

			var bobFirstReplyID, newReplyID int32
			// The thread has 3 comments: alice welcome, bob reply 1, bob reply 2
			err = sqlDB.QueryRow("SELECT idcomments FROM comments WHERE forumthread_id = ? ORDER BY written ASC LIMIT 1 OFFSET 1", staffWelcomeThreadID).Scan(&bobFirstReplyID)
			require.NoError(t, err)
			err = sqlDB.QueryRow("SELECT idcomments FROM comments WHERE forumthread_id = ? ORDER BY written DESC LIMIT 1", staffWelcomeThreadID).Scan(&newReplyID)
			require.NoError(t, err)

			require.Equal(t, bobFirstReplyID, aliceMarker, "Alice's marker remains exactly Bob's first reply")
			require.Equal(t, newReplyID, bobMarker, "Bob's marker equals the new reply")

			var staffRoomTopicID int32
			errStaff := sqlDB.QueryRow("SELECT idforumtopic FROM forumtopic WHERE title = 'Staff Room'").Scan(&staffRoomTopicID)
			require.NoError(t, errStaff)

			bobThreadPattern := fmt.Sprintf("reply:/private/topic/%d/thread/%d/*", staffRoomTopicID, staffWelcomeThreadID)
			hasSub := bobCD.HasSubscription(bobThreadPattern, "internal")
			require.True(t, hasSub, "Bob should be automatically subscribed to staff-welcome thread after replying")

			var daveID int32
			errDave := sqlDB.QueryRow("SELECT idusers FROM users WHERE username = 'dave'").Scan(&daveID)
			require.NoError(t, errDave)
			daveCD := coreDataForUser(context.Background(), cd, daveID)
			hasSub = daveCD.HasSubscription(bobThreadPattern, "internal")
			require.False(t, hasSub, "Dave should not be subscribed to staff-welcome thread")
		})
	})

	t.Run("Unsubscribe", func(t *testing.T) {
		runScenarioAndAssert(t, unsubEventCount, func(t *testing.T, client *http.Client, url string, sqlDB *sql.DB, cd *common.CoreData) {
			var aliceID int32
			errAlice := sqlDB.QueryRow("SELECT idusers FROM users WHERE username = 'alice'").Scan(&aliceID)
			require.NoError(t, errAlice)

			aliceCD := coreDataForUser(context.Background(), cd, aliceID)

			var staffRoomTopicID, coordinationTopicID int32
			errStaff := sqlDB.QueryRow("SELECT idforumtopic FROM forumtopic WHERE title = 'Staff Room'").Scan(&staffRoomTopicID)
			require.NoError(t, errStaff)
			errCoord := sqlDB.QueryRow("SELECT idforumtopic FROM forumtopic WHERE title = 'Coordination'").Scan(&coordinationTopicID)
			require.NoError(t, errCoord)

			pattern1 := fmt.Sprintf("create thread:/private/topic/%d/*", staffRoomTopicID)
			hasSub := aliceCD.HasSubscription(pattern1, "internal")
			require.False(t, hasSub, "Staff Room private-topic subscription is absent")

			pattern2 := fmt.Sprintf("create thread:/private/topic/%d/*", coordinationTopicID)
			hasSub = aliceCD.HasSubscription(pattern2, "internal")
			require.True(t, hasSub, "Coordination subscription is unchanged")
		})
	})

	t.Run("Subscribe", func(t *testing.T) {
		runScenarioAndAssert(t, subEventCount, func(t *testing.T, client *http.Client, url string, sqlDB *sql.DB, cd *common.CoreData) {
			var aliceID int32
			errAlice := sqlDB.QueryRow("SELECT idusers FROM users WHERE username = 'alice'").Scan(&aliceID)
			require.NoError(t, errAlice)

			aliceCD := coreDataForUser(context.Background(), cd, aliceID)

			var staffRoomTopicID, coordinationTopicID int32
			errStaff := sqlDB.QueryRow("SELECT idforumtopic FROM forumtopic WHERE title = 'Staff Room'").Scan(&staffRoomTopicID)
			require.NoError(t, errStaff)
			errCoord := sqlDB.QueryRow("SELECT idforumtopic FROM forumtopic WHERE title = 'Coordination'").Scan(&coordinationTopicID)
			require.NoError(t, errCoord)

			pattern1 := fmt.Sprintf("create thread:/private/topic/%d/*", staffRoomTopicID)
			hasSub := aliceCD.HasSubscription(pattern1, "internal")
			require.True(t, hasSub, "Staff Room subscription is restored with the exact private pattern")

			pattern2 := fmt.Sprintf("create thread:/private/topic/%d/*", coordinationTopicID)
			hasSub = aliceCD.HasSubscription(pattern2, "internal")
			require.True(t, hasSub, "Coordination subscription is unchanged")
		})
	})
}
