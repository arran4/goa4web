package auth

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestForgotPasswordNoEmail(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.SystemGetLoginRow = &db.SystemGetLoginRow{
		Idusers:  1,
		Username: sql.NullString{String: "u", Valid: true},
	}
	q.SystemListVerifiedEmailsByUserIDReturn = []*db.UserEmail{}
	q.GetLoginRoleForUserReturns = 1
	q.GetPasswordResetByUserErr = sql.ErrNoRows
	q.CreatePasswordResetForUserFn = func(ctx context.Context, arg db.CreatePasswordResetForUserParams) error {
		return nil
	}

	evt := &eventbus.TaskEvent{}
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig(), common.WithEvent(evt))
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)

	form := url.Values{"username": {"u"}, "password": {"pw"}}
	req := httptest.NewRequest("POST", "/forgot", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handlers.TaskHandler(forgotPasswordTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestEmailAssociationRequestTask(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.SystemGetUserByUsernameRow = &db.SystemGetUserByUsernameRow{
		Idusers:  1,
		Username: sql.NullString{String: "u", Valid: true},
	}
	q.SystemListVerifiedEmailsByUserIDReturn = []*db.UserEmail{}
	q.AdminInsertRequestQueueFn = func(ctx context.Context, arg db.AdminInsertRequestQueueParams) (sql.Result, error) {
		return db.FakeSQLResult{}, nil
	}
	q.AdminInsertRequestCommentFn = func(ctx context.Context, arg db.AdminInsertRequestCommentParams) error {
		return nil
	}

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(context.Background(), consts.KeyCoreData, cd)

	form := url.Values{"username": {"u"}, "email": {"a@test.com"}, "reason": {"help"}}
	req := httptest.NewRequest("POST", "/forgot", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handlers.TaskHandler(emailAssociationRequestTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}
