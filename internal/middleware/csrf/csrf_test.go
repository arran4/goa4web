package csrf

import (
	"crypto/sha256"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/runtimeconfig"

	"github.com/arran4/goa4web/core"
)

var (
	sessionName = "my-session"
	store       *sessions.CookieStore
)

func TestCSRFLoginFlow(t *testing.T) {
	store = sessions.NewCookieStore([]byte("testsecret"))
	core.Store = store
	core.SessionName = sessionName
	key := sha256.Sum256([]byte("testsecret"))

	r := mux.NewRouter()
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Token", csrf.Token(r))
	}).Methods("GET")
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("POST")

	handler := csrf.Protect(key[:], csrf.Secure(false), csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("failure: %v", csrf.FailureReason(r))
		http.Error(w, "fail", http.StatusForbidden)
	})))(r)

	req := csrf.PlaintextHTTPRequest(httptest.NewRequest("GET", "/login", nil))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("GET status %d", rr.Code)
	}
	token := rr.Header().Get("X-Token")
	if token == "" {
		t.Fatal("missing token")
	}
	cookieHeader := rr.Header().Get("Set-Cookie")

	data := url.Values{"gorilla.csrf.Token": {token}}
	req2 := csrf.PlaintextHTTPRequest(httptest.NewRequest("POST", "/login", strings.NewReader(data.Encode())))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("Cookie", cookieHeader)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d (%v)", rr2.Code, csrf.FailureReason(req2))
	}
}

func TestCSRFMismatchedReferer(t *testing.T) {
	store = sessions.NewCookieStore([]byte("testsecret"))
	core.Store = store
	core.SessionName = sessionName
	key := sha256.Sum256([]byte("testsecret"))

	r := mux.NewRouter()
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Token", csrf.Token(r))
	}).Methods("GET")
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("POST")

	handler := csrf.Protect(key[:], csrf.Secure(true), csrf.TrustedOrigins([]string{"example.com"}))(r)

	req := httptest.NewRequest("GET", "https://example.com/login", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	token := rr.Header().Get("X-Token")
	cookie := rr.Header().Get("Set-Cookie")

	data := url.Values{"gorilla.csrf.Token": {token}}
	req2 := httptest.NewRequest("POST", "https://example.com/login", strings.NewReader(data.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("Cookie", cookie)
	req2.Header.Set("Referer", "https://bad.com/login")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden got %d", rr2.Code)
	}

	req3 := httptest.NewRequest("POST", "https://example.com/login", strings.NewReader(data.Encode()))
	req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req3.Header.Set("Cookie", cookie)
	req3.Header.Set("Referer", "https://example.com/login")
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)
	if rr3.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d (%v)", rr3.Code, csrf.FailureReason(req3))
	}
}

func TestCSRFDisabled(t *testing.T) {
	store = sessions.NewCookieStore([]byte("testsecret"))
	core.Store = store
	core.SessionName = sessionName

	r := mux.NewRouter()
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("POST")

	orig := runtimeconfig.AppRuntimeConfig
	runtimeconfig.AppRuntimeConfig.CSRFEnabled = false
	t.Cleanup(func() { runtimeconfig.AppRuntimeConfig = orig })

	var handler http.Handler = r
	if CSRFEnabled() {
		key := sha256.Sum256([]byte("testsecret"))
		handler = csrf.Protect(key[:], csrf.Secure(false))(handler)
	}

	req := httptest.NewRequest("POST", "http://example.com/login", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
}
