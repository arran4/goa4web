package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/tasks"
)

func TestRequireRole_CacheControl(t *testing.T) {
	tests := []struct {
		name         string
		roles        []string
		expectedHdr  bool
		allowedRoles []string
	}{
		{
			name:         "Public role anyone",
			roles:        []string{"anyone"},
			expectedHdr:  false,
			allowedRoles: []string{"anyone"},
		},
		{
			name:         "Restricted role user",
			roles:        []string{"user"},
			expectedHdr:  true,
			allowedRoles: []string{"user"},
		},
		{
			name:         "Restricted role admin",
			roles:        []string{"admin"},
			expectedHdr:  true,
			allowedRoles: []string{"admin"},
		},
		{
			name:         "Mixed roles with anyone",
			roles:        []string{"user", "anyone"},
			expectedHdr:  false,
			allowedRoles: []string{"anyone"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := RequireRole(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}, fmt.Errorf("forbidden"), tt.roles...)

			req := httptest.NewRequest("GET", "/", nil)
			cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles(tt.allowedRoles))
			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

			rr := httptest.NewRecorder()
			h(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("expected status OK, got %d", rr.Code)
			}

			cc := rr.Header().Get("Cache-Control")
			if tt.expectedHdr {
				if !strings.Contains(cc, "no-cache") {
					t.Errorf("expected Cache-Control: no-cache, got %q", cc)
				}
				if !strings.Contains(cc, "no-store") {
					t.Errorf("expected Cache-Control: no-store, got %q", cc)
				}
			} else {
				if cc != "" {
					t.Errorf("expected no Cache-Control header, got %q", cc)
				}
			}
		})
	}
}

func TestDisableCaching(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		rr := httptest.NewRecorder()
		DisableCaching(rr)

		if got := rr.Header().Get("Cache-Control"); got != "no-cache, no-store, must-revalidate" {
			t.Errorf("Cache-Control = %q; want no-cache, no-store, must-revalidate", got)
		}
		if got := rr.Header().Get("Pragma"); got != "no-cache" {
			t.Errorf("Pragma = %q; want no-cache", got)
		}
		if got := rr.Header().Get("Expires"); got != "0" {
			t.Errorf("Expires = %q; want 0", got)
		}
	})
}

func TestErrorHandlers_CacheControl(t *testing.T) {
	// Mock tasks.Handle to avoid template execution errors
	originalHandle := tasks.Handle
	defer func() { tasks.Handle = originalHandle }()
	tasks.Handle = func(w http.ResponseWriter, r *http.Request, p tasks.Template, data any) error {
		return nil
	}

	t.Run("Happy Path", func(t *testing.T) {
		t.Run("RenderPermissionDenied", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig())
			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
			rr := httptest.NewRecorder()

			RenderPermissionDenied(rr, req)

			cc := rr.Header().Get("Cache-Control")
			if !strings.Contains(cc, "no-cache") {
				t.Errorf("expected Cache-Control: no-cache, got %q", cc)
			}
		})

		t.Run("RenderNotFoundOrLogin", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig())
			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
			rr := httptest.NewRecorder()

			RenderNotFoundOrLogin(rr, req)

			cc := rr.Header().Get("Cache-Control")
			if !strings.Contains(cc, "no-cache") {
				t.Errorf("expected Cache-Control: no-cache, got %q", cc)
			}
		})
	})
}
