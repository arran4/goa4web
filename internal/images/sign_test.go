package images

import (
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
)

func TestMapURLUploading(t *testing.T) {
	signer := NewSigner(&config.RuntimeConfig{}, "k")
	got := signer.MapURL("img", "uploading:abc")
	if got != "uploading:abc" {
		t.Fatalf("expected placeholder unchanged, got %s", got)
	}
}

func TestMapURLImage(t *testing.T) {
	cfg := &config.RuntimeConfig{HTTPHostname: "http://mysite"}
	signer := NewSigner(cfg, "k")
	got := signer.MapURL("img", "image:foo")
	if got == "image:foo" || !strings.Contains(got, "/images/image/") {
		t.Fatalf("expected signed image url, got %s", got)
	}
}

func TestSignedURLTTL(t *testing.T) {
	cfg := &config.RuntimeConfig{HTTPHostname: "http://example.com"}
	signer := NewSigner(cfg, "k")
	ttl := 2 * time.Hour
	surl := signer.SignedURLTTL("img123.jpg", ttl)
	if !strings.Contains(surl, "ts=") || !strings.Contains(surl, "sig=") {
		t.Fatalf("missing signature params in %s", surl)
	}
	u, err := url.Parse(surl)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	ts := u.Query().Get("ts")
	sig := u.Query().Get("sig")
	if !signer.Verify("image:"+"img123.jpg", ts, sig) {
		t.Fatalf("signature did not verify")
	}
	exp, _ := strconv.ParseInt(ts, 10, 64)
	if exp < time.Now().Add(ttl-time.Minute).Unix() || exp > time.Now().Add(ttl+time.Minute).Unix() {
		t.Fatalf("expiry not roughly ttl: %d", exp)
	}
}
