//go:build sqlite || sqlite3

package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"net/url"
	"testing"

	"github.com/arran4/goa4web/internal/db"
	"github.com/stretchr/testify/require"
)

func TestPrivateForumAPIKeyStopsAuthorizingWhenOwnerIsDeactivated(t *testing.T) {
	env, cleanup := setupAPITestEnv(t)
	defer cleanup()
	ctx := context.Background()
	aliceID, ok := env.registry.ResolveUser("alice")
	require.True(t, ok)
	const bearer = "deactivation-scenario-bearer" // Credential local to this test.
	hash := sha256.Sum256([]byte(bearer))
	_, err := env.cd.Queries().CreateAPIKey(ctx, db.CreateAPIKeyParams{
		UsersIdusers: aliceID,
		ApiKey:       hex.EncodeToString(hash[:]),
		Name:         "deactivation",
		Scopes:       "private_forum:read,private_forum:write",
	})
	require.NoError(t, err)
	resp, _ := apiRequest(t, env, http.MethodGet, "/private/api/topics", bearer, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// The deactivation command scrubs the account without deleting its grants or API keys.
	require.NoError(t, env.cd.Queries().AdminScrubUser(ctx, db.AdminScrubUserParams{
		Idusers: aliceID, Username: sql.NullString{String: "deactivated-alice", Valid: true},
	}))
	resp, body := apiRequest(t, env, http.MethodGet, "/private/api/topics", bearer, nil)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	require.Equal(t, "unauthorized", body["message"])
	require.NotContains(t, body, "topics")
	resp, body = apiRequest(t, env, http.MethodPost, "/private/api/topics", bearer, url.Values{"title": {"must not create"}})
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	require.Equal(t, "unauthorized", body["message"])
	keys, err := env.cd.Queries().ListAPIKeysByUser(ctx, aliceID)
	require.NoError(t, err)
	require.Len(t, keys, 1)
	require.False(t, keys[0].RevokedAt.Valid, "deactivation preserves the existing key record")
}
