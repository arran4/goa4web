package user

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestUserEmailVerifyCodePage_Invalid(t *testing.T) {
	t.Run("Unhappy Path", func(t *testing.T) {
		code := "abc"
		queries := testhelpers.NewQuerierStub()
		queries.GetUserEmailByCodeFn = func(ctx context.Context, c sql.NullString) (*db.UserEmail, error) {
			if c.String != code {
				t.Errorf("unexpected code: %s", c.String)
			}
			return &db.UserEmail{ID: 1, UserID: 1, Email: "e@example.com", LastVerificationCode: sql.NullString{String: code, Valid: true}}, nil
		}

		store := sessions.NewCookieStore([]byte("test"))
		sess := sessions.NewSession(store, "test")
		sess.Values = map[interface{}]interface{}{"UID": int32(2)} // Different User ID
		core.Store = store
		core.SessionName = "test"

		ctx := context.Background()
		ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
		cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithSession(sess))
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

		req := httptest.NewRequest(http.MethodGet, "/usr/email/verify?code="+code, nil).WithContext(ctx)
		rr := httptest.NewRecorder()
		userEmailVerifyCodePage(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("status=%d", rr.Code)
		}
	})
}

func TestUserEmailVerifyCodePage_Success(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		code := "xyz"
		var marked []db.SystemMarkUserEmailVerifiedParams
		var prioritySetArgs []db.SetNotificationPriorityForListerParams
		var deletedArgs []db.SystemDeleteUserEmailsByEmailExceptIDParams

		queries := testhelpers.NewQuerierStub()
		queries.GetUserEmailByCodeFn = func(ctx context.Context, c sql.NullString) (*db.UserEmail, error) {
			if c.String != code {
				t.Errorf("unexpected code: %s", c.String)
			}
			return &db.UserEmail{ID: 1, UserID: 1, Email: "e@example.com", LastVerificationCode: sql.NullString{String: code, Valid: true}, VerificationExpiresAt: sql.NullTime{Time: time.Now().Add(time.Hour), Valid: true}}, nil
		}
		queries.SystemMarkUserEmailVerifiedFn = func(ctx context.Context, arg db.SystemMarkUserEmailVerifiedParams) error {
			marked = append(marked, arg)
			return nil
		}
		queries.GetMaxNotificationPriorityFn = func(ctx context.Context, userID int32) (interface{}, error) {
			return int64(0), nil
		}
		queries.SetNotificationPriorityForListerFn = func(ctx context.Context, arg db.SetNotificationPriorityForListerParams) error {
			prioritySetArgs = append(prioritySetArgs, arg)
			return nil
		}
		queries.SystemDeleteUserEmailsByEmailExceptIDFn = func(ctx context.Context, arg db.SystemDeleteUserEmailsByEmailExceptIDParams) error {
			deletedArgs = append(deletedArgs, arg)
			return nil
		}

		store := sessions.NewCookieStore([]byte("test"))
		sess := sessions.NewSession(store, "test")
		sess.Values = map[interface{}]interface{}{"UID": int32(1)}
		core.Store = store
		core.SessionName = "test"

		ctx := context.Background()
		ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
		cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithSession(sess))
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

		form := url.Values{"code": {code}}
		req := httptest.NewRequest(http.MethodPost, "/usr/email/verify", strings.NewReader(form.Encode())).WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		userEmailVerifyCodePage(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Fatalf("status=%d", rr.Code)
		}
		if loc := rr.Header().Get("Location"); loc != "/usr/email" {
			t.Fatalf("location=%q", loc)
		}
		if len(marked) != 1 || marked[0].ID != 1 {
			t.Fatalf("unexpected mark args: %#v", marked)
		}
		if len(prioritySetArgs) != 1 {
			t.Fatalf("unexpected priority updates: %#v", prioritySetArgs)
		}
		if arg := prioritySetArgs[0]; arg.ListerID != 1 || arg.NotificationPriority != 1 || arg.ID != 1 {
			t.Fatalf("unexpected priority args: %#v", arg)
		}
		if len(deletedArgs) != 1 {
			t.Fatalf("unexpected delete args: %#v", deletedArgs)
		}
		if arg := deletedArgs[0]; arg.Email != "e@example.com" || arg.ID != 1 {
			t.Fatalf("unexpected delete args: %#v", arg)
		}
	})
}
