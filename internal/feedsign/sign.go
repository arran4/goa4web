package feedsign

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/sign"
)

// Signer signs feed links so they can be accessed with authentication in a read-only manner.
type Signer struct {
	cfg    *config.RuntimeConfig
	signer *sign.Signer
}

// NewSigner returns a Signer using cfg for hostname resolution and key for HMAC.
func NewSigner(cfg *config.RuntimeConfig, key string) *Signer {
	return &Signer{cfg: cfg, signer: &sign.Signer{Key: key}}
}

// SignedURL generates a URL for the given feed path, query params, and username.
// path should be the base path (e.g. "/blogs/rss").
// query should be the encoded query string (e.g. "rss=bob"), or empty.
// The resulting URL will be "/{section}/u/{username}/{rest}?ts={ts}&sig={sig}&{query}"
func (s *Signer) SignedURL(path, query, username string, ops ...any) string {
	data := fmt.Sprintf("feed:%s:%s", username, path)
	if query != "" {
		data += "?" + query
	}

	// Set default option if none provided
	if len(ops) == 0 {
		ops = append(ops, sign.WithOutNonce())
	}

	sig := s.signer.Sign(data, ops...)

	// Check if path ends with slash, if so remove it to avoid double slashes
	path = strings.TrimSuffix(path, "/")

	parts := strings.Split(path, "/")
	if len(parts) > 1 {
		// Inject /u/{username} after the first segment (section)
		// parts[0] is empty for absolute paths
		newPath := fmt.Sprintf("/%s/u/%s", parts[1], url.QueryEscape(username))
		if len(parts) > 2 {
			newPath += "/" + strings.Join(parts[2:], "/")
		}
		path = newPath
	} else {
		// Fallback for root paths
		path = fmt.Sprintf("/u/%s%s", url.QueryEscape(username), path)
	}

	res := fmt.Sprintf("%s?sig=%s", path, sig)
	if query != "" {
		res += "&" + query
	}
	return res
}

// Verify checks the provided signature matches the username, path, and query.
func (s *Signer) Verify(path, query, username, ts, sig string) bool {
	// Reconstruct the data string: feed:{username}:{path}?{query}
	data := fmt.Sprintf("feed:%s:%s", username, path)
	if query != "" {
		data += "?" + query
	}
	valid, _ := s.signer.Verify(data, sig, sign.WithExpiryTimestamp(ts))
	return valid
}

// StripSignatureParams removes ts and sig from values and returns encoded string.
func StripSignatureParams(v url.Values) string {
	c := url.Values{}
	for k, val := range v {
		if k != "ts" && k != "sig" {
			c[k] = val
		}
	}
	return c.Encode()
}
