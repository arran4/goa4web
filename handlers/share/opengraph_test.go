package share

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/sharesign"
	"github.com/gorilla/mux"
)

func TestVerifyAndGetPathPreservesQueryOrder(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.HTTPHostname = "http://example.com"
	signer := sharesign.NewSigner(cfg, "secret")

	signed := signer.SignedURLQuery("/blogs/blog/5?b=1&a=2")
	req, err := http.NewRequest(http.MethodGet, signed, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	got := VerifyAndGetPath(req, signer)
	want := "/blogs/shared/blog/5?b=1&a=2"
	if got != want {
		t.Fatalf("expected path %q got %q", want, got)
	}
}

func TestSignatureStyleFromRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/private/shared/topic/1/thread/2?ts=1&sig=abc", nil)
	if got := SignatureStyleFromRequest(req); got != SignatureStyleQuery {
		t.Fatalf("expected query signature style, got %q", got)
	}

	pathReq := httptest.NewRequest(http.MethodGet, "/private/shared/topic/1/thread/2/ts/1/sign/abc", nil)
	pathReq = mux.SetURLVars(pathReq, map[string]string{"ts": "1", "sign": "abc"})
	if got := SignatureStyleFromRequest(pathReq); got != SignatureStylePath {
		t.Fatalf("expected path signature style, got %q", got)
	}
}

func TestMakeImageURLWithPathSignature(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.HTTPHostname = "http://example.com"
	signer := sharesign.NewSigner(cfg, "secret")

	exp := time.Unix(1000, 0)
	got := MakeImageURLWithStyle(cfg.HTTPHostname, "Title", signer, exp, SignatureStylePath, "")
	if !strings.Contains(got, "/api/og-image/ts/1000/sign/") {
		t.Fatalf("expected path signature in url, got %q", got)
	}
}
