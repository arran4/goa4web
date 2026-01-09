package sharesign

import (
	"fmt"
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
func NewSigner(cfg *config.RuntimeConfig, key string) *Signer {
	return &Signer{cfg: cfg, signer: &sign.Signer{Key: key}}
}

// SignedURL generates a redirect URL for the given link.
func (s *Signer) SignedURL(link string, exp ...time.Time) string {
	host := strings.TrimSuffix(s.cfg.HTTPHostname, "/")
	ts, sig := s.signer.Sign("share:"+link, exp...)
	return fmt.Sprintf("%s/shared%s?ts=%d&sig=%s", host, link, ts, sig)
}

// Sign returns the timestamp and signature for link using the optional expiry time.
func (s *Signer) Sign(link string, exp ...time.Time) (int64, string) {
	return s.signer.Sign("share:"+link, exp...)
}

// Verify checks the provided signature matches the link.
func (s *Signer) Verify(link, ts, sig string) bool {
	return s.signer.Verify("share:"+link, ts, sig)
}
