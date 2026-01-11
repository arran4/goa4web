package sharesign_test

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/sharesign"
	"github.com/arran4/goa4web/internal/sign"
)

func TestSigner(t *testing.T) {
	cfg := &config.RuntimeConfig{
		HTTPHostname: "http://localhost:8080",
	}
	s := sharesign.NewSigner(cfg, "secret")
	link := "/news/news/1"
	sig := s.Sign(link, sign.WithNonce("testnonce"))

	if valid, err := s.Verify(link, sig, sign.WithNonce("testnonce")); !valid || err != nil {
		t.Errorf("Verify failed: %v", err)
	}
	if valid, _ := s.Verify(link, "invalid", sign.WithNonce("testnonce")); valid {
		t.Errorf("Verify succeeded with invalid signature")
	}
	if valid, _ := s.Verify("invalid", sig, sign.WithNonce("testnonce")); valid {
		t.Errorf("Verify succeeded with invalid link")
	}
	if valid, _ := s.Verify(link, sig, sign.WithNonce("wrongnonce")); valid {
		t.Errorf("Verify succeeded with wrong nonce")
	}
	// Test expiry
	oldTime := time.Now().Add(-48 * time.Hour)
	oldSig := s.Sign(link, sign.WithExpiry(oldTime))

	if valid, err := s.Verify(link, oldSig, sign.WithExpiryTime(oldTime)); valid {
		t.Errorf("Verify succeeded with expired timestamp")
	} else if err == nil || !strings.Contains(err.Error(), "expired") {
		// It should fail with expired error
		t.Logf("Verify failed as expected but error was: %v", err)
	}
}

func TestSignedURL(t *testing.T) {
	cfg := &config.RuntimeConfig{
		HTTPHostname: "http://localhost:8080",
	}
	s := sharesign.NewSigner(cfg, "secret")

	// Test Path based (default)
	link := "/private/topic/1/thread/2"
	signed, err := s.SignedURL(link)
	if err != nil {
		t.Fatalf("SignedURL error: %v", err)
	}
	// expected: http://localhost:8080/private/shared/topic/1/thread/2/ts/.../sign/...
	if !strings.Contains(signed, "/private/shared/topic/1/thread/2/nonce/") {
		t.Errorf("Path signature format incorrect: %s", signed)
	}
	if !strings.Contains(signed, "/sign/") {
		t.Errorf("Path signature missing sign: %s", signed)
	}

	// Test Query based
	signedQuery, err := s.SignedURLQuery(link)
	if err != nil {
		t.Fatalf("SignedURLQuery error: %v", err)
	}
	if !strings.Contains(signedQuery, "/private/shared/topic/1/thread/2?nonce=") {
		t.Errorf("Query signature format incorrect: %s", signedQuery)
	}
}

func TestSignedURLQueryWithParams(t *testing.T) {
	cfg := &config.RuntimeConfig{
		HTTPHostname: "http://localhost:8080",
	}
	s := sharesign.NewSigner(cfg, "secret")
	link := "/private/topic/1/thread/2?from=share"
	signed, err := s.SignedURLQuery(link)
	if err != nil {
		t.Fatalf("SignedURLQueryURL error: %v", err)
	}
	parsed, err := url.Parse(signed)
	if err != nil {
		t.Fatalf("parse signed url: %v", err)
	}
	ts := parsed.Query().Get("ts")
	sig := parsed.Query().Get("sig")
	nonce := parsed.Query().Get("nonce")
	query := parsed.Query()
	query.Del("ts")
	query.Del("sig")
	query.Del("nonce")
	dataPath := parsed.Path
	if encoded := query.Encode(); encoded != "" {
		dataPath = dataPath + "?" + encoded
	}
	// Verify needs to strip "share:" prefix? No, Signer.Verify adds it.
	// We pass dataPath which is "/private/shared/..." (injected).
	// But `prepareSharedLink` injected "shared".
	// s.Verify adds "share:" prefix.
	// Wait, TestSignedURLQueryWithParams verification logic:
	// `dataPath` extracted from URL is correct path.
	// `s.Verify(dataPath...)`.
	var valid bool
	if nonce != "" {
		valid, err = s.Verify(dataPath, sig, sign.WithNonce(nonce))
	} else {
		valid, err = s.Verify(dataPath, sig, sign.WithExpiryTimestamp(ts))
	}
	if !valid || err != nil {
		t.Fatalf("signature did not verify for %s. Err: %v", signed, err)
	}
}
