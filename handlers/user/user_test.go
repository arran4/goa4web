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
	"github.com/arran4/goa4web/internal/email"
	logProv "github.com/arran4/goa4web/internal/email/log"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func newEmailReg() *email.Registry {
	r := email.NewRegistry()
	logProv.Register(r)
	return r
}

var (
	store       *sessions.CookieStore
	sessionName = "my-session"
)

func TestUserEmailTestAction(t *testing.T) {
	t.Run("No Provider", func(t *testing.T) {
		cfg := config.NewRuntimeConfig()
		cfg.EmailProvider = ""

		queries := testhelpers.NewQuerierStub()
		queries.SystemGetUserByIDFn = func(ctx context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
			return &db.SystemGetUserByIDRow{
				Idusers:  1,
				Email:    sql.NullString{String: "e", Valid: true},
				Username: sql.NullString{String: "u", Valid: true},
			}, nil
		}

		req := httptest.NewRequest("POST", "/email", nil)
		ctx := req.Context()
		reg := newEmailReg()
		p, _ := reg.ProviderFromConfig(cfg)
		cd := common.NewCoreData(ctx, queries, cfg, common.WithEmailProvider(p), common.WithUserRoles([]string{}))
		cd.UserID = 1
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(testMailTask)(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		body := rr.Body.String()
		if !strings.Contains(body, "Refresh") && !strings.Contains(body, "Redirect") && !strings.Contains(body, "<a href=") {
			t.Logf("Body: %s", body)
		}
	})

	t.Run("With Provider", func(t *testing.T) {
		cfg := config.NewRuntimeConfig()
		cfg.EmailProvider = "log"

		queries := testhelpers.NewQuerierStub()
		queries.SystemGetUserByIDFn = func(ctx context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
			return &db.SystemGetUserByIDRow{
				Idusers:  1,
				Email:    sql.NullString{String: "e", Valid: true},
				Username: sql.NullString{String: "u", Valid: true},
			}, nil
		}

		req := httptest.NewRequest("POST", "/email", nil)
		ctx := req.Context()
		reg := newEmailReg()
		p, _ := reg.ProviderFromConfig(cfg)
		cd := common.NewCoreData(ctx, queries, cfg, common.WithEmailProvider(p), common.WithUserRoles([]string{}))
		cd.UserID = 1
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(testMailTask)(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
	})
}

func TestUserEmailPage(t *testing.T) {
	t.Run("Show Error", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		queries.SystemGetUserByIDFn = func(ctx context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
			return &db.SystemGetUserByIDRow{
				Idusers:  1,
				Email:    sql.NullString{String: "e", Valid: true},
				Username: sql.NullString{String: "u", Valid: true},
			}, nil
		}
		queries.GetPreferenceForListerFn = func(ctx context.Context, id int32) (*db.Preference, error) {
			return nil, sql.ErrNoRows
		}
		queries.ListUserEmailsForListerFn = func(ctx context.Context, arg db.ListUserEmailsForListerParams) ([]*db.UserEmail, error) {
			return []*db.UserEmail{{ID: 1, UserID: 1, Email: "e"}}, nil
		}

		req := httptest.NewRequest("GET", "/usr/email?error=missing", nil)
		ctx := req.Context()
		cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{}))
		cd.UserID = 1
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		userEmailPage(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "missing") {
			t.Fatalf("body=%q", rr.Body.String())
		}
	})

	t.Run("No Unverified", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		queries.SystemGetUserByIDFn = func(ctx context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
			return &db.SystemGetUserByIDRow{
				Idusers:  1,
				Email:    sql.NullString{String: "e", Valid: true},
				Username: sql.NullString{String: "u", Valid: true},
			}, nil
		}
		queries.GetPreferenceForListerFn = func(ctx context.Context, id int32) (*db.Preference, error) {
			return nil, sql.ErrNoRows
		}
		queries.ListUserEmailsForListerFn = func(ctx context.Context, arg db.ListUserEmailsForListerParams) ([]*db.UserEmail, error) {
			return []*db.UserEmail{{
				ID:         1,
				UserID:     1,
				Email:      "e",
				VerifiedAt: sql.NullTime{Time: time.Now(), Valid: true},
			}}, nil
		}

		req := httptest.NewRequest("GET", "/usr/email", nil)
		ctx := req.Context()
		cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{}))
		cd.UserID = 1
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		userEmailPage(rr, req)

		body := rr.Body.String()
		if strings.Contains(body, "Unverified Emails") {
			t.Fatalf("unverified section should be hidden: %q", body)
		}
		if !strings.Contains(body, "Verified Emails") {
			t.Fatalf("missing verified section: %q", body)
		}
	})

	t.Run("No Verified", func(t *testing.T) {
		queries := testhelpers.NewQuerierStub()
		queries.SystemGetUserByIDFn = func(ctx context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
			return &db.SystemGetUserByIDRow{
				Idusers:  1,
				Email:    sql.NullString{String: "e", Valid: true},
				Username: sql.NullString{String: "u", Valid: true},
			}, nil
		}
		queries.GetPreferenceForListerFn = func(ctx context.Context, id int32) (*db.Preference, error) {
			return nil, sql.ErrNoRows
		}
		queries.ListUserEmailsForListerFn = func(ctx context.Context, arg db.ListUserEmailsForListerParams) ([]*db.UserEmail, error) {
			return []*db.UserEmail{}, nil
		}

		req := httptest.NewRequest("GET", "/usr/email", nil)
		ctx := req.Context()
		cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{}))
		cd.UserID = 1
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		userEmailPage(rr, req)

		body := rr.Body.String()
		if !strings.Contains(body, "No verified emails") {
			t.Fatalf("missing warning message: %q", body)
		}
		if strings.Contains(body, "Unverified Emails") {
			t.Fatalf("unexpected unverified section: %q", body)
		}
	})
}

func TestUserLangSave(t *testing.T) {
	t.Run("Save All New Pref", func(t *testing.T) {
		var deleted bool
		var insertedLangs []db.InsertUserLangParams
		var insertedPrefs []db.InsertPreferenceForListerParams

		queries := testhelpers.NewQuerierStub()
		queries.SystemListLanguagesReturns = []*db.Language{
			{ID: 1, Nameof: sql.NullString{String: "en", Valid: true}},
			{ID: 2, Nameof: sql.NullString{String: "fr", Valid: true}},
		}
		queries.GetPreferenceForListerFn = func(ctx context.Context, id int32) (*db.Preference, error) {
			return nil, sql.ErrNoRows
		}
		queries.DeleteUserLanguagesForUserFn = func(ctx context.Context, id int32) error {
			deleted = true
			return nil
		}
		queries.InsertUserLangFn = func(ctx context.Context, arg db.InsertUserLangParams) error {
			insertedLangs = append(insertedLangs, arg)
			return nil
		}
		queries.InsertPreferenceForListerFn = func(ctx context.Context, arg db.InsertPreferenceForListerParams) error {
			insertedPrefs = append(insertedPrefs, arg)
			return nil
		}

		store = sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = sessionName

		form := url.Values{}
		form.Set("dothis", "Save all")
		form.Set("language1", "on")
		form.Set("defaultLanguage", "2")

		req := httptest.NewRequest("POST", "/usr/lang", strings.NewReader(form.Encode()))
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
		cfg := config.NewRuntimeConfig()
		cfg.PageSizeDefault = 15
		cd := common.NewCoreData(ctx, queries, cfg, common.WithSession(sess))
		cd.UserID = 1
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		saveAllTask.Action(rr, req)

		if rr.Result().StatusCode != http.StatusOK {
			t.Fatalf("status=%d", rr.Result().StatusCode)
		}
		if !deleted {
			t.Fatal("expected user languages to be deleted")
		}
		if len(insertedLangs) != 1 || insertedLangs[0].LanguageID != 1 {
			t.Fatalf("unexpected inserted languages: %#v", insertedLangs)
		}
		if len(insertedPrefs) != 1 {
			t.Fatalf("unexpected preference inserts: %#v", insertedPrefs)
		}
		insertedPref := insertedPrefs[0]
		if insertedPref.ListerID != 1 || insertedPref.LanguageID.Int32 != 2 || insertedPref.PageSize != int32(cfg.PageSizeDefault) {
			t.Fatalf("unexpected preference insert: %#v", insertedPref)
		}
	})

	t.Run("Save Languages", func(t *testing.T) {
		var deleted bool
		var insertedLangs []db.InsertUserLangParams

		queries := testhelpers.NewQuerierStub()
		queries.SystemListLanguagesReturns = []*db.Language{{ID: 1, Nameof: sql.NullString{String: "en", Valid: true}}}
		queries.DeleteUserLanguagesForUserFn = func(ctx context.Context, id int32) error {
			deleted = true
			return nil
		}
		queries.InsertUserLangFn = func(ctx context.Context, arg db.InsertUserLangParams) error {
			insertedLangs = append(insertedLangs, arg)
			return nil
		}

		store = sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = sessionName

		form := url.Values{}
		form.Set("dothis", "Save languages")
		form.Set("language1", "on")

		req := httptest.NewRequest("POST", "/usr/lang", strings.NewReader(form.Encode()))
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

		saveLanguagesTask.Action(rr, req)

		if rr.Result().StatusCode != http.StatusOK {
			t.Fatalf("status=%d", rr.Result().StatusCode)
		}
		if !deleted {
			t.Fatal("expected user languages to be deleted")
		}
		if len(insertedLangs) != 1 || insertedLangs[0].LanguageID != 1 {
			t.Fatalf("unexpected inserted languages: %#v", insertedLangs)
		}
	})

	t.Run("Save Language Update Pref", func(t *testing.T) {
		var updatedPrefs []db.UpdatePreferenceForListerParams

		queries := testhelpers.NewQuerierStub()
		queries.GetPreferenceForListerFn = func(ctx context.Context, id int32) (*db.Preference, error) {
			return &db.Preference{
				Idpreferences:        1,
				LanguageID:           sql.NullInt32{Int32: 1, Valid: true},
				UsersIdusers:         1,
				Emailforumupdates:    sql.NullBool{},
				PageSize:             int32(15),
				AutoSubscribeReplies: true,
				Timezone:             sql.NullString{},
			}, nil
		}
		queries.UpdatePreferenceForListerFn = func(ctx context.Context, arg db.UpdatePreferenceForListerParams) error {
			updatedPrefs = append(updatedPrefs, arg)
			return nil
		}

		store = sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = sessionName

		form := url.Values{}
		form.Set("dothis", "Save language")
		form.Set("defaultLanguage", "2")

		req := httptest.NewRequest("POST", "/usr/lang", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		sess, _ := store.Get(req, sessionName)
		sess.Values["UID"] = int32(1)
		w := httptest.NewRecorder()
		sess.Save(req, w)
		for _, c := range w.Result().Cookies() {
			req.AddCookie(c)
		}
		rr := httptest.NewRecorder()
		cfg := config.NewRuntimeConfig()
		cfg.PageSizeDefault = 15

		ctx := req.Context()
		cd := common.NewCoreData(ctx, queries, cfg, common.WithSession(sess))
		cd.UserID = 1
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		saveLanguageTask.Action(rr, req)

		if rr.Result().StatusCode != http.StatusOK {
			t.Fatalf("status=%d", rr.Result().StatusCode)
		}
		if len(updatedPrefs) != 1 {
			t.Fatalf("unexpected preference updates: %#v", updatedPrefs)
		}
		updatedPref := updatedPrefs[0]
		if updatedPref.LanguageID.Int32 != 2 || updatedPref.ListerID != 1 {
			t.Fatalf("unexpected preference update: %#v", updatedPref)
		}
	})
}
