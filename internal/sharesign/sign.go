package sharesign

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
func NewSigner(cfg *config.RuntimeConfig, key string) *Signer {
	return &Signer{cfg: cfg, signer: &sign.Signer{Key: key}}
}

// SignedURL generates a redirect URL for the given link.
// Defaults to path-based signature.
// For module paths like "/private/topic/2/thread/1", it becomes "/private/shared/topic/2/thread/1/ts/{ts}/sign/{sig}"
func (s *Signer) SignedURL(link string, exp ...time.Time) string {
	return s.SignedURLPath(link, exp...)
}

// SignedURLQuery generates a redirect URL with query parameters.
func (s *Signer) SignedURLQuery(link string, exp ...time.Time) string {
	host := strings.TrimSuffix(s.cfg.HTTPHostname, "/")
	sharedPath, rawQuery, fragment := s.prepareSharedLink(link)
	data := s.signatureData(sharedPath, rawQuery)
	var ops []any
	for _, t := range exp {
		ops = append(ops, sign.WithExpiry(t))
	}
	ts, sig := s.signer.Sign(data, ops...)
	if rawQuery != "" {
		return fmt.Sprintf("%s%s?%s&ts=%d&sig=%s%s", host, sharedPath, rawQuery, ts, sig, fragment)
	}
	return fmt.Sprintf("%s%s?ts=%d&sig=%s%s", host, sharedPath, ts, sig, fragment)
}

// SignedURLPath generates a redirect URL with path parameters.
func (s *Signer) SignedURLPath(link string, exp ...time.Time) string {
	host := strings.TrimSuffix(s.cfg.HTTPHostname, "/")
	sharedPath, rawQuery, fragment := s.prepareSharedLink(link)
	data := s.signatureData(sharedPath, rawQuery)
	var ops []any
	for _, t := range exp {
		ops = append(ops, sign.WithExpiry(t))
	}
	ts, sig := s.signer.Sign(data, ops...)
	if rawQuery != "" {
		return fmt.Sprintf("%s%s/ts/%d/sign/%s?%s%s", host, sharedPath, ts, sig, rawQuery, fragment)
	}
	return fmt.Sprintf("%s%s/ts/%d/sign/%s%s", host, sharedPath, ts, sig, fragment)
}

func (s *Signer) signatureData(path, rawQuery string) string {
	if rawQuery == "" {
		return "share:" + path
	}
	return "share:" + path + "?" + rawQuery
}

func (s *Signer) prepareSharedLink(link string) (string, string, string) {
	sharedPath := link
	rawQuery := ""
	fragment := ""
	if parsed, err := url.Parse(link); err == nil {
		if parsed.Path != "" {
			sharedPath = parsed.Path
		}
		rawQuery = parsed.RawQuery
		if parsed.Fragment != "" {
			fragment = "#" + parsed.Fragment
		}
	}
	sharedPath = s.injectShared(sharedPath)
	return sharedPath, rawQuery, fragment
}

func (s *Signer) injectShared(link string) string {
	// Inject "/shared" after the first path segment (module name)
	// e.g., "/private/topic/2/thread/1" → "/private/shared/topic/2/thread/1"
	parts := strings.SplitN(link, "/", 3)
	if len(parts) >= 3 && parts[0] == "" && parts[1] != "" {
		// Avoid double injection if "shared" is already the next segment
		if strings.HasPrefix(parts[2], "shared/") || parts[2] == "shared" {
			return link
		}
		// parts: ["", "private", "topic/2/thread/1"]
		return "/" + parts[1] + "/shared/" + parts[2]
	}
	return link
}

// Sign returns the timestamp and signature for link using the optional expiry time.
func (s *Signer) Sign(link string, ops ...any) (int64, string) {
	var newOps []any
	for _, op := range ops {
		if t, ok := op.(time.Time); ok {
			newOps = append(newOps, sign.WithExpiry(t))
		} else {
			newOps = append(newOps, op)
		}
	}
	return s.signer.Sign("share:"+link, newOps...)
}

// Verify checks the provided signature matches the link.
// You must provide either sign.WithExpiryTimestamp(ts) or sign.WithNoExpiry().
func (s *Signer) Verify(link, sig string, ops ...any) (bool, error) {
	return s.signer.Verify("share:"+link, sig, ops...)
}
