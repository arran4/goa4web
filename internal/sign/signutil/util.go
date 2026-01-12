package signutil

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	"github.com/arran4/goa4web/internal/sign"
)

// SignAndAddQuery signs data and adds the signature to the URL as query parameters.
// If the URL already has query params, they are preserved.
// The data should typically be the path + un-signed query params.
func SignAndAddQuery(urlStr string, data string, key string, opts ...sign.SignOption) (string, error) {
	sig := sign.Sign(data, key, opts...)
	return sign.AddQuerySig(urlStr, sig, opts...)
}

// SignAndAddPath signs data and adds the signature to the URL path.
// The data should typically be just the path without signature parts.
func SignAndAddPath(urlStr string, data string, key string, opts ...sign.SignOption) (string, error) {
	sig := sign.Sign(data, key, opts...)
	return sign.AddPathSig(urlStr, sig, opts...)
}

// VerifyQueryURL extracts signature from query params and verifies it.
// Returns the clean data (path + non-auth query params) if valid, error otherwise.
func VerifyQueryURL(urlStr string, key string) (string, error) {
	cleanURL, sig, opts, err := sign.ExtractQuerySig(urlStr)
	if err != nil {
		return "", fmt.Errorf("extract query sig: %w", err)
	}

	if sig == "" {
		return "", fmt.Errorf("missing signature")
	}

	// Parse clean URL to get path + query for verification
	u, err := url.Parse(cleanURL)
	if err != nil {
		return "", fmt.Errorf("parse clean url: %w", err)
	}

	data := u.Path
	if u.RawQuery != "" {
		data += "?" + u.RawQuery
	}

	if err := sign.Verify(data, sig, key, opts...); err != nil {
		return "", fmt.Errorf("verify: %w", err)
	}

	return data, nil
}

// VerifyPathURL extracts signature from path vars and verifies it.
// pathVars should contain the mux variables from the request.
// Returns the clean path if valid, error otherwise.
func VerifyPathURL(fullPath string, pathVars map[string]string, key string, additionalQuery string) (string, error) {
	cleanPath, sig, opts, err := sign.ExtractPathSig(fullPath, pathVars)
	if err != nil {
		return "", fmt.Errorf("extract path sig: %w", err)
	}

	if sig == "" {
		return "", fmt.Errorf("missing signature")
	}

	// Build data including additional query params if present
	data := cleanPath
	if additionalQuery != "" {
		data += "?" + additionalQuery
	}

	if err := sign.Verify(data, sig, key, opts...); err != nil {
		return "", fmt.Errorf("verify: %w", err)
	}

	return data, nil
}

// InjectShared injects "/shared" after the first path segment (module name).
// e.g., "/private/topic/2/thread/1" → "/private/shared/topic/2/thread/1"
func InjectShared(path string) string {
	parts := strings.SplitN(path, "/", 3)
	if len(parts) >= 3 && parts[0] == "" && parts[1] != "" {
		// Avoid double injection if "shared" is already the next segment
		if strings.HasPrefix(parts[2], "shared/") || parts[2] == "shared" {
			return path
		}
		// parts: ["", "private", "topic/2/thread/1"]
		return "/" + parts[1] + "/shared/" + parts[2]
	}
	return path
}

// GenerateNonce creates a random hex-encoded nonce.
func GenerateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
