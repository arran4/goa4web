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

func TestSignedURLQueryParams(t *testing.T) {
	cfg := &config.RuntimeConfig{HTTPHostname: "http://example.com"}
	signer := NewSigner(cfg, "k")

	// Case 1: ID with existing params but no sig/ts
	idWithParams := "img123.jpg?foo=bar"
	surl := signer.SignedURLTTL(idWithParams, time.Hour)
	if strings.Count(surl, "?") > 1 {
		t.Errorf("url has multiple '?': %s", surl)
	}
	if !strings.Contains(surl, "&ts=") {
		t.Errorf("url missing chained ts param: %s", surl)
	}

	// Case 2: ID that looks like it's already signed
	// We construct a fake signed URL string as the ID
	idAlreadySigned := "img123.jpg?ts=12345&sig=abcde"
	surl2 := signer.SignedURLTTL(idAlreadySigned, time.Hour)
	if strings.Count(surl2, "?") > 1 {
		t.Errorf("url has multiple '?': %s", surl2)
	}
	if strings.Count(surl2, "ts=") > 1 || strings.Count(surl2, "sig=") > 1 {
		t.Errorf("url has duplicate sig/ts params: %s", surl2)
	}

	// Case 3: Just to be sure, verify parsing
	u, err := url.Parse(surl2)
	if err != nil {
		t.Errorf("result is not a valid url: %s", surl2)
	}
	q := u.Query()
	if len(q["ts"]) != 1 {
		t.Errorf("expected exactly 1 ts param, got %d", len(q["ts"]))
	}
}
