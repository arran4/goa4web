//go:build sqlite
// +build sqlite

package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestScenarioServeCmd_CSRFCachingGuarantees(t *testing.T) {
	ctx := context.Background()

	t.Setenv("GOA4WEB_CSRF_ENABLED", "true")
	t.Setenv("GOA4WEB_SESSION_SECRET", "testsecret")

	root, err := parseRoot([]string{"goa4web", "scenario", "serve", "../../testdata/scenarios/100-private-forum"})
	if err != nil {
		t.Fatalf("parseRoot: %v", err)
	}
	defer root.Close()

	parent, err := parseScenarioCmd(root, []string{"serve", "../../testdata/scenarios/100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}

	cmd, err := parseScenarioServeCmd(parent, []string{"../../testdata/scenarios/100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}

	srv, _, cleanup, err := cmd.Bootstrap(ctx)
	if err != nil {
		t.Fatalf("scenarioServeCmd Bootstrap failed: %v", err)
	}
	defer cleanup()

	router := srv.Router

	t.Run("anonymous_read_only", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://localhost/faq", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		res := rr.Result()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK for /faq, got %d", res.StatusCode)
		}

		for _, c := range res.Header.Values("Set-Cookie") {
			if strings.HasPrefix(c, "my-session=") || strings.HasPrefix(c, "goa4web_session=") {
				t.Errorf("Anonymous read-only GET incorrectly emitted application session cookie: %s", c)
			}
		}

		cc := res.Header.Get("Cache-Control")
		if cc != "public, max-age=3600" {
			t.Errorf("Expected Cache-Control to be public, max-age=3600; got %q", cc)
		}
	})

	t.Run("anonymous_csrf_form", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://localhost/login", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		res := rr.Result()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 OK for /login, got %d", res.StatusCode)
		}

		hasAppCookie := false
		hasMainCookie := false
		for _, c := range res.Header.Values("Set-Cookie") {
			if strings.HasPrefix(c, "my-session_csrf=") || strings.HasPrefix(c, "goa4web_session_csrf=") {
				hasAppCookie = true
			}
			if strings.HasPrefix(c, "my-session=") || strings.HasPrefix(c, "goa4web_session=") {
				hasMainCookie = true
			}
		}
		if !hasAppCookie {
			t.Errorf("Expected CSRF session cookie for /login because it contains a CSRF form")
		}
		if hasMainCookie {
			t.Errorf("Expected NO main application cookie for /login because it only renders a CSRF form")
		}

		cc := res.Header.Get("Cache-Control")
		expectedCC := "no-cache, no-store, must-revalidate"
		if cc != expectedCC {
			t.Errorf("Expected Cache-Control %q, got %q", expectedCC, cc)
		}
		cfcc := res.Header.Get("Cloudflare-CDN-Cache-Control")
		expectedCFCC := "no-store"
		if cfcc != expectedCFCC {
			t.Errorf("Expected Cloudflare-CDN-Cache-Control %q, got %q", expectedCFCC, cfcc)
		}
	})
	t.Run("logout_csrf_lifecycle", func(t *testing.T) {
		// Start by logging in as Alice (from the 100-private-forum scenario which gets seeded)
		// We'll extract the CSRF token from the login page, then POST /login

		// 1. GET /login
		req := httptest.NewRequest(http.MethodGet, "http://localhost/login", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Result().StatusCode != http.StatusOK {
			t.Fatalf("GET /login failed")
		}

		var csrfCookie *http.Cookie
		for _, c := range rr.Result().Cookies() {
			if strings.HasPrefix(c.Name, "goa4web_session_csrf") || strings.HasPrefix(c.Name, "my-session_csrf") {
				csrfCookie = c
				// Fix the cookie if it's broken
				if strings.Contains(csrfCookie.Value, ";") {
					csrfCookie.Value = strings.Split(csrfCookie.Value, ";")[0]
				}
			}
		}

		// Extract token from body
		bodyStr := rr.Body.String()
		tokenPrefix := "name=\"gorilla.csrf.Token\" value=\""
		idx := strings.Index(bodyStr, tokenPrefix)
		if idx == -1 {
			t.Fatalf("Could not find CSRF token in login page")
		}
		val := bodyStr[idx+len(tokenPrefix):]
		endIdx := strings.Index(val, "\"")
		loginToken := val[:endIdx]

		// 2. POST /login
		loginForm := url.Values{

			"task":               {"Login"},
			"username":           {"alice"},
			"password":           {"alice-test"},
			"gorilla.csrf.Token": {loginToken},
		}
		req = httptest.NewRequest(http.MethodPost, "http://localhost/login", strings.NewReader(loginForm.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("X-CSRF-Token", loginToken)
		if csrfCookie != nil {
			req.AddCookie(csrfCookie)
		}
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Result().StatusCode != http.StatusSeeOther {

			// extract the error message from the form to understand why
			t.Fatalf("POST /login failed: %s", rr.Body.String())
		}

		var authCookie *http.Cookie
		for _, c := range rr.Result().Cookies() {
			if c.Name == "goa4web_session" || c.Name == "my-session" {
				authCookie = c
			}
			if strings.HasPrefix(c.Name, "goa4web_session_csrf") || strings.HasPrefix(c.Name, "my-session_csrf") {
				csrfCookie = c // updated csrf cookie
			}
		}

		if authCookie == nil {
			t.Fatalf("No auth cookie generated upon login")
		}

		// Ensure we are logged in
		req = httptest.NewRequest(http.MethodGet, "http://localhost/usr", nil)
		req.AddCookie(authCookie)
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Result().StatusCode != http.StatusOK {
			t.Fatalf("Failed to verify authenticated state")
		}

		// 3. GET /usr/logout non-mutating & provides token
		req = httptest.NewRequest(http.MethodGet, "http://localhost/usr/logout", nil)
		req.AddCookie(authCookie)
		if csrfCookie != nil {
			req.AddCookie(csrfCookie)
		}
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Result().StatusCode != http.StatusOK {
			t.Fatalf("GET /usr/logout failed")
		}

		// Need to get the NEW token after login because identity change rotates CSRF token!
		// Since logout.gohtml fails to render here, we will fetch /usr which also has a CSRF form (e.g. for user settings).
		req = httptest.NewRequest(http.MethodGet, "http://localhost/usr", nil)
		req.AddCookie(authCookie)
		if csrfCookie != nil {
			req.AddCookie(csrfCookie)
		}
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Instead of /usr, query /user/appearance which has a CSRF form.
		req = httptest.NewRequest(http.MethodGet, "http://localhost/usr/appearance", nil)
		req.AddCookie(authCookie)
		if csrfCookie != nil {
			req.AddCookie(csrfCookie)
		}
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		for _, c := range rr.Result().Cookies() {
			if strings.HasPrefix(c.Name, "goa4web_session_csrf") || strings.HasPrefix(c.Name, "my-session_csrf") {
				csrfCookie = c
				if strings.Contains(csrfCookie.Value, ";") {
					csrfCookie.Value = strings.Split(csrfCookie.Value, ";")[0]
				}
			}
		}

		bodyStr = rr.Body.String()
		idx = strings.Index(bodyStr, tokenPrefix)
		if idx == -1 {
			t.Fatalf("Could not find CSRF token in /usr/appearance page")
		}
		val = bodyStr[idx+len(tokenPrefix):]
		endIdx = strings.Index(val, "\"")
		logoutToken := val[:endIdx]

		// 4. POST /usr/logout without token => forbidden
		req = httptest.NewRequest(http.MethodPost, "http://localhost/usr/logout", nil)
		req.AddCookie(authCookie)
		if csrfCookie != nil {
			req.AddCookie(csrfCookie)
		}
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Result().StatusCode != http.StatusForbidden {
			t.Fatalf("POST /usr/logout without CSRF should be 403, got %d", rr.Result().StatusCode)
		}

		// Session should remain valid
		req = httptest.NewRequest(http.MethodGet, "http://localhost/usr", nil)
		req.AddCookie(authCookie)
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Result().StatusCode != http.StatusOK {
			t.Fatalf("Session should remain valid after failed CSRF logout")
		}

		// 5. POST /usr/logout with invalid token => forbidden
		logoutForm := url.Values{"gorilla.csrf.Token": {"invalidtoken"}}
		req = httptest.NewRequest(http.MethodPost, "http://localhost/usr/logout", strings.NewReader(logoutForm.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(authCookie)
		if csrfCookie != nil {
			req.AddCookie(csrfCookie)
		}
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Result().StatusCode != http.StatusForbidden {
			t.Fatalf("POST /usr/logout with invalid CSRF should be 403, got %d", rr.Result().StatusCode)
		}

		// 6. POST /usr/logout with valid token
		logoutForm = url.Values{"gorilla.csrf.Token": {logoutToken}}
		req = httptest.NewRequest(http.MethodPost, "http://localhost/usr/logout", strings.NewReader(logoutForm.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("X-CSRF-Token", logoutToken)
		req.AddCookie(authCookie)
		if csrfCookie != nil {
			req.AddCookie(csrfCookie)
		}
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Result().StatusCode != http.StatusSeeOther {
			t.Fatalf("POST /usr/logout with valid CSRF should be 303, got %d", rr.Result().StatusCode)
		}

		// Verify cookies were cleared
		authCleared := false
		csrfCleared := false
		for _, c := range rr.Result().Cookies() {
			if (c.Name == "goa4web_session" || c.Name == "my-session") && c.MaxAge < 0 {
				authCleared = true
			}
			if (strings.HasPrefix(c.Name, "goa4web_session_csrf") || strings.HasPrefix(c.Name, "my-session_csrf")) && c.MaxAge < 0 {
				csrfCleared = true
			}
		}
		if !authCleared || !csrfCleared {
			t.Fatalf("Logout did not emit clear cookies for both auth and CSRF sessions")
		}

		// 7. Replay captured pre-logout cookie
		req = httptest.NewRequest(http.MethodGet, "http://localhost/usr", nil)
		req.AddCookie(authCookie)
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Result().StatusCode == http.StatusOK {
			t.Fatalf("Replay of pre-logout cookie succeeded, expected failure")
		}
	})
	t.Run("independent_csrf_jars", func(t *testing.T) {
		// Jar A: Alice
		reqA := httptest.NewRequest(http.MethodGet, "http://localhost/login", nil)
		rrA := httptest.NewRecorder()
		router.ServeHTTP(rrA, reqA)

		var csrfCookieA *http.Cookie
		for _, c := range rrA.Result().Cookies() {
			if strings.HasPrefix(c.Name, "goa4web_session_csrf") || strings.HasPrefix(c.Name, "my-session_csrf") {
				csrfCookieA = c
				if strings.Contains(csrfCookieA.Value, ";") {
					csrfCookieA.Value = strings.Split(csrfCookieA.Value, ";")[0]
				}
			}
		}

		bodyStr := rrA.Body.String()
		tokenPrefix := "name=\"gorilla.csrf.Token\" value=\""
		idx := strings.Index(bodyStr, tokenPrefix)
		if idx == -1 {
			t.Fatalf("Could not find CSRF token in login page A")
		}
		val := bodyStr[idx+len(tokenPrefix):]
		tokenA := val[:strings.Index(val, "\"")]

		// Jar B: Bob
		reqB := httptest.NewRequest(http.MethodGet, "http://localhost/login", nil)
		rrB := httptest.NewRecorder()
		router.ServeHTTP(rrB, reqB)

		var csrfCookieB *http.Cookie
		for _, c := range rrB.Result().Cookies() {
			if strings.HasPrefix(c.Name, "goa4web_session_csrf") || strings.HasPrefix(c.Name, "my-session_csrf") {
				csrfCookieB = c
				if strings.Contains(csrfCookieB.Value, ";") {
					csrfCookieB.Value = strings.Split(csrfCookieB.Value, ";")[0]
				}
			}
		}

		bodyStr = rrB.Body.String()
		idx = strings.Index(bodyStr, tokenPrefix)
		if idx == -1 {
			t.Fatalf("Could not find CSRF token in login page B")
		}
		val = bodyStr[idx+len(tokenPrefix):]
		tokenB := val[:strings.Index(val, "\"")]

		// Ensure token A does not work with Jar B's cookie
		loginForm := url.Values{
			"task":               {"Login"},
			"username":           {"alice"},
			"password":           {"alice-test"},
			"gorilla.csrf.Token": {tokenA},
		}
		reqMismatched := httptest.NewRequest(http.MethodPost, "http://localhost/login", strings.NewReader(loginForm.Encode()))
		reqMismatched.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		reqMismatched.Header.Set("X-CSRF-Token", tokenA)
		if csrfCookieB != nil {
			reqMismatched.AddCookie(csrfCookieB)
		}
		rrMismatched := httptest.NewRecorder()
		router.ServeHTTP(rrMismatched, reqMismatched)

		if rrMismatched.Result().StatusCode != http.StatusForbidden {
			t.Fatalf("POST /login with mismatched CSRF jar should be 403, got %d", rrMismatched.Result().StatusCode)
		}

		// Ensure token B works with Jar B's cookie
		loginForm.Set("gorilla.csrf.Token", tokenB)
		loginForm.Set("username", "bob")
		loginForm.Set("password", "bob-test")
		reqValid := httptest.NewRequest(http.MethodPost, "http://localhost/login", strings.NewReader(loginForm.Encode()))
		reqValid.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		reqValid.Header.Set("X-CSRF-Token", tokenB)
		if csrfCookieB != nil {
			reqValid.AddCookie(csrfCookieB)
		}
		rrValid := httptest.NewRecorder()
		router.ServeHTTP(rrValid, reqValid)

		if rrValid.Result().StatusCode != http.StatusSeeOther {
			t.Fatalf("POST /login with valid CSRF jar should be 303, got %d", rrValid.Result().StatusCode)
		}
	})
}
