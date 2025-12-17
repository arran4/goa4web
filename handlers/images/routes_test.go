package images

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	intimages "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/navigation"
)

func TestValidID(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"abcd", true},
		{"1234", true},
		{"a1b2c3", true},
		{"a.b", false},
		{"abc!", false},
		{".", false},
		{"..", false},
		{"hi/hi", false},
		{"text.text", true},
		{"a/bc", false},
		{"abc", false},
	}
	for _, tt := range tests {
		if got := intimages.ValidID(tt.id); got != tt.valid {
			t.Errorf("ValidID(%q) = %v want %v", tt.id, got, tt.valid)
		}
	}
}

func TestImageRouteInvalidID(t *testing.T) {
	r := mux.NewRouter()
	cfg := config.NewRuntimeConfig()
	navReg := navigation.NewRegistry()
	RegisterRoutes(r, cfg, navReg)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/images/image/abc!", nil)
	cd := common.NewCoreData(req.Context(), nil, cfg)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)

	r.ServeHTTP(rr, req.WithContext(ctx))

	if rr.Code != http.StatusForbidden {
		t.Fatalf("want %d got %d", http.StatusForbidden, rr.Code)
	}
}

func TestCacheRouteInvalidID(t *testing.T) {
	r := mux.NewRouter()
	cfg := config.NewRuntimeConfig()
	navReg := navigation.NewRegistry()
	RegisterRoutes(r, cfg, navReg)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/images/cache/abc!", nil)
	cd := common.NewCoreData(req.Context(), nil, cfg)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)

	r.ServeHTTP(rr, req.WithContext(ctx))

	if rr.Code != http.StatusForbidden {
		t.Fatalf("want %d got %d", http.StatusForbidden, rr.Code)
	}
}

func TestVerifyMiddlewareUnauthorized(t *testing.T) {
	called := false
	h := verifyMiddleware("image:")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	req := httptest.NewRequest("GET", "/images/image/abcd", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abcd"})
	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req.WithContext(ctx))

	if called {
		t.Fatalf("next handler was called")
	}
	if rr.Code != http.StatusForbidden {
		t.Fatalf("want %d got %d", http.StatusForbidden, rr.Code)
	}
}
