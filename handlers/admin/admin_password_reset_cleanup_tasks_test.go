package admin

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestClearExpiredPasswordResetsTask(t *testing.T) {
	qs := testhelpers.NewQuerierStub()
	qs.SystemPurgePasswordResetsBeforeResult = db.FakeSQLResult{RowsAffectedValue: 3}

	form := url.Values{}
	form.Set("hours", "24")
	form.Set("back", "/admin/password_resets?status=pending")
	req := httptest.NewRequest(http.MethodPost, "/admin/password_resets", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	cd := common.NewCoreData(req.Context(), qs, config.NewRuntimeConfig(), common.WithEvent(evt))
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	start := time.Now()
	result := clearExpiredPasswordResetsTask.Action(nil, req)
	end := time.Now()

	if len(qs.SystemPurgePasswordResetsBeforeCalls) != 1 {
		t.Fatalf("expected purge to be called once, got %d", len(qs.SystemPurgePasswordResetsBeforeCalls))
	}
	called := qs.SystemPurgePasswordResetsBeforeCalls[0]
	expectedLower := start.Add(-24*time.Hour - time.Second)
	expectedUpper := end.Add(-24*time.Hour + time.Second)
	if called.Before(expectedLower) || called.After(expectedUpper) {
		t.Errorf("expected expiry between %s and %s, got %s", expectedLower, expectedUpper, called)
	}

	if evt.Data["DeletedCount"] != int64(3) {
		t.Errorf("expected DeletedCount to be 3, got %v", evt.Data["DeletedCount"])
	}
	if evt.Data["Hours"] != 24 {
		t.Errorf("expected Hours to be 24, got %v", evt.Data["Hours"])
	}

	redirect, ok := result.(handlers.RedirectHandler)
	if !ok {
		t.Fatalf("expected redirect result, got %T", result)
	}
	if !strings.Contains(string(redirect), "cleanup_action=expired") {
		t.Errorf("expected cleanup_action in redirect, got %s", redirect)
	}
}

func TestClearUserPasswordResetsTask(t *testing.T) {
	qs := testhelpers.NewQuerierStub()
	qs.SystemGetUserByUsernameRow = &db.SystemGetUserByUsernameRow{
		Idusers:  7,
		Username: sql.NullString{String: "target", Valid: true},
	}
	qs.SystemDeletePasswordResetsByUserResult = db.FakeSQLResult{RowsAffectedValue: 2}

	form := url.Values{}
	form.Set("user", "target")
	form.Set("back", "/admin/password_resets?status=pending")
	req := httptest.NewRequest(http.MethodPost, "/admin/password_resets", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	evt := &eventbus.TaskEvent{Data: map[string]any{}}
	cd := common.NewCoreData(req.Context(), qs, config.NewRuntimeConfig(), common.WithEvent(evt))
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	result := clearUserPasswordResetsTask.Action(nil, req)

	if len(qs.SystemDeletePasswordResetsByUserCalls) != 1 {
		t.Fatalf("expected delete resets to be called once, got %d", len(qs.SystemDeletePasswordResetsByUserCalls))
	}
	if qs.SystemDeletePasswordResetsByUserCalls[0] != 7 {
		t.Errorf("expected delete resets for user 7, got %d", qs.SystemDeletePasswordResetsByUserCalls[0])
	}
	if evt.Data["DeletedCount"] != int64(2) {
		t.Errorf("expected DeletedCount to be 2, got %v", evt.Data["DeletedCount"])
	}
	if evt.Data["Username"] != "target" {
		t.Errorf("expected Username to be target, got %v", evt.Data["Username"])
	}

	redirect, ok := result.(handlers.RedirectHandler)
	if !ok {
		t.Fatalf("expected redirect result, got %T", result)
	}
	if !strings.Contains(string(redirect), "cleanup_action=user") {
		t.Errorf("expected cleanup_action in redirect, got %s", redirect)
	}
}
