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
)

type appearanceQueries struct {
	db.Querier
	pref    *db.Preference
	prefErr error
	updated bool
	css     string
}

func (q *appearanceQueries) GetPreferenceForLister(context.Context, int32) (*db.Preference, error) {
	if q.prefErr != nil {
		return nil, q.prefErr
	}
	return q.pref, nil
}

func (q *appearanceQueries) UpdateCustomCssForLister(ctx context.Context, arg db.UpdateCustomCssForListerParams) error {
	q.updated = true
	q.css = arg.CustomCss.String
	return nil
}

func (q *appearanceQueries) InsertPreferenceForLister(ctx context.Context, arg db.InsertPreferenceForListerParams) error {
	return nil
}

func (q *appearanceQueries) GetPermissionsByUserID(context.Context, int32) ([]*db.GetPermissionsByUserIDRow, error) {
	return nil, nil
}

func TestUserAppearancePage_Get(t *testing.T) {
	queries := &appearanceQueries{
		pref: &db.Preference{
			CustomCss: sql.NullString{String: "body { color: red; }", Valid: true},
		},
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
}

func TestAppearanceSaveTask_Action(t *testing.T) {
	queries := &appearanceQueries{
		pref: &db.Preference{
			CustomCss: sql.NullString{String: "old", Valid: true},
		},
	}
	store = sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = sessionName

	form := url.Values{}
	form.Set("task", "Save appearance")
	form.Set("custom_css", "new css")

	req := httptest.NewRequest("POST", "/usr/appearance", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	sess, _ := store.Get(req, sessionName)
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
	if !queries.updated {
		t.Fatal("expected update")
	}
	if queries.css != "new css" {
		t.Fatalf("expected 'new css', got %q", queries.css)
	}
	// Verify it re-renders
	body := rr.Body.String()
	if !strings.Contains(body, "Appearance Settings") {
		t.Fatalf("expected re-render: %q", body)
	}
}
