package user

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

func TestResendVerificationEmailTask(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		var updated []db.SetVerificationCodeForListerParams
		queries := testhelpers.NewQuerierStub()
		queries.GetUserEmailByIDFn = func(ctx context.Context, id int32) (*db.UserEmail, error) {
			if id != 1 {
				return nil, sql.ErrNoRows
			}
			return &db.UserEmail{ID: 1, UserID: 1, Email: "a@example.com"}, nil
		}
		queries.SetVerificationCodeForListerFn = func(ctx context.Context, arg db.SetVerificationCodeForListerParams) error {
			updated = append(updated, arg)
			return nil
		}
		queries.SystemGetUserByIDFn = func(ctx context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
			return &db.SystemGetUserByIDRow{Idusers: 1, Username: sql.NullString{String: "alice", Valid: true}}, nil
		}

		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test"

		req := httptest.NewRequest("POST", "http://example.com/usr/email/resend", strings.NewReader("id=1"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		sess, _ := store.Get(req, core.SessionName)
		sess.Values["UID"] = int32(1)
		w := httptest.NewRecorder()
		_ = sess.Save(req, w)
		for _, c := range w.Result().Cookies() {
			req.AddCookie(c)
		}

		evt := &eventbus.TaskEvent{Data: map[string]any{}}
		ctx := context.Background()
		ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
		cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithSession(sess), common.WithEvent(evt))
		cd.UserID = 1
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

		addEmailTask.codeGenerator = func() (string, error) { return "deadbeef", nil }
		defer func() { addEmailTask.codeGenerator = nil }()

		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		handlers.TaskHandler(resendVerificationEmailTask)(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		if _, ok := evt.Data["page"]; !ok {
			t.Fatalf("missing page event data: %+v", evt.Data)
		}
		if _, ok := evt.Data["email"]; !ok {
			t.Fatalf("missing email event data: %+v", evt.Data)
		}
		if evt.Data["VerificationCode"] != "deadbeef" {
			t.Fatalf("verification code missing: %+v", evt.Data)
		}
		if evt.Data["Token"] != "deadbeef" {
			t.Fatalf("token missing: %+v", evt.Data)
		}
		if len(updated) != 1 {
			t.Fatalf("expected verification update, got %d", len(updated))
		}
		if arg := updated[0]; arg.ListerID != 1 || arg.ID != 1 {
			t.Fatalf("unexpected verification args: %#v", arg)
		}
	})
}
