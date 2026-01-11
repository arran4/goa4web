package share

import (
	"net/http"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/sharesign"
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
