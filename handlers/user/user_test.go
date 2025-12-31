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

// helper to setup session cookie
func newRequestWithSession(method, target string, values map[string]interface{}) (*http.Request, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(method, target, nil)
	w := httptest.NewRecorder()
	sess, _ := store.Get(r, sessionName)
	for k, v := range values {
		sess.Values[k] = v
	}
	sess.Save(r, w)
	for _, c := range w.Result().Cookies() {
		r.AddCookie(c)
	}
	return r, httptest.NewRecorder()
}

type testMailQueries struct {
	db.Querier
	user *db.SystemGetUserByIDRow
}

func (q *testMailQueries) SystemGetUserByID(context.Context, int32) (*db.SystemGetUserByIDRow, error) {
	return q.user, nil
}

type userEmailPageQueries struct {
	db.Querier
	user    *db.SystemGetUserByIDRow
	pref    *db.Preference
	prefErr error
	emails  []*db.UserEmail
}

func (q *userEmailPageQueries) SystemGetUserByID(context.Context, int32) (*db.SystemGetUserByIDRow, error) {
	return q.user, nil
}

func (q *userEmailPageQueries) GetPreferenceForLister(context.Context, int32) (*db.Preference, error) {
	if q.prefErr != nil {
		return nil, q.prefErr
	}
	return q.pref, nil
}

func (q *userEmailPageQueries) ListUserEmailsForLister(context.Context, db.ListUserEmailsForListerParams) ([]*db.UserEmail, error) {
	return q.emails, nil
}

func (q *userEmailPageQueries) GetPermissionsByUserID(context.Context, int32) ([]*db.GetPermissionsByUserIDRow, error) {
	return nil, nil
}

func (q *userEmailPageQueries) UpdateCustomCssForLister(context.Context, db.UpdateCustomCssForListerParams) error {
	return nil
}

type languageSaveQueries struct {
	db.Querier
	languages     []*db.Language
	deleted       bool
	insertedLangs []db.InsertUserLangParams
	pref          *db.Preference
	prefErr       error
	insertedPrefs []db.InsertPreferenceForListerParams
	updatedPrefs  []db.UpdatePreferenceForListerParams
}

func (q *languageSaveQueries) DeleteUserLanguagesForUser(context.Context, int32) error {
	q.deleted = true
	return nil
}

func (q *languageSaveQueries) SystemListLanguages(context.Context) ([]*db.Language, error) {
	return q.languages, nil
}

func (q *languageSaveQueries) InsertUserLang(_ context.Context, arg db.InsertUserLangParams) error {
	q.insertedLangs = append(q.insertedLangs, arg)
	return nil
}

func (q *languageSaveQueries) GetPreferenceForLister(context.Context, int32) (*db.Preference, error) {
	if q.prefErr != nil {
		return nil, q.prefErr
	}
	return q.pref, nil
}

func (q *languageSaveQueries) InsertPreferenceForLister(_ context.Context, arg db.InsertPreferenceForListerParams) error {
	q.insertedPrefs = append(q.insertedPrefs, arg)
	return nil
}

func (q *languageSaveQueries) UpdatePreferenceForLister(_ context.Context, arg db.UpdatePreferenceForListerParams) error {
	q.updatedPrefs = append(q.updatedPrefs, arg)
	return nil
}

func (q *languageSaveQueries) UpdateCustomCssForLister(context.Context, db.UpdateCustomCssForListerParams) error {
	return nil
}

func TestUserEmailTestAction_NoProvider(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.EmailProvider = ""
	queries := &testMailQueries{
		user: &db.SystemGetUserByIDRow{
			Idusers:  1,
			Email:    sql.NullString{String: "e", Valid: true},
			Username: sql.NullString{String: "u", Valid: true},
		},
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
	want := url.QueryEscape(ErrMailNotConfigured.Error())
	if req.URL.RawQuery != "error="+want {
		t.Fatalf("query=%q", req.URL.RawQuery)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "<a href=") {
		t.Fatalf("body=%q", body)
	}
}

func TestUserEmailTestAction_WithProvider(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.EmailProvider = "log"

	queries := &testMailQueries{
		user: &db.SystemGetUserByIDRow{
			Idusers:  1,
			Email:    sql.NullString{String: "e", Valid: true},
			Username: sql.NullString{String: "u", Valid: true},
		},
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
}

func TestUserEmailPage_ShowError(t *testing.T) {
	queries := &userEmailPageQueries{
		user: &db.SystemGetUserByIDRow{
			Idusers:  1,
			Email:    sql.NullString{String: "e", Valid: true},
			Username: sql.NullString{String: "u", Valid: true},
		},
		prefErr: sql.ErrNoRows,
		emails:  []*db.UserEmail{{ID: 1, UserID: 1, Email: "e"}},
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
}

func TestUserEmailPage_NoUnverified(t *testing.T) {
	queries := &userEmailPageQueries{
		user: &db.SystemGetUserByIDRow{
			Idusers:  1,
			Email:    sql.NullString{String: "e", Valid: true},
			Username: sql.NullString{String: "u", Valid: true},
		},
		prefErr: sql.ErrNoRows,
		emails: []*db.UserEmail{{
			ID:         1,
			UserID:     1,
			Email:      "e",
			VerifiedAt: sql.NullTime{Time: time.Now(), Valid: true},
		}},
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
}

func TestUserEmailPage_NoVerified(t *testing.T) {
	queries := &userEmailPageQueries{
		user: &db.SystemGetUserByIDRow{
			Idusers:  1,
			Email:    sql.NullString{String: "e", Valid: true},
			Username: sql.NullString{String: "u", Valid: true},
		},
		prefErr: sql.ErrNoRows,
		emails:  []*db.UserEmail{},
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
}

func TestUserLangSaveAllActionPage_NewPref(t *testing.T) {
	queries := &languageSaveQueries{
		languages: []*db.Language{
			{ID: 1, Nameof: sql.NullString{String: "en", Valid: true}},
			{ID: 2, Nameof: sql.NullString{String: "fr", Valid: true}},
		},
		prefErr: sql.ErrNoRows,
	}
	store = sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = sessionName
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
	if !queries.deleted {
		t.Fatal("expected user languages to be deleted")
	}
	if len(queries.insertedLangs) != 1 || queries.insertedLangs[0].LanguageID != 1 {
		t.Fatalf("unexpected inserted languages: %#v", queries.insertedLangs)
	}
	if len(queries.insertedPrefs) != 1 {
		t.Fatalf("unexpected preference inserts: %#v", queries.insertedPrefs)
	}
	insertedPref := queries.insertedPrefs[0]
	if insertedPref.ListerID != 1 || insertedPref.LanguageID.Int32 != 2 || insertedPref.PageSize != int32(cfg.PageSizeDefault) {
		t.Fatalf("unexpected preference insert: %#v", insertedPref)
	}
}

func TestUserLangSaveLanguagesActionPage(t *testing.T) {
	queries := &languageSaveQueries{
		languages: []*db.Language{{ID: 1, Nameof: sql.NullString{String: "en", Valid: true}}},
	}
	store = sessions.NewCookieStore([]byte("test"))

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
	if !queries.deleted {
		t.Fatal("expected user languages to be deleted")
	}
	if len(queries.insertedLangs) != 1 || queries.insertedLangs[0].LanguageID != 1 {
		t.Fatalf("unexpected inserted languages: %#v", queries.insertedLangs)
	}
}

func TestUserLangSaveLanguageActionPage_UpdatePref(t *testing.T) {
	queries := &languageSaveQueries{
		pref: &db.Preference{
			Idpreferences:        1,
			LanguageID:           sql.NullInt32{Int32: 1, Valid: true},
			UsersIdusers:         1,
			Emailforumupdates:    sql.NullBool{},
			PageSize:             int32(15),
			AutoSubscribeReplies: true,
			Timezone:             sql.NullString{},
		},
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
	if len(queries.updatedPrefs) != 1 {
		t.Fatalf("unexpected preference updates: %#v", queries.updatedPrefs)
	}
	updatedPref := queries.updatedPrefs[0]
	if updatedPref.LanguageID.Int32 != 2 || updatedPref.ListerID != 1 {
		t.Fatalf("unexpected preference update: %#v", updatedPref)
	}
}
