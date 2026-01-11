package sharesign_test

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/sharesign"
)

func TestSigner(t *testing.T) {
	cfg := &config.RuntimeConfig{
		HTTPHostname: "http://localhost:8080",
	}
	s := sharesign.NewSigner(cfg, "secret")
	link := "/news/news/1"
	ts, sig := s.Sign(link)
	if !s.Verify(link, fmt.Sprint(ts), sig) {
		t.Errorf("Verify failed")
	}
	if s.Verify(link, fmt.Sprint(ts), "invalid") {
		t.Errorf("Verify succeeded with invalid signature")
	}
	if s.Verify("invalid", fmt.Sprint(ts), sig) {
		t.Errorf("Verify succeeded with invalid link")
	}
	if s.Verify(link, fmt.Sprint(ts+1), sig) {
		t.Errorf("Verify succeeded with invalid timestamp")
	}
	if s.Verify(link, fmt.Sprint(time.Now().Add(-48*time.Hour).Unix()), sig) {
		t.Errorf("Verify succeeded with expired timestamp")
	}
}

func TestSignedURL(t *testing.T) {
	cfg := &config.RuntimeConfig{
		HTTPHostname: "http://localhost:8080",
	}
	s := sharesign.NewSigner(cfg, "secret")

	// Test Path based (default)
	link := "/private/topic/1/thread/2"
	signed := s.SignedURL(link)
	// expected: http://localhost:8080/private/shared/topic/1/thread/2/ts/.../sign/...
	if !strings.Contains(signed, "/private/shared/topic/1/thread/2/ts/") {
		t.Errorf("Path signature format incorrect: %s", signed)
	}
	if !strings.Contains(signed, "/sign/") {
		t.Errorf("Path signature missing sign: %s", signed)
	}

	// Test Query based
	signedQuery := s.SignedURLQuery(link)
	if !strings.Contains(signedQuery, "/private/shared/topic/1/thread/2?ts=") {
		t.Errorf("Query signature format incorrect: %s", signedQuery)
	}
}

func TestSignedURLQueryWithParams(t *testing.T) {
	cfg := &config.RuntimeConfig{
		HTTPHostname: "http://localhost:8080",
	}
	s := sharesign.NewSigner(cfg, "secret")
	link := "/private/topic/1/thread/2?from=share"
	signed := s.SignedURLQuery(link)
	parsed, err := url.Parse(signed)
	if err != nil {
		t.Fatalf("parse signed url: %v", err)
	}
	ts := parsed.Query().Get("ts")
	sig := parsed.Query().Get("sig")
	query := parsed.Query()
	query.Del("ts")
	query.Del("sig")
	dataPath := parsed.Path
	if encoded := query.Encode(); encoded != "" {
		dataPath = dataPath + "?" + encoded
	}
	if !s.Verify(dataPath, ts, sig) {
		t.Fatalf("signature did not verify for %s", signed)
	}
}
