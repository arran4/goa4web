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
// For module paths like "/private/topic/2/thread/1", it becomes "/private/shared/topic/2/thread/1"
func (s *Signer) SignedURL(link string, exp ...time.Time) string {
	host := strings.TrimSuffix(s.cfg.HTTPHostname, "/")

	// Inject "/shared" after the first path segment (module name)
	// e.g., "/private/topic/2/thread/1" → "/private/shared/topic/2/thread/1"
	parts := strings.SplitN(link, "/", 3)
	sharedLink := link
	if len(parts) >= 3 && parts[0] == "" && parts[1] != "" {
		// parts: ["", "private", "topic/2/thread/1"]
		sharedLink = "/" + parts[1] + "/shared/" + parts[2]
	}

	ts, sig := s.signer.Sign("share:"+sharedLink, exp...)
	return fmt.Sprintf("%s%s?ts=%d&sig=%s", host, sharedLink, ts, sig)
}

// Sign returns the timestamp and signature for link using the optional expiry time.
func (s *Signer) Sign(link string, exp ...time.Time) (int64, string) {
	return s.signer.Sign("share:"+link, exp...)
}

// Verify checks the provided signature matches the link.
func (s *Signer) Verify(link, ts, sig string) bool {
	return s.signer.Verify("share:"+link, ts, sig)
}
