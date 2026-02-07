package user

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestUserAppearancePage(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		queries.GetPreferenceForListerFn = func(ctx context.Context, id int32) (*db.Preference, error) {
			return &db.Preference{
				CustomCss: sql.NullString{String: "body { color: red; }", Valid: true},
			}, nil
		}

		req := httptest.NewRequest("GET", "/usr/appearance", nil)
		ctx := req.Context()
		cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{}))
		cd.UserID = 1
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		userAppearancePage(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		body := rr.Body.String()
		if !strings.Contains(body, "body { color: red; }") {
			t.Fatalf("body missing css: %q", body)
		}
	})
}

func TestAppearanceSaveTask(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		var updated bool
		var css string

		queries := testhelpers.NewQuerierStub()
		queries.GetPreferenceForListerFn = func(ctx context.Context, id int32) (*db.Preference, error) {
			return &db.Preference{
				CustomCss: sql.NullString{String: "old", Valid: true},
			}, nil
		}
		queries.UpdateCustomCssForListerFn = func(ctx context.Context, arg db.UpdateCustomCssForListerParams) error {
			updated = true
			css = arg.CustomCss.String
			return nil
		}

		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test"

		form := url.Values{}
		form.Set("task", "Save appearance")
		form.Set("custom_css", "new css")

		req := httptest.NewRequest("POST", "/usr/appearance", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		sess, _ := store.Get(req, core.SessionName)
		sess.Values["UID"] = int32(1)
		w := httptest.NewRecorder()
		sess.Save(req, w)
		for _, c := range w.Result().Cookies() {
			req.AddCookie(c)
		}
		rr := httptest.NewRecorder()

		ctx := req.Context()
		cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithSession(sess))
		cd.UserID = 1
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		appearanceSaveTask.Action(rr, req)

		if rr.Result().StatusCode != http.StatusOK {
			t.Fatalf("status=%d", rr.Result().StatusCode)
		}
		if !updated {
			t.Fatal("expected update")
		}
		if css != "new css" {
			t.Fatalf("expected 'new css', got %q", css)
		}
		// Verify it re-renders
		body := rr.Body.String()
		if !strings.Contains(body, "Appearance Settings") {
			t.Fatalf("expected re-render: %q", body)
		}
	})
}
