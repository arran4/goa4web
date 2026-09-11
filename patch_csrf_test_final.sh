cat << 'INNER_EOF' > tests/csrf/csrf_prod_test.go
package csrf_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/middleware/csrf"
	"github.com/arran4/goa4web/internal/tasks"
    "github.com/arran4/goa4web/internal/app/server"
    "github.com/arran4/goa4web/internal/db"
    "github.com/arran4/goa4web/handlers"
)

func TestCSRFAnonymousPublicCacheable(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.SessionSecret = "testsecret"
    cfg.CSRFEnabled = true

	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	core.Store = store
	core.SessionName = cfg.SessionName

	s := server.New(server.WithConfig(cfg), server.WithDB(&db.QuerierStub{}))

	r := s.Router()

	// A purely anonymous route that doesn't need CSRF state.
	r.HandleFunc("/public-page", func(w http.ResponseWriter, req *http.Request) {
		if err := handlers.TemplateHandler(w, req, tasks.Template(""), nil); err != nil {
			t.Fatalf("TemplateHandler: %v", err)
		}
	}).Methods(http.MethodGet)

	originalTemplateExecute := tasks.TemplateExecute
	t.Cleanup(func() { tasks.TemplateExecute = originalTemplateExecute })
	tasks.TemplateExecute = func(http.ResponseWriter, *http.Request, tasks.Template, any) error {
		return nil
	}

    req := httptest.NewRequest(http.MethodGet, "http://example.com/public-page", nil)
	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rr.Code)
	}

    res := rr.Result()
	cookies := res.Header.Values("Set-Cookie")
	for _, c := range cookies {
		if strings.HasPrefix(c, cfg.SessionName+"=") {
			t.Errorf("Anonymous read-only GET incorrectly emitted application session cookie: %s", c)
		}
	}

	cc := res.Header.Get("Cache-Control")
	if cc != "public, max-age=3600" {
		t.Errorf("Expected Cache-Control to be public, max-age=3600; got %q", cc)
	}
}

func TestCSRFAnonymousWithFormNoStore(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.SessionSecret = "testsecret"
    cfg.CSRFEnabled = true

	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	core.Store = store
	core.SessionName = cfg.SessionName

	s := server.New(server.WithConfig(cfg), server.WithDB(&db.QuerierStub{}))

	r := s.Router()

	// A route that explicitly renders a CSRF form.
	r.HandleFunc("/login-test", func(w http.ResponseWriter, req *http.Request) {
		originalTemplateExecute := tasks.TemplateExecute
		t.Cleanup(func() { tasks.TemplateExecute = originalTemplateExecute })
		tasks.TemplateExecute = func(w2 http.ResponseWriter, req2 *http.Request, tmpl tasks.Template, data any) error {
			_, _ = w2.Write([]byte("start form... "))
			// Simulating {{ csrfField }} being evaluated AFTER some output has been written
			_ = csrf.TemplateField(req2)
			_, _ = w2.Write([]byte("form rendered"))
			return nil
		}

		if err := handlers.TemplateHandler(w, req, tasks.Template("test"), nil); err != nil {
			t.Fatalf("TemplateHandler: %v", err)
		}
	}).Methods(http.MethodGet)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/login-test", nil)
	rr := httptest.NewRecorder()

	s.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rr.Code)
	}

    res := rr.Result()
	hasAppCookie := false
	cookies := res.Header.Values("Set-Cookie")
	for _, c := range cookies {
		if strings.HasPrefix(c, cfg.SessionName+"=") {
			hasAppCookie = true
		}
	}
	if !hasAppCookie {
		t.Errorf("Expected application session cookie to be generated when CSRF token is requested")
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
}

func TestCSRFLazyBypassPath(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.SessionSecret = "testsecret"
    cfg.CSRFEnabled = true

	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	core.Store = store
	core.SessionName = cfg.SessionName

	s := server.New(server.WithConfig(cfg), server.WithDB(&db.QuerierStub{}))

    r := s.Router()

    r.HandleFunc("/direct-write", func(w http.ResponseWriter, req *http.Request) {
        _, _ = w.Write([]byte("start directly "))
        _ = csrf.TemplateField(req)
        _, _ = w.Write([]byte("end directly"))
    }).Methods(http.MethodGet)

    req := httptest.NewRequest(http.MethodGet, "http://example.com/direct-write", nil)
	rr := httptest.NewRecorder()

    s.ServeHTTP(rr, req)

    _ = rr.Result()
    if rr.Code != http.StatusOK {
        t.Fatalf("Expected 200, got %d", rr.Code)
    }
}
INNER_EOF
