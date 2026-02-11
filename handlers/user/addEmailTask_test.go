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
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestAddEmailTask(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		t.Run("Config Expiry", func(t *testing.T) {
			var inserted []db.InsertUserEmailParams
			queries := testhelpers.NewQuerierStub()
			queries.GetUserEmailByEmailFn = func(ctx context.Context, email string) (*db.UserEmail, error) {
				if email != "new@example.com" {
					return nil, sql.ErrNoRows
				}
				return nil, sql.ErrNoRows
			}
			queries.InsertUserEmailFn = func(ctx context.Context, arg db.InsertUserEmailParams) error {
				inserted = append(inserted, arg)
				return nil
			}
			queries.SystemGetUserByIDFn = func(ctx context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
				if id != 1 {
					t.Errorf("unexpected user id: %d", id)
				}
				return &db.SystemGetUserByIDRow{
					Idusers:                1,
					Email:                  sql.NullString{String: "primary@example.com", Valid: true},
					Username:               sql.NullString{String: "tester", Valid: true},
					PublicProfileEnabledAt: sql.NullTime{},
				}, nil
			}

			store := sessions.NewCookieStore([]byte("test"))
			core.Store = store
			core.SessionName = "test"
			sess, _ := store.Get(httptest.NewRequest(http.MethodGet, "http://example.com", nil), core.SessionName)
			sess.Values["UID"] = int32(1)

			addEmailTask.codeGenerator = func() (string, error) { return "deadbeef", nil }
			defer func() { addEmailTask.codeGenerator = nil }()

			evt := &eventbus.TaskEvent{Data: map[string]any{}}
			cfg := config.NewRuntimeConfig()
			cfg.EmailVerificationExpiryHours = 72
			ctx := context.Background()
			cd := common.NewCoreData(ctx, queries, cfg, common.WithSession(sess), common.WithEvent(evt))
			cd.UserID = 1

			ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
			ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

			expectedExpire := time.Now().Add(time.Duration(cfg.EmailVerificationExpiryHours) * time.Hour)

			form := url.Values{"new_email": {"new@example.com"}}
			req := httptest.NewRequest(http.MethodPost, "http://example.com/usr/email", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			handlers.TaskHandler(addEmailTask)(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status=%d", rr.Code)
			}

			expiresAt, ok := evt.Data["ExpiresAt"].(time.Time)
			if !ok {
				t.Fatalf("expected expiry time in event data, got %#v", evt.Data["ExpiresAt"])
			}
			margin := 2 * time.Second
			if expiresAt.Before(expectedExpire.Add(-margin)) || expiresAt.After(expectedExpire.Add(margin)) {
				t.Fatalf("expiry %v not within margin of expected %v", expiresAt, expectedExpire)
			}

			if _, ok := evt.Data["URL"]; !ok {
				t.Fatalf("missing URL event data: %+v", evt.Data)
			}
			if evt.Data["VerificationCode"] != "deadbeef" {
				t.Fatalf("verification code not set: %+v", evt.Data)
			}
			if evt.Data["Token"] != "deadbeef" {
				t.Fatalf("token not set: %+v", evt.Data)
			}
			if evt.Data["Username"] != "tester" {
				t.Fatalf("username not set: %+v", evt.Data)
			}

			if len(inserted) != 1 {
				t.Fatalf("expected insert, got %d", len(inserted))
			}
			insertedArg := inserted[0]
			if insertedArg.UserID != 1 || insertedArg.Email != "new@example.com" {
				t.Fatalf("unexpected insert args: %#v", insertedArg)
			}
			if insertedArg.VerificationExpiresAt.Time.Before(expectedExpire.Add(-margin)) || insertedArg.VerificationExpiresAt.Time.After(expectedExpire.Add(margin)) {
				t.Fatalf("insert expiry %v not within margin of expected %v", insertedArg.VerificationExpiresAt.Time, expectedExpire)
			}
		})
	})

	t.Run("Unhappy Path", func(t *testing.T) {
		t.Run("Invalid Email", func(t *testing.T) {
			queries := testhelpers.NewQuerierStub()

			store := sessions.NewCookieStore([]byte("test"))
			core.Store = store
			core.SessionName = "test"
			sess, _ := store.Get(httptest.NewRequest(http.MethodGet, "http://example.com", nil), core.SessionName)
			sess.Values["UID"] = int32(1)

			evt := &eventbus.TaskEvent{Data: map[string]any{}}
			ctx := context.Background()
			cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithSession(sess), common.WithEvent(evt))
			cd.UserID = 1
			ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
			ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

			form := url.Values{"new_email": {"foo@bar..com"}}
			req := httptest.NewRequest(http.MethodPost, "http://example.com/usr/email", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			handlers.TaskHandler(addEmailTask)(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status=%d", rr.Code)
			}
			if len(evt.Data) != 0 {
				t.Fatalf("unexpected event data: %+v", evt.Data)
			}
			if cd.AutoRefresh == "" || !strings.Contains(cd.AutoRefresh, "invalid+email") {
				t.Fatalf("auto refresh=%q", cd.AutoRefresh)
			}
		})
	})
}
