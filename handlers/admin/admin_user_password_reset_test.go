package admin

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestUserPasswordResetClearsPendingResets(t *testing.T) {
	qs := testhelpers.NewQuerierStub()
	qs.SystemGetUserByIDRow = &db.SystemGetUserByIDRow{
		Idusers:  42,
		Username: sql.NullString{String: "target", Valid: true},
	}
	qs.SystemDeletePasswordResetsByUserResult = db.FakeSQLResult{RowsAffectedValue: 1}

	req := httptest.NewRequest(http.MethodPost, "/admin/user/42/reset", nil)
	req = mux.SetURLVars(req, map[string]string{"user": "42"})

	cd := common.NewCoreData(req.Context(), qs, config.NewRuntimeConfig())
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rr := httptest.NewRecorder()
	userForcePasswordChangeTask.Action(rr, req)

	if len(qs.InsertPasswordCalls) != 1 {
		t.Fatalf("expected InsertPassword to be called once, got %d", len(qs.InsertPasswordCalls))
	}
	if len(qs.SystemDeletePasswordResetsByUserCalls) != 1 {
		t.Fatalf("expected SystemDeletePasswordResetsByUser to be called once, got %d", len(qs.SystemDeletePasswordResetsByUserCalls))
	}
	if qs.SystemDeletePasswordResetsByUserCalls[0] != 42 {
		t.Errorf("expected SystemDeletePasswordResetsByUser to use user ID 42, got %d", qs.SystemDeletePasswordResetsByUserCalls[0])
	}
}

func TestUserPasswordResetStopsOnPendingResetCleanupError(t *testing.T) {
	qs := testhelpers.NewQuerierStub()
	qs.SystemGetUserByIDRow = &db.SystemGetUserByIDRow{
		Idusers:  42,
		Username: sql.NullString{String: "target", Valid: true},
	}
	qs.SystemDeletePasswordResetsByUserErr = errors.New("cleanup failed")

	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	req := httptest.NewRequest(http.MethodPost, "/admin/user/42/reset", nil)
	req = mux.SetURLVars(req, map[string]string{"user": "42"})

	cd := common.NewCoreData(req.Context(), qs, config.NewRuntimeConfig(), common.WithEvent(evt))
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rr := httptest.NewRecorder()
	userForcePasswordChangeTask.Action(rr, req)

	if len(qs.InsertPasswordCalls) != 1 {
		t.Fatalf("expected InsertPassword to be called once, got %d", len(qs.InsertPasswordCalls))
	}
	if _, ok := evt.Data["targetUserID"]; ok {
		t.Errorf("expected no event data on cleanup failure")
	}
}

func TestUserSendResetEmail(t *testing.T) {
	qs := testhelpers.NewQuerierStub()
	qs.SystemGetUserByIDRow = &db.SystemGetUserByIDRow{
		Idusers:  42,
		Username: sql.NullString{String: "target", Valid: true},
	}

	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	req := httptest.NewRequest(http.MethodPost, "/admin/user/42/reset", nil)
	req = mux.SetURLVars(req, map[string]string{"user": "42"})

	// Mock DB calls for token generation
	qs.GetPasswordResetByUserFn = func(ctx context.Context, arg db.GetPasswordResetByUserParams) (*db.PendingPassword, error) {
		return nil, sql.ErrNoRows // No existing reset
	}
	qs.CreatePasswordResetForUserFn = func(ctx context.Context, arg db.CreatePasswordResetForUserParams) error {
		return nil
	}
	qs.AdminInsertRequestQueueFn = func(ctx context.Context, arg db.AdminInsertRequestQueueParams) (sql.Result, error) {
		return db.FakeSQLResult{}, nil
	}

	// Needed for signing link
	cd := common.NewCoreData(req.Context(), qs, config.NewRuntimeConfig(), common.WithEvent(evt), common.WithLinkSignKey("secret"))
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rr := httptest.NewRecorder()
	userSendResetEmailTask.Action(rr, req)

	if len(qs.InsertPasswordCalls) != 0 {
		t.Fatalf("expected InsertPassword NOT to be called, got %d", len(qs.InsertPasswordCalls))
	}
	if _, ok := evt.Data["ResetURL"]; !ok {
		t.Errorf("expected ResetURL in event data")
	}
	if username, ok := evt.Data["Username"]; !ok || username != "target" {
		t.Errorf("expected Username 'target' in event data, got %v", username)
	}
}
