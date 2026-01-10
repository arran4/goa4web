package linksign

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/sign"
)

// Signer signs external links so they are redirected through a confirmation page.
type Signer struct {
	cfg    *config.RuntimeConfig
	signer *sign.Signer
}

// NewSigner returns a Signer using cfg for hostname resolution and key for HMAC.
func NewSigner(cfg *config.RuntimeConfig, key string, expiry ...any) *Signer {
	var e time.Duration
	if len(expiry) > 0 {
		if v, ok := expiry[0].(time.Duration); ok {
			e = v
		}
	}
	return &Signer{cfg: cfg, signer: &sign.Signer{Key: key, DefaultExpiry: e}}
}

// SignedURL generates a redirect URL for the given link.
func (s *Signer) SignedURL(link string, exp ...time.Time) string {
	host := strings.TrimSuffix(s.cfg.HTTPHostname, "/")
	ts, sig := s.signer.Sign("link:"+link, exp...)
	return fmt.Sprintf("%s/goto?u=%s&ts=%d&sig=%s", host, url.QueryEscape(link), ts, sig)
}

// Sign returns the timestamp and signature for link using the optional expiry time.
func (s *Signer) Sign(link string, exp ...time.Time) (int64, string) {
	return s.signer.Sign("link:"+link, exp...)
}

// Verify checks the provided signature matches the link.
func (s *Signer) Verify(link, ts, sig string) bool {
	return s.signer.Verify("link:"+link, ts, sig)
}

// MapURL rewrites outbound links to SignedURL when the host is external.
func (s *Signer) MapURL(tag, val string) string {
	if tag != "a" {
		return val
	}
	if strings.HasPrefix(val, "http://") || strings.HasPrefix(val, "https://") {
		u, err := url.Parse(val)
		if err != nil {
			return val
		}
		allowed := false
		for _, h := range strings.Fields(s.cfg.HTTPHostname) {
			h = strings.TrimSpace(h)
			if h == "" {
				continue
			}
			if pu, err := url.Parse(h); err == nil && pu.Host != "" {
				h = pu.Host
			} else {
				h = strings.TrimSuffix(h, "/")
			}
			if strings.EqualFold(h, u.Host) {
				allowed = true
				break
			}
		}
		if !allowed {
			return s.SignedURL(val)
		}
	}
	return val
}
