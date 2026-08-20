package common

import (
	"net/url"
	"strings"
)

// signatureKeys defines query parameter names indicating a cryptographic signature
// or signed URL that would be broken if query parameters are removed or reordered.
var signatureKeys = []string{
	"x-amz-signature",
	"x-amz-credential",
	"signature",
	"sig",
	"hash",
	"hmac",
	"x-goog-signature",
	"x-ms-signature",
}

// knownTrackingPrefixes defines prefixes of query parameters used strictly for tracking/analytics.
var knownTrackingPrefixes = []string{
	"utm_",
}

// knownTrackingExactKeys defines exact parameter names used strictly for tracking/analytics.
var knownTrackingExactKeys = map[string]bool{
	"fbclid":   true,
	"gclid":    true,
	"gbraid":   true,
	"wbraid":   true,
	"mc_cid":   true,
	"mc_eid":   true,
	"igshid":   true,
	"msclkid":  true,
	"twclid":   true,
	"yclid":    true,
	"click_id": true,
	"clickid":  true,
	"_hsenc":   true,
	"_hsmi":    true,
	"mkt_tok":  true,
}

// isTrackingParam checks if a query parameter key is a known tracking parameter.
func isTrackingParam(key string) bool {
	lower := strings.ToLower(key)
	for _, prefix := range knownTrackingPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return knownTrackingExactKeys[lower]
}

// isSignatureParam checks if a query parameter key indicates a cryptographic signature.
func isSignatureParam(key string) bool {
	lower := strings.ToLower(key)
	for _, sigKey := range signatureKeys {
		if lower == sigKey {
			return true
		}
	}
	return false
}

// CanonicalizeExternalURL removes known tracking parameters while preserving functional ones,
// original parameter order and escaping, repeated parameters, blank values, and cryptographically signed URLs.
func CanonicalizeExternalURL(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return raw
	}

	if u.RawQuery == "" {
		return raw
	}

	// Split RawQuery into components preserving exact original order and encoding
	parts := strings.Split(u.RawQuery, "&")
	var retained []string
	modified := false

	// First check if any query parameter is a signature indicator.
	// If so, preserve the URL completely byte-for-byte to avoid invalidating the signature.
	for _, part := range parts {
		if part == "" {
			continue
		}
		key := part
		if idx := strings.IndexByte(part, '='); idx != -1 {
			key = part[:idx]
		}
		unescapedKey, err := url.QueryUnescape(key)
		if err != nil {
			unescapedKey = key
		}
		if isSignatureParam(unescapedKey) {
			return raw
		}
	}

	// Filter out tracking parameters while keeping everything else exactly intact
	for _, part := range parts {
		if part == "" {
			continue
		}
		key := part
		if idx := strings.IndexByte(part, '='); idx != -1 {
			key = part[:idx]
		}
		unescapedKey, err := url.QueryUnescape(key)
		if err != nil {
			unescapedKey = key
		}
		if isTrackingParam(unescapedKey) {
			modified = true
		} else {
			retained = append(retained, part)
		}
	}

	if !modified {
		return raw
	}

	if len(retained) == 0 {
		u.RawQuery = ""
	} else {
		u.RawQuery = strings.Join(retained, "&")
	}
	return u.String()
}
