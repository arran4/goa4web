package apiauth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/stretchr/testify/assert"
)

type mockQuerier struct {
	db.Querier
}

func (m mockQuerier) GetAPIKeyByHash(ctx context.Context, apiKey string) (*db.ApiKey, error) {
	return &db.ApiKey{
		ID:           1,
		UsersIdusers: 42,
		Scopes:       "read,write",
	}, nil
}

func (m mockQuerier) UpdateAPIKeyLastUsed(ctx context.Context, id int32) error {
	return nil
}

func TestAPIKeyAuthMiddlewareEventPreservation(t *testing.T) {
	q := mockQuerier{}
	cd := common.NewCoreData(context.Background(), q, &config.RuntimeConfig{})

	cd.UserID = 0 // anonymous

	// Pre-existing event created by TaskEventMiddleware for the anonymous user
	initialEvent := &eventbus.TaskEvent{
		UserID: 0,
	}
	cd.SetEvent(initialEvent)

	req := httptest.NewRequest("POST", "/api/action", nil)

	// Create a dummy token. GetAPIKeyByHash will always return success above.
	token := "valid_token_123"
	req.Header.Set("Authorization", "Bearer "+token)

	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	var nextCD *common.CoreData
	var nextScopes map[string]bool

	handler := APIKeyAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCD = r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		nextScopes = r.Context().Value(consts.KeyAPIScopes).(map[string]bool)
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.NotNil(t, nextCD)
	assert.Equal(t, int32(42), nextCD.UserID)

	// The event should be the EXACT same pointer
	assert.Same(t, initialEvent, nextCD.Event())

	// And the event's UserID should be correctly updated to the authenticated user's ID
	assert.Equal(t, int32(42), nextCD.Event().UserID)

	// Validate scopes were successfully parsed
	assert.True(t, nextScopes["read"])
	assert.True(t, nextScopes["write"])
}
