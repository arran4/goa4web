package signutil

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/sign"
	"github.com/gorilla/mux"
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

// SignedData contains the result of a signature verification
type SignedData struct {
	Valid bool
}

// GetSignedData extracts and verifies signature from the request using the signer
func GetSignedData(r *http.Request, key string) (*SignedData, error) {
	// Try query params first
	if r.URL.Query().Get("sig") != "" {
		cleanURL, sig, opts, err := sign.ExtractQuerySig(r.URL.String())
		if err != nil {
			return &SignedData{Valid: false}, nil
		}

		vars := mux.Vars(r)
		tsPath := vars["ts"]
		noncePath := vars["nonce"]

		u, err := url.Parse(cleanURL)
		if err != nil {
			return &SignedData{Valid: false}, nil
		}
		cleanPath := u.Path

		if tsPath != "" {
			if ts, err := strconv.ParseInt(tsPath, 10, 64); err == nil {
				opts = append(opts, sign.WithExpiry(time.Unix(ts, 0)))
				cleanPath = strings.Replace(cleanPath, "/ts/"+tsPath, "", 1)
			}
		}
		if noncePath != "" {
			opts = append(opts, sign.WithNonce(noncePath))
			cleanPath = strings.Replace(cleanPath, "/nonce/"+noncePath, "", 1)
		}

		data := cleanPath
		if u.RawQuery != "" {
			data += "?" + u.RawQuery
		}

		if err := sign.Verify(data, sig, key, opts...); err != nil {
			return &SignedData{Valid: false}, nil
		}

		return &SignedData{Valid: true}, nil
	}

	// Try path params
	vars := mux.Vars(r)
	if vars["sig"] != "" || vars["sign"] != "" {
		// Handle path verification
		// We need to reconstruct what the full path was including query params if any
		cleanPath, sig, opts, err := sign.ExtractPathSig(r.URL.Path, vars)
		if err != nil {
			return &SignedData{Valid: false}, nil
		}

		if sig == "" {
			return &SignedData{Valid: false}, nil
		}

		// Reconstruct data to be verified
		data := cleanPath
		q := r.URL.Query()
		q.Del("sig")
		q.Del("nonce")
		q.Del("ts")
		if encoded := q.Encode(); encoded != "" {
			data += "?" + encoded
		}

		if err := sign.Verify(data, sig, key, opts...); err != nil {
			return &SignedData{Valid: false}, nil
		}

		return &SignedData{Valid: true}, nil
	}

	return &SignedData{Valid: false}, nil
}
