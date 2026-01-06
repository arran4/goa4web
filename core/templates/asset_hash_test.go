package templates

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestGetAssetHash(t *testing.T) {
	// Test with a known asset
	path := "/static/site.js"
	hashedPath := GetAssetHash(path)

	if !strings.HasPrefix(hashedPath, path+"?v=") {
		t.Errorf("Expected hashed path to start with %s?v=, got %s", path, hashedPath)
	}

	// Verify hash correctness
	content, err := getAssetContent("site.js", "")
	if err != nil {
		t.Fatalf("Failed to read asset content: %v", err)
	}
	sum := sha256.Sum256(content)
	expectedHash := hex.EncodeToString(sum[:])[:16]

	if !strings.HasSuffix(hashedPath, expectedHash) {
		t.Errorf("Expected hash suffix %s, got %s", expectedHash, hashedPath)
	}

	// Test with non-existent asset
	badPath := "/static/nonexistent.js"
	hashedBadPath := GetAssetHash(badPath)
	if hashedBadPath != badPath {
		t.Errorf("Expected path %s to be returned unchanged, got %s", badPath, hashedBadPath)
	}
}
