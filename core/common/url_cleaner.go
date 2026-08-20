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

// isSignedURL checks if the URL query contains parameters that indicate a signed URL.
func isSignedURL(q url.Values) bool {
	for k := range q {
		lower := strings.ToLower(k)
		for _, sigKey := range signatureKeys {
			if lower == sigKey {
				return true
			}
		}
	}
	return false
}

// CanonicalizeExternalURL removes known tracking parameters while preserving functional ones,
// unknown parameters, and cryptographically signed URLs.
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

	q := u.Query()
	if len(q) == 0 {
		return raw
	}

	// If the URL has a signature parameter, preserve query string intact to avoid invalidating the signature
	if isSignedURL(q) {
		return raw
	}

	modified := false
	for k := range q {
		if isTrackingParam(k) {
			q.Del(k)
			modified = true
		}
	}

	if !modified {
		return raw
	}

	if len(q) == 0 {
		u.RawQuery = ""
	} else {
		u.RawQuery = q.Encode()
	}
	return u.String()
}
