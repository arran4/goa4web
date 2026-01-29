package images

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	intimages "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/sign"
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
	RegisterRoutes(r, cfg, navReg, nil, nil)
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
	RegisterRoutes(r, cfg, navReg, nil, nil)
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

func TestVerifyMiddlewareAllowsQuerySignedImage(t *testing.T) {
	called := false
	cfg := config.NewRuntimeConfig()
	cfg.HTTPHostname = "http://localhost"
	key := "k"
	signedURL := "http://localhost/images/image/abcd.png?size=small&sig=" + sign.Sign("image:abcd.png?size=small", key, sign.WithOutNonce())
	parsed, err := url.Parse(signedURL)
	if err != nil {
		t.Fatalf("parse signed url: %v", err)
	}
	h := verifyMiddleware("image:")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	req := httptest.NewRequest("GET", parsed.Path+"?"+parsed.RawQuery, nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abcd.png"})
	cd := common.NewCoreData(req.Context(), nil, cfg, common.WithImageSignKey(key))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req.WithContext(ctx))

	if !called {
		t.Fatalf("next handler was not called")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("want %d got %d", http.StatusOK, rr.Code)
	}
}
