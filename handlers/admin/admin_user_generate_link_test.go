package admin

import (
	"context"
	"database/sql"
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

func TestHappyPathUserGenerateResetLink(t *testing.T) {
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
		return nil, sql.ErrNoRows
	}
	qs.CreatePasswordResetForUserFn = func(ctx context.Context, arg db.CreatePasswordResetForUserParams) error {
		if arg.PasswdAlgorithm != "magic" {
			t.Errorf("expected algorithm 'magic', got %s", arg.PasswdAlgorithm)
		}
		if arg.Passwd != "magic-link" {
			t.Errorf("expected password 'magic-link', got %s", arg.Passwd)
		}
		return nil
	}
	qs.AdminInsertRequestQueueFn = func(ctx context.Context, arg db.AdminInsertRequestQueueParams) (sql.Result, error) {
		return db.FakeSQLResult{}, nil
	}

	cd := common.NewCoreData(req.Context(), qs, config.NewRuntimeConfig(), common.WithEvent(evt), common.WithLinkSignKey("secret"))
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rr := httptest.NewRecorder()
	userGenerateResetLinkTask.Action(rr, req)

	if len(qs.CreatePasswordResetForUserCalls) != 1 {
		t.Fatalf("expected CreatePasswordResetForUser to be called once")
	}
	if _, ok := evt.Data["ResetURL"]; !ok {
		t.Errorf("expected ResetURL in event data")
	}

	// Check that the link is in the response messages (implied by template data)
	// We can't easily check template data as it's passed to handler.
	// But we can check if it didn't error.
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}
