package images

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/sign"
)

// Signer generates and verifies image signatures.
type Signer struct {
	cfg    *config.RuntimeConfig
	signer *sign.Signer
}

// NewSigner returns a Signer using cfg for hostname resolution and key for HMAC.
func NewSigner(cfg *config.RuntimeConfig, key string) *Signer {
	return &Signer{cfg: cfg, signer: &sign.Signer{Key: key}}
}

func (s *Signer) defaultExpiry() time.Time { return time.Now().Add(24 * time.Hour) }

// SignedURL maps an image identifier to a signed URL.
func (s *Signer) SignedURL(id string) string {
	return s.SignedURLTTL(id, 24*time.Hour)
}

// SignedURLTTL maps an image identifier to a signed URL that expires after ttl.
func (s *Signer) SignedURLTTL(id string, ttl time.Duration) string {
	id = strings.TrimPrefix(strings.TrimPrefix(id, "image:"), "img:")
	host := strings.TrimSuffix(s.cfg.HTTPHostname, "/")
	cleanId, sep := s.cleanParams(id)
	ts, sig := s.signer.Sign("image:"+cleanId, time.Now().Add(ttl))
	return fmt.Sprintf("%s/images/image/%s%sts=%d&sig=%s", host, cleanId, sep, ts, sig)
}

func (s *Signer) cleanParams(id string) (string, string) {
	u, err := url.Parse(id)
	if err != nil {
		if strings.Contains(id, "?") {
			return id, "&"
		}
		return id, "?"
	}
	q := u.Query()
	if q.Has("ts") || q.Has("sig") {
		log.Printf("[Signer] Double signing detected for ID %q. Removing existing ts/sig params.", id)
		q.Del("ts")
		q.Del("sig")
	}
	u.RawQuery = q.Encode()
	cleanId := u.String()
	if strings.Contains(cleanId, "?") {
		return cleanId, "&"
	}
	return cleanId, "?"
}

// SignedCacheURL maps a cache identifier to a signed URL.
func (s *Signer) SignedCacheURL(id string) string {
	return s.SignedCacheURLTTL(id, 24*time.Hour)
}

// SignedCacheURLTTL maps a cache identifier to a signed URL that expires after ttl.
func (s *Signer) SignedCacheURLTTL(id string, ttl time.Duration) string {
	host := strings.TrimSuffix(s.cfg.HTTPHostname, "/")
	cleanId, sep := s.cleanParams(id)
	ts, sig := s.signer.Sign("cache:"+cleanId, time.Now().Add(ttl))
	return fmt.Sprintf("%s/images/cache/%s%sts=%d&sig=%s", host, cleanId, sep, ts, sig)
}

// Verify checks the provided signature matches data.
func (s *Signer) Verify(data, ts, sig string) bool { return s.signer.Verify(data, ts, sig) }

// SignedRef appends a signature to an image or cache reference.
// The input should start with "image:", "img:", or "cache:".
func (s *Signer) SignedRef(ref string) string {
	var prefix, id string
	switch {
	case strings.HasPrefix(ref, "image:"):
		prefix = "image:"
		id = strings.TrimPrefix(ref, "image:")
	case strings.HasPrefix(ref, "img:"):
		prefix = "image:"
		id = strings.TrimPrefix(ref, "img:")
	case strings.HasPrefix(ref, "cache:"):
		prefix = "cache:"
		id = strings.TrimPrefix(ref, "cache:")
	default:
		return ref
	}

	cleanId, sep := s.cleanParams(id)
	ts, sig := s.signer.Sign(prefix + cleanId)
	return fmt.Sprintf("%s%s%sts=%d&sig=%s", prefix, cleanId, sep, ts, sig)
}

// MapURL converts image references to signed HTTP URLs.
func (s *Signer) MapURL(tag, val string) string {
	if tag != "img" {
		return val
	}
	switch {
	case strings.HasPrefix(val, "uploading:"):
		return val
	case strings.HasPrefix(val, "image:") || strings.HasPrefix(val, "img:"):
		return s.SignedURL(val)
	case strings.HasPrefix(val, "cache:"):
		return s.SignedCacheURL(strings.TrimPrefix(val, "cache:"))
	default:
		return val
	}
}
