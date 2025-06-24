package goa4web

import (
	"crypto/sha256"
	"net/url"
	"os"
	"strings"

	config "github.com/arran4/goa4web/config"
	"github.com/gorilla/csrf"
)

// csrfEnabled reports if CSRF protection should be active.
func csrfEnabled() bool {
	v := strings.ToLower(os.Getenv(config.EnvCSRFEnabled))
	if v == "" {
		return true
	}
	switch v {
	case "0", "false", "off", "no":
		return false
	default:
		return true
	}
}

// newCSRFMiddleware returns a wrapper enforcing CSRF protection using the
// provided session secret and HTTP configuration.
func newCSRFMiddleware(secret string, hostname string, version string) routerWrapper {
	key := sha256.Sum256([]byte(secret))
	origins := []string{}
	if u, err := url.Parse(hostname); err == nil && u.Host != "" {
		origins = append(origins, u.Host)
	}
	return routerWrapperFunc(csrf.Protect(key[:], csrf.Secure(version != "dev"), csrf.TrustedOrigins(origins)))
}
