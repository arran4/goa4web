package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/gorilla/sessions"
)

func signBackURL(key, u string, ts int64) string {
	return sign.Sign("back:"+u, key, sign.WithOutNonce())
}

type loginAttemptRecord struct {
	username  string
	ip        string
	createdAt time.Time
}

type loginQuerierFake struct {
	db.Querier

	mu                sync.Mutex
	users             map[string]*db.SystemGetLoginRow
	loginAttempts     []loginAttemptRecord
	passwordResets    map[int32][]db.PendingPassword
	loginRoleByUserID map[int32]bool
	insertedPasswords []db.InsertPasswordParams
	now               func() time.Time
}

func newLoginQuerierFake() *loginQuerierFake {
	return &loginQuerierFake{
		users:             map[string]*db.SystemGetLoginRow{},
		passwordResets:    map[int32][]db.PendingPassword{},
		loginRoleByUserID: map[int32]bool{},
		now:               time.Now,
	}
}

func (f *loginQuerierFake) SystemCountRecentLoginAttempts(_ context.Context, arg db.SystemCountRecentLoginAttemptsParams) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var count int64
	for _, attempt := range f.loginAttempts {
		if attempt.createdAt.After(arg.CreatedAt) && (attempt.username == arg.Username || attempt.ip == arg.IpAddress) {
			count++
		}
	}
	return count, nil
}

func (f *loginQuerierFake) SystemInsertLoginAttempt(_ context.Context, arg db.SystemInsertLoginAttemptParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.loginAttempts = append(f.loginAttempts, loginAttemptRecord{username: arg.Username, ip: arg.IpAddress, createdAt: f.now()})
	return nil
}

func (f *loginQuerierFake) SystemGetLogin(_ context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	user, ok := f.users[username.String]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return user, nil
}

func (f *loginQuerierFake) GetPasswordResetByUser(_ context.Context, arg db.GetPasswordResetByUserParams) (*db.PendingPassword, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	resets := f.passwordResets[arg.UserID]
	var newest *db.PendingPassword
	for i := range resets {
		reset := resets[i]
		if reset.CreatedAt.After(arg.CreatedAt) {
			if newest == nil || reset.CreatedAt.After(newest.CreatedAt) {
				cp := reset
				newest = &cp
			}
		}
	}
	if newest == nil {
		return nil, sql.ErrNoRows
	}
	return newest, nil
}

func (f *loginQuerierFake) SystemCheckRoleGrant(context.Context, db.SystemCheckRoleGrantParams) (int32, error) {
	return 0, sql.ErrNoRows
}

func (f *loginQuerierFake) GetLoginRoleForUser(_ context.Context, usersIdusers int32) (int32, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.loginRoleByUserID[usersIdusers] {
		return 1, nil
	}
	return 0, sql.ErrNoRows
}

func (f *loginQuerierFake) InsertPassword(_ context.Context, arg db.InsertPasswordParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.insertedPasswords = append(f.insertedPasswords, arg)
	for username, user := range f.users {
		if user.Idusers == arg.UsersIdusers {
			cp := *user
			cp.Passwd = sql.NullString{String: arg.Passwd, Valid: true}
			cp.PasswdAlgorithm = arg.PasswdAlgorithm
			f.users[username] = &cp
			break
		}
	}
	return nil
}

func TestLoginAction_NoSuchUser(t *testing.T) {
	queries := newLoginQuerierFake()

	form := url.Values{"username": {"bob"}, "password": {"pw"}}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "1.2.3.4:1111"
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(loginTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if len(queries.loginAttempts) != 1 {
		t.Fatalf("expected login attempt recorded, got %d", len(queries.loginAttempts))
	}
	if queries.loginAttempts[0].username != "bob" || queries.loginAttempts[0].ip != "1.2.3.4" {
		t.Fatalf("unexpected login attempt record: %+v", queries.loginAttempts[0])
	}
	body := rr.Body.String()
	if !strings.Contains(body, "No such user") {
		t.Fatalf("body=%q", body)
	}
}

func TestLoginAction_InvalidPassword(t *testing.T) {
	queries := newLoginQuerierFake()
	queries.users["bob"] = &db.SystemGetLoginRow{
		Idusers:         1,
		Passwd:          sql.NullString{String: "7c4f29407893c334a6cb7a87bf045c0d", Valid: true},
		PasswdAlgorithm: sql.NullString{String: "md5", Valid: true},
		Username:        sql.NullString{String: "bob", Valid: true},
	}

	form := url.Values{"username": {"bob"}, "password": {"wrong"}}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "1.2.3.4:1111"
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(loginTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if len(queries.loginAttempts) != 1 {
		t.Fatalf("expected login attempt recorded, got %d", len(queries.loginAttempts))
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Invalid password") {
		t.Fatalf("body=%q", body)
	}
}

func TestLoginAction_InvalidPasswordPreservesBackData(t *testing.T) {
	queries := newLoginQuerierFake()
	queries.users["bob"] = &db.SystemGetLoginRow{
		Idusers:         1,
		Passwd:          sql.NullString{String: "7c4f29407893c334a6cb7a87bf045c0d", Valid: true},
		PasswdAlgorithm: sql.NullString{String: "md5", Valid: true},
		Username:        sql.NullString{String: "bob", Valid: true},
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
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(loginTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if len(queries.loginAttempts) != 1 {
		t.Fatalf("expected login attempt recorded, got %d", len(queries.loginAttempts))
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
}

func TestLoginPageHiddenFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/login?code=abc&back=%2Ffoo&method=POST&data=x", nil)
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
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
}

func TestLoginFormHandler_ActionTarget(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/register", nil)
	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	loginFormHandler{msg: "approval is pending"}.ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "<form method=\"post\" action=\"/login\">") {
		t.Fatalf("expected login form to post to /login: %q", body)
	}
}

func TestLoginAction_PendingResetPrompt(t *testing.T) {
	q := newLoginQuerierFake()
	pwHash, alg, _ := HashPassword("newpw")
	q.users["bob"] = &db.SystemGetLoginRow{
		Idusers:         1,
		Passwd:          sql.NullString{String: "oldhash", Valid: true},
		PasswdAlgorithm: sql.NullString{String: "md5", Valid: true},
		Username:        sql.NullString{String: "bob", Valid: true},
	}
	q.passwordResets[1] = []db.PendingPassword{{
		ID:               2,
		UserID:           1,
		Passwd:           sql.NullString{String: pwHash, Valid: true},
		PasswdAlgorithm:  sql.NullString{String: alg, Valid: true},
		VerificationCode: "code",
		CreatedAt:        time.Now(),
	}}

	form := url.Values{"username": {"bob"}, "password": {"newpw"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "1.2.3.4:1111"

	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlers.TaskHandler(loginTask)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if len(q.loginAttempts) != 0 {
		t.Fatalf("unexpected login attempts recorded: %v", q.loginAttempts)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "name=\"id\" value=\"2\"") {
		t.Fatalf("missing id field: %q", body)
	}
}

func TestSanitizeBackURL(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "example.com"
	cfg := config.NewRuntimeConfig()
	cfg.LoginAttemptThreshold = 10
	cfg.HTTPHostname = ""
	cd := common.NewCoreData(req.Context(), db.New(nil), cfg)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if got := cd.SanitizeBackURL(req, "/foo"); got != "/foo" {
		t.Fatalf("relative got %q", got)
	}
	if got := cd.SanitizeBackURL(req, "https://example.com/bar?x=1"); got != "/bar?x=1" {
		t.Fatalf("host match got %q", got)
	}
	if got := cd.SanitizeBackURL(req, "https://evil.com/"); got != "" {
		t.Fatalf("evil got %q", got)
	}

	cfg.HTTPHostname = "https://example.com"
	cd = common.NewCoreData(req.Context(), db.New(nil), cfg)
	ctx = context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	if got := cd.SanitizeBackURL(req, "https://example.com/baz"); got != "/baz" {
		t.Fatalf("cfg host got %q", got)
	}
}

func TestSanitizeBackURLSigned(t *testing.T) {
	raw := "https://evil.com/x"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "example.com"
	cfg := config.NewRuntimeConfig()
	cfg.LoginAttemptThreshold = 10
	key := "k"
	cd := common.NewCoreData(req.Context(), db.New(nil), config.NewRuntimeConfig(), common.WithImageSignKey(key))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	ts := time.Now().Add(time.Hour).Unix()
	sig := signBackURL("k", raw, ts)
	q := req.URL.Query()
	q.Set("back_ts", fmt.Sprint(ts))
	q.Set("back_sig", sig)
	req.URL.RawQuery = q.Encode()
	if got := cd.SanitizeBackURL(req, raw); got != raw {
		t.Fatalf("signed got %q", got)
	}
}

func TestLoginPageInvalidBackURL(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/login?back=https://evil.com/x", nil)
	req.Host = "example.com"
	cd := common.NewCoreData(req.Context(), newLoginQuerierFake(), config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	loginTask.Page(rr, req)

	body := rr.Body.String()
	if strings.Contains(body, "name=\"back\"") {
		t.Fatalf("back field present: %q", body)
	}
}

func TestLoginPageSignedBackURL(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.LoginAttemptThreshold = 10
	raw := "https://evil.com/x"
	ts := time.Now().Add(time.Hour).Unix()
	sig := signBackURL("k", raw, ts)
	req := httptest.NewRequest(http.MethodGet, "/login?back="+url.QueryEscape(raw)+"&back_ts="+fmt.Sprint(ts)+"&back_sig="+sig, nil)
	req.Host = "example.com"
	key := "k"
	cd := common.NewCoreData(req.Context(), newLoginQuerierFake(), cfg, common.WithImageSignKey(key))
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
}

func TestLoginAction_ExternalBackURLIgnored(t *testing.T) {
	q := newLoginQuerierFake()
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"
	pwHash, alg, _ := HashPassword("pw")
	q.users["bob"] = &db.SystemGetLoginRow{
		Idusers:         1,
		Passwd:          sql.NullString{String: pwHash, Valid: true},
		PasswdAlgorithm: sql.NullString{String: alg, Valid: true},
		Username:        sql.NullString{String: "bob", Valid: true},
	}
	q.loginRoleByUserID[1] = true

	form := url.Values{"username": {"bob"}, "password": {"pw"}, "back": {"https://evil.com"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "1.2.3.4:1111"
	req.Host = "example.com"

	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
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
}

func TestLoginAction_SignedExternalBackURL(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.LoginAttemptThreshold = 10
	q := newLoginQuerierFake()
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"
	pwHash, alg, _ := HashPassword("pw")
	q.users["bob"] = &db.SystemGetLoginRow{
		Idusers:         1,
		Passwd:          sql.NullString{String: pwHash, Valid: true},
		PasswdAlgorithm: sql.NullString{String: alg, Valid: true},
		Username:        sql.NullString{String: "bob", Valid: true},
	}
	q.loginRoleByUserID[1] = true

	raw := "https://example.org/ok"
	ts := time.Now().Add(time.Hour).Unix()
	sig := signBackURL("k", raw, ts)
	key := "k"
	form := url.Values{"username": {"bob"}, "password": {"pw"}, "back": {raw}, "back_ts": {fmt.Sprint(ts)}, "back_sig": {sig}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "1.2.3.4:1111"
	req.Host = "example.com"

	cd := common.NewCoreData(req.Context(), q, cfg, common.WithImageSignKey(key))
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
}

func TestLoginAction_Throttle(t *testing.T) {
	q := newLoginQuerierFake()
	q.loginAttempts = append(q.loginAttempts,
		loginAttemptRecord{username: "bob", ip: "1.2.3.4", createdAt: time.Now()},
		loginAttemptRecord{username: "bob", ip: "1.2.3.4", createdAt: time.Now()},
		loginAttemptRecord{username: "bob", ip: "1.2.3.4", createdAt: time.Now()},
		loginAttemptRecord{username: "bob", ip: "1.2.3.4", createdAt: time.Now()},
		loginAttemptRecord{username: "bob", ip: "1.2.3.4", createdAt: time.Now()},
	)

	cfg := config.NewRuntimeConfig()
	cfg.LoginAttemptThreshold = 3
	cfg.LoginAttemptWindow = 15

	form := url.Values{"username": {"bob"}, "password": {"pw"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "1.2.3.4:1111"
	cd := common.NewCoreData(req.Context(), q, cfg)
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
}

func TestRedirectBackPageHandlerGET(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
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
}

func TestRedirectBackPageHandlerEmptyMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
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
}

func TestRedirectBackPageHandler(t *testing.T) {
	cases := map[string]string{"empty": "", "get": http.MethodGet}
	for name, method := range cases {
		t.Run(name, func(t *testing.T) {
			cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			h := redirectBackPageHandler{BackURL: "/back", Method: method, Values: url.Values{"a": {"b"}}}
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
	}

	t.Run("post", func(t *testing.T) {
		cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
		req := httptest.NewRequest(http.MethodGet, "/", nil)
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
