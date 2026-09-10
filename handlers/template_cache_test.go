package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/tasks"
)

func TestTemplateHandlerCachePolicyPrecedence(t *testing.T) {
	originalTemplateExecute := tasks.TemplateExecute
	t.Cleanup(func() { tasks.TemplateExecute = originalTemplateExecute })
	tasks.TemplateExecute = func(http.ResponseWriter, *http.Request, tasks.Template, any) error {
		return nil
	}

	tests := []struct {
		name           string
		userID         int32
		cookie         bool
		initialPolicy  string
		initialEdge    string
		wantPolicy     string
		wantCloudflare string
	}{
		{
			name:       "anonymous public output",
			wantPolicy: "public, max-age=3600",
		},
		{
			name:           "authenticated output",
			userID:         42,
			wantPolicy:     "no-cache, no-store, must-revalidate",
			wantCloudflare: "no-store",
		},
		{
			name:           "session cookie output",
			cookie:         true,
			wantPolicy:     "no-cache, no-store, must-revalidate",
			wantCloudflare: "no-store",
		},
		{
			name:          "explicit sensitive policy",
			initialPolicy: "private, max-age=0",
			wantPolicy:    "private, max-age=0",
		},
		{
			name:           "explicit Cloudflare sensitive policy",
			initialEdge:    "no-store",
			wantCloudflare: "no-store",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewRuntimeConfig()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.cookie {
				req.AddCookie(&http.Cookie{Name: cfg.SessionName, Value: "stale"})
			}
			cd := common.NewCoreData(context.Background(), nil, cfg)
			cd.UserID = tt.userID
			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

			rr := httptest.NewRecorder()
			rr.Header().Set("Cache-Control", tt.initialPolicy)
			rr.Header().Set("Cloudflare-CDN-Cache-Control", tt.initialEdge)
			if err := TemplateHandler(rr, req, tasks.Template("test"), nil); err != nil {
				t.Fatalf("TemplateHandler: %v", err)
			}

			if got := rr.Header().Get("Cache-Control"); got != tt.wantPolicy {
				t.Errorf("Cache-Control = %q; want %q", got, tt.wantPolicy)
			}
			if got := rr.Header().Get("Cloudflare-CDN-Cache-Control"); got != tt.wantCloudflare {
				t.Errorf("Cloudflare-CDN-Cache-Control = %q; want %q", got, tt.wantCloudflare)
			}
		})
	}
}
