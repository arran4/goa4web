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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/app/server"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/scenario"
	"github.com/arran4/goa4web/testdata/scenarios"
	"github.com/arran4/goa4web/workers/postcountworker"
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

func apiRequest(t *testing.T, env *privateForumAPITestEnv, method, path, token string, form url.Values) (*http.Response, map[string]any) {
	t.Helper()
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequest(method, env.http.URL+path, body)
	require.NoError(t, err)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := env.client.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { resp.Body.Close() })
	require.Contains(t, resp.Header.Get("Content-Type"), "application/json")
	decoded := map[string]any{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&decoded))
	return resp, decoded
}

func topicIDs(response map[string]any) []int32 {
	topics, _ := response["topics"].([]any)
	ids := make([]int32, 0, len(topics))
	for _, raw := range topics {
		topic, _ := raw.(map[string]any)
		if id, ok := topic["id"].(float64); ok {
			ids = append(ids, int32(id))
		}
	}
	return ids
}

func requireContainsID(t *testing.T, ids []int32, want int32) {
	t.Helper()
	assert.Contains(t, ids, want)
}

func requireNotContainsID(t *testing.T, ids []int32, unwanted int32) {
	t.Helper()
	assert.NotContains(t, ids, unwanted)
}

func TestE2EPrivateForumAPI(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()

	// Get users and topic IDs from scenario
	aliceID, ok := env.registry.ResolveUser("alice")
	require.True(t, ok)

	bobID, ok := env.registry.ResolveUser("bob")
	require.True(t, ok)

	ctx := context.Background()

	// We need to create an API key for Alice
	apiKey := "test-api-key-alice"
	apiKeyHash := sha256.Sum256([]byte(apiKey))
	_, err := env.cd.Queries().CreateAPIKey(ctx, db.CreateAPIKeyParams{
		UsersIdusers: aliceID,
		ApiKey:       hex.EncodeToString(apiKeyHash[:]),
		Name:         "testkey",
		Scopes:       "private_forum:read,private_forum:write",
		ExpiresAt:    sql.NullTime{Time: time.Now().Add(24 * time.Hour), Valid: true},
	})
	require.NoError(t, err)
	storedAliceKey, err := env.cd.Queries().GetAPIKeyByHash(ctx, hex.EncodeToString(apiKeyHash[:]))
	require.NoError(t, err)
	assert.Equal(t, aliceID, storedAliceKey.UsersIdusers)
	assert.Equal(t, "private_forum:read,private_forum:write", storedAliceKey.Scopes)

	// Create API key for Bob
	bobApiKey := "test-api-key-bob"
	bobApiKeyHash := sha256.Sum256([]byte(bobApiKey))
	_, err = env.cd.Queries().CreateAPIKey(ctx, db.CreateAPIKeyParams{
		UsersIdusers: bobID,
		ApiKey:       hex.EncodeToString(bobApiKeyHash[:]),
		Name:         "bobkey",
		Scopes:       "private_forum:read,private_forum:write",
		ExpiresAt:    sql.NullTime{Time: time.Now().Add(24 * time.Hour), Valid: true},
	})
	require.NoError(t, err)

	// Create scoped API keys for testing layer limits
	aliceReadApiKey := "test-api-key-alice-read"
	aliceReadApiKeyHash := sha256.Sum256([]byte(aliceReadApiKey))
	_, err = env.cd.Queries().CreateAPIKey(ctx, db.CreateAPIKeyParams{
		UsersIdusers: aliceID,
		ApiKey:       hex.EncodeToString(aliceReadApiKeyHash[:]),
		Name:         "alicereadkey",
		Scopes:       "private_forum:read",
		ExpiresAt:    sql.NullTime{Time: time.Now().Add(24 * time.Hour), Valid: true},
	})
	require.NoError(t, err)

	aliceWriteApiKey := "test-api-key-alice-write"
	aliceWriteApiKeyHash := sha256.Sum256([]byte(aliceWriteApiKey))
	_, err = env.cd.Queries().CreateAPIKey(ctx, db.CreateAPIKeyParams{
		UsersIdusers: aliceID,
		ApiKey:       hex.EncodeToString(aliceWriteApiKeyHash[:]),
		Name:         "alicewritekey",
		Scopes:       "private_forum:write",
		ExpiresAt:    sql.NullTime{Time: time.Now().Add(24 * time.Hour), Valid: true},
	})
	require.NoError(t, err)

	daveID, ok := env.registry.ResolveUser("dave")
	require.True(t, ok)

	// Create API key for Dave (Non-Participant in Staff Room, but with valid forum API scopes)
	daveApiKey := "test-api-key-dave"
	daveApiKeyHash := sha256.Sum256([]byte(daveApiKey))
	_, err = env.cd.Queries().CreateAPIKey(ctx, db.CreateAPIKeyParams{
		UsersIdusers: daveID,
		ApiKey:       hex.EncodeToString(daveApiKeyHash[:]),
		Name:         "davekey",
		Scopes:       "private_forum:read,private_forum:write",
		ExpiresAt:    sql.NullTime{Time: time.Now().Add(24 * time.Hour), Valid: true},
	})
	require.NoError(t, err)

	carolID, ok := env.registry.ResolveUser("carol")
	require.True(t, ok)
	staffTopicID := requireScenarioID(t, env.registry, scenario.RefTypeForum, "staff-room")
	projectTopicID := requireScenarioID(t, env.registry, scenario.RefTypeForum, "project-room")
	coordinationTopicID := requireScenarioID(t, env.registry, scenario.RefTypeForum, "coordination")
	staffThreadID := requireScenarioID(t, env.registry, scenario.RefTypeThread, "staff-welcome")
	projectThreadID := requireScenarioID(t, env.registry, scenario.RefTypeThread, "project-kickoff")

	t.Run("scope-and-private-resource-isolation", func(t *testing.T) {
		t.Run("visibility-matrix", func(t *testing.T) {
			for _, tc := range []struct {
				name      string
				token     string
				visible   []int32
				invisible []int32
			}{
				{name: "alice", token: apiKey, visible: []int32{staffTopicID, coordinationTopicID}, invisible: []int32{projectTopicID}},
				{name: "bob", token: bobApiKey, visible: []int32{staffTopicID}, invisible: []int32{projectTopicID, coordinationTopicID}},
				{name: "dave", token: daveApiKey, visible: []int32{projectTopicID}, invisible: []int32{staffTopicID, coordinationTopicID}},
			} {
				t.Run(tc.name, func(t *testing.T) {
					resp, body := apiRequest(t, env, http.MethodGet, "/private/api/topics", tc.token, nil)
					require.Equal(t, http.StatusOK, resp.StatusCode)
					ids := topicIDs(body)
					for _, id := range tc.visible {
						requireContainsID(t, ids, id)
					}
					for _, id := range tc.invisible {
						requireNotContainsID(t, ids, id)
					}
				})
			}

			carolKey := "test-api-key-carol"
			carolHash := sha256.Sum256([]byte(carolKey))
			_, err := env.cd.Queries().CreateAPIKey(ctx, db.CreateAPIKeyParams{
				UsersIdusers: carolID,
				ApiKey:       hex.EncodeToString(carolHash[:]),
				Name:         "carolkey",
				Scopes:       "private_forum:read",
				ExpiresAt:    sql.NullTime{Time: time.Now().Add(24 * time.Hour), Valid: true},
			})
			require.NoError(t, err)
			resp, body := apiRequest(t, env, http.MethodGet, "/private/api/topics", carolKey, nil)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			ids := topicIDs(body)
			requireContainsID(t, ids, projectTopicID)
			requireContainsID(t, ids, coordinationTopicID)
			requireNotContainsID(t, ids, staffTopicID)
		})

		t.Run("scopes-are-independent-of-resource-grants", func(t *testing.T) {
			resp, _ := apiRequest(t, env, http.MethodGet, "/private/api/topic/"+strconv.Itoa(int(staffTopicID))+"/threads", aliceReadApiKey, nil)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			resp, body := apiRequest(t, env, http.MethodPost, "/private/api/topics", aliceReadApiKey, url.Values{"title": {"read-only must not create"}, "participants": {"bob"}})
			require.Equal(t, http.StatusForbidden, resp.StatusCode)
			assert.Equal(t, "error", body["status"])

			resp, body = apiRequest(t, env, http.MethodGet, "/private/api/topics", aliceWriteApiKey, nil)
			require.Equal(t, http.StatusForbidden, resp.StatusCode)
			assert.Equal(t, "error", body["status"])

			resp, _ = apiRequest(t, env, http.MethodGet, "/private/api/topic/"+strconv.Itoa(int(projectTopicID))+"/threads", daveApiKey, nil)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			resp, body = apiRequest(t, env, http.MethodGet, "/private/api/topic/"+strconv.Itoa(int(staffTopicID))+"/threads", daveApiKey, nil)
			require.Equal(t, http.StatusNotFound, resp.StatusCode)
			assert.Equal(t, "error", body["status"])

			resp, _ = apiRequest(t, env, http.MethodGet, "/private/api/topic/"+strconv.Itoa(int(staffTopicID))+"/thread/"+strconv.Itoa(int(staffThreadID)), daveApiKey, nil)
			require.Equal(t, http.StatusNotFound, resp.StatusCode)
			resp, _ = apiRequest(t, env, http.MethodGet, "/private/api/topic/"+strconv.Itoa(int(projectTopicID))+"/thread/"+strconv.Itoa(int(staffThreadID)), daveApiKey, nil)
			require.Equal(t, http.StatusNotFound, resp.StatusCode)

			beforeReplies := int64(0)
			require.NoError(t, env.db.QueryRow("SELECT COUNT(*) FROM comments WHERE forumthread_id = ?", staffThreadID).Scan(&beforeReplies))
			resp, _ = apiRequest(t, env, http.MethodPost, "/private/api/topic/"+strconv.Itoa(int(staffTopicID))+"/thread/"+strconv.Itoa(int(staffThreadID))+"/reply", daveApiKey, url.Values{"text": {"unauthorized dave reply"}})
			require.Equal(t, http.StatusNotFound, resp.StatusCode)
			afterReplies := int64(0)
			require.NoError(t, env.db.QueryRow("SELECT COUNT(*) FROM comments WHERE forumthread_id = ?", staffThreadID).Scan(&afterReplies))
			assert.Equal(t, beforeReplies, afterReplies)

			resp, _ = apiRequest(t, env, http.MethodGet, "/private/api/topic/"+strconv.Itoa(int(projectTopicID))+"/thread/"+strconv.Itoa(int(projectThreadID)), daveApiKey, nil)
			require.Equal(t, http.StatusOK, resp.StatusCode)
		})
	})

	t.Run("create-topic", func(t *testing.T) {
		t.Run("permission-first-username-nonenumeration", func(t *testing.T) {
			var before, after int
			require.NoError(t, env.db.QueryRow("SELECT COUNT(*) FROM forumtopic").Scan(&before))
			for _, username := range []string{"alice", "username-that-does-not-exist-3146"} {
				resp, body := apiRequest(t, env, http.MethodPost, "/private/api/topics", bobApiKey,
					url.Values{"title": {"must not create"}, "participants": {username}})
				require.Equal(t, http.StatusForbidden, resp.StatusCode, "username=%s", username)
				require.Equal(t, "permission denied", body["message"], "username=%s", username)
			}
			require.NoError(t, env.db.QueryRow("SELECT COUNT(*) FROM forumtopic").Scan(&after))
			require.Equal(t, before, after)
		})

		t.Run("api-numeric-id-and-page-boundaries", func(t *testing.T) {
			hugeID := "4294967297"
			staffTopic := strconv.Itoa(int(staffTopicID))
			staffThread := strconv.Itoa(int(staffThreadID))
			for _, tc := range []struct {
				method string
				path string
				form url.Values
			}{
				{http.MethodGet, "/private/api/topic/" + hugeID + "/threads", nil},
				{http.MethodGet, "/private/api/topic/" + hugeID + "/thread/" + staffThread, nil},
				{http.MethodGet, "/private/api/topic/" + staffTopic + "/thread/" + hugeID, nil},
				{http.MethodPost, "/private/api/topic/" + hugeID + "/thread/" + staffThread + "/reply", url.Values{"text": {"must not post"}}},
				{http.MethodPost, "/private/api/topic/" + staffTopic + "/thread/" + hugeID + "/reply", url.Values{"text": {"must not post"}}},
			} {
				resp, body := apiRequest(t, env, tc.method, tc.path, apiKey, tc.form)
				require.Equal(t, http.StatusBadRequest, resp.StatusCode, "request %s %s: %#v", tc.method, tc.path, body)
			}

			// Make sure invalid pages cannot overflow any of the three list slices.
			// Scenario page size is expected to exceed one.
			require.Greater(t, env.cd.PageSize(), 1)
			hugePage := strconv.Itoa(int(^uint(0) >> 1))
			for _, path := range []string{
				"/private/api/topics?page=" + hugePage,
				"/private/api/topic/" + staffTopic + "/threads?page=" + hugePage,
				"/private/api/topic/" + staffTopic + "/thread/" + staffThread + "?page=" + hugePage,
			} {
				resp, body := apiRequest(t, env, http.MethodGet, path, apiKey, nil)
				require.Equal(t, http.StatusBadRequest, resp.StatusCode, "request %s: %#v", path, body)
			}
		})

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

			// We want to test event publication by capturing from the actual event bus.
			sub := env.server.Bus.SubscribeWithOptions(
				eventbus.WithTypes(eventbus.TaskMessageType),
				eventbus.WithBufferSize(8),
			)
			defer sub.Close()

			form := url.Values{}
			form.Add("title", "API Test Topic")
			form.Add("description", "Created via API")
			form.Add("participants", "bob")

			req, _ := http.NewRequest("POST", env.http.URL+"/private/api/topics", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Authorization", "Bearer "+aliceWriteApiKey)

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
			createdTopicID := int32(respBody["topicID"].(float64))

			var participantGrantCount int
			require.NoError(t, env.db.QueryRow(`
				SELECT COUNT(*) FROM grants
				WHERE section = 'privateforum' AND item = 'topic' AND item_id = ?
				  AND user_id IN (?, ?) AND active = 1 AND rule_type = 'allow'
			`, createdTopicID, aliceID, bobID).Scan(&participantGrantCount))
			assert.GreaterOrEqual(t, participantGrantCount, 8, "creator and invitee receive topic grants")
			var subscriptionCount int
			require.NoError(t, env.db.QueryRow(`
				SELECT COUNT(*) FROM subscriptions
				WHERE users_idusers IN (?, ?) AND pattern = ? AND method = 'internal'
			`, aliceID, bobID, fmt.Sprintf("create thread:/private/topic/%d/*", createdTopicID)).Scan(&subscriptionCount))
			assert.Equal(t, 2, subscriptionCount, "creator and invitee subscribe to the created topic")

			expectedTopicPath := "/private/api/topics"

			select {
			case envelope, open := <-sub.Channel():
				require.True(t, open, "subscription closed")
				envelope.Ack()

				evt, ok := envelope.Msg.(eventbus.TaskEvent)
				require.Truef(t, ok, "unexpected event type %T", envelope.Msg)

				require.Equal(t, expectedTopicPath, evt.Path)
				require.Equal(t, aliceID, evt.UserID)
				require.Equal(t, eventbus.TaskOutcomeSuccess, evt.Outcome)
				require.NotNil(t, evt.Task)
				require.NotNil(t, evt.Data["privateforum.topic-create"])

			case <-time.After(2 * time.Second):
				t.Fatal("No task event published")
			}
		})

		t.Run("invalid-create-input-leaves-no-partial-state", func(t *testing.T) {
			outsiderID, err := env.cd.CreateUserWithEmail("private-api-outsider", "private-api-outsider@example.test", "hash", "bcrypt")
			require.NoError(t, err)
			require.NotZero(t, outsiderID)

			assertRejectedCreate := func(t *testing.T, participants string) {
				t.Helper()
				var topicsBefore, grantsBefore, subscriptionsBefore int
				require.NoError(t, env.db.QueryRow("SELECT COUNT(*) FROM forumtopic").Scan(&topicsBefore))
				require.NoError(t, env.db.QueryRow("SELECT COUNT(*) FROM grants").Scan(&grantsBefore))
				require.NoError(t, env.db.QueryRow("SELECT COUNT(*) FROM subscriptions").Scan(&subscriptionsBefore))
				resp, body := apiRequest(t, env, http.MethodPost, "/private/api/topics", apiKey, url.Values{"title": {"must not persist"}, "participants": {participants}})
				require.Equal(t, http.StatusBadRequest, resp.StatusCode)
				assert.Equal(t, "error", body["status"])
				var topicsAfter, grantsAfter, subscriptionsAfter int
				require.NoError(t, env.db.QueryRow("SELECT COUNT(*) FROM forumtopic").Scan(&topicsAfter))
				require.NoError(t, env.db.QueryRow("SELECT COUNT(*) FROM grants").Scan(&grantsAfter))
				require.NoError(t, env.db.QueryRow("SELECT COUNT(*) FROM subscriptions").Scan(&subscriptionsAfter))
				assert.Equal(t, topicsBefore, topicsAfter)
				assert.Equal(t, grantsBefore, grantsAfter)
				assert.Equal(t, subscriptionsBefore, subscriptionsAfter)
			}

			t.Run("nonexistent-user", func(t *testing.T) { assertRejectedCreate(t, "does-not-exist") })
			t.Run("ineligible-existing-user", func(t *testing.T) { assertRejectedCreate(t, "private-api-outsider") })
			t.Run("no-other-participant", func(t *testing.T) { assertRejectedCreate(t, "") })
		})

		t.Run("bob-denied-by-domain-permission", func(t *testing.T) {
			var topicsBefore, grantsBefore, subscriptionsBefore int
			require.NoError(t, env.db.QueryRow("SELECT COUNT(*) FROM forumtopic").Scan(&topicsBefore))
			require.NoError(t, env.db.QueryRow("SELECT COUNT(*) FROM grants").Scan(&grantsBefore))
			require.NoError(t, env.db.QueryRow("SELECT COUNT(*) FROM subscriptions").Scan(&subscriptionsBefore))

			form := url.Values{}
			form.Add("title", "Bob API Test Topic")
			form.Add("description", "Created via API")
			form.Add("participants", "alice")

			req, _ := http.NewRequest("POST", env.http.URL+"/private/api/topics", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Authorization", "Bearer "+bobApiKey) // Bob lacks domain grant to create topics in this setup

			resp, err := env.client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Bob should be denied creating topics by domain permission")
			assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")

			var topicsAfter, grantsAfter, subscriptionsAfter int
			require.NoError(t, env.db.QueryRow("SELECT COUNT(*) FROM forumtopic").Scan(&topicsAfter))
			require.NoError(t, env.db.QueryRow("SELECT COUNT(*) FROM grants").Scan(&grantsAfter))
			require.NoError(t, env.db.QueryRow("SELECT COUNT(*) FROM subscriptions").Scan(&subscriptionsAfter))
			assert.Equal(t, topicsBefore, topicsAfter)
			assert.Equal(t, grantsBefore, grantsAfter)
			assert.Equal(t, subscriptionsBefore, subscriptionsAfter)
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

	t.Run("list-threads-pagination-and-private-boundaries", func(t *testing.T) {
		pageSize := env.cd.PageSize()
		for i := 0; i < pageSize; i++ {
			_, err := env.cd.ForUser(aliceID).CreateForumThread(ctx, common.CreateForumThreadParams{
				ActorID:        aliceID,
				TopicID:        staffTopicID,
				LanguageID:     1,
				Text:           fmt.Sprintf("pagination fixture thread %d", i),
				Private:        true,
				EnforceHandler: true,
				BasePath:       "/private",
			})
			require.NoError(t, err)
		}

		threadIDs := func(t *testing.T, body map[string]any) []int32 {
			t.Helper()
			rows, ok := body["threads"].([]any)
			require.True(t, ok)
			ids := make([]int32, 0, len(rows))
			for _, raw := range rows {
				row, ok := raw.(map[string]any)
				require.True(t, ok)
				id, ok := row["Idforumthread"].(float64)
				require.True(t, ok, "thread response should expose its ID")
				ids = append(ids, int32(id))
			}
			return ids
		}

		resp, firstPage := apiRequest(t, env, http.MethodGet, fmt.Sprintf("/private/api/topic/%d/threads?page=1", staffTopicID), apiKey, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, true, firstPage["has_more"])
		firstIDs := threadIDs(t, firstPage)
		assert.Len(t, firstIDs, pageSize)

		resp, secondPage := apiRequest(t, env, http.MethodGet, fmt.Sprintf("/private/api/topic/%d/threads?page=2", staffTopicID), apiKey, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, false, secondPage["has_more"])
		secondIDs := threadIDs(t, secondPage)
		assert.NotEmpty(t, secondIDs)
		for _, id := range secondIDs {
			assert.NotContains(t, firstIDs, id)
		}

		resp, outOfRange := apiRequest(t, env, http.MethodGet, fmt.Sprintf("/private/api/topic/%d/threads?page=99", staffTopicID), apiKey, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		threads, ok := outOfRange["threads"].([]any)
		require.True(t, ok, "out-of-range page must encode threads as [] rather than null")
		assert.Empty(t, threads)

		resp, _ = apiRequest(t, env, http.MethodGet, fmt.Sprintf("/private/api/topic/%d/threads", projectTopicID), apiKey, nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		resp, _ = apiRequest(t, env, http.MethodGet, "/private/api/topic/999999/threads", apiKey, nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
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

		sub := env.server.Bus.SubscribeWithOptions(
			eventbus.WithTypes(eventbus.TaskMessageType),
			eventbus.WithBufferSize(8),
		)
		defer sub.Close()

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

		commentID := int32(respBody["comment_id"].(float64))
		expectedReplyPath := fmt.Sprintf("/private/api/topic/%d/thread/%d/reply", topicID, threadID)

		select {
		case envelope, open := <-sub.Channel():
			require.True(t, open, "subscription closed")
			envelope.Ack()

			evt, ok := envelope.Msg.(eventbus.TaskEvent)
			require.Truef(t, ok, "unexpected event type %T", envelope.Msg)

			require.Equal(t, expectedReplyPath, evt.Path)
			require.Equal(t, aliceID, evt.UserID)
			require.Equal(t, eventbus.TaskOutcomeSuccess, evt.Outcome)
			require.NotNil(t, evt.Task)

			post, ok := evt.Data[postcountworker.EventKey].(postcountworker.UpdateEventData)
			require.Truef(t, ok, "missing/wrong postcount event payload: %#v", evt.Data)
			require.Equal(t, threadID, post.ThreadID)
			require.Equal(t, topicID, post.TopicID)
			require.Equal(t, commentID, post.CommentID)

		case <-time.After(2 * time.Second):
			t.Fatal("no reply event: check HTTP status, event task attachment, and bus identity")
		}

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
