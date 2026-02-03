package user

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type mockQuerier struct {
	db.QuerierStub
}

func (m *mockQuerier) GetNotificationForLister(ctx context.Context, arg db.GetNotificationForListerParams) (*db.Notification, error) {
	return &db.Notification{
		ID:      arg.ID,
		Message: sql.NullString{String: "Test Notification Message", Valid: true},
		Link:    sql.NullString{String: "", Valid: false},
	}, nil
}

func TestUserNotificationOpenPage_SetsTitle(t *testing.T) {
	q := &mockQuerier{}

	req := httptest.NewRequest("GET", "/usr/notifications/open/123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "123"})

	// Setup session
	store := sessions.NewCookieStore([]byte("secret"))
	core.Store = store
	session, _ := store.New(req, "session")
	session.Values["UID"] = int32(1)
	ctx := context.WithValue(req.Context(), core.ContextValues("session"), session)

	// Setup CoreData
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig())
	cd.Config.NotificationsEnabled = true
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	userNotificationOpenPage(rr, req)

	if cd.PageTitle != "Notification" {
		t.Errorf("Expected PageTitle to be 'Notification', got '%s'", cd.PageTitle)
	}
}
