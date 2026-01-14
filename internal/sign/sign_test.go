package sign_test

import (
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/internal/sign"
)

const testKey = "test-secret-key"

func TestSign_WithNonce(t *testing.T) {
	data := "test-data"
	nonce := "test-nonce-123"

	sig := sign.Sign(data, testKey, sign.WithNonce(nonce))
	if sig == "" {
		t.Fatal("signature should not be empty")
	}

	// Verify should succeed
	err := sign.Verify(data, sig, testKey, sign.WithNonce(nonce))
	if err != nil {
		t.Errorf("verify failed: %v", err)
	}

	// Wrong nonce should fail
	err = sign.Verify(data, sig, testKey, sign.WithNonce("wrong-nonce"))
	if err == nil {
		t.Error("verify should fail with wrong nonce")
	}

	// Wrong data should fail
	err = sign.Verify("wrong-data", sig, testKey, sign.WithNonce(nonce))
	if err == nil {
		t.Error("verify should fail with wrong data")
	}

	// Wrong key should fail
	err = sign.Verify(data, sig, "wrong-key", sign.WithNonce(nonce))
	if err == nil {
		t.Error("verify should fail with wrong key")
	}
}

func TestSign_WithExpiry(t *testing.T) {
	data := "test-data"
	expiry := time.Now().Add(1 * time.Hour)

	sig := sign.Sign(data, testKey, sign.WithExpiry(expiry))
	if sig == "" {
		t.Fatal("signature should not be empty")
	}

	// Verify should succeed
	err := sign.Verify(data, sig, testKey, sign.WithExpiry(expiry))
	if err != nil {
		t.Errorf("verify failed: %v", err)
	}

	// Expired signature should fail
	pastExpiry := time.Now().Add(-1 * time.Hour)
	pastSig := sign.Sign(data, testKey, sign.WithExpiry(pastExpiry))
	err = sign.Verify(data, pastSig, testKey, sign.WithExpiry(pastExpiry))
	if err == nil {
		t.Error("verify should fail with expired signature")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("error should mention expiry, got: %v", err)
	}
}

func TestSign_WithHostname(t *testing.T) {
	data := "test-data"
	hostname := "example.com"
	nonce := "nonce123"

	sig := sign.Sign(data, testKey, sign.WithHostname(hostname), sign.WithNonce(nonce))
	if sig == "" {
		t.Fatal("signature should not be empty")
	}

	// Verify with correct hostname should succeed
	err := sign.Verify(data, sig, testKey, sign.WithHostname(hostname), sign.WithNonce(nonce))
	if err != nil {
		t.Errorf("verify failed: %v", err)
	}

	// Verify without hostname should fail
	err = sign.Verify(data, sig, testKey, sign.WithNonce(nonce))
	if err == nil {
		t.Error("verify should fail without hostname")
	}

	// Verify with wrong hostname should fail
	err = sign.Verify(data, sig, testKey, sign.WithHostname("wrong.com"), sign.WithNonce(nonce))
	if err == nil {
		t.Error("verify should fail with wrong hostname")
	}
}

func TestAddQuerySig_WithNonce(t *testing.T) {
	baseURL := "http://example.com/path"
	nonce := "nonce123"
	sig := "abc123"

	result, err := sign.AddQuerySig(baseURL, sig, sign.WithNonce(nonce))
	if err != nil {
		t.Fatalf("AddQuerySig failed: %v", err)
	}

	if !strings.Contains(result, "sig="+sig) {
		t.Errorf("result should contain sig parameter: %s", result)
	}
	if !strings.Contains(result, "nonce="+nonce) {
		t.Errorf("result should contain nonce parameter: %s", result)
	}
	if !strings.HasPrefix(result, baseURL+"?") {
		t.Errorf("result should start with base URL: %s", result)
	}
}

func TestAddQuerySig_WithExpiry(t *testing.T) {
	baseURL := "http://example.com/path"
	expiry := time.Unix(1234567890, 0)
	sig := "abc123"

	result, err := sign.AddQuerySig(baseURL, sig, sign.WithExpiry(expiry))
	if err != nil {
		t.Fatalf("AddQuerySig failed: %v", err)
	}

	if !strings.Contains(result, "sig="+sig) {
		t.Errorf("result should contain sig parameter: %s", result)
	}
	if !strings.Contains(result, "ts=1234567890") {
		t.Errorf("result should contain ts parameter: %s", result)
	}
}

func TestAddQuerySig_PreservesExistingParams(t *testing.T) {
	baseURL := "http://example.com/path?foo=bar&baz=qux"
	nonce := "nonce123"
	sig := "abc123"

	result, err := sign.AddQuerySig(baseURL, sig, sign.WithNonce(nonce))
	if err != nil {
		t.Fatalf("AddQuerySig failed: %v", err)
	}

	if !strings.Contains(result, "foo=bar") {
		t.Errorf("result should preserve foo parameter: %s", result)
	}
	if !strings.Contains(result, "baz=qux") {
		t.Errorf("result should preserve baz parameter: %s", result)
	}
	if !strings.Contains(result, "sig="+sig) {
		t.Errorf("result should contain sig parameter: %s", result)
	}
	if !strings.Contains(result, "nonce="+nonce) {
		t.Errorf("result should contain nonce parameter: %s", result)
	}
}

func TestAddPathSig_WithNonce(t *testing.T) {
	baseURL := "http://example.com/path"
	nonce := "nonce123"
	sig := "abc123"

	result, err := sign.AddPathSig(baseURL, sig, sign.WithNonce(nonce))
	if err != nil {
		t.Fatalf("AddPathSig failed: %v", err)
	}

	expected := baseURL + "/nonce/" + nonce + "/sign/" + sig
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestAddPathSig_WithExpiry(t *testing.T) {
	baseURL := "http://example.com/path"
	expiry := time.Unix(1234567890, 0)
	sig := "abc123"

	result, err := sign.AddPathSig(baseURL, sig, sign.WithExpiry(expiry))
	if err != nil {
		t.Fatalf("AddPathSig failed: %v", err)
	}

	expected := baseURL + "/ts/1234567890/sign/" + sig
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestExtractQuerySig_WithNonce(t *testing.T) {
	urlStr := "http://example.com/path?foo=bar&nonce=nonce123&sig=abc123"

	cleanURL, sig, opts, err := sign.ExtractQuerySig(urlStr)
	if err != nil {
		t.Fatalf("ExtractQuerySig failed: %v", err)
	}

	if sig != "abc123" {
		t.Errorf("expected sig abc123, got %s", sig)
	}

	if len(opts) != 1 {
		t.Fatalf("expected 1 option, got %d", len(opts))
	}

	if nonce, ok := opts[0].(sign.WithNonce); !ok || string(nonce) != "nonce123" {
		t.Errorf("expected WithNonce(nonce123), got %v", opts[0])
	}

	if !strings.Contains(cleanURL, "foo=bar") {
		t.Errorf("cleanURL should preserve foo parameter: %s", cleanURL)
	}
	if strings.Contains(cleanURL, "nonce=") {
		t.Errorf("cleanURL should not contain nonce: %s", cleanURL)
	}
	if strings.Contains(cleanURL, "sig=") {
		t.Errorf("cleanURL should not contain sig: %s", cleanURL)
	}
}

func TestExtractQuerySig_WithExpiry(t *testing.T) {
	urlStr := "http://example.com/path?ts=1234567890&sig=abc123"

	cleanURL, sig, opts, err := sign.ExtractQuerySig(urlStr)
	if err != nil {
		t.Fatalf("ExtractQuerySig failed: %v", err)
	}

	if sig != "abc123" {
		t.Errorf("expected sig abc123, got %s", sig)
	}

	if len(opts) != 1 {
		t.Fatalf("expected 1 option, got %d", len(opts))
	}

	if expiry, ok := opts[0].(sign.WithExpiry); !ok || time.Time(expiry).Unix() != 1234567890 {
		t.Errorf("expected WithExpiry(1234567890), got %v", opts[0])
	}

	if strings.Contains(cleanURL, "ts=") {
		t.Errorf("cleanURL should not contain ts: %s", cleanURL)
	}
}

func TestExtractPathSig_WithNonce(t *testing.T) {
	path := "/api/og-image/data/nonce/nonce123/sign/abc123"
	pathVars := map[string]string{
		"nonce": "nonce123",
		"sign":  "abc123",
	}

	cleanPath, sig, opts, err := sign.ExtractPathSig(path, pathVars)
	if err != nil {
		t.Fatalf("ExtractPathSig failed: %v", err)
	}

	if cleanPath != "/api/og-image/data" {
		t.Errorf("expected /api/og-image/data, got %s", cleanPath)
	}

	if sig != "abc123" {
		t.Errorf("expected sig abc123, got %s", sig)
	}

	if len(opts) != 1 {
		t.Fatalf("expected 1 option, got %d", len(opts))
	}

	if nonce, ok := opts[0].(sign.WithNonce); !ok || string(nonce) != "nonce123" {
		t.Errorf("expected WithNonce(nonce123), got %v", opts[0])
	}
}

func TestExtractPathSig_WithExpiry(t *testing.T) {
	path := "/api/og-image/data/ts/1234567890/sign/abc123"
	pathVars := map[string]string{
		"ts":   "1234567890",
		"sign": "abc123",
	}

	cleanPath, sig, opts, err := sign.ExtractPathSig(path, pathVars)
	if err != nil {
		t.Fatalf("ExtractPathSig failed: %v", err)
	}

	if cleanPath != "/api/og-image/data" {
		t.Errorf("expected /api/og-image/data, got %s", cleanPath)
	}

	if sig != "abc123" {
		t.Errorf("expected sig abc123, got %s", sig)
	}

	if len(opts) != 1 {
		t.Fatalf("expected 1 option, got %d", len(opts))
	}

	if expiry, ok := opts[0].(sign.WithExpiry); !ok || time.Time(expiry).Unix() != 1234567890 {
		t.Errorf("expected WithExpiry(1234567890), got %v", opts[0])
	}
}

// End-to-end test combining Sign + AddQuerySig + ExtractQuerySig + Verify
func TestEndToEnd_Query(t *testing.T) {
	data := "/api/resource?id=123"
	nonce := "e2e-nonce"

	// Sign
	sig := sign.Sign(data, testKey, sign.WithNonce(nonce))

	// Add to URL
	signedURL, err := sign.AddQuerySig("http://example.com"+data, sig, sign.WithNonce(nonce))
	if err != nil {
		t.Fatalf("AddQuerySig failed: %v", err)
	}

	// Extract
	cleanURL, extractedSig, opts, err := sign.ExtractQuerySig(signedURL)
	if err != nil {
		t.Fatalf("ExtractQuerySig failed: %v", err)
	}

	// Verify using extracted data
	// Need to reconstruct the data path from cleanURL
	// cleanURL is full URL, we need just the path + query
	cleanData := strings.TrimPrefix(cleanURL, "http://example.com")
	err = sign.Verify(cleanData, extractedSig, testKey, opts...)
	if err != nil {
		t.Errorf("verify failed: %v", err)
	}
}

// End-to-end test with path-based signature
func TestEndToEnd_Path(t *testing.T) {
	data := "/api/resource"
	nonce := "e2e-nonce"

	// Sign
	sig := sign.Sign(data, testKey, sign.WithNonce(nonce))

	// Add to URL
	signedURL, err := sign.AddPathSig("http://example.com"+data, sig, sign.WithNonce(nonce))
	if err != nil {
		t.Fatalf("AddPathSig failed: %v", err)
	}

	// Simulate path vars extraction (what mux would do)
	pathVars := map[string]string{
		"nonce": nonce,
		"sign":  sig,
	}

	// Extract path from URL
	fullPath := strings.TrimPrefix(signedURL, "http://example.com")

	// Extract signature info
	cleanPath, extractedSig, opts, err := sign.ExtractPathSig(fullPath, pathVars)
	if err != nil {
		t.Fatalf("ExtractPathSig failed: %v", err)
	}

	// Verify
	err = sign.Verify(cleanPath, extractedSig, testKey, opts...)
	if err != nil {
		t.Errorf("verify failed: %v", err)
	}
}
