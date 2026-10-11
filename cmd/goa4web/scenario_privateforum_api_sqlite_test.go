//go:build sqlite || sqlite3
package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/app/server"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/scenario"
	"github.com/arran4/goa4web/testdata/scenarios"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type privateForumAPITestEnv struct {
	server   *server.Server
	db       *sql.DB
	http     *httptest.Server
	client   *http.Client
	registry *scenario.RefRegistry
	cd       *common.CoreData
}

func requireScenarioID(
	t *testing.T,
	registry *scenario.RefRegistry,
	typ scenario.RefType,
	ref string,
) int32 {
	t.Helper()

	value, ok := registry.Resolve(typ, ref)
	require.True(t, ok, "scenario %s ref %q should resolve", typ, ref)

	switch v := value.(type) {
	case int32:
		return v
	case int:
		return int32(v)
	case int64:
		return int32(v)
	default:
		t.Fatalf("scenario %s ref %q has unexpected type %T", typ, ref, value)
		return 0
	}
}

func setupAPITestEnv(t *testing.T) (*privateForumAPITestEnv, func()) {
	ctx := context.Background()

	root, err := parseRoot([]string{
		"goa4web",
		"scenario",
		"serve",
		"100-private-forum",
	})
	require.NoError(t, err)

	parent, err := parseScenarioCmd(root, []string{
		"serve",
		"100-private-forum",
	})
	require.NoError(t, err)

	serveCmd, err := parseScenarioServeCmd(parent, []string{
		"100-private-forum",
	})
	require.NoError(t, err)

	serveCmd.fsys = scenarios.FS

	srv, sqlDB, bootstrapCleanup, err := serveCmd.Bootstrap(ctx)
	require.NoError(t, err)

	require.Same(t, sqlDB, srv.DB)
	require.True(t, srv.Config.CSRFEnabled)

	httpServer := httptest.NewServer(srv.Router)

	client := httpServer.Client()

	env := &privateForumAPITestEnv{
		server:   srv,
		db:       sqlDB,
		http:     httpServer,
		client:   client,
		registry: serveCmd.Registry,
		cd:       serveCmd.CoreData,
	}

	cleanup := func() {
		httpServer.Close()
		bootstrapCleanup()
		root.Close()
	}

	return env, cleanup
}

func TestE2EPrivateForumAPI(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()

	// Get users and topic IDs from scenario
	aliceID, ok := env.registry.ResolveUser("alice")
	require.True(t, ok)

	bobID, ok := env.registry.ResolveUser("bob")
	require.True(t, ok)

	// We need to create an API key for Alice
	apiKey := "test-api-key-alice"
	apiKeyHash := sha256.Sum256([]byte(apiKey))
	_, err := env.db.Exec(`
		INSERT INTO api_keys (users_idusers, api_key, name, scopes, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		aliceID, hex.EncodeToString(apiKeyHash[:]), "testkey", "private_forum:read,private_forum:write", time.Now().Add(24*time.Hour), time.Now(),
	)
	require.NoError(t, err)

	// Create API key for Bob
	bobApiKey := "test-api-key-bob"
	bobApiKeyHash := sha256.Sum256([]byte(bobApiKey))
	_, err = env.db.Exec(`
		INSERT INTO api_keys (users_idusers, api_key, name, scopes, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		bobID, hex.EncodeToString(bobApiKeyHash[:]), "testkeybob", "private_forum:read,private_forum:write", time.Now().Add(24*time.Hour), time.Now(),
	)
	require.NoError(t, err)

	t.Run("create-topic", func(t *testing.T) {
		t.Run("alice-api-key-persisted-create-grant", func(t *testing.T) {
			grantRow, err := env.cd.Queries().SystemCheckGrant(context.Background(), db.SystemCheckGrantParams{
				ViewerID:               aliceID,
				Section:                "privateforum",
				Item:                   sql.NullString{String: "topic", Valid: true},
				Action:                 "create",
				ItemID:                 sql.NullInt32{Valid: false},
				IsSpecificPrivateForum: false,
				UserID:                 sql.NullInt32{Int32: aliceID, Valid: true},
			})
			require.NoError(t, err)
			assert.Equal(t, int64(1), int64(grantRow))

			form := url.Values{}
			form.Add("title", "API Test Topic")
			form.Add("description", "Created via API")
			form.Add("participants", "bob")

			req, _ := http.NewRequest("POST", env.http.URL+"/private/api/topics", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Authorization", "Bearer "+apiKey)

			resp, err := env.client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			if !assert.Equal(t, http.StatusOK, resp.StatusCode) {
				bodyBytes, _ := io.ReadAll(resp.Body)
				t.Logf("Response body: %s", string(bodyBytes))
				t.FailNow()
			}

			var respBody map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			require.NoError(t, err)
			assert.Equal(t, "success", respBody["status"])
			assert.NotNil(t, respBody["topicID"])
		})
	})

	t.Run("list-topics", func(t *testing.T) {
		req, _ := http.NewRequest("GET", env.http.URL+"/private/api/topics", nil)
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := env.client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		if !assert.Equal(t, http.StatusOK, resp.StatusCode) {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Logf("Response body: %s", string(bodyBytes))
			t.FailNow()
		}
		var respBody map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		require.NoError(t, err)

		topics, ok := respBody["topics"].([]interface{})
		assert.True(t, ok)
		assert.GreaterOrEqual(t, len(topics), 1)
	})

	t.Run("reply-thread", func(t *testing.T) {
		// Use symbolic references from scenario where practical
		topicID := requireScenarioID(t, env.registry, scenario.RefTypeForum, "staff-room")
		threadID := requireScenarioID(t, env.registry, scenario.RefTypeThread, "staff-welcome")

		// Assert relationship and grants on same DB
		var forumtopic_idforumtopic int32
		err = env.db.QueryRow("SELECT forumtopic_idforumtopic FROM forumthread WHERE idforumthread = ?", threadID).Scan(&forumtopic_idforumtopic)
		require.NoError(t, err)
		assert.Equal(t, topicID, forumtopic_idforumtopic)

		var grantCount int
		err = env.db.QueryRow(`
			SELECT COUNT(*)
			FROM grants
			WHERE section = 'privateforum_thread'
			  AND item = 'thread'
			  AND action = 'reply'
			  AND item_id = ?
			  AND user_id = ?
			  AND active = 1
			  AND rule_type = 'allow'
		`, threadID, aliceID).Scan(&grantCount)
		require.NoError(t, err)
		require.GreaterOrEqual(t, grantCount, 1, "Alice must have reply grant on the thread")

		// Direct CanReplyForumThread check using exact parameters
		allowed, err := common.CanReplyForumThread(
			context.Background(),
			env.cd.Queries(),
			consts.PermissionSectionPrivateForumThread,
			consts.PermissionItemThread,
			int32(threadID),
			int32(threadID),
			aliceID,
		)
		require.NoError(t, err)
		require.True(t, allowed, "Direct CanReplyForumThread must be true")

		// Post reply as participant (Alice)
		form := url.Values{}
		form.Add("text", "This is an API reply")
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/private/api/topic/%d/thread/%d/reply", env.http.URL, topicID, threadID), strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := env.client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		if !assert.Equal(t, http.StatusOK, resp.StatusCode) {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Logf("Response body: %s", string(bodyBytes))
			t.FailNow()
		}

		var respBody map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		require.NoError(t, err)
		assert.Equal(t, "success", respBody["status"])
		assert.Equal(t, float64(threadID), respBody["thread_id"])
		assert.Equal(t, float64(topicID), respBody["topic_id"])
		assert.NotNil(t, respBody["comment_id"])

		// Another participant (Bob) can read the thread and see the reply
		req, _ = http.NewRequest("GET", fmt.Sprintf("%s/private/api/topic/%d/thread/%d", env.http.URL, topicID, threadID), nil)
		req.Header.Set("Authorization", "Bearer "+bobApiKey)

		resp, err = env.client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		require.NoError(t, err)

		comments := respBody["comments"].([]interface{})
		assert.GreaterOrEqual(t, len(comments), 2, "Bob should see the initial post and Alice's reply")

		foundReply := false
		for _, c := range comments {
			comment := c.(map[string]interface{})
			if comment["text"] == "This is an API reply" && comment["username"] == "alice" {
				foundReply = true
				break
			}
		}
		assert.True(t, foundReply, "Alice's reply should be visible to Bob")

		// Unauthorized user (non-participant without API key) cannot reply
		form = url.Values{}
		form.Add("text", "This is an unauthorized API reply")
		req, _ = http.NewRequest("POST", fmt.Sprintf("%s/private/api/topic/%d/thread/%d/reply", env.http.URL, topicID, threadID), strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err = env.client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Unauthorized reply should be forbidden")

		// Topic/thread mismatch is rejected
		otherTopicID := requireScenarioID(t, env.registry, scenario.RefTypeForum, "project-room")
		form = url.Values{}
		form.Add("text", "This is an API reply to the wrong topic")
		req, _ = http.NewRequest("POST", fmt.Sprintf("%s/private/api/topic/%d/thread/%d/reply", env.http.URL, otherTopicID, threadID), strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err = env.client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Mismatched topic and thread ID should return 404")

		// Malformed thread ID is rejected
		req, _ = http.NewRequest("POST", fmt.Sprintf("%s/private/api/topic/%d/thread/malformed/reply", env.http.URL, topicID), strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err = env.client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Malformed thread ID should return 400")
	})

	t.Run("read-thread", func(t *testing.T) {
		topicID := requireScenarioID(t, env.registry, scenario.RefTypeForum, "staff-room")
		threadID := requireScenarioID(t, env.registry, scenario.RefTypeThread, "staff-welcome")

		// Authorized participant can see comments
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/private/api/topic/%d/thread/%d", env.http.URL, topicID, threadID), nil)
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := env.client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var respBody map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		require.NoError(t, err)

		comments := respBody["comments"].([]interface{})
		assert.GreaterOrEqual(t, len(comments), 1, "Authorized participant should see comments")

		// Unrelated topic ID mapping should be rejected
		otherTopicID := requireScenarioID(t, env.registry, scenario.RefTypeForum, "project-room")
		req, _ = http.NewRequest("GET", fmt.Sprintf("%s/private/api/topic/%d/thread/%d", env.http.URL, otherTopicID, threadID), nil)
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err = env.client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Note: we can choose to assert 404 or 403 based on API requirements for mismatched topics.
		// For the post request we currently return 404. Since we have not yet fully refactored `APIShowComments`
		// we just check it is not successful if it currently returns 500 or 404. Let's write the test based on what we'll fix next.
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Mismatched topic and thread ID should return 404")

		// Unauthorized user cannot read thread comments
		req, _ = http.NewRequest("GET", fmt.Sprintf("%s/private/api/topic/%d/thread/%d", env.http.URL, topicID, threadID), nil)
		// No API key
		resp, err = env.client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// The API key middleware currently returns 401 Unauthorized if no API key is provided, which is correct
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Unauthorized read should return 401")
	})
}
