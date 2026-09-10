package server

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
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/handlers/auth"
	"github.com/arran4/goa4web/handlers/news"
	"github.com/arran4/goa4web/handlers/user"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	nav "github.com/arran4/goa4web/internal/navigation"
	routerpkg "github.com/arran4/goa4web/internal/router"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func TestIssue3095ProductionAuthTransitions(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	passwordHash, passwordAlgorithm, err := auth.HashPassword("correcthorse")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	q.SystemGetLoginFn = func(_ context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
		if username.String != "testuser" {
			return nil, sql.ErrNoRows
		}
		return &db.SystemGetLoginRow{
			Idusers:         10,
			Passwd:          sql.NullString{String: passwordHash, Valid: true},
			PasswdAlgorithm: sql.NullString{String: passwordAlgorithm, Valid: true},
			Username:        username,
		}, nil
	}
	q.GetLoginRoleForUserFn = func(context.Context, int32) (int32, error) { return 1, nil }
	q.SystemGetUserByUsernameErr = sql.ErrNoRows
	q.SystemGetUserByEmailErr = sql.ErrNoRows
	q.SystemInsertUserReturns = 20
	q.GetPasswordResetByUserErr = sql.ErrNoRows

	cfg := config.NewRuntimeConfig()
	cfg.SessionName = "issue3095_session"
	store := sessions.NewCookieStore([]byte("01234567890123456789012345678901"))

	originalStore, originalSessionName := core.Store, core.SessionName
	core.Store, core.SessionName = store, cfg.SessionName
	t.Cleanup(func() {
		core.Store, core.SessionName = originalStore, originalSessionName
	})

	routerRegistry := routerpkg.NewRegistry()
	auth.Register(routerRegistry)
	user.Register(routerRegistry)
	news.Register(routerRegistry)
	navigationRegistry := nav.NewRegistry()
	r := mux.NewRouter()
	routerpkg.RegisterRoutes(r, routerRegistry, cfg, navigationRegistry)

	sessionManager := &sessionManagerStub{}
	srv := New(
		WithStore(store),
		WithQuerier(q),
		WithConfig(cfg),
		WithRouterRegistry(routerRegistry),
		WithNavRegistry(navigationRegistry),
		WithEmailRegistry(email.NewRegistry()),
		WithSessionManager(sessionManager),
	)
	handler := srv.CoreDataMiddleware()(r)

	request := func(method, target string, form url.Values, cookie *http.Cookie) *httptest.ResponseRecorder {
		t.Helper()
		var body *strings.Reader
		if form == nil {
			body = strings.NewReader("")
		} else {
			body = strings.NewReader(form.Encode())
		}
		req := httptest.NewRequest(method, target, body)
		if form != nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		if cookie != nil {
			req.AddCookie(cookie)
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		return rr
	}

	for _, target := range []string{"/login", "/register", "/forgot"} {
		sensitivePage := request(http.MethodGet, target, nil, nil)
		if sensitivePage.Code != http.StatusOK {
			t.Fatalf("anonymous GET %s status = %d; want 200", target, sensitivePage.Code)
		}
		assertIssue3095NoStore(t, sensitivePage)
		if target == "/login" && !strings.Contains(sensitivePage.Body.String(), "window.addEventListener('pageshow'") {
			t.Error("anonymous login page is missing the BFCache restoration safeguard")
		}
	}

	registrationPage := request(http.MethodGet, "/register?back=%2Fnews%3Fview%3Dfull&method=POST&data=secret", nil, nil)
	if !strings.Contains(registrationPage.Body.String(), `name="back" value="/news?view=full"`) {
		t.Error("registration form did not retain its sanitized safe GET continuation")
	}
	for _, replayField := range []string{`name="method"`, `name="data"`} {
		if strings.Contains(registrationPage.Body.String(), replayField) {
			t.Errorf("registration form contains legacy replay field %q", replayField)
		}
	}

	registration := request(http.MethodPost, "/register", url.Values{
		"task":     {"Register"},
		"username": {"new-user"},
		"password": {"new-password"},
		"email":    {"new-user@example.com"},
		"back":     {"/news?view=full"},
	}, nil)
	if registration.Code != http.StatusSeeOther {
		t.Fatalf("registration status = %d; want direct 303; body: %s", registration.Code, registration.Body.String())
	}
	registrationLocation, err := url.Parse(registration.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse registration Location: %v", err)
	}
	if registrationLocation.Path != "/login" || registrationLocation.Query().Get("notice") != "approval is pending" || registrationLocation.Query().Get("back") != "/news?view=full" {
		t.Errorf("registration Location = %q; want login notice and safe back", registrationLocation.String())
	}
	if registrationLocation.Query().Has("method") || registrationLocation.Query().Has("data") {
		t.Errorf("registration Location contains legacy replay state: %q", registrationLocation.String())
	}
	assertIssue3095NoStore(t, registration)
	if strings.Contains(strings.ToLower(registration.Body.String()), "http-equiv=\"refresh\"") {
		t.Error("registration rendered a meta-refresh transition")
	}

	failedRegistration := request(http.MethodPost, "/register", url.Values{
		"task":     {"Register"},
		"password": {"new-password"},
		"email":    {"invalid-registration@example.com"},
	}, nil)
	assertIssue3095NoStore(t, failedRegistration)

	replayRegistration := request(http.MethodPost, "/register", url.Values{
		"task":     {"Register"},
		"username": {"replay-user"},
		"password": {"new-password"},
		"email":    {"replay-user@example.com"},
		"back":     {"/news"},
		"method":   {"POST"},
		"data":     {"secret=must-not-leak"},
	}, nil)
	assertIssue3095NoStore(t, replayRegistration)
	if strings.Contains(replayRegistration.Header().Get("Location"), "secret") || strings.Contains(replayRegistration.Body.String(), "secret=must-not-leak") {
		t.Error("registration response exposed rejected arbitrary POST replay data")
	}

	for _, authPost := range []url.Values{
		{"task": {"Password Reset"}, "username": {"testuser"}},
		{"task": {"Email Association Request"}, "username": {"testuser"}},
		{"task": {"Login"}, "username": {"missing-user"}, "password": {"wrong"}},
	} {
		authPostResponse := request(http.MethodPost, map[string]string{
			"Password Reset":            "/forgot",
			"Email Association Request": "/forgot",
			"Login":                     "/login",
		}[authPost.Get("task")], authPost, nil)
		assertIssue3095NoStore(t, authPostResponse)
		if got := authPostResponse.Header().Get("Cache-Control"); got == "public, max-age=3600" {
			t.Errorf("auth POST %q was publicly cacheable", authPost.Get("task"))
		}
	}

	protected := request(http.MethodGet, "/usr?from=protected", nil, nil)
	assertIssue3095Redirect(t, protected, "/login", "/usr?from=protected")
	assertIssue3095NoStore(t, protected)
	journeyCookie := issue3095Cookie(t, protected, cfg.SessionName)

	login := request(http.MethodPost, "/login", url.Values{
		"task":     {"Login"},
		"username": {"testuser"},
		"password": {"correcthorse"},
		"back":     {"/usr?from=protected"},
	}, journeyCookie)
	if login.Code != http.StatusSeeOther || login.Header().Get("Location") != "/usr?from=protected" {
		t.Fatalf("login response = %d Location %q; want 303 to protected page", login.Code, login.Header().Get("Location"))
	}
	assertIssue3095NoStore(t, login)
	if strings.Contains(strings.ToLower(login.Body.String()), "http-equiv=\"refresh\"") {
		t.Error("successful login rendered a meta-refresh transition")
	}
	journeyCookie = issue3095Cookie(t, login, cfg.SessionName)

	authenticatedLoginPage := request(http.MethodGet, "/login?code=abc&back=%2Fusr%3Fview%3Dsettings", nil, journeyCookie)
	if authenticatedLoginPage.Code != http.StatusOK {
		t.Fatalf("authenticated GET /login status = %d; want 200", authenticatedLoginPage.Code)
	}
	assertIssue3095NoStore(t, authenticatedLoginPage)
	if !strings.Contains(authenticatedLoginPage.Body.String(), "window.addEventListener('pageshow'") {
		t.Error("authenticated login page is missing the BFCache restoration safeguard")
	}
	for _, field := range []string{`name="code" value="abc"`, `name="back" value="/usr?view=settings"`} {
		if !strings.Contains(authenticatedLoginPage.Body.String(), field) {
			t.Errorf("authenticated login page is missing %q", field)
		}
	}
	for _, replayField := range []string{`name="method"`, `name="data"`} {
		if strings.Contains(authenticatedLoginPage.Body.String(), replayField) {
			t.Errorf("authenticated login page contains legacy POST replay field %q", replayField)
		}
	}

	logout := request(http.MethodGet, "/usr/logout", nil, journeyCookie)
	if logout.Code != http.StatusSeeOther || logout.Header().Get("Location") != "/" {
		t.Fatalf("logout response = %d Location %q; want 303 to /", logout.Code, logout.Header().Get("Location"))
	}
	assertIssue3095NoStore(t, logout)
	journeyCookie = issue3095Cookie(t, logout, cfg.SessionName)
	assertIssue3095SessionUnauthenticated(t, store, cfg.SessionName, journeyCookie)

	afterLogout := request(http.MethodGet, "/usr?after=logout", nil, journeyCookie)
	assertIssue3095Redirect(t, afterLogout, "/login", "/usr?after=logout")
	assertIssue3095NoStore(t, afterLogout)

	publicPage := request(http.MethodGet, "/", nil, nil)
	if publicPage.Code != http.StatusOK {
		t.Fatalf("anonymous public page status = %d; want 200; body: %s", publicPage.Code, publicPage.Body.String())
	}
	if got := publicPage.Header().Get("Cache-Control"); got != "public, max-age=3600" {
		t.Errorf("anonymous public page Cache-Control = %q; want public, max-age=3600", got)
	}
	if strings.Contains(publicPage.Body.String(), "window.addEventListener('pageshow'") {
		t.Error("anonymous public page unexpectedly includes the BFCache restoration safeguard")
	}

	publicWithSessionCookie := request(http.MethodGet, "/", nil, journeyCookie)
	assertIssue3095NoStore(t, publicWithSessionCookie)

	corruptCookie := &http.Cookie{Name: cfg.SessionName, Value: "corrupt-signed-session"}
	corrupt := request(http.MethodGet, "/usr?corrupt=yes", nil, corruptCookie)
	assertIssue3095Redirect(t, corrupt, "/login", "/usr?corrupt=yes")
	assertIssue3095NoStore(t, corrupt)
	clearedCorruptCookie := issue3095Cookie(t, corrupt, cfg.SessionName)
	if clearedCorruptCookie.MaxAge >= 0 {
		t.Errorf("corrupt session cookie MaxAge = %d; want a deletion cookie", clearedCorruptCookie.MaxAge)
	}
	corruptRecoveryLogin := request(http.MethodPost, "/login", url.Values{
		"task":     {"Login"},
		"username": {"testuser"},
		"password": {"correcthorse"},
		"back":     {"/usr?corrupt=yes"},
	}, clearedCorruptCookie)
	if corruptRecoveryLogin.Code != http.StatusSeeOther || corruptRecoveryLogin.Header().Get("Location") != "/usr?corrupt=yes" {
		t.Fatalf("corrupt-session recovery login = %d Location %q; want 303 to original URI", corruptRecoveryLogin.Code, corruptRecoveryLogin.Header().Get("Location"))
	}
	assertIssue3095NoStore(t, corruptRecoveryLogin)
	recoveredCookie := issue3095Cookie(t, corruptRecoveryLogin, cfg.SessionName)
	corruptContinuation := request(http.MethodGet, "/usr?corrupt=yes", nil, recoveredCookie)
	if corruptContinuation.Code != http.StatusOK || corruptContinuation.Header().Get("Location") != "" {
		t.Fatalf("corrupt-session continuation = %d Location %q; want 200 without a login loop", corruptContinuation.Code, corruptContinuation.Header().Get("Location"))
	}

	expiredCookie := issue3095ExpiredCookie(t, store, cfg.SessionName)
	expired := request(http.MethodGet, "/usr?important=yes", nil, expiredCookie)
	assertIssue3095Redirect(t, expired, "/login", "/usr?important=yes")
	assertIssue3095NoStore(t, expired)
	expiredCookie = issue3095Cookie(t, expired, cfg.SessionName)
	assertIssue3095SessionUnauthenticated(t, store, cfg.SessionName, expiredCookie)

	afterExpiryLogin := request(http.MethodPost, "/login", url.Values{
		"task":     {"Login"},
		"username": {"testuser"},
		"password": {"correcthorse"},
		"back":     {"/usr?important=yes"},
	}, expiredCookie)
	if afterExpiryLogin.Code != http.StatusSeeOther || afterExpiryLogin.Header().Get("Location") != "/usr?important=yes" {
		t.Fatalf("post-expiry login response = %d Location %q; want 303 to original URI", afterExpiryLogin.Code, afterExpiryLogin.Header().Get("Location"))
	}
	assertIssue3095NoStore(t, afterExpiryLogin)
	expiredCookie = issue3095Cookie(t, afterExpiryLogin, cfg.SessionName)

	continued := request(http.MethodGet, "/usr?important=yes", nil, expiredCookie)
	if continued.Code != http.StatusOK || continued.Header().Get("Location") != "" {
		t.Fatalf("continued protected request = %d Location %q; want 200 without another login redirect", continued.Code, continued.Header().Get("Location"))
	}
	assertIssue3095NoStore(t, continued)
}

func issue3095Cookie(t *testing.T, rr *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("response did not set %q cookie", name)
	return nil
}

func issue3095ExpiredCookie(t *testing.T, store *sessions.CookieStore, name string) *http.Cookie {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	session, err := store.New(req, name)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	session.Values["UID"] = int32(10)
	session.Values["LoginTime"] = time.Now().Add(-2 * time.Hour).Unix()
	session.Values["ExpiryTime"] = time.Now().Add(-time.Hour).Unix()
	rr := httptest.NewRecorder()
	if err := session.Save(req, rr); err != nil {
		t.Fatalf("save expired session: %v", err)
	}
	return issue3095Cookie(t, rr, name)
}

func assertIssue3095SessionUnauthenticated(t *testing.T, store *sessions.CookieStore, name string, cookie *http.Cookie) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookie)
	session, err := store.Get(req, name)
	if err != nil {
		t.Fatalf("decode transitioned session: %v", err)
	}
	for _, key := range []string{"UID", "LoginTime", "ExpiryTime"} {
		if _, ok := session.Values[key]; ok {
			t.Errorf("transitioned session still contains %q", key)
		}
	}
}

func assertIssue3095Redirect(t *testing.T, rr *httptest.ResponseRecorder, path, back string) {
	t.Helper()
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d; want 303", rr.Code)
	}
	location, err := url.Parse(rr.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse redirect: %v", err)
	}
	if location.Path != path || location.Query().Get("back") != back {
		t.Errorf("redirect = %q; want %s with back=%q", location.String(), path, back)
	}
}

func assertIssue3095NoStore(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()
	if got := rr.Header().Get("Cache-Control"); got != "no-cache, no-store, must-revalidate" {
		t.Errorf("Cache-Control = %q; want no-cache, no-store, must-revalidate", got)
	}
	if got := rr.Header().Get("Cloudflare-CDN-Cache-Control"); got != "no-store" {
		t.Errorf("Cloudflare-CDN-Cache-Control = %q; want no-store", got)
	}
}
