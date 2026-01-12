package linksign

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/sign"
)

// Signer signs external links so they are redirected through a confirmation page.
type Signer struct {
	cfg    *config.RuntimeConfig
	signer *sign.Signer
}

// NewSigner returns a Signer using cfg for hostname resolution and key for HMAC.
func NewSigner(cfg *config.RuntimeConfig, key string) *Signer {
	return &Signer{cfg: cfg, signer: &sign.Signer{Key: key}}
}

// SignedURL generates a redirect URL for the given link.
func (s *Signer) SignedURL(link string, ops ...any) string {
	host := strings.TrimSuffix(s.cfg.HTTPHostname, "/")
	// Generate nonce if not provided
	if len(ops) == 0 {
		ops = append(ops, sign.WithOutNonce())
	}
	sig := s.signer.Sign("link:"+link, ops...)
	return fmt.Sprintf("%s/goto?u=%s&sig=%s", host, url.QueryEscape(link), sig)
}

// Sign returns the timestamp and signature for link using the optional expiry time.
// Sign returns the timestamp and signature for link using the provided options.
func (s *Signer) Sign(link string, ops ...any) string {
	return s.signer.Sign("link:"+link, ops...)
}

// Verify checks the provided signature matches the link.
func (s *Signer) Verify(link, ts, sig string) bool {
	valid, _ := s.signer.Verify("link:"+link, sig, sign.WithExpiryTimestamp(ts))
	return valid
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
			return s.SignedURL(val, sign.WithOutNonce())
		}
	}
	return val
}
