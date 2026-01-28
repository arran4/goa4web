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
		Idusers:         1,
		Passwd:          sql.NullString{String: "", Valid: true},
		PasswdAlgorithm: sql.NullString{String: "", Valid: true},
		Username:        sql.NullString{String: "u", Valid: true},
	}
	q.SystemListVerifiedEmailsByUserIDReturn = []*db.UserEmail{}
	q.SystemGetUserByEmailRow = &db.SystemGetUserByEmailRow{
		Idusers:  1,
		Email:    "u@test.com",
		Username: sql.NullString{String: "u", Valid: true},
	}
	q.GetLoginRoleForUserReturns = 1
	q.GetPasswordResetByUserErr = sql.ErrNoRows

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
		Idusers:                1,
		Username:               sql.NullString{String: "u", Valid: true},
		PublicProfileEnabledAt: sql.NullTime{},
	}
	q.SystemListVerifiedEmailsByUserIDReturn = []*db.UserEmail{}
	q.AdminInsertRequestQueueReturns = db.FakeSQLResult{LastInsertIDValue: 1, RowsAffectedValue: 1}

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
