package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
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

func TestCSRFLazyErrorBypass(t *testing.T) {
	// A deterministic regression test for the exact writer-bypass failure
	// where `csrfField` evaluates late, triggers a session error (RedirectToLogin),
	// and verifies the final committed response status/headers come from the
	// buffered redirect rather than the outer response.
	r := setupTestRouter()

	r.HandleFunc("/late-error", func(w http.ResponseWriter, req *http.Request) {
		originalTemplateExecute := tasks.TemplateExecute
		t.Cleanup(func() { tasks.TemplateExecute = originalTemplateExecute })

		tasks.TemplateExecute = func(w2 http.ResponseWriter, req2 *http.Request, tmpl tasks.Template, data any) error {
			if tmpl == TaskErrorAcknowledgementPageTmpl {
				return nil
			}
			_, _ = w2.Write([]byte("some initial output... "))
			// Evaluate csrfField. Because the cookie is invalid, it will call SessionErrorRedirect
			_ = csrf.TemplateField(req2)
			_, _ = w2.Write([]byte("form rendered"))
			return nil
		}

		if err := TemplateHandler(w, req, tasks.Template("test"), nil); err != nil {
			t.Fatalf("TemplateHandler: %v", err)
		}
	}).Methods(http.MethodGet)

	handler := csrf.NewCSRFMiddleware("testsecret", "http://example.com", "dev")(r)

	req, _ := mockRequestWithCoreData(http.MethodGet, "http://example.com/late-error")

	// Inject an intentionally corrupt session cookie to trigger core.SessionErrorRedirect inside getToken
	req.AddCookie(&http.Cookie{
		Name:  "goa4web_session",
		Value: "corrupt-undecodable-garbage-value",
	})

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	res := rr.Result()

	// The final committed response must be the 303 Redirect to login, not a 200 OK.
	if res.StatusCode != http.StatusSeeOther {
		t.Fatalf("Expected status 303 See Other, got %d", res.StatusCode)
	}

	loc := res.Header.Get("Location")
	if !strings.HasPrefix(loc, "/login") {
		t.Errorf("Expected redirect location to start with /login, got %q", loc)
	}

	cc := res.Header.Get("Cache-Control")
	expectedCC := "no-cache, no-store, must-revalidate"
	if cc != expectedCC {
		t.Errorf("Expected Cache-Control %q for error redirect, got %q", expectedCC, cc)
	}
}
