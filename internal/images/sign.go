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

// ImageSigner signs and verifies image URLs.
type ImageSigner struct{ key []byte }

// NewSigner creates a new ImageSigner.
func NewSigner(k string) *ImageSigner { return &ImageSigner{key: []byte(k)} }

func (s *ImageSigner) sign(data string) (int64, string) {
	expires := time.Now().Add(24 * time.Hour).Unix()
	mac := hmac.New(sha256.New, s.key)
	io.WriteString(mac, fmt.Sprintf("%s:%d", data, expires))
	return expires, hex.EncodeToString(mac.Sum(nil))
}

// SignedURL maps an image identifier to a signed URL.
func (s *ImageSigner) SignedURL(id string) string {
	id = strings.TrimPrefix(strings.TrimPrefix(id, "image:"), "img:")
	host := strings.TrimSuffix(config.AppRuntimeConfig.HTTPHostname, "/")
	ts, sig := s.sign("image:" + id)
	return fmt.Sprintf("%s/images/image/%s?ts=%d&sig=%s", host, id, ts, sig)
}

// SignedCacheURL maps a cache identifier to a signed URL.
func (s *ImageSigner) SignedCacheURL(id string) string {
	host := strings.TrimSuffix(config.AppRuntimeConfig.HTTPHostname, "/")
	ts, sig := s.sign("cache:" + id)
	return fmt.Sprintf("%s/images/cache/%s?ts=%d&sig=%s", host, id, ts, sig)
}

func (s *ImageSigner) Verify(data, tsStr, sig string) bool {
	exp, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return false
	}
	mac := hmac.New(sha256.New, s.key)
	io.WriteString(mac, fmt.Sprintf("%s:%d", data, exp))
	want := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(want), []byte(sig))
}

// SignedRef appends a signature to an image or cache reference.
// The input should start with "image:", "img:", or "cache:".
func (s *ImageSigner) SignedRef(ref string) string {
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
	ts, sig := s.sign(prefix + id)
	return fmt.Sprintf("%s%s?ts=%d&sig=%s", prefix, id, ts, sig)
}

// MapURL converts image references to signed HTTP URLs.
func (s *ImageSigner) MapURL(tag, val string) string {
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
