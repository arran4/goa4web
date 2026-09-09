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

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/middleware"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/mux"
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

	// It's unexported loginFormHandler, check state message directly via string format
	resStr := fmt.Sprintf("%#v", result)
	if !strings.Contains(resStr, "Invalid username or password. You remain logged in as your current account.") {
		t.Errorf("Expected error message indicating A remains logged in, got %v", resStr)
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

	resStr := fmt.Sprintf("%#v", result)
	if !strings.Contains(resStr, "Approval is pending. You remain logged in as your current account.") {
		t.Errorf("Expected pending message indicating A remains logged in, got %v", resStr)
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

	resStr := fmt.Sprintf("%#v", result)
	if !strings.Contains(resStr, "Too many failed attempts. You remain logged in as your current account.") {
		t.Errorf("Expected throttled message indicating A remains logged in, got %v", resStr)
	}
}

func TestIssue3095_SessionErrorRedirect_CorruptSession(t *testing.T) {
	// A request with a corrupt session that returns SessionFetchFail via the TaskHandler layer
	// should hit core.SessionErrorRedirect -> core.RedirectToLogin.

	req := httptest.NewRequest("GET", "/protected-page?foo=bar", nil)
	rr := httptest.NewRecorder()

	core.SessionName = "test_session"
	core.Store = sessions.NewCookieStore([]byte("secret"))

	core.SessionErrorRedirect(rr, req, fmt.Errorf("corrupt session signature"))

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

func TestIssue3095_LoginPageDeterministicAndNoMethodData(t *testing.T) {
	// A user who is already authenticated hitting /login via GET.
	// Since .MatcherFunc(gml.Not(handlers.RequiresAnAccount())) was removed, they get the page.

	req := httptest.NewRequest(http.MethodGet, "/login?code=abc&back=%2Ffoo", nil)

	core.SessionName = "test_session"
	store := sessions.NewCookieStore([]byte("secret"))
	core.Store = store
	session, _ := store.New(req, core.SessionName)
	session.Values["UID"] = int32(5) // User is logged in

	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}), common.WithSession(session))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	loginTask.Page(rr, req)

	body := rr.Body.String()

	// Must have code and back
	if !strings.Contains(body, "name=\"code\" value=\"abc\"") {
		t.Errorf("missing code field: %q", body)
	}
	if !strings.Contains(body, "name=\"back\" value=\"/foo\"") {
		t.Errorf("missing back field: %q", body)
	}

	// Must not have POST replay method or data fields
	if strings.Contains(body, "name=\"method\"") || strings.Contains(body, "name=\"data\"") {
		t.Errorf("unexpected method or data fields in form: %q", body)
	}
}

// Adding integrated matrix tests to cover actual routing, logout, and caching

func TestIssue3095_RouteLevelTransitions(t *testing.T) {
	// Setup mocked database, session store, and full application router for end-to-end tests
	q := testhelpers.NewQuerierStub()
	pwHash, alg, _ := HashPassword("correcthorse")
	q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
		if username.String == "testuser" {
			return &db.SystemGetLoginRow{
				Idusers:         10,
				Passwd:          sql.NullString{String: pwHash, Valid: true},
				PasswdAlgorithm: sql.NullString{String: alg, Valid: true},
				Username:        username,
			}, nil
		}
		return nil, sql.ErrNoRows
	}
	q.GetLoginRoleForUserFn = func(ctx context.Context, id int32) (int32, error) {
		return 1, nil // Approved
	}

	cfg := config.NewRuntimeConfig()
	core.SessionName = "test_session"
	core.Store = sessions.NewCookieStore([]byte("secret"))

	// Create a dummy router that uses the real handlers and middleware.
	// We only need the pieces relevant to auth transitions and caching.

	r := mux.NewRouter()

	// Middleware setup identical to production
	coreDataMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			session, _ := core.GetSession(req)
			cd := common.NewCoreData(req.Context(), q, cfg, common.WithSession(session))
			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
			next.ServeHTTP(w, req)
		})
	}

	sessionContextMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			session, _ := core.Store.Get(req, core.SessionName)
			ctx := context.WithValue(req.Context(), core.ContextValues("session"), session)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}

	r.Use(sessionContextMiddleware)
	r.Use(coreDataMiddleware)

	// Ensure we register the /login routes appropriately for test using the correct paths
	loginRouter := r.PathPrefix("/login").Subrouter()
	loginRouter.HandleFunc("", handlers.WithNoCache(loginTask.Page)).Methods("GET")
	loginRouter.HandleFunc("", handlers.TaskHandler(loginTask)).Methods("POST")

	// A dummy protected route
	r.HandleFunc("/protected", func(w http.ResponseWriter, req *http.Request) {
		cd := req.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if cd.UserID == 0 {
			middleware.RedirectToLogin(w, req, cd.GetSession())
			return
		}
		handlers.DisableCaching(w)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Welcome user %d", cd.UserID)))
	}).Methods("GET")

	// A dummy anonymous public route utilizing TemplateHandler logic (simulated)
	r.HandleFunc("/public", func(w http.ResponseWriter, req *http.Request) {
		// Mock what TemplateHandler does regarding caching
		cd := req.Context().Value(consts.KeyCoreData).(*common.CoreData)
		_, err := req.Cookie(core.SessionName)
		hasCookie := err == nil

		if (cd != nil && cd.UserID != 0) || hasCookie {
			handlers.DisableCaching(w)
		} else {
			w.Header().Set("Cache-Control", "public, max-age=3600")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Public content"))
	}).Methods("GET")

	// A dummy logout route
	r.HandleFunc("/logout", func(w http.ResponseWriter, req *http.Request) {
		session, _ := core.GetSession(req)
		delete(session.Values, "UID")
		session.Save(req, w)
		handlers.DisableCaching(w)
		http.Redirect(w, req, "/", http.StatusSeeOther)
	}).Methods("GET")

	server := httptest.NewServer(r)
	defer server.Close()

	// Use a custom client that doesn't follow redirects automatically so we can inspect them
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	var sessionCookie *http.Cookie

	t.Run("1. Protected GET redirects to Login with back param", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/protected?id=123", nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("Expected 303 See Other, got %d", resp.StatusCode)
		}
		loc := resp.Header.Get("Location")
		if !strings.Contains(loc, "/login") || !strings.Contains(loc, "back=%2Fprotected%3Fid%3D123") {
			t.Errorf("Expected location with encoded back param, got %s", loc)
		}
	})

	t.Run("2. Successful POST /login transition back to original destination", func(t *testing.T) {
		form := url.Values{}
		form.Set("username", "testuser")
		form.Set("password", "correcthorse")
		form.Set("back", "/protected?id=123")

		req, _ := http.NewRequest("POST", server.URL+"/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("Expected 303 See Other on successful login, got %d", resp.StatusCode)
		}
		loc := resp.Header.Get("Location")
		if loc != "/protected?id=123" {
			t.Errorf("Expected redirect back to /protected?id=123, got %q", loc)
		}

		// Ensure auth-changing response has no-store cache headers
		if resp.Header.Get("Cache-Control") != "no-cache, no-store, must-revalidate" {
			t.Errorf("Missing no-store Cache-Control on login success")
		}
		if resp.Header.Get("Cloudflare-CDN-Cache-Control") != "no-store" {
			t.Errorf("Missing Cloudflare no-store on login success")
		}

		// Capture session cookie for subsequent requests
		for _, cookie := range resp.Cookies() {
			if cookie.Name == core.SessionName {
				sessionCookie = cookie
				break
			}
		}
		if sessionCookie == nil {
			t.Fatal("Expected session cookie to be set")
		}
	})

	t.Run("3. GET /login while authenticated is deterministic and hits route", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/login", nil)
		if sessionCookie != nil {
			req.AddCookie(sessionCookie)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		// Should return 200 OK rendering the login page, NOT 404 or fall-through
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200 OK for /login when authenticated, got %d", resp.StatusCode)
		}
	})

	t.Run("4. Logout correctly clears session and redirects", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/logout", nil)
		if sessionCookie != nil {
			req.AddCookie(sessionCookie)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("Expected 303 See Other on logout, got %d", resp.StatusCode)
		}

		// Ensure logout has no-store cache headers
		if resp.Header.Get("Cache-Control") != "no-cache, no-store, must-revalidate" {
			t.Errorf("Missing no-store Cache-Control on logout")
		}

		// Check that the returned cookie invalidates the session
		for _, cookie := range resp.Cookies() {
			if cookie.Name == core.SessionName {
				sessionCookie = cookie // It's now empty/invalidated
			}
		}
	})

	t.Run("5. Protected route rejects access after logout", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/protected", nil)
		if sessionCookie != nil {
			req.AddCookie(sessionCookie) // The invalidated cookie
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("Expected 303 See Other redirecting to login, got %d", resp.StatusCode)
		}
		if !strings.Contains(resp.Header.Get("Location"), "/login") {
			t.Errorf("Expected redirect to /login, got %s", resp.Header.Get("Location"))
		}
	})

	t.Run("6. Anonymous public caching is explicit when no cookie is present", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/public", nil)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
		}

		if resp.Header.Get("Cache-Control") != "public, max-age=3600" {
			t.Errorf("Expected Cache-Control: public, max-age=3600 on anonymous public route, got %q", resp.Header.Get("Cache-Control"))
		}
	})

	t.Run("7. Stale/corrupt cookie triggers no-store on public routes", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/public", nil)
		// Add the invalidated/stale session cookie
		if sessionCookie != nil {
			req.AddCookie(sessionCookie)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
		}

		// Even though UID is 0 (anonymous logic), the mere presence of the cookie must trigger DisableCaching
		if resp.Header.Get("Cache-Control") != "no-cache, no-store, must-revalidate" {
			t.Errorf("Expected Cache-Control: no-store when stale cookie is present, got %q", resp.Header.Get("Cache-Control"))
		}
	})
}
