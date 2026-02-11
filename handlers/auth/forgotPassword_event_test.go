package auth

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

func TestHappyPathForgotPasswordEventData(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.SystemGetLoginRow = &db.SystemGetLoginRow{
		Idusers:         1,
		Passwd:          sql.NullString{String: "", Valid: true},
		PasswdAlgorithm: sql.NullString{String: "", Valid: true},
		Username:        sql.NullString{String: "u", Valid: true},
	}
	q.SystemListVerifiedEmailsByUserIDReturn = []*db.UserEmail{{
		ID:                   1,
		UserID:               1,
		Email:                "a@test.com",
		VerifiedAt:           sql.NullTime{Time: time.Now(), Valid: true},
		NotificationPriority: 0,
	}}
	q.SystemGetUserByEmailRow = &db.SystemGetUserByEmailRow{
		Idusers:  1,
		Email:    "a@test.com",
		Username: sql.NullString{String: "u", Valid: true},
	}
	q.GetLoginRoleForUserReturns = 1
	q.GetPasswordResetByUserErr = sql.ErrNoRows

	evt := &eventbus.TaskEvent{Data: map[string]any{}}
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
	if _, ok := evt.Data["Username"]; !ok {
		t.Fatalf("missing Username data")
	}
	if _, ok := evt.Data["Code"]; !ok {
		t.Fatalf("missing Code data")
	}
	if _, ok := evt.Data["ResetURL"]; !ok {
		t.Fatalf("missing ResetURL data")
	}
	if _, ok := evt.Data["UserURL"]; !ok {
		t.Fatalf("missing UserURL data")
	}
}
