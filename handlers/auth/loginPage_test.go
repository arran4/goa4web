package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/sessions"
)

func TestLoginTask_Action(t *testing.T) {
	t.Run("Happy Path - Pending Reset Prompt", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		pwHash, alg, _ := HashPassword("newpw")
		q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
			return &db.SystemGetLoginRow{
				Idusers:         1,
				Passwd:          sql.NullString{String: "oldhash", Valid: true},
				PasswdAlgorithm: sql.NullString{String: "md5", Valid: true},
				Username:        sql.NullString{String: "bob", Valid: true},
			}, nil
		}
		q.GetPasswordResetByUserReturns = &db.PendingPassword{
			ID:               2,
			UserID:           1,
			Passwd:           pwHash,
			PasswdAlgorithm:  alg,
			VerificationCode: "code",
			CreatedAt:        time.Now(),
		}

		form := url.Values{"username": {"bob"}, "password": {"newpw"}}
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "1.2.3.4:1111"

		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)

		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithSession(session))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(loginTask)(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		// If SystemInsertLoginAttemptCalls is missing, this will fail compilation.
		// If it is present, it should work.
		if len(q.SystemInsertLoginAttemptCalls) != 0 {
			t.Fatalf("unexpected login attempts recorded: %v", q.SystemInsertLoginAttemptCalls)
		}
		body := rr.Body.String()
		if !strings.Contains(body, "name=\"id\" value=\"2\"") {
			t.Fatalf("missing id field: %q", body)
		}
	})

	t.Run("Happy Path - Signed External Back URL", func(t *testing.T) {
		cfg := config.NewRuntimeConfig()
		cfg.LoginAttemptThreshold = 10
		q := testhelpers.NewQuerierStub()
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		pwHash, alg, _ := HashPassword("pw")
		q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
			return &db.SystemGetLoginRow{
				Idusers:         1,
				Passwd:          sql.NullString{String: pwHash, Valid: true},
				PasswdAlgorithm: sql.NullString{String: alg, Valid: true},
				Username:        sql.NullString{String: "bob", Valid: true},
			}, nil
		}
		q.GetLoginRoleForUserReturns = 1

		raw := "https://example.org/ok"
		ts := time.Now().Add(time.Hour).Unix()
		sig := SignBackURL("k", raw, ts)
		key := "k"
		form := url.Values{"username": {"bob"}, "password": {"pw"}, "back": {raw}, "back_ts": {fmt.Sprint(ts)}, "back_sig": {sig}}
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "1.2.3.4:1111"
		req.Host = "example.com"

		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(req.Context(), q, cfg, common.WithImageSignKey(key), common.WithSession(session))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(loginTask)(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		if cd.AutoRefresh == "" || !strings.Contains(cd.AutoRefresh, "url="+raw) {
			t.Fatalf("auto refresh=%q", cd.AutoRefresh)
		}
	})

	t.Run("Unhappy Path - No Such User", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		q.SystemGetLoginErr = sql.ErrNoRows

		form := url.Values{"username": {"bob"}, "password": {"pw"}}
		req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "1.2.3.4:1111"
		ctx := req.Context()
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithSession(session))
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(loginTask)(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		if len(q.SystemInsertLoginAttemptCalls) != 1 {
			t.Fatalf("expected login attempt recorded, got %d", len(q.SystemInsertLoginAttemptCalls))
		}
		body := rr.Body.String()
		if !strings.Contains(body, "No such user") {
			t.Fatalf("body=%q", body)
		}
	})

	t.Run("Unhappy Path - Invalid Password", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		q.GetPasswordResetByUserErr = sql.ErrNoRows
		q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
			return &db.SystemGetLoginRow{
				Idusers:         1,
				Passwd:          sql.NullString{String: "7c4f29407893c334a6cb7a87bf045c0d", Valid: true},
				PasswdAlgorithm: sql.NullString{String: "md5", Valid: true},
				Username:        sql.NullString{String: "bob", Valid: true},
			}, nil
		}

		form := url.Values{"username": {"bob"}, "password": {"wrong"}}
		req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "1.2.3.4:1111"
		ctx := req.Context()
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithSession(session))
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(loginTask)(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		if len(q.SystemInsertLoginAttemptCalls) != 1 {
			t.Fatalf("expected login attempt recorded, got %d", len(q.SystemInsertLoginAttemptCalls))
		}
		body := rr.Body.String()
		if !strings.Contains(body, "Invalid password") {
			t.Fatalf("body=%q", body)
		}
	})

	t.Run("Unhappy Path - Invalid Password Preserves Back Data", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		q.GetPasswordResetByUserErr = sql.ErrNoRows
		q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
			return &db.SystemGetLoginRow{
				Idusers:         1,
				Passwd:          sql.NullString{String: "7c4f29407893c334a6cb7a87bf045c0d", Valid: true},
				PasswdAlgorithm: sql.NullString{String: "md5", Valid: true},
				Username:        sql.NullString{String: "bob", Valid: true},
			}, nil
		}

		form := url.Values{
			"username": {"bob"},
			"password": {"wrong"},
			"back":     {"/target"},
			"method":   {http.MethodPost},
			"data":     {"a=1&b=2"},
		}
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "1.2.3.4:1111"
		ctx := req.Context()
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithSession(session))
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(loginTask)(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		if len(q.SystemInsertLoginAttemptCalls) != 1 {
			t.Fatalf("expected login attempt recorded, got %d", len(q.SystemInsertLoginAttemptCalls))
		}
		body := rr.Body.String()
		if !strings.Contains(body, "Invalid password") {
			t.Fatalf("body=%q", body)
		}
		if !strings.Contains(body, "name=\"back\" value=\"/target\"") {
			t.Fatalf("missing back field: %q", body)
		}
		if !strings.Contains(body, "name=\"method\" value=\"POST\"") {
			t.Fatalf("missing method field: %q", body)
		}
		if !strings.Contains(body, "name=\"data\" value=\"a=1&amp;b=2\"") {
			t.Fatalf("missing data field: %q", body)
		}
	})

	t.Run("Unhappy Path - External Back URL Ignored", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		pwHash, alg, _ := HashPassword("pw")
		q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
			return &db.SystemGetLoginRow{
				Idusers:         1,
				Passwd:          sql.NullString{String: pwHash, Valid: true},
				PasswdAlgorithm: sql.NullString{String: alg, Valid: true},
				Username:        sql.NullString{String: "bob", Valid: true},
			}, nil
		}
		q.GetLoginRoleForUserReturns = 1

		form := url.Values{"username": {"bob"}, "password": {"pw"}, "back": {"https://evil.com"}}
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "1.2.3.4:1111"
		req.Host = "example.com"

		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithSession(session))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(loginTask)(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		body := rr.Body.String()
		if !strings.Contains(body, "url=/") {
			t.Fatalf("missing refresh to root: %q", body)
		}
	})

	t.Run("Unhappy Path - Throttle", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		// Mock SystemCountRecentLoginAttempts to return threshold
		q.SystemCountRecentLoginAttemptsReturns = 5

		cfg := config.NewRuntimeConfig()
		cfg.LoginAttemptThreshold = 3
		cfg.LoginAttemptWindow = 15

		form := url.Values{"username": {"bob"}, "password": {"pw"}}
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "1.2.3.4:1111"
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(req.Context(), q, cfg, common.WithSession(session))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		handlers.TaskHandler(loginTask)(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Too many failed attempts") {
			t.Fatalf("body=%q", rr.Body.String())
		}
	})
}

func TestLoginTask_Page(t *testing.T) {
	t.Run("Happy Path - Hidden Fields", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/login?code=abc&back=%2Ffoo&method=POST&data=x", nil)
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}), common.WithSession(session))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		loginTask.Page(rr, req)

		body := rr.Body.String()
		if !strings.Contains(body, "name=\"code\" value=\"abc\"") {
			t.Fatalf("missing code field: %q", body)
		}
		if !strings.Contains(body, "name=\"back\" value=\"/foo\"") {
			t.Fatalf("missing back field: %q", body)
		}
		if !strings.Contains(body, "name=\"method\" value=\"POST\"") {
			t.Fatalf("missing method field: %q", body)
		}
		if !strings.Contains(body, "name=\"data\" value=\"x\"") {
			t.Fatalf("missing data field: %q", body)
		}
		if strings.Contains(body, "back_sig") || strings.Contains(body, "back_ts") {
			t.Fatalf("unexpected signature fields: %q", body)
		}
	})

	t.Run("Happy Path - Signed Back URL", func(t *testing.T) {
		cfg := config.NewRuntimeConfig()
		cfg.LoginAttemptThreshold = 10
		raw := "https://evil.com/x"
		ts := time.Now().Add(time.Hour).Unix()
		sig := SignBackURL("k", raw, ts)
		req := httptest.NewRequest(http.MethodGet, "/login?back="+url.QueryEscape(raw)+"&back_ts="+fmt.Sprint(ts)+"&back_sig="+sig, nil)
		req.Host = "example.com"
		key := "k"
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(req.Context(), testhelpers.NewQuerierStub(), cfg, common.WithImageSignKey(key), common.WithSession(session))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		loginTask.Page(rr, req)

		body := rr.Body.String()
		if !strings.Contains(body, "name=\"back\" value=\""+raw+"\"") {
			t.Fatalf("missing back field: %q", body)
		}
		if !strings.Contains(body, "name=\"back_sig\" value=") {
			t.Fatalf("missing back_sig field: %q", body)
		}
		if !strings.Contains(body, "name=\"back_ts\" value=") {
			t.Fatalf("missing back_ts field: %q", body)
		}
	})

	t.Run("Unhappy Path - Invalid Back URL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/login?back=https://evil.com/x", nil)
		req.Host = "example.com"
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(req.Context(), testhelpers.NewQuerierStub(), config.NewRuntimeConfig(), common.WithSession(session))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		loginTask.Page(rr, req)

		body := rr.Body.String()
		if strings.Contains(body, "name=\"back\"") {
			t.Fatalf("back field present: %q", body)
		}
	})
}

func TestHappyPathLoginFormHandler_ActionTarget(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/register", nil)
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"
	session, _ := store.New(req, core.SessionName)
	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}), common.WithSession(session))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	loginFormHandler{msg: "approval is pending"}.ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "<form method=\"post\" action=\"/login\">") {
		t.Fatalf("expected login form to post to /login: %q", body)
	}
}

func TestHappyPathSanitizeBackURL(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "example.com"
	cfg := config.NewRuntimeConfig()
	cfg.LoginAttemptThreshold = 10
	cfg.BaseURL = ""
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"
	session, _ := store.New(req, core.SessionName)
	cd := common.NewCoreData(req.Context(), db.New(nil), cfg, common.WithSession(session))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if got, _ := cd.SanitizeBackURL(req, "/foo"); got != "/foo" {
		t.Fatalf("relative got %q", got)
	}
	if got, _ := cd.SanitizeBackURL(req, "https://example.com/bar?x=1"); got != "/bar?x=1" {
		t.Fatalf("host match got %q", got)
	}
	if got, _ := cd.SanitizeBackURL(req, "https://evil.com/"); got != "" {
		t.Fatalf("evil got %q", got)
	}

	cfg.BaseURL = "https://example.com"
	cd = common.NewCoreData(req.Context(), db.New(nil), cfg, common.WithSession(session))
	ctx = context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	if got, _ := cd.SanitizeBackURL(req, "https://example.com/baz"); got != "/baz" {
		t.Fatalf("cfg host got %q", got)
	}
}

func TestHappyPathSanitizeBackURLSigned(t *testing.T) {
	raw := "https://evil.com/x"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "example.com"
	cfg := config.NewRuntimeConfig()
	cfg.LoginAttemptThreshold = 10
	key := "k"
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"
	session, _ := store.New(req, core.SessionName)
	cd := common.NewCoreData(req.Context(), db.New(nil), config.NewRuntimeConfig(), common.WithImageSignKey(key), common.WithSession(session))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	ts := time.Now().Add(time.Hour).Unix()
	sig := SignBackURL("k", raw, ts)
	q := req.URL.Query()
	q.Set("back_ts", fmt.Sprint(ts))
	q.Set("back_sig", sig)
	req.URL.RawQuery = q.Encode()
	if got, _ := cd.SanitizeBackURL(req, raw); got != raw {
		t.Fatalf("signed got %q", got)
	}
}

func TestRedirectBackPageHandler(t *testing.T) {
	t.Run("Happy Path - GET", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}), common.WithSession(session))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h := redirectBackPageHandler{BackURL: "/foo", Method: http.MethodGet, Values: url.Values{"x": {"1"}}}
		h.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		if cd.AutoRefresh == "" || !strings.Contains(cd.AutoRefresh, "url=/foo?x=1") {
			t.Fatalf("auto refresh=%q", cd.AutoRefresh)
		}
	})

	t.Run("Happy Path - Empty Method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}), common.WithSession(session))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h := redirectBackPageHandler{BackURL: "/bar", Method: "", Values: url.Values{"y": {"2"}}}
		h.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		if cd.AutoRefresh == "" || !strings.Contains(cd.AutoRefresh, "url=/bar?y=2") {
			t.Fatalf("auto refresh=%q", cd.AutoRefresh)
		}
	})

	t.Run("Happy Path - Back URL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}), common.WithSession(session))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h := redirectBackPageHandler{BackURL: "/back", Method: http.MethodGet, Values: url.Values{"a": {"b"}}}
		h.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		if cd.AutoRefresh == "" || !strings.Contains(cd.AutoRefresh, "url=/back") {
			t.Fatalf("auto refresh=%q", cd.AutoRefresh)
		}
		if strings.Contains(rr.Body.String(), "<form") {
			t.Fatalf("unexpected form: %q", rr.Body.String())
		}
	})

	t.Run("Happy Path - Post", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}), common.WithSession(session))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h := redirectBackPageHandler{BackURL: "/back", Method: http.MethodPost, Values: url.Values{"a": {"b"}}}
		h.ServeHTTP(rr, req)

		if cd.AutoRefresh != "" {
			t.Fatalf("unexpected refresh: %q", cd.AutoRefresh)
		}
		body := rr.Body.String()
		if !strings.Contains(body, "<form") || !strings.Contains(body, "action=\"/back\"") {
			t.Fatalf("missing form: %q", body)
		}
	})
}
