package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/middleware/csrf"
	"github.com/arran4/goa4web/internal/tasks"
)

// Tests addressing Issue #3104: Prevent CSRF session creation from making
// anonymous HTML publicly cacheable.

func setupTestRouter() *mux.Router {
	store := sessions.NewCookieStore([]byte("testsecret"))
	core.Store = store
	core.SessionName = "goa4web_session"

	r := mux.NewRouter()
	return r
}

func mockRequestWithCoreData(method, path string) (*http.Request, *common.CoreData) {
	req := httptest.NewRequest(method, path, nil)
	cfg := config.NewRuntimeConfig()
	cfg.SessionName = "goa4web_session"
	cd := common.NewCoreData(context.Background(), nil, cfg)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	return req, cd
}

func TestCSRFAnonymousPublicCacheable(t *testing.T) {
	r := setupTestRouter()

	// A purely anonymous route that doesn't need CSRF state.
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// Just render empty template. Since it doesn't use csrfField,
		// the lazyCSRF token is not generated.
		if err := TemplateHandler(w, req, tasks.Template(""), nil); err != nil {
			t.Fatalf("TemplateHandler: %v", err)
		}
	}).Methods(http.MethodGet)

	handler := csrf.NewCSRFMiddleware("testsecret", "http://example.com", "dev")(r)

	req, _ := mockRequestWithCoreData(http.MethodGet, "http://example.com/")
	rr := httptest.NewRecorder()

	originalTemplateExecute := tasks.TemplateExecute
	t.Cleanup(func() { tasks.TemplateExecute = originalTemplateExecute })
	tasks.TemplateExecute = func(http.ResponseWriter, *http.Request, tasks.Template, any) error {
		return nil
	}

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rr.Code)
	}

	cookies := rr.Header().Values("Set-Cookie")
	for _, c := range cookies {
		if strings.HasPrefix(c, "goa4web_session=") {
			t.Errorf("Anonymous read-only GET incorrectly emitted application session cookie: %s", c)
		}
	}

	cc := rr.Header().Get("Cache-Control")
	if cc != "public, max-age=3600" {
		t.Errorf("Expected Cache-Control to be public, max-age=3600; got %q", cc)
	}
}

func TestCSRFAnonymousWithFormNoStore(t *testing.T) {
	r := setupTestRouter()

	// A route that explicitly renders a CSRF form.
	r.HandleFunc("/login", func(w http.ResponseWriter, req *http.Request) {
		// Mock a template executing that uses TemplateField -> generating the token lazily.
		originalTemplateExecute := tasks.TemplateExecute
		t.Cleanup(func() { tasks.TemplateExecute = originalTemplateExecute })
		tasks.TemplateExecute = func(w2 http.ResponseWriter, req2 *http.Request, tmpl tasks.Template, data any) error {
			// Simulating {{ csrfField }}
			_ = csrf.TemplateField(req2)
			_, _ = w2.Write([]byte("form rendered"))
			return nil
		}

		if err := TemplateHandler(w, req, tasks.Template("test"), nil); err != nil {
			t.Fatalf("TemplateHandler: %v", err)
		}
	}).Methods(http.MethodGet)

	handler := csrf.NewCSRFMiddleware("testsecret", "http://example.com", "dev")(r)

	req, _ := mockRequestWithCoreData(http.MethodGet, "http://example.com/login")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rr.Code)
	}

	hasAppCookie := false
	cookies := rr.Header().Values("Set-Cookie")
	for _, c := range cookies {
		if strings.HasPrefix(c, "goa4web_session=") {
			hasAppCookie = true
		}
	}
	if !hasAppCookie {
		t.Errorf("Expected application session cookie to be generated when CSRF token is requested")
	}

	cc := rr.Header().Get("Cache-Control")
	expectedCC := "no-cache, no-store, must-revalidate"
	if cc != expectedCC {
		t.Errorf("Expected Cache-Control %q, got %q", expectedCC, cc)
	}
	cfcc := rr.Header().Get("Cloudflare-CDN-Cache-Control")
	expectedCFCC := "no-store"
	if cfcc != expectedCFCC {
		t.Errorf("Expected Cloudflare-CDN-Cache-Control %q, got %q", expectedCFCC, cfcc)
	}
}

func TestCSRFValidTokenSucceeds(t *testing.T) {
	r := setupTestRouter()

	r.HandleFunc("/post", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Success"))
	}).Methods(http.MethodPost)

	handler := csrf.NewCSRFMiddleware("testsecret", "http://example.com", "dev")(r)

	req, _ := mockRequestWithCoreData(http.MethodGet, "http://example.com/")
	rr := httptest.NewRecorder()

	// GET to fetch token
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		token := csrf.Token(req)
		w.Header().Set("X-Token", token)
	}).Methods(http.MethodGet)

	handler.ServeHTTP(rr, req)

	token := rr.Header().Get("X-Token")
	if token == "" {
		t.Fatalf("Failed to retrieve token from GET request")
	}

	var appCookie string
	var gorillaCookie string
	cookies := rr.Header().Values("Set-Cookie")
	for _, c := range cookies {
		if strings.HasPrefix(c, "goa4web_session=") {
			appCookie = strings.Split(c, ";")[0]
		}
		if strings.HasPrefix(c, "_gorilla_csrf=") {
			gorillaCookie = strings.Split(c, ";")[0]
		}
	}

	// POST request with token
	data := url.Values{"gorilla.csrf.Token": {token}}
	req2, _ := mockRequestWithCoreData(http.MethodPost, "http://example.com/post")
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("X-CSRF-Token", token)
	// actually use strings.NewReader
	req2 = httptest.NewRequest(http.MethodPost, "http://example.com/post", strings.NewReader(data.Encode()))
	// Reattach coredata context because NewRequest clears it
	cfg := config.NewRuntimeConfig()
	cfg.SessionName = "goa4web_session"
	cd := common.NewCoreData(context.Background(), nil, cfg)
	req2 = req2.WithContext(context.WithValue(req2.Context(), consts.KeyCoreData, cd))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("X-CSRF-Token", token)

	if appCookie != "" {
		req2.Header.Add("Cookie", appCookie)
	}
	if gorillaCookie != "" {
		req2.Header.Add("Cookie", gorillaCookie)
	}

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("Expected POST to succeed with 200, got %d", rr2.Code)
	}
}
