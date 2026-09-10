package auth

import (
	"context"
	"database/sql"
	"errors"
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

// Tests transition matrices for #3095.

func TestIssue3095_LoginRedirectOriginalDestination(t *testing.T) {
	form := url.Values{}
	form.Set("username", "testuser")
	form.Set("password", "correcthorse")
	form.Set("back", "/protected-page")

	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	core.SessionName = "test_session"
	core.Store = sessions.NewCookieStore([]byte("secret"))
	session, _ := core.Store.New(req, core.SessionName)

	q := testhelpers.NewQuerierStub()
	hash, alg, _ := HashPassword("correcthorse")
	q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
		return &db.SystemGetLoginRow{
			Idusers:         10,
			Passwd:          sql.NullString{String: hash, Valid: true},
			PasswdAlgorithm: sql.NullString{String: alg, Valid: true},
		}, nil
	}
	q.GetLoginRoleForUserFn = func(ctx context.Context, id int32) (int32, error) {
		return 1, nil
	}

	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), q, cfg, common.WithSession(session))
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rr := httptest.NewRecorder()
	result := loginTask.Action(rr, req)

	handler, ok := result.(http.HandlerFunc)
	if !ok {
		t.Fatalf("Expected http.HandlerFunc from LoginTask.Action, got %T", result)
	}

	handler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303 See Other, got %d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/protected-page" {
		t.Errorf("Expected redirect to /protected-page, got %q", loc)
	}

	// Cache headers verification
	if rr.Header().Get("Cache-Control") != "no-cache, no-store, must-revalidate" {
		t.Errorf("Missing expected Cache-Control header for auth change")
	}
	if rr.Header().Get("Cloudflare-CDN-Cache-Control") != "no-store" {
		t.Errorf("Missing expected Cloudflare-CDN-Cache-Control header")
	}
}

func TestIssue3095_AccountSwitchingSuccess(t *testing.T) {
	form := url.Values{}
	form.Set("username", "userB")
	form.Set("password", "correcthorse")

	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	core.SessionName = "test_session"
	core.Store = sessions.NewCookieStore([]byte("secret"))
	session, _ := core.Store.New(req, core.SessionName)

	// Pre-populate session as User A (ID = 5)
	session.Values["UID"] = int32(5)
	session.Values["LoginTime"] = time.Now().Unix()
	session.Values["ExpiryTime"] = time.Now().Add(time.Hour).Unix()
	session.Values["user_a_sensitive_state"] = "should-be-gone"

	q := testhelpers.NewQuerierStub()
	hash, alg, _ := HashPassword("correcthorse")
	q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
		return &db.SystemGetLoginRow{
			Idusers:         20, // User B
			Passwd:          sql.NullString{String: hash, Valid: true},
			PasswdAlgorithm: sql.NullString{String: alg, Valid: true},
		}, nil
	}
	q.GetLoginRoleForUserFn = func(ctx context.Context, id int32) (int32, error) {
		return 1, nil
	}

	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), q, cfg, common.WithSession(session))
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rr := httptest.NewRecorder()
	result := loginTask.Action(rr, req)

	handler, ok := result.(http.HandlerFunc)
	if !ok {
		t.Fatalf("Expected http.HandlerFunc, got %T", result)
	}
	handler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303, got %d", rr.Code)
	}

	// Check session is updated to User B
	if uid, ok := session.Values["UID"].(int32); !ok || uid != 20 {
		t.Errorf("Expected session UID to be 20, got %v", session.Values["UID"])
	}

	// Check user A state is cleared
	if _, ok := session.Values["user_a_sensitive_state"]; ok {
		t.Errorf("Expected user_a_sensitive_state to be removed from session on account switch")
	}
}

func TestIssue3095_AccountSwitchingFailure(t *testing.T) {
	form := url.Values{}
	form.Set("username", "userB")
	form.Set("password", "wrong")

	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	core.SessionName = "test_session"
	core.Store = sessions.NewCookieStore([]byte("secret"))
	session, _ := core.Store.New(req, core.SessionName)

	// Pre-populate session as User A (ID = 5)
	session.Values["UID"] = int32(5)
	session.Values["LoginTime"] = int64(100)
	session.Values["ExpiryTime"] = int64(200)
	session.Values["user_a_sensitive_state"] = "should-stay"

	q := testhelpers.NewQuerierStub()
	hash, alg, _ := HashPassword("correcthorse")
	q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
		return &db.SystemGetLoginRow{
			Idusers:         20,
			Passwd:          sql.NullString{String: hash, Valid: true},
			PasswdAlgorithm: sql.NullString{String: alg, Valid: true},
		}, nil
	}
	q.SystemInsertLoginAttemptFn = func(ctx context.Context, params db.SystemInsertLoginAttemptParams) error {
		return nil
	}
	q.GetPasswordResetByUserFn = func(ctx context.Context, arg db.GetPasswordResetByUserParams) (*db.PendingPassword, error) {
		return nil, sql.ErrNoRows
	}

	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), q, cfg, common.WithSession(session))
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rr := httptest.NewRecorder()
	result := loginTask.Action(rr, req)

	// Verify error handler returned
	h, ok := result.(http.Handler)
	if !ok {
		t.Fatalf("Expected http.Handler from failed login, got %T", result)
	}

	// Actually serve it to see if headers are applied for failed anon/auth login
	h.ServeHTTP(rr, req)

	if rr.Header().Get("Cache-Control") != "no-cache, no-store, must-revalidate" {
		t.Errorf("Missing expected Cache-Control header for failed auth change")
	}

	if !strings.Contains(rr.Body.String(), "Invalid username or password. You remain logged in as your current account.") {
		t.Errorf("Expected rendered error to state that A remains logged in, got %q", rr.Body.String())
	}

	// Check session A is untouched
	if uid, ok := session.Values["UID"].(int32); !ok || uid != 5 {
		t.Errorf("Expected session UID to remain 5, got %v", session.Values["UID"])
	}
	if session.Values["user_a_sensitive_state"] != "should-stay" {
		t.Errorf("Expected A state to persist, got %v", session.Values["user_a_sensitive_state"])
	}
}

func TestIssue3095_AccountSwitchingNoRolePending(t *testing.T) {
	form := url.Values{}
	form.Set("username", "userB")
	form.Set("password", "correcthorse")

	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	core.SessionName = "test_session"
	core.Store = sessions.NewCookieStore([]byte("secret"))
	session, _ := core.Store.New(req, core.SessionName)

	// Pre-populate session as User A (ID = 5)
	session.Values["UID"] = int32(5)
	session.Values["LoginTime"] = time.Now().Unix()

	q := testhelpers.NewQuerierStub()
	hash, alg, _ := HashPassword("correcthorse")
	q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
		return &db.SystemGetLoginRow{
			Idusers:         20,
			Passwd:          sql.NullString{String: hash, Valid: true},
			PasswdAlgorithm: sql.NullString{String: alg, Valid: true},
		}, nil
	}
	// Simulate no role / pending approval
	q.GetLoginRoleForUserFn = func(ctx context.Context, id int32) (int32, error) {
		return 0, sql.ErrNoRows
	}

	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), q, cfg, common.WithSession(session))
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rr := httptest.NewRecorder()
	result := loginTask.Action(rr, req)

	h, ok := result.(http.Handler)
	if !ok {
		t.Fatalf("Expected http.Handler from pending login, got %T", result)
	}
	h.ServeHTTP(rr, req)
	if !strings.Contains(rr.Body.String(), "Approval is pending. You remain logged in as your current account.") {
		t.Errorf("Expected rendered pending message to state that A remains logged in, got %q", rr.Body.String())
	}

	if uid, ok := session.Values["UID"].(int32); !ok || uid != 5 {
		t.Errorf("Expected session UID to remain 5, got %v", session.Values["UID"])
	}
}

func TestIssue3095_AccountSwitchingThrottled(t *testing.T) {
	form := url.Values{}
	form.Set("username", "userB")
	form.Set("password", "wrong")

	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	core.SessionName = "test_session"
	core.Store = sessions.NewCookieStore([]byte("secret"))
	session, _ := core.Store.New(req, core.SessionName)

	// Pre-populate session as User A (ID = 5)
	session.Values["UID"] = int32(5)

	q := testhelpers.NewQuerierStub()
	// Mock SystemCountRecentLoginAttempts to return threshold
	q.SystemCountRecentLoginAttemptsReturns = 5

	cfg := config.NewRuntimeConfig()
	cfg.LoginAttemptThreshold = 3
	cfg.LoginAttemptWindow = 15

	cd := common.NewCoreData(req.Context(), q, cfg, common.WithSession(session))
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rr := httptest.NewRecorder()
	result := loginTask.Action(rr, req)

	h, ok := result.(http.Handler)
	if !ok {
		t.Fatalf("Expected http.Handler from throttled login, got %T", result)
	}
	h.ServeHTTP(rr, req)
	if !strings.Contains(rr.Body.String(), "Too many failed attempts. You remain logged in as your current account.") {
		t.Errorf("Expected rendered throttling message to state that A remains logged in, got %q", rr.Body.String())
	}
}

func TestIssue3095_SessionErrorRedirect_CorruptSession(t *testing.T) {
	req := httptest.NewRequest("GET", "/protected-page?foo=bar", nil)
	rr := httptest.NewRecorder()

	core.SessionName = "test_session"
	core.Store = sessions.NewCookieStore([]byte("secret"))

	core.SessionErrorRedirect(rr, req, errors.New("corrupt session signature"))

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303 See Other, got %d", rr.Code)
	}

	loc := rr.Header().Get("Location")
	if !strings.HasPrefix(loc, "/login?back=") {
		t.Errorf("Expected redirect to login, got %q", loc)
	}

	parsed, _ := url.Parse(loc)
	back := parsed.Query().Get("back")
	if back != "/protected-page?foo=bar" {
		t.Errorf("Expected back param to be /protected-page?foo=bar, got %q", back)
	}

	sc := rr.Header().Get("Set-Cookie")
	if !strings.Contains(sc, "Max-Age=-1") && !strings.Contains(sc, "Max-Age=0") && !strings.Contains(sc, "Expires=") {
		t.Errorf("Expected cleared cookie for corrupt session, got %q", sc)
	}

	if rr.Header().Get("Cache-Control") != "no-cache, no-store, must-revalidate" {
		t.Errorf("Missing expected Cache-Control header for session error redirect")
	}
}

func TestIssue3095_OpenRedirectRejection(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(context.Background(), q, cfg)

	req := httptest.NewRequest("GET", "/", nil)

	// Test external URL sanitization
	safeUrl, _ := cd.SanitizeBackURL(req, "https://evil.example.com/steal-creds")
	if safeUrl != "/" && safeUrl != "" {
		t.Errorf("Expected external URL to be sanitized to /, got %q", safeUrl)
	}

	// Test protocol-relative sanitization
	safeUrl2, _ := cd.SanitizeBackURL(req, "//evil.example.com/steal-creds")
	if safeUrl2 != "/" && safeUrl2 != "" {
		t.Errorf("Expected protocol-relative URL to be sanitized to /, got %q", safeUrl2)
	}

	// Test valid local URL
	safeUrl3, _ := cd.SanitizeBackURL(req, "/my-account/settings?page=2")
	if safeUrl3 != "/my-account/settings?page=2" {
		t.Errorf("Expected local URL to be preserved, got %q", safeUrl3)
	}
}
