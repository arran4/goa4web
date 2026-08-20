package common

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"
)

func GenerateURLHash(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])
}

// CanonicalizeExternalURL removes known tracking parameters while preserving functional ones.
func CanonicalizeExternalURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" {
		return raw
	}

	q := u.Query()
	modified := false
	for k := range q {
		lowerK := strings.ToLower(k)
		if strings.HasPrefix(lowerK, "utm_") || lowerK == "click_id" || lowerK == "yclid" || lowerK == "fbclid" || lowerK == "gclid" {
			q.Del(k)
			modified = true
		}
	}
	if modified {
		u.RawQuery = q.Encode()
	}
	return u.String()
}
