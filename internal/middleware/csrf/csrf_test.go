package csrf

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/testhelpers"
)

var (
	sessionName = "my-session"
	store       *sessions.CookieStore
)

func TestCSRFLoginFlow(t *testing.T) {
	store = sessions.NewCookieStore([]byte("testsecret"))
	core.Store = store
	core.SessionName = sessionName

	r := mux.NewRouter()
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Token", Token(r))
	}).Methods("GET")
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("POST")

	handler := NewCSRFMiddleware("testsecret", "http://example.com", "dev")(r)

	req := httptest.NewRequest("GET", "http://example.com/login", nil)
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
	req2 := httptest.NewRequest("POST", "http://example.com/login", strings.NewReader(data.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("X-CSRF-Token", token)
	req2.Header.Set("Cookie", cookieHeader)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr2.Code)
	}
}

func TestCSRFCrossSite(t *testing.T) {
	store = sessions.NewCookieStore([]byte("testsecret"))
	core.Store = store
	core.SessionName = sessionName

	r := mux.NewRouter()
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Token", Token(r))
	}).Methods("GET")
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("POST")

	handler := NewCSRFMiddleware("testsecret", "https://example.com", "prod")(r)

	req := httptest.NewRequest("GET", "https://example.com/login", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	token := rr.Header().Get("X-Token")
	cookie := rr.Header().Get("Set-Cookie")

	data := url.Values{"gorilla.csrf.Token": {token}}
	req2 := httptest.NewRequest("POST", "https://example.com/login", strings.NewReader(data.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("X-CSRF-Token", token)
	req2.Header.Set("Cookie", cookie)
	req2.Header.Set("Origin", "https://bad.com")
	req2.Header.Set("Sec-Fetch-Site", "cross-site")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden got %d", rr2.Code)
	}

	req3 := httptest.NewRequest("POST", "https://example.com/login", strings.NewReader(data.Encode()))
	req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req3.Header.Set("X-CSRF-Token", token)
	req3.Header.Set("Cookie", cookie)
	req3.Header.Set("Origin", "https://example.com")
	req3.Header.Set("Sec-Fetch-Site", "same-origin")
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)
	if rr3.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr3.Code)
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

	cfg := config.RuntimeConfig{CSRFEnabled: false}

	var handler http.Handler = r
	if cfg.CSRFEnabled {
		handler = NewCSRFMiddleware("testsecret", "http://example.com", "dev")(handler)
	}

	req := httptest.NewRequest("POST", "http://example.com/login", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
}

func TestCSRFRotatesAfterAuthentication(t *testing.T) {
	store = sessions.NewCookieStore([]byte("testsecret"))
	core.Store = store
	core.SessionName = sessionName

	r := mux.NewRouter()
	r.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Token", Token(r))
	}).Methods(http.MethodGet)
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		session := testhelpers.Must(core.GetSession(r))
		session.Values["UID"] = int32(42)
		if err := session.Save(r, w); err != nil {
			t.Fatalf("save session: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}).Methods(http.MethodPost)

	handler := NewCSRFMiddleware("testsecret", "http://example.com", "dev")(r)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/token", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	token := rr.Header().Get("X-Token")
	if token == "" {
		t.Fatalf("missing token on initial request")
	}
	cookie := rr.Header().Get("Set-Cookie")

	req2 := httptest.NewRequest(http.MethodGet, "http://example.com/token", nil)
	req2.Header.Set("Cookie", cookie)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if token != rr2.Header().Get("X-Token") {
		t.Fatalf("token changed before authentication")
	}

	loginForm := url.Values{"gorilla.csrf.Token": {token}}
	loginReq := httptest.NewRequest(http.MethodPost, "http://example.com/login", strings.NewReader(loginForm.Encode()))
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginReq.Header.Set("X-CSRF-Token", token)
	loginReq.Header.Set("Cookie", cookie)
	loginRR := httptest.NewRecorder()
	handler.ServeHTTP(loginRR, loginReq)
	if loginRR.Code != http.StatusNoContent {
		t.Fatalf("login failed with status %d", loginRR.Code)
	}
	if updated := loginRR.Header().Get("Set-Cookie"); updated != "" {
		cookie = updated
	}

	req3 := httptest.NewRequest(http.MethodGet, "http://example.com/token", nil)
	req3.Header.Set("Cookie", cookie)
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)
	rotated := rr3.Header().Get("X-Token")
	if rotated == "" {
		t.Fatalf("missing token after login")
	}
	if rotated == token {
		t.Fatalf("expected token rotation after authentication")
	}
}
