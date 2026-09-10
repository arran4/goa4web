package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

func TestHappyPathAuthPages_CacheControl(t *testing.T) {
	originalTemplateExecute := tasks.TemplateExecute
	defer func() {
		tasks.TemplateExecute = originalTemplateExecute
	}()

	// Keep the production TemplateHandler path while avoiding unrelated
	// template data requirements in this cache-policy regression.
	tasks.TemplateExecute = func(w http.ResponseWriter, r *http.Request, p tasks.Template, data any) error {
		return nil
	}

	tests := []struct {
		name    string
		handler func(w http.ResponseWriter, r *http.Request)
	}{
		{
			name: "Login Page",
			handler: func(w http.ResponseWriter, r *http.Request) {
				handlers.WithNoCache(loginTask.Page)(w, r)
			},
		},
		{
			name: "Register Page",
			handler: func(w http.ResponseWriter, r *http.Request) {
				handlers.WithNoCache(registerTask.Page)(w, r)
			},
		},
		{
			name: "Forgot Password Page",
			handler: func(w http.ResponseWriter, r *http.Request) {
				handlers.WithNoCache(forgotPasswordTask.Page)(w, r)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig())
			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

			rr := httptest.NewRecorder()
			tt.handler(rr, req)

			cc := rr.Header().Get("Cache-Control")
			if !strings.Contains(cc, "no-cache") {
				t.Errorf("expected Cache-Control: no-cache, got %q", cc)
			}
			if !strings.Contains(cc, "no-store") {
				t.Errorf("expected Cache-Control: no-store, got %q", cc)
			}
			if got := rr.Header().Get("Cloudflare-CDN-Cache-Control"); got != "no-store" {
				t.Errorf("expected Cloudflare-CDN-Cache-Control: no-store, got %q", got)
			}
			if pragma := rr.Header().Get("Pragma"); pragma != "no-cache" {
				t.Errorf("expected Pragma: no-cache, got %q", pragma)
			}
			if expires := rr.Header().Get("Expires"); expires != "0" {
				t.Errorf("expected Expires: 0, got %q", expires)
			}
		})
	}
}
