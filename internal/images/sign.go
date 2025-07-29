package images

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/config"
)

// Signer generates and verifies image signatures without relying on globals.
type Signer struct {
	cfg *config.RuntimeConfig
	key string
}

// NewSigner returns a Signer using cfg for hostname resolution and key for HMAC.
func NewSigner(cfg *config.RuntimeConfig, key string) *Signer {
	return &Signer{cfg: cfg, key: key}
}

func (s *Signer) sign(data string, exp time.Time) (int64, string) {
	expires := exp.Unix()
	mac := hmac.New(sha256.New, []byte(s.key))
	io.WriteString(mac, fmt.Sprintf("%s:%d", data, expires))
	return expires, hex.EncodeToString(mac.Sum(nil))
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
	ts, sig := s.sign("image:"+id, time.Now().Add(ttl))
	return fmt.Sprintf("%s/images/image/%s?ts=%d&sig=%s", host, id, ts, sig)
}

// SignedCacheURL maps a cache identifier to a signed URL.
func (s *Signer) SignedCacheURL(id string) string {
	return s.SignedCacheURLTTL(id, 24*time.Hour)
}

// SignedCacheURLTTL maps a cache identifier to a signed URL that expires after ttl.
func (s *Signer) SignedCacheURLTTL(id string, ttl time.Duration) string {
	host := strings.TrimSuffix(s.cfg.HTTPHostname, "/")
	ts, sig := s.sign("cache:"+id, time.Now().Add(ttl))
	return fmt.Sprintf("%s/images/cache/%s?ts=%d&sig=%s", host, id, ts, sig)
}

func (s *Signer) Verify(data, tsStr, sig string) bool {
	exp, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return false
	}
	mac := hmac.New(sha256.New, []byte(s.key))
	io.WriteString(mac, fmt.Sprintf("%s:%d", data, exp))
	want := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(want), []byte(sig))
}

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
	ts, sig := s.sign(prefix+id, s.defaultExpiry())
	return fmt.Sprintf("%s%s?ts=%d&sig=%s", prefix, id, ts, sig)
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
