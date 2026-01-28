package externallink

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/sign"
)

func TestRedirectHandlerSignedURLParam(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	key := "k"
	link := "https://example.com/foo"
	sig := sign.Sign("link:"+link, key)
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/goto?u=%s&sig=%s&go=1", url.QueryEscape(link), sig), nil)
	cd := common.NewCoreData(context.Background(), nil, cfg, func(cd *common.CoreData) {
		cd.LinkSignKey = key
	})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rec := httptest.NewRecorder()

	RedirectHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, res.StatusCode)
	}
	if got := res.Header.Get("Location"); got != link {
		t.Fatalf("expected redirect to %s, got %s", link, got)
	}
}

func TestRedirectHandlerSignedURLParamWithQuery(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	key := "k"
	link := "https://example.com?a=1&b=2"
	sig := sign.Sign("link:"+link, key)
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/goto?u=%s&sig=%s&go=1", url.QueryEscape(link), sig), nil)
	cd := common.NewCoreData(context.Background(), nil, cfg, func(cd *common.CoreData) {
		cd.LinkSignKey = key
	})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rec := httptest.NewRecorder()

	RedirectHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, res.StatusCode)
	}
	if got := res.Header.Get("Location"); got != link {
		t.Fatalf("expected redirect to %s, got %s", link, got)
	}
}
