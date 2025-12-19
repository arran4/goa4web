package feedsign

import (
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestSigner_SignedURL_and_Verify(t *testing.T) {
	cfg := &config.RuntimeConfig{
		HTTPHostname: "http://example.com",
	}
	key := "secret-key"
	signer := NewSigner(cfg, key)

	path := "/blogs/rss"
	query := "rss=bob"
	username := "testuser"

	// Generate signed URL
	signedURL := signer.SignedURL(path, query, username)

	// Expect format: /blogs/u/testuser/rss?ts=...&sig=...&rss=bob
	expectedPrefix := "/blogs/u/testuser/rss?"
	if !strings.HasPrefix(signedURL, expectedPrefix) {
		t.Errorf("Expected URL to start with %s, got %s", expectedPrefix, signedURL)
	}
	if !strings.Contains(signedURL, "&rss=bob") {
		t.Errorf("Expected URL to contain query params, got %s", signedURL)
	}

	// Parse URL to extract ts and sig
	parts := strings.Split(signedURL, "?")
	if len(parts) != 2 {
		t.Fatalf("Invalid URL format: %s", signedURL)
	}
	queryParams := parts[1]

	var ts, sig string
	for _, pair := range strings.Split(queryParams, "&") {
		kv := strings.Split(pair, "=")
		if len(kv) == 2 {
			if kv[0] == "ts" {
				ts = kv[1]
			} else if kv[0] == "sig" {
				sig = kv[1]
			}
		}
	}

	if ts == "" || sig == "" {
		t.Fatalf("Missing ts or sig in URL: %s", signedURL)
	}

	// Verify
	if !signer.Verify(path, query, username, ts, sig) {
		t.Errorf("Verification failed for valid signature")
	}

	// Verify with wrong username
	if signer.Verify(path, query, "wronguser", ts, sig) {
		t.Errorf("Verification succeeded for wrong username")
	}

	// Verify with wrong path
	if signer.Verify("/blogs/atom", query, username, ts, sig) {
		t.Errorf("Verification succeeded for wrong path")
	}

	// Verify with wrong query
	if signer.Verify(path, "rss=alice", username, ts, sig) {
		t.Errorf("Verification succeeded for wrong query")
	}
}
