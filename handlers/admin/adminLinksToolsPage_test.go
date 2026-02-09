package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestHappyPathAdminLinksToolsPage(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.LinkSignSecret = "test-key"
	cfg.HTTPHostname = "http://example.com"
	queries := testhelpers.NewQuerierStub()

	req := httptest.NewRequest(http.MethodGet, "/admin/links/tools", nil)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, common.NewCoreData(req.Context(), queries, cfg, common.WithUserRoles([]string{"administrator"})))
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	AdminLinksToolsPage(rr, req)

	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
	if !strings.Contains(rr.Body.String(), "Link Signing Tools") {
		t.Fatalf("expected page content, got %q", rr.Body.String())
	}
}

func TestHappyPathAdminLinksToolsPageSign(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.LinkSignSecret = "test-key"
	cfg.HTTPHostname = "http://example.com"
	queries := testhelpers.NewQuerierStub()

	form := url.Values{}
	form.Set("action", "sign")
	form.Set("sign_url", "https://example.com")
	form.Set("sign_duration", "1h")

	req := httptest.NewRequest(http.MethodPost, "/admin/links/tools", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, common.NewCoreData(req.Context(), queries, cfg, common.WithUserRoles([]string{"administrator"})))
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	AdminLinksToolsPage(rr, req)

	body := rr.Body.String()
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
	if !strings.Contains(body, "u=https%3A%2F%2Fexample.com") {
		t.Fatalf("expected signed url in body: %s", body)
	}
	if !strings.Contains(body, "sig=") {
		t.Fatalf("expected signature in body: %s", body)
	}
}

func TestHappyPathAdminLinksToolsPageVerify(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.LinkSignSecret = "test-key"
	queries := testhelpers.NewQuerierStub()

	ts := time.Now().Add(1 * time.Hour).Unix()
	tsStr := strconv.FormatInt(ts, 10)
	urlToVerify := "https://example.com/resource"
	sig := sign.Sign(urlToVerify, cfg.LinkSignSecret, sign.WithExpiry(time.Unix(ts, 0)))

	form := url.Values{}
	form.Set("action", "verify")
	form.Set("verify_url", urlToVerify)
	form.Set("verify_sig", sig)
	form.Set("verify_ts", tsStr)

	req := httptest.NewRequest(http.MethodPost, "/admin/links/tools", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, common.NewCoreData(req.Context(), queries, cfg, common.WithUserRoles([]string{"administrator"})))
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	AdminLinksToolsPage(rr, req)

	body := rr.Body.String()
	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
	if !strings.Contains(body, "valid") {
		t.Fatalf("expected valid verification, got body: %s", body)
	}
}
