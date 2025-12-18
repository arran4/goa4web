package feedsign

import (
	"fmt"
	"net/url"
	"strings"
	"time"

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
// The resulting URL will be "/{section}/private/{username}/{rest}?ts={ts}&sig={sig}&{query}"
func (s *Signer) SignedURL(path, query, username string, exp ...time.Time) string {
	data := fmt.Sprintf("feed:%s:%s", username, path)
	if query != "" {
		data += "?" + query
	}
	ts, sig := s.signer.Sign(data, exp...)
	// Check if path ends with slash, if so remove it to avoid double slashes
	path = strings.TrimSuffix(path, "/")

	parts := strings.Split(path, "/")
	if len(parts) > 1 {
		// Inject /private/{username} after the first segment (section)
		// parts[0] is empty for absolute paths
		newPath := fmt.Sprintf("/%s/private/%s", parts[1], url.QueryEscape(username))
		if len(parts) > 2 {
			newPath += "/" + strings.Join(parts[2:], "/")
		}
		path = newPath
	} else {
		// Fallback for root paths
		path = fmt.Sprintf("/private/%s%s", url.QueryEscape(username), path)
	}

	res := fmt.Sprintf("%s?ts=%d&sig=%s", path, ts, sig)
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
	return s.signer.Verify(data, ts, sig)
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
